package gaian_test

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"testing"

	gaian "github.com/gaiannetwork/gaian-go"
)

// Env vars read by newTestClient and the getTest* fixture helpers below.
// Set these (e.g. via a gitignored .env file loaded by TestMain) to run
// the *_integration_test.go files in this package against the real Gaian
// sandbox instead of skipping them.
const (
	envSandboxBaseURL   = "GAIAN_SANDBOX_BASE_URL"
	envSandboxAPIKey    = "GAIAN_SANDBOX_API_KEY"
	envSandboxSecretKey = "GAIAN_SANDBOX_SECRET_KEY"

	defaultSandboxBaseURL = "https://sandbox.gaian.network"

	// testAPIKey/testSecretKey are used by *_unit_test.go files — any
	// non-empty pair works since newMockClient never talks to a real
	// server.
	testAPIKey    = "test-api-key"
	testSecretKey = "test-secret-key"

	envSandboxUserID          = "GAIAN_SANDBOX_USER_ID"
	envSandboxWalletAddress   = "GAIAN_SANDBOX_WALLET_ADDRESS"
	envSandboxOrderID         = "GAIAN_SANDBOX_ORDER_ID"
	envSandboxQuoteID         = "GAIAN_SANDBOX_QUOTE_ID"
	envSandboxPrefundQuoteID  = "GAIAN_SANDBOX_PREFUND_QUOTE_ID"
	envSandboxTransactionHash = "GAIAN_SANDBOX_TX_HASH"
	envSandboxQRString        = "GAIAN_SANDBOX_QR_STRING"
	envSandboxBankCode        = "GAIAN_SANDBOX_BANK_CODE"
	envSandboxAccountNumber   = "GAIAN_SANDBOX_ACCOUNT_NUMBER"
)

// TestMain loads a gitignored .env file (if present) before any test
// runs, so sandbox credentials/fixtures can live in a local file instead
// of being exported by hand every time.
func TestMain(m *testing.M) {
	loadDotEnv(".env")
	os.Exit(m.Run())
}

// loadDotEnv reads KEY=VALUE pairs from path into the process
// environment. Blank lines and lines starting with "#" are skipped; a
// key already set in the real environment is left untouched. A missing
// file is silently ignored — .env is optional.
func loadDotEnv(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		if _, exists := os.LookupEnv(key); exists {
			continue
		}

		os.Setenv(key, strings.Trim(strings.TrimSpace(value), `"'`))
	}
}

// newTestClient builds a *gaian.Client wired to the real Gaian sandbox
// environment, using credentials from GAIAN_SANDBOX_API_KEY /
// GAIAN_SANDBOX_SECRET_KEY (GAIAN_SANDBOX_BASE_URL optionally overrides
// the default sandbox host). It skips the calling test when credentials
// aren't set, so `go test ./...` stays green without network access or
// secrets configured.
func newTestClient(t *testing.T) *gaian.Client {
	t.Helper()

	apiKey := strings.TrimSpace(os.Getenv(envSandboxAPIKey))
	if apiKey == "" {
		t.Skipf("skipping test: %s not set", envSandboxAPIKey)
	}

	secretKey := strings.TrimSpace(os.Getenv(envSandboxSecretKey))
	if secretKey == "" {
		t.Skipf("skipping test: %s not set", envSandboxSecretKey)
	}

	baseURL := strings.TrimSpace(os.Getenv(envSandboxBaseURL))
	if baseURL == "" {
		baseURL = defaultSandboxBaseURL
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	client, err := gaian.NewClient(
		baseURL,
		apiKey,
		secretKey,
		gaian.WithDebug(true),
		gaian.WithLogger(logger),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	return client
}

// ── fixture getters ─────────────────────────────────────────────────────
//
// These identify real, pre-existing sandbox resources (a registered
// user, a linked wallet, an open order, an unexpired quote, a broadcast
// tx hash, a real bank account) that no single test can create on its
// own. Each skips the calling test if the corresponding env var isn't
// set, rather than sending made-up values and hoping for a particular
// error — set them to real sandbox values to exercise the actual flow.

func getTestUserID(t *testing.T) string {
	t.Helper()
	v := strings.TrimSpace(os.Getenv(envSandboxUserID))
	if v == "" {
		t.Skipf("skipping test: %s not set", envSandboxUserID)
	}
	return v
}

func getTestWalletAddress(t *testing.T) string {
	t.Helper()
	v := strings.TrimSpace(os.Getenv(envSandboxWalletAddress))
	if v == "" {
		t.Skipf("skipping test: %s not set", envSandboxWalletAddress)
	}
	return v
}

func getTestOrderID(t *testing.T) string {
	t.Helper()
	v := strings.TrimSpace(os.Getenv(envSandboxOrderID))
	if v == "" {
		t.Skipf("skipping test: %s not set", envSandboxOrderID)
	}
	return v
}

func getTestQuoteID(t *testing.T) string {
	t.Helper()
	v := strings.TrimSpace(os.Getenv(envSandboxQuoteID))
	if v == "" {
		t.Skipf("skipping test: %s not set", envSandboxQuoteID)
	}
	return v
}

func getTestPrefundQuoteID(t *testing.T) string {
	t.Helper()
	v := strings.TrimSpace(os.Getenv(envSandboxPrefundQuoteID))
	if v == "" {
		t.Skipf("skipping test: %s not set", envSandboxPrefundQuoteID)
	}
	return v
}

func getTestTransactionHash(t *testing.T) string {
	t.Helper()
	v := strings.TrimSpace(os.Getenv(envSandboxTransactionHash))
	if v == "" {
		t.Skipf("skipping test: %s not set", envSandboxTransactionHash)
	}
	return v
}

func getTestQRString(t *testing.T) string {
	t.Helper()
	v := strings.TrimSpace(os.Getenv(envSandboxQRString))
	if v == "" {
		t.Skipf("skipping test: %s not set", envSandboxQRString)
	}
	return v
}

func getTestBankCode(t *testing.T) string {
	t.Helper()
	v := strings.TrimSpace(os.Getenv(envSandboxBankCode))
	if v == "" {
		t.Skipf("skipping test: %s not set", envSandboxBankCode)
	}
	return v
}

func getTestAccountNumber(t *testing.T) string {
	t.Helper()
	v := strings.TrimSpace(os.Getenv(envSandboxAccountNumber))
	if v == "" {
		t.Skipf("skipping test: %s not set", envSandboxAccountNumber)
	}
	return v
}

// generateTestWalletAddress returns a random, never-before-seen address
// so TestSandbox_CreateWallet doesn't collide with a wallet linked by a
// previous run.
func generateTestWalletAddress(t *testing.T) string {
	t.Helper()

	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		t.Fatalf("failed to generate random address: %v", err)
	}

	return "0x" + hex.EncodeToString(b)
}

// assertServerResponded accepts either success or a real API-level
// rejection (gaian.ErrUnexpectedStatus): sandbox may legitimately reject
// a request built from these fixtures (an expired quote, a
// production-only endpoint, a tenant not on the prefund plan, ...), and
// that's a fine outcome here — it proves the request reached the server
// and got a real response back. A transport failure or any other
// client-side error means the SDK itself is broken and must fail loudly.
func assertServerResponded(t *testing.T, err error) {
	t.Helper()

	if err == nil {
		return
	}
	if errors.Is(err, gaian.ErrUnexpectedStatus) {
		t.Logf("sandbox rejected the request: %v", err)
		return
	}
	t.Fatalf("unexpected client-side error talking to sandbox: %v", err)
}

// ── unit-test support ────────────────────────────────────────────────────
//
// mockHTTPClient is a canned gaian.HTTPClient for *_unit_test.go files in
// this package that need to verify request/response wiring
// deterministically (e.g. exact marshaled JSON, envelope unmarshaling)
// without depending on network access or sandbox state. It also records
// the last request it was asked to send (and its body, since
// http.Request.Body can only be read once) so tests can assert on what
// the client actually sent.
type mockHTTPClient struct {
	response *http.Response
	err      error

	lastReq     *http.Request
	lastReqBody []byte
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	m.lastReq = req
	if req.Body != nil {
		b, _ := io.ReadAll(req.Body)
		m.lastReqBody = b
	}

	return m.response, m.err
}

// lastPath returns the URL path of the last request sent through m.
func (m *mockHTTPClient) lastPath() string {
	if m.lastReq == nil {
		return ""
	}
	return m.lastReq.URL.Path
}

// calledPath reports whether the last request sent through m had the
// given URL path.
func (m *mockHTTPClient) calledPath(path string) bool {
	return m.lastPath() == path
}

// lastBody returns the body of the last request sent through m.
func (m *mockHTTPClient) lastBody() []byte {
	return m.lastReqBody
}

func mockResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		Status:        http.StatusText(statusCode),
		StatusCode:    statusCode,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Header:        make(http.Header),
		Body:          io.NopCloser(bytes.NewBufferString(body)),
		ContentLength: int64(len(body)),
	}
}

// newMockClient builds a *gaian.Client backed by mock — a deterministic,
// network-free client for *_unit_test.go files.
func newMockClient(t *testing.T, mock gaian.HTTPClient) *gaian.Client {
	t.Helper()

	client, err := gaian.NewClient(defaultSandboxBaseURL, testAPIKey, testSecretKey,
		gaian.WithHTTPClient(mock),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	return client
}

// errConnectionRefused simulates a transport-level failure (server
// unreachable) for HTTPClient.Do, as opposed to a well-formed HTTP
// error response.
var errConnectionRefused = errors.New("connection refused")
