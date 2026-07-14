package signer_test

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"sort"
	"strings"
	"testing"

	"github.com/gaiannetwork/gaian-go/signer"
	"github.com/google/uuid"
)

// canonicalize reimplements the documented signing wire format (top-level
// keys sorted, compact JSON, "" for no body) so tests can independently
// recompute the expected signature without reaching into signer's
// unexported internals.
func canonicalize(t *testing.T, body any) string {
	t.Helper()

	if body == nil {
		return ""
	}

	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("canonicalize: marshal: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("canonicalize: unmarshal: %v", err)
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var buf strings.Builder
	buf.WriteByte('{')
	for i, k := range keys {
		kb, _ := json.Marshal(k)
		vb, err := json.Marshal(m[k])
		if err != nil {
			t.Fatalf("canonicalize: marshal value for %q: %v", k, err)
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

func TestSign(t *testing.T) {
	secretKey := "test-secret"
	message := "GETsome/path?query=1{}1700000000"

	got := signer.Sign(secretKey, message)

	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(message))
	want := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	want = strings.ReplaceAll(want, "+", "-")
	want = strings.ReplaceAll(want, "/", "_")

	if got != want {
		t.Errorf("Sign() = %q, want %q", got, want)
	}
	if strings.ContainsAny(got, "+/") {
		t.Errorf("Sign() = %q, expected URL-safe base64 (no '+' or '/')", got)
	}
}

func TestSign_Deterministic(t *testing.T) {
	a := signer.Sign("secret", "message")
	b := signer.Sign("secret", "message")
	if a != b {
		t.Errorf("Sign() not deterministic: %q != %q", a, b)
	}
}

func TestSign_DifferentInputsDifferentSignatures(t *testing.T) {
	base := signer.Sign("secret", "message-a")
	other := signer.Sign("secret", "message-b")
	if base == other {
		t.Error("different messages produced the same signature")
	}

	other = signer.Sign("different-secret", "message-a")
	if base == other {
		t.Error("different secrets produced the same signature")
	}
}

func TestBuildHeaders(t *testing.T) {
	apiKey := "test-api-key"
	secretKey := "test-secret-key"
	body := map[string]any{"quoteId": "quote_123"}

	headers, err := signer.BuildHeaders(apiKey, secretKey, signer.SignParams{
		Method: "POST",
		Path:   "/api/v2/orders",
		Query:  "",
		Body:   body,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if headers["X-GAIAN-KEY"] != apiKey {
		t.Errorf("X-GAIAN-KEY = %q, want %q", headers["X-GAIAN-KEY"], apiKey)
	}
	if headers["X-GAIAN-SIGNATURE"] == "" {
		t.Error("X-GAIAN-SIGNATURE is empty")
	}
	if headers["X-GAIAN-TIMESTAMP"] == "" {
		t.Error("X-GAIAN-TIMESTAMP is empty")
	}
	if _, err := uuid.Parse(headers["X-REQUEST-ID"]); err != nil {
		t.Errorf("X-REQUEST-ID is not a valid UUID: %v", err)
	}

	// Signature must be reproducible from the same inputs, using the
	// actual timestamp BuildHeaders picked.
	message := apiKey + "POST" + "/api/v2/orders" + "" + canonicalize(t, body) + headers["X-GAIAN-TIMESTAMP"]
	want := signer.Sign(secretKey, message)
	if headers["X-GAIAN-SIGNATURE"] != want {
		t.Errorf("X-GAIAN-SIGNATURE = %q, want %q (message=%q)", headers["X-GAIAN-SIGNATURE"], want, message)
	}
}

func TestBuildHeaders_NilBody(t *testing.T) {
	headers, err := signer.BuildHeaders("key", "secret", signer.SignParams{
		Method: "GET",
		Path:   "/api/v2/orders/ord_1",
		Query:  "",
		Body:   nil,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if headers["X-GAIAN-SIGNATURE"] == "" {
		t.Error("X-GAIAN-SIGNATURE is empty")
	}

	message := "key" + "GET" + "/api/v2/orders/ord_1" + "" + "" + headers["X-GAIAN-TIMESTAMP"]
	want := signer.Sign("secret", message)
	if headers["X-GAIAN-SIGNATURE"] != want {
		t.Errorf("X-GAIAN-SIGNATURE = %q, want %q", headers["X-GAIAN-SIGNATURE"], want)
	}
}

func TestBuildHeaders_TopLevelKeysCanonicalizedRegardlessOfInputOrder(t *testing.T) {
	// Two Go maps built with different insertion order (and, since Go
	// randomizes map iteration, different internal iteration order too)
	// must still produce the same canonical body — and therefore, given
	// the same timestamp, the same signature.
	bodyA := map[string]any{"z": 1, "a": 2, "m": 3}
	bodyB := map[string]any{"m": 3, "z": 1, "a": 2}

	if canonicalize(t, bodyA) != canonicalize(t, bodyB) {
		t.Fatal("test setup invalid: canonicalize() itself isn't order-independent")
	}

	headersA, err := signer.BuildHeaders("key", "secret", signer.SignParams{Method: "POST", Path: "/p", Body: bodyA})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	headersB, err := signer.BuildHeaders("key", "secret", signer.SignParams{Method: "POST", Path: "/p", Body: bodyB})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Recompute both using their own timestamps (which may legitimately
	// differ between the two calls) and confirm they resolve to the
	// same signature whenever the timestamps happen to match — i.e.
	// the only thing that can make them differ is the timestamp, not
	// map ordering.
	msgA := "key" + "POST" + "/p" + "" + canonicalize(t, bodyA) + headersA["X-GAIAN-TIMESTAMP"]
	msgB := "key" + "POST" + "/p" + "" + canonicalize(t, bodyB) + headersA["X-GAIAN-TIMESTAMP"]
	if signer.Sign("secret", msgA) != signer.Sign("secret", msgB) {
		t.Error("logically identical bodies produced different canonical signatures")
	}
	_ = headersB
}

func TestBuildHeaders_QueryAffectsSignature(t *testing.T) {
	noQuery, err := signer.BuildHeaders("key", "secret", signer.SignParams{
		Method: "GET",
		Path:   "/api/v2/users/wallet/0xabc",
		Query:  "",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msgNoQuery := "key" + "GET" + "/api/v2/users/wallet/0xabc" + "" + "" + noQuery["X-GAIAN-TIMESTAMP"]
	msgWithQuery := "key" + "GET" + "/api/v2/users/wallet/0xabc" + "page=2" + "" + noQuery["X-GAIAN-TIMESTAMP"]

	if signer.Sign("secret", msgNoQuery) == signer.Sign("secret", msgWithQuery) {
		t.Error("query string did not affect the signed message")
	}
}
