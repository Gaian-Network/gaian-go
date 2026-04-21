// Package gaian provides a Go client for the Gaian Network API.
//
// Two base services are exposed: User (registration, KYC, orders) and
// Payment (place/verify orders, QR parsing, exchange calculation).
// Both require an API key issued via the Gaian Client Admin portal.
//
// Quick start:
//
//	c := gaian.New("your-api-key", gaian.Sandbox)
//	user, err := c.User.Register(ctx, "0xWalletAddress")
package gaian

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Environment selects which set of base URLs the client uses.
type Environment int

const (
	Sandbox    Environment = iota // https://dev-*.gaian-dev.network
	Production                    // https://*.gaian-dev.network
)

const (
	sandboxUserBase    = "https://dev-user.gaian-dev.network"
	sandboxPaymentBase = "https://dev-payments.gaian-dev.network"
	prodUserBase       = "https://user.gaian-dev.network"
	prodPaymentBase    = "https://payments.gaian-dev.network"
)

// HTTPClient is the interface satisfied by *http.Client and any test double.
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// Client is the root Gaian API client. Use New to construct one.
type Client struct {
	User    *UserService
	Payment *PaymentService
	Tenant  *TenantService
	Policy  *PolicyService
}

// UserService groups user, order, and KYC endpoints (User API base URL).
type UserService struct{ c *core }

// PaymentService groups payment, QR, and exchange endpoints (Payment API base URL).
type PaymentService struct{ c *core }

// TenantService groups tenant balance and spend endpoints (User API base URL).
type TenantService struct{ c *core }

// PolicyService exposes the wallet policy check endpoint (Payment API base URL).
type PolicyService struct{ c *core }

// core holds the shared transport used by all services.
type core struct {
	httpClient  HTTPClient
	apiKey      string
	userBaseURL string
	pmtBaseURL  string
	logger      Logger
	debug       bool
}

// Option configures the client.
type Option func(*core)

// WithHTTPClient replaces the default HTTP client.
func WithHTTPClient(hc HTTPClient) Option {
	return func(c *core) { c.httpClient = hc }
}

// WithTimeout sets a request timeout on the default HTTP client (default: 30s).
// Has no effect when WithHTTPClient is also provided.
func WithTimeout(d time.Duration) Option {
	return func(c *core) {
		if hc, ok := c.httpClient.(*http.Client); ok {
			hc.Timeout = d
		}
	}
}

// WithLogger injects a custom logger.
func WithLogger(l Logger) Option {
	return func(c *core) { c.logger = l }
}

// WithDebug enables debug-level request logging.
func WithDebug(enabled bool) Option {
	return func(c *core) { c.debug = enabled }
}

// New creates a Client for the given environment using the provided API key.
func New(apiKey string, env Environment, opts ...Option) *Client {
	userBase, pmtBase := sandboxUserBase, sandboxPaymentBase
	if env == Production {
		userBase, pmtBase = prodUserBase, prodPaymentBase
	}
	return NewWithURLs(apiKey, userBase, pmtBase, opts...)
}

// NewWithURLs creates a Client with explicit base URLs. Useful for tests.
func NewWithURLs(apiKey, userBaseURL, paymentBaseURL string, opts ...Option) *Client {
	cr := &core{
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		apiKey:      apiKey,
		userBaseURL: userBaseURL,
		pmtBaseURL:  paymentBaseURL,
		logger:      nopLogger{},
	}
	for _, o := range opts {
		o(cr)
	}
	return &Client{
		User:    &UserService{c: cr},
		Payment: &PaymentService{c: cr},
		Tenant:  &TenantService{c: cr},
		Policy:  &PolicyService{c: cr},
	}
}

// envelope is the generic JSON wrapper used by most User-service responses.
type envelope[T any] struct {
	Status  any    `json:"status"`
	Message string `json:"message"`
	Data    *T     `json:"data"`
}

// do executes an authenticated request and decodes the response into out.
func (c *core) do(ctx context.Context, method, baseURL, path string, body, out any) error {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("gaian: marshal request: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("gaian: build request: %w", err)
	}
	req.Header.Set("x-api-key", c.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.debug {
		c.logger.Debugf("gaian: → %s %s%s", method, baseURL, path)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("gaian: http: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("gaian: read body: %w", err)
	}

	if c.debug {
		c.logger.Debugf("gaian: ← %d %s", resp.StatusCode, baseURL+path)
	}

	if resp.StatusCode >= 400 {
		return parseAPIError(resp.StatusCode, raw)
	}

	if out != nil {
		if err := json.Unmarshal(raw, out); err != nil {
			return fmt.Errorf("gaian: decode response: %w", err)
		}
	}
	return nil
}

func (c *core) get(ctx context.Context, baseURL, path string, q url.Values, out any) error {
	p := path
	if len(q) > 0 {
		p += "?" + q.Encode()
	}
	return c.do(ctx, http.MethodGet, baseURL, p, nil, out)
}

func (c *core) post(ctx context.Context, baseURL, path string, body, out any) error {
	return c.do(ctx, http.MethodPost, baseURL, path, body, out)
}
