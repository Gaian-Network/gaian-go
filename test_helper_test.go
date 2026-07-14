package gaian

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	"github.com/gaiannetwork/gaian-go/signer"
	"github.com/google/uuid"
)

const (
	testAPIKey    = "test-api-key"
	testSecretKey = "test-secret-key"
)

// newTestClient spins up an httptest.Server driven by handler and
// returns a *Client wired to it with a fixed API key/secret. The server
// is closed automatically at the end of the test.
func newTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	client, err := NewClient(srv.URL, testAPIKey, testSecretKey)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	return client, srv
}

// writeJSON writes v as a JSON response body with the given status code.
func writeJSON(t *testing.T, w http.ResponseWriter, status int, v any) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Fatalf("encode response: %v", err)
	}
}

// readBody reads and returns r's body, restoring it so the handler's
// caller (httptest's request plumbing) can still be inspected afterward.
func readBody(t *testing.T, r *http.Request) []byte {
	t.Helper()

	if r.Body == nil {
		return nil
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		t.Fatalf("read request body: %v", err)
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	return body
}

// assertSignedHeadersPresent checks that a request carries the expected
// HMAC signing headers, without recomputing the signature itself. Use
// verifySignature instead when the test wants full cryptographic
// verification.
func assertSignedHeadersPresent(t *testing.T, r *http.Request) {
	t.Helper()

	if got := r.Header.Get("X-Gaian-Key"); got != testAPIKey {
		t.Errorf("X-Gaian-Key = %q, want %q", got, testAPIKey)
	}
	if r.Header.Get("X-Gaian-Signature") == "" {
		t.Error("X-Gaian-Signature header missing")
	}
	if r.Header.Get("X-Gaian-Timestamp") == "" {
		t.Error("X-Gaian-Timestamp header missing")
	}
	if _, err := uuid.Parse(r.Header.Get("X-Request-Id")); err != nil {
		t.Errorf("X-Request-Id is not a valid UUID: %v", err)
	}
}

// canonicalBody reimplements the wire-format canonicalization documented
// for the signing recipe (top-level keys sorted, compact JSON, "" for no
// body) so tests can independently recompute the expected signature
// without reaching into the unexported signer internals.
func canonicalBody(t *testing.T, raw []byte) string {
	t.Helper()

	if len(raw) == 0 {
		return ""
	}

	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("canonicalBody: unmarshal: %v", err)
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, k := range keys {
		kb, _ := json.Marshal(k)
		vb, err := json.Marshal(m[k])
		if err != nil {
			t.Fatalf("canonicalBody: marshal value for %q: %v", k, err)
		}
		buf.Write(kb)
		buf.WriteByte(':')
		buf.Write(vb)
		if i < len(keys)-1 {
			buf.WriteByte(',')
		}
	}
	buf.WriteByte('}')
	return buf.String()
}

// verifySignature fully recomputes the expected HMAC signature for r
// (using the timestamp r actually sent) and asserts it matches
// X-Gaian-Signature. This proves the whole Client -> signer pipeline is
// wired correctly end to end, not just that "some" signature was sent.
func verifySignature(t *testing.T, r *http.Request, rawBody []byte) {
	t.Helper()

	assertSignedHeadersPresent(t, r)

	timestamp := r.Header.Get("X-Gaian-Timestamp")
	message := testAPIKey + r.Method + r.URL.Path + r.URL.RawQuery + canonicalBody(t, rawBody) + timestamp
	want := signer.Sign(testSecretKey, message)

	if got := r.Header.Get("X-Gaian-Signature"); got != want {
		t.Errorf("X-Gaian-Signature = %q, want %q (message=%q)", got, want, message)
	}
}

// recordingHTTPClient is a fake HTTPClient that captures the last request
// it was asked to send (and its body, since http.Request.Body can only
// be read once) and returns a canned response. Used by unit tests that
// need to inspect exactly what Client would send over the wire without
// spinning up a real server.
type recordingHTTPClient struct {
	req     *http.Request
	reqBody []byte

	status int
	body   []byte
	err    error
}

func (c *recordingHTTPClient) Do(req *http.Request) (*http.Response, error) {
	c.req = req
	if req.Body != nil {
		b, _ := io.ReadAll(req.Body)
		c.reqBody = b
	}

	if c.err != nil {
		return nil, c.err
	}

	status := c.status
	if status == 0 {
		status = http.StatusOK
	}

	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(bytes.NewReader(c.body)),
		Header:     make(http.Header),
	}, nil
}
