package gaian_test

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"

	gaian "github.com/Gaian-Network/gaian-go/v2"
)

// recordingLogger keeps every message and attribute it is given, at any
// level, as one flat string per call.
type recordingLogger struct {
	mu      sync.Mutex
	entries []string
}

func (l *recordingLogger) record(level, msg string, args ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.entries = append(l.entries, fmt.Sprint(append([]any{level, msg}, args...)...))
}

func (l *recordingLogger) Info(msg string, args ...any)  { l.record("INFO", msg, args...) }
func (l *recordingLogger) Error(msg string, args ...any) { l.record("ERROR", msg, args...) }
func (l *recordingLogger) Debug(msg string, args ...any) { l.record("DEBUG", msg, args...) }

func (l *recordingLogger) all() string {
	l.mu.Lock()
	defer l.mu.Unlock()

	return strings.Join(l.entries, "\n")
}

func TestDebugLogRedacts(t *testing.T) {
	t.Parallel()

	const (
		apiKey       = "gk_debug_API_KEY_sentinel"
		secretKey    = "gs_debug_SECRET_KEY_sentinel"
		requestValue = "REQUEST_BODY_SENTINEL_0xABCDEF"
		respValue    = "RESPONSE_BODY_SENTINEL_42"
	)

	tests := []struct {
		name   string
		status int
		body   string
	}{
		{"success response", http.StatusOK, `{"data":{"id":"` + respValue + `"},"requestId":"r1"}`},
		{"error response", http.StatusBadRequest, `{"message":"` + respValue + `"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mock := &mockHTTPClient{response: mockResponse(tt.status, tt.body)} //nolint:bodyclose // response body closed by client
			logger := &recordingLogger{}

			client, err := gaian.NewClient(defaultSandboxBaseURL, apiKey, secretKey,
				gaian.WithHTTPClient(mock),
				gaian.WithLogger(logger),
				gaian.WithDebug(true),
			)
			if err != nil {
				t.Fatalf("NewClient: %v", err)
			}

			// The error (if any) is returned to the caller and may carry the
			// body; only what the SDK writes to the logger is checked here.
			_, _ = client.CreateWallet(context.Background(), &gaian.CreateWalletRequest{
				UserID:  "u1",
				Address: requestValue,
				Chain:   "solana",
			})

			logged := logger.all()
			if !strings.Contains(logged, "/api/v2/users/u1/wallets") {
				t.Fatalf("debug logging is on but nothing about the request was logged: %q", logged)
			}

			signature := mock.lastReq.Header.Get("X-GAIAN-SIGNATURE")
			if signature == "" {
				t.Fatal("request was sent without a signature header")
			}

			secrets := map[string]string{
				"signature header value": signature,
				"secret key":             secretKey,
				"api key":                apiKey,
				"signature header name":  "X-GAIAN-SIGNATURE",
				"request body":           requestValue,
				"response body":          respValue,
			}
			for what, value := range secrets {
				if strings.Contains(logged, value) {
					t.Errorf("debug log contains the %s (%q):\n%s", what, value, logged)
				}
			}
		})
	}
}

func TestDebugOffLogsNothing(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{response: mockResponse(http.StatusOK, `{"data":{},"requestId":"r1"}`)} //nolint:bodyclose // response body closed by client
	logger := &recordingLogger{}

	client, err := gaian.NewClient(defaultSandboxBaseURL, testAPIKey, testSecretKey,
		gaian.WithHTTPClient(mock),
		gaian.WithLogger(logger),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	if _, err := client.CreateUser(context.Background(), &gaian.CreateUserRequest{}); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	if logged := logger.all(); logged != "" {
		t.Errorf("debug is off but the SDK logged: %q", logged)
	}
}

func TestDebugLogOmitsQuery(t *testing.T) {
	t.Parallel()

	const queryValue = "QUERY_VALUE_SENTINEL"

	tests := []struct {
		name   string
		status int
		body   string
	}{
		{"success response", http.StatusOK, `{"data":[],"requestId":"r1"}`},
		{"error response", http.StatusBadRequest, `{"message":"rejected"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mock := &mockHTTPClient{response: mockResponse(tt.status, tt.body)} //nolint:bodyclose // response body closed by client
			logger := &recordingLogger{}

			client, err := gaian.NewClient(defaultSandboxBaseURL, testAPIKey, testSecretKey,
				gaian.WithHTTPClient(mock),
				gaian.WithLogger(logger),
				gaian.WithDebug(true),
			)
			if err != nil {
				t.Fatalf("NewClient: %v", err)
			}

			_, _ = client.ListUserOrders(context.Background(), &gaian.ListUserOrdersRequest{
				UserID: "u1",
				Status: queryValue,
				Page:   7,
			})

			if got := mock.lastReq.URL.RawQuery; !strings.Contains(got, queryValue) {
				t.Fatalf("query was not sent: %q", got)
			}

			logged := logger.all()
			if !strings.Contains(logged, "/api/v2/users/u1/orders") {
				t.Fatalf("debug logging is on but the path was not logged: %q", logged)
			}

			for _, unwanted := range []string{queryValue, "status=", "page=", "?", defaultSandboxBaseURL} {
				if strings.Contains(logged, unwanted) {
					t.Errorf("debug log contains %q:\n%s", unwanted, logged)
				}
			}
		})
	}
}
