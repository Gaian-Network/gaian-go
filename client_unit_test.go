package gaian

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
)

const (
	testAPIKey    = "test-api-key"
	testSecretKey = "test-secret-key"
)

// recordingHTTPClient is a fake HTTPClient that captures the last request
// it was asked to send (and its body, since http.Request.Body can only
// be read once) and returns a canned response. It's used below to
// inspect exactly what Client would send over the wire — this validates
// the SDK's own request-construction/signing logic, not any particular
// server's behavior, which is why it lives in this internal (white-box)
// test file rather than the sandbox-hitting *_integration_test.go files
// in package gaian_test.
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

// ── NewClient validation ────────────────────────────────────────────────────
//
// Pure constructor logic, no HTTP involved — a clean unit-test case.

func TestNewClient_Validation(t *testing.T) {
	tests := []struct {
		name      string
		baseURL   string
		apiKey    string
		secretKey string
		wantErr   error
	}{
		{"empty base URL", "", "key", "secret", ErrEmptyBaseURL},
		{"empty api key", "https://sandbox.gaian.network", "", "secret", ErrEmptyAPIKey},
		{"empty secret key", "https://sandbox.gaian.network", "key", "", ErrEmptySecretKey},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(tt.baseURL, tt.apiKey, tt.secretKey)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("NewClient() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewClient_Defaults(t *testing.T) {
	c, err := NewClient("https://sandbox.gaian.network", "key", "secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.httpClient != http.DefaultClient {
		t.Error("expected default httpClient to be http.DefaultClient")
	}
	if c.debug {
		t.Error("expected debug to default to false")
	}
}

func TestNewClient_OptionsApplied(t *testing.T) {
	rc := &recordingHTTPClient{}
	c, err := NewClient("https://sandbox.gaian.network", "key", "secret",
		WithHTTPClient(rc),
		WithDebug(true),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.httpClient != rc {
		t.Error("WithHTTPClient option was not applied")
	}
	if !c.debug {
		t.Error("WithDebug(true) option was not applied")
	}
}

// ── paramsToMap ──────────────────────────────────────────────────────────────
//
// Pure function feeding the GET query-encoding path — worth locking down
// directly since a bug here silently corrupts every GET request's query
// string.

func TestParamsToMap_Nil(t *testing.T) {
	m, err := paramsToMap(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m) != 0 {
		t.Errorf("paramsToMap(nil) = %v, want empty map", m)
	}
}

func TestParamsToMap_MapPassthrough(t *testing.T) {
	in := map[string]any{"a": 1, "b": "two"}
	m, err := paramsToMap(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m) != 2 || m["b"] != "two" {
		t.Errorf("paramsToMap() = %v, want passthrough of %v", m, in)
	}
}

func TestParamsToMap_StructUsesJSONTags(t *testing.T) {
	type params struct {
		UserID string `json:"userId"`
		Hidden string `json:"-"`
		Empty  string `json:"empty,omitempty"`
	}

	m, err := paramsToMap(params{UserID: "usr_1", Hidden: "should-not-appear"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m["userId"] != "usr_1" {
		t.Errorf(`m["userId"] = %v, want "usr_1"`, m["userId"])
	}
	if _, ok := m["Hidden"]; ok {
		t.Error(`json:"-"` + ` field leaked into the map`)
	}
	if _, ok := m["empty"]; ok {
		t.Error(`empty omitempty field leaked into the map`)
	}
}

// ── buildRequest / buildGETRequest ──────────────────────────────────────────
//
// These assemble the outgoing *http.Request; a bug here breaks every
// endpoint at once, so it's worth testing directly against a
// recordingHTTPClient rather than only indirectly through one specific
// endpoint's own test.

func TestBuildRequest_GET_PathTaggedFieldExcludedFromQuery(t *testing.T) {
	rc := &recordingHTTPClient{status: http.StatusOK, body: []byte(`{}`)}
	c, err := NewClient("https://sandbox.gaian.network", testAPIKey, testSecretKey, WithHTTPClient(rc))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	type listQuery struct {
		UserID string `param:"userId" json:"-"`
		Status string `json:"status,omitempty"`
	}

	req := &request{
		Method:   http.MethodGet,
		Endpoint: "/api/v2/users/usr_1/orders",
		Params:   listQuery{UserID: "usr_1", Status: "completed"},
	}

	if _, err := c.excute(context.Background(), req); err != nil {
		t.Fatalf("excute: %v", err)
	}

	q := rc.req.URL.Query()
	if q.Get("status") != "completed" {
		t.Errorf(`query "status" = %q, want "completed"`, q.Get("status"))
	}
	if q.Has("userId") || q.Has("UserID") {
		t.Errorf("path-tagged field leaked into the query string: %s", rc.req.URL.RawQuery)
	}
}

func TestBuildRequest_GET_NoParams_NoQueryString(t *testing.T) {
	rc := &recordingHTTPClient{status: http.StatusOK, body: []byte(`{}`)}
	c, err := NewClient("https://sandbox.gaian.network", testAPIKey, testSecretKey, WithHTTPClient(rc))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := &request{Method: http.MethodGet, Endpoint: "/api/v2/users/usr_1"}
	if _, err := c.excute(context.Background(), req); err != nil {
		t.Fatalf("excute: %v", err)
	}

	if rc.req.URL.RawQuery != "" {
		t.Errorf("expected no query string, got %q", rc.req.URL.RawQuery)
	}
}

func TestBuildRequest_POST_SetsContentTypeAndUserAgent(t *testing.T) {
	rc := &recordingHTTPClient{status: http.StatusOK, body: []byte(`{}`)}
	c, err := NewClient("https://sandbox.gaian.network", testAPIKey, testSecretKey, WithHTTPClient(rc))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := &request{Method: http.MethodPost, Endpoint: "/api/v2/users", Params: map[string]any{"a": 1}}
	if _, err := c.excute(context.Background(), req); err != nil {
		t.Fatalf("excute: %v", err)
	}

	if got := rc.req.Header.Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}
	if got := rc.req.Header.Get("User-Agent"); got != UserArgent {
		t.Errorf("User-Agent = %q, want %q", got, UserArgent)
	}
}

func TestBuildRequest_POST_BodyMatchesParams(t *testing.T) {
	rc := &recordingHTTPClient{status: http.StatusOK, body: []byte(`{}`)}
	c, err := NewClient("https://sandbox.gaian.network", testAPIKey, testSecretKey, WithHTTPClient(rc))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := &request{
		Method:   http.MethodPost,
		Endpoint: "/api/v2/orders",
		Params:   map[string]any{"quoteId": "quote_123"},
	}
	if _, err := c.excute(context.Background(), req); err != nil {
		t.Fatalf("excute: %v", err)
	}

	if got := string(rc.reqBody); got != `{"quoteId":"quote_123"}` {
		t.Errorf("request body = %q, want %q", got, `{"quoteId":"quote_123"}`)
	}
}

// TestBuildRequest_POST_NilParams_EmptyBodyMatchesSignedBody guards
// against a real bug: json.Marshal(nil) produces the literal 4-byte
// string "null", but signer.BuildHeaders treats a nil Body as "" per the
// documented canonicalization rule. If buildRequest sent "null" as the
// wire body while signing "", the gateway would recompute a different
// signature than the one sent and reject the request.
func TestBuildRequest_POST_NilParams_EmptyBodyMatchesSignedBody(t *testing.T) {
	rc := &recordingHTTPClient{status: http.StatusOK, body: []byte(`{}`)}
	c, err := NewClient("https://sandbox.gaian.network", testAPIKey, testSecretKey, WithHTTPClient(rc))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := &request{Method: http.MethodPost, Endpoint: "/api/v2/users"}
	if _, err := c.excute(context.Background(), req); err != nil {
		t.Fatalf("excute: %v", err)
	}

	if len(rc.reqBody) != 0 {
		t.Errorf("request body = %q, want empty (not the literal string %q)", rc.reqBody, "null")
	}
}

func TestSign_PublicRequestSkipsSigning(t *testing.T) {
	rc := &recordingHTTPClient{status: http.StatusOK, body: []byte(`{}`)}
	c, err := NewClient("https://sandbox.gaian.network", testAPIKey, testSecretKey, WithHTTPClient(rc))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := &request{Method: http.MethodGet, Endpoint: "/health", Public: true}
	if _, err := c.excute(context.Background(), req); err != nil {
		t.Fatalf("excute: %v", err)
	}

	if rc.req.Header.Get("X-Gaian-Signature") != "" {
		t.Error("expected no signature header on a Public request")
	}
}

func TestBuildRequest_NilRequest(t *testing.T) {
	c, err := NewClient("https://sandbox.gaian.network", testAPIKey, testSecretKey)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := c.buildRequest(nil); !errors.Is(err, ErrNilRequest) {
		t.Errorf("buildRequest(nil) error = %v, want %v", err, ErrNilRequest)
	}
}

// ── excute status-code handling ─────────────────────────────────────────────

func TestExcute_NonSuccessStatusReturnsError(t *testing.T) {
	rc := &recordingHTTPClient{status: http.StatusNotFound, body: []byte(`{"message":"user not found"}`)}
	c, err := NewClient("https://sandbox.gaian.network", testAPIKey, testSecretKey, WithHTTPClient(rc))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := &request{Method: http.MethodGet, Endpoint: "/api/v2/users/does-not-exist"}
	_, err = c.excute(context.Background(), req)
	if !errors.Is(err, ErrUnexpectedStatus) {
		t.Errorf("excute() error = %v, want wrapping %v", err, ErrUnexpectedStatus)
	}
}

func TestExcute_SuccessReturnsRawBody(t *testing.T) {
	body := `{"data":{"userId":"usr_1"},"requestId":"req_1"}`
	rc := &recordingHTTPClient{status: http.StatusOK, body: []byte(body)}
	c, err := NewClient("https://sandbox.gaian.network", testAPIKey, testSecretKey, WithHTTPClient(rc))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := &request{Method: http.MethodPost, Endpoint: "/api/v2/users"}
	got, err := c.excute(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != body {
		t.Errorf("excute() = %q, want %q", got, body)
	}
}

func TestExcute_TransportError(t *testing.T) {
	rc := &recordingHTTPClient{err: errors.New("connection refused")}
	c, err := NewClient("https://sandbox.gaian.network", testAPIKey, testSecretKey, WithHTTPClient(rc))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := &request{Method: http.MethodGet, Endpoint: "/api/v2/users/usr_1"}
	_, err = c.excute(context.Background(), req)
	if !errors.Is(err, ErrHTTPFailure) {
		t.Errorf("excute() error = %v, want wrapping %v", err, ErrHTTPFailure)
	}
}
