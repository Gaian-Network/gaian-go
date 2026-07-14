// Package gaian provides a Go client for the Gaian Network v2 API — a
// single signed gateway for user onboarding, KYC, wallets, and
// crypto-to-fiat payments across Vietnam, Thailand, and other supported
// markets.
//
// Every method hangs directly off *Client (there are no per-resource
// sub-clients): c.CreateUser, c.Quote, c.PlaceOrder, and so on. Every
// request is HMAC-signed with an API key + secret pair issued from the
// Client Admin portal — see the signer subpackage for the signing
// mechanics, and https://gaian.network/authentication-v2 for the wire
// format this implements.
//
// # Quick start
//
//	client, err := gaian.NewClient("https://sandbox.gaian.network", apiKey, secretKey)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	resp, err := client.CreateUser(ctx, &gaian.CreateUserRequest{})
//	if err != nil {
//		log.Fatal(err)
//	}
//	userID := resp.Data.UserID
//
// # Payment flow
//
// The v2 payment flow is quote-first: Quote (or one of its ParseQR/
// prefund/direct variants) reserves a rate for ~120s, PlaceOrder consumes
// it and returns a transaction to sign and broadcast, VerifyOrder submits
// the resulting transaction hash, and GetOrderStatus is polled until the
// order reaches a terminal state. See the doc comments on Quote,
// PlaceOrder, VerifyOrder, and GetOrderStatus for the full sequence and
// its prefund/direct-transfer variants.
//
// # Response envelopes
//
// Every method returns one of two generic wrappers from response.go:
// UserResposne[T] for User/KYC/Wallet/Organization endpoints, or
// PaymentResponse[T] for Payment endpoints (quote/orders/verify/status/
// parseQr) — the two backing services shape their responses (and errors)
// differently. Check the doc comment on each method for which one it
// returns.
package gaian

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/gaiannetwork/gaian-go/signer"
)

const UserArgent = "gaian-go-sdk"

var (
	// Client initialization errors.
	ErrEmptyBaseURL   = errors.New("baseURL is required")
	ErrEmptySecretKey = errors.New("secretKey is required")
	ErrEmptyAPIKey    = errors.New("apiKey is required")

	// Request lifecycle errors.
	ErrNilRequest    = errors.New("request is nil")
	ErrRequestBuild  = errors.New("failed to build request")
	ErrRequestSign   = errors.New("failed to sign request")
	ErrRequestEncode = errors.New("failed to encode request body")
	ErrInvalidParams = errors.New("invalid request params")

	// HTTP / transport errors.
	ErrHTTPFailure      = errors.New("http request failed")
	ErrUnexpectedStatus = errors.New("unexpected http status code")
)

type Logger interface {
	Info(msg string, args ...any)
	Error(msg string, args ...any)
	Debug(msg string, args ...any)
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Option func(*Client)

type Client struct {
	baseURL    string
	apiKey     string
	secretKey  string
	logger     Logger
	debug      bool
	httpClient HTTPClient
}

func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}

func WithLogger(logger Logger) Option {
	return func(c *Client) {
		c.logger = logger
	}
}

func WithDebug(debug bool) Option {
	return func(c *Client) {
		c.debug = debug
	}
}

func WithHTTPClient(client HTTPClient) Option {
	return func(c *Client) {
		c.httpClient = client
	}
}

func NewClient(
	baseURL string,
	apiKey string,
	secretKey string,
	opts ...Option,
) (*Client, error) {
	if baseURL == "" {
		return nil, ErrEmptyBaseURL
	}

	if apiKey == "" {
		return nil, ErrEmptyAPIKey
	}

	if secretKey == "" {
		return nil, ErrEmptySecretKey
	}

	client := &Client{
		baseURL:    baseURL,
		apiKey:     apiKey,
		secretKey:  secretKey,
		logger:     slog.Default(),
		debug:      false,
		httpClient: http.DefaultClient,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client, nil
}

func (c *Client) logDebug(msg string, attrs ...any) {
	if c.debug {
		c.logger.Debug(msg, attrs...)
	}
}

func paramsToMap(params any) (map[string]any, error) {
	if params == nil {
		return map[string]any{}, nil
	}

	if m, ok := params.(map[string]any); ok {
		return m, nil
	}

	data, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	var result map[string]any
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) sign(req *request, headers http.Header, query string) error {
	if req.Public {
		return nil
	}

	signHeaders, err := signer.BuildHeaders(c.apiKey, c.secretKey, signer.SignParams{
		Method: req.Method,
		Path:   req.Endpoint,
		Query:  query,
		Body:   req.Params,
	})
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRequestSign, err)
	}

	for k, v := range signHeaders {
		headers.Set(k, v)
	}

	return nil
}

func (c *Client) buildGETRequest(req *request, fullURL string, headers http.Header) error {
	paramsMap, err := paramsToMap(req.Params)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidParams, err)
	}

	query := ""
	if len(paramsMap) > 0 {
		queryParams := url.Values{}
		for key, value := range paramsMap {
			queryParams.Add(key, fmt.Sprintf("%v", value))
		}

		query = queryParams.Encode()
		fullURL += "?" + query
	}

	if err := c.sign(req, headers, query); err != nil {
		return err
	}

	c.logDebug("http request", "url", fullURL)

	req.FullURL = fullURL
	req.Header = headers
	req.Body = nil

	return nil
}

func (c *Client) buildRequest(req *request) error {
	if req == nil {
		return ErrNilRequest
	}

	fullURL := c.baseURL + req.Endpoint

	headers := http.Header{}
	if req.Header != nil {
		headers = req.Header.Clone()
	}

	headers.Set("Content-Type", "application/json")
	headers.Set("User-Agent", UserArgent)

	if req.Method == http.MethodGet {
		return c.buildGETRequest(req, fullURL, headers)
	}

	// Params == nil means "no body" — leave bodyBytes empty rather than
	// marshaling to the literal 4-byte JSON "null". Otherwise the wire
	// body wouldn't match what sign() signs: signer.BuildHeaders treats
	// a nil Body as the empty string per the documented canonicalization
	// rule ("" for GET/empty body), so sending "null" here would produce
	// a body that doesn't match its own signature.
	var bodyBytes []byte
	if req.Params != nil {
		var err error
		bodyBytes, err = json.Marshal(req.Params)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrRequestEncode, err)
		}
	}

	if err := c.sign(req, headers, ""); err != nil {
		return err
	}

	c.logDebug("http request", "method", req.Method, "url", fullURL)

	req.FullURL = fullURL
	req.Header = headers
	req.Body = bytes.NewReader(bodyBytes)
	req.SigningData = bodyBytes

	return nil
}

func (c *Client) excute(ctx context.Context, req *request) ([]byte, error) {
	if err := c.buildRequest(req); err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		req.Method,
		req.FullURL,
		req.Body,
	)
	if err != nil {
		return nil, err
	}

	httpReq.Header = req.Header

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrHTTPFailure, err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	c.logDebug("http response", "status", resp.StatusCode)
	c.logDebug("http response body", "body", string(responseBody))

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf(
			"%w: status=%d body=%s",
			ErrUnexpectedStatus,
			resp.StatusCode,
			string(responseBody),
		)
	}

	return responseBody, nil
}
