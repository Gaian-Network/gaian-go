package signer

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidAPIKey    = errors.New("invalid api key")
	ErrInvalidSecretKey = errors.New("invalid secret key")
)

// SignParams describes the request to sign. Path is the request path
// exactly as sent, percent-encoded. Query is the raw query string exactly
// as sent, without the leading '?'. Body is the JSON body value; leave it
// nil for GET requests and requests without a body.
type SignParams struct {
	Method string
	Path   string
	Query  string
	Body   any
}

// BuildHeaders returns the X-GAIAN-KEY, X-GAIAN-SIGNATURE,
// X-GAIAN-TIMESTAMP and X-REQUEST-ID headers for one request.
//
// payload.Path must be the percent-encoded path as sent on the wire.
// BuildHeaders percent-decodes it once, so a path that was already
// decoded and contains a literal '%' is decoded a second time and the
// gateway rejects the signature.
//
// The signature covers the decoded path without ':' and the body with
// its keys sorted; a nil body and an empty JSON object both sign as "".
func BuildHeaders(apiKey, secretKey string, payload SignParams) (map[string]string, error) {
	timestamp := fmt.Sprintf("%d", time.Now().Unix())

	sortedBody, err := marshalSorted(payload.Body)
	if err != nil {
		return nil, fmt.Errorf("gaian signer: %w", err)
	}

	message := apiKey + payload.Method + signedPath(payload.Path) + payload.Query + sortedBody + timestamp
	signature := Sign(secretKey, message)

	requestID := uuid.New().String()

	headers := map[string]string{
		"X-GAIAN-KEY":       apiKey,
		"X-GAIAN-SIGNATURE": signature,
		"X-GAIAN-TIMESTAMP": timestamp,
		"X-REQUEST-ID":      requestID,
	}

	return headers, nil
}

// signedPath returns the path the gateway signs: percent-decoded, with
// every ':' removed. A path that is not valid percent-encoding is used
// as given.
func signedPath(path string) string {
	if decoded, err := url.PathUnescape(path); err == nil {
		path = decoded
	}

	return strings.ReplaceAll(path, ":", "")
}

func marshalSorted(v any) (string, error) {
	if v == nil {
		return "", nil
	}

	raw, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		return string(raw), err
	}

	if len(body) == 0 {
		return "", nil
	}

	return buildSortedBody(body)
}

func buildSortedBody(body map[string]any) (string, error) {
	keys := make([]string, 0, len(body))
	for key := range body {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	out := make([]byte, 0, 128)
	out = append(out, '{')

	for i, k := range keys {
		kBytes, err := json.Marshal(k)
		if err != nil {
			return "", err
		}

		vBytes, err := json.Marshal(body[k])
		if err != nil {
			return "", err
		}
		out = append(out, kBytes...)
		out = append(out, ':')
		out = append(out, vBytes...)
		if i < len(keys)-1 {
			out = append(out, ',')
		}
	}

	out = append(out, '}')
	return string(out), nil
}

func Sign(secretKey, message string) string {
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(message))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	signature = strings.ReplaceAll(signature, "+", "-")
	signature = strings.ReplaceAll(signature, "/", "_")

	return signature
}
