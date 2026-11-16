package signer_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/Gaian-Network/gaian-go/v2/signer"
)

type requestVector struct {
	Name          string          `json:"name"`
	Method        string          `json:"method"`
	Path          string          `json:"path"`
	Query         string          `json:"query"`
	Body          json.RawMessage `json:"body"`
	SignedPath    string          `json:"signedPath"`
	CanonicalBody string          `json:"canonicalBody"`
	Signature     string          `json:"signature"`
}

type vectorFile struct {
	APIKey    string          `json:"apiKey"`
	SecretKey string          `json:"secretKey"`
	Timestamp string          `json:"timestamp"`
	Requests  []requestVector `json:"requests"`
}

func loadVectors(t *testing.T) vectorFile {
	t.Helper()

	raw, err := os.ReadFile("../testdata/signature-vectors.json")
	if err != nil {
		t.Fatalf("read vectors: %v", err)
	}

	var f vectorFile
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatalf("parse vectors: %v", err)
	}
	if len(f.Requests) == 0 {
		t.Fatal("vector file has no requests")
	}

	return f
}

// bodyFromVector turns the vector's JSON body into the Go value an SDK
// caller would hand to BuildHeaders: nil for "no body", otherwise the
// decoded JSON object.
func bodyFromVector(t *testing.T, v requestVector) any {
	t.Helper()

	if string(v.Body) == "null" {
		return nil
	}

	var m map[string]any
	if err := json.Unmarshal(v.Body, &m); err != nil {
		t.Fatalf("%s: decode body: %v", v.Name, err)
	}

	return m
}

// TestSignVectors checks that BuildHeaders signs every vector the way the
// gateway does. BuildHeaders stamps the current time, so the vector's
// fixed timestamp cannot be injected. Instead the test (1) checks Sign
// against the vector's own signature, then (2) rebuilds the expected
// signature from the vector's signedPath/canonicalBody and the timestamp
// BuildHeaders actually returned.
func TestSignVectors(t *testing.T) {
	f := loadVectors(t)

	for _, v := range f.Requests {
		t.Run(v.Name, func(t *testing.T) {
			vectorMessage := f.APIKey + v.Method + v.SignedPath + v.Query + v.CanonicalBody + f.Timestamp
			if got := signer.Sign(f.SecretKey, vectorMessage); got != v.Signature {
				t.Fatalf("Sign() = %q, want vector signature %q", got, v.Signature)
			}

			headers, err := signer.BuildHeaders(f.APIKey, f.SecretKey, signer.SignParams{
				Method: v.Method,
				Path:   v.Path,
				Query:  v.Query,
				Body:   bodyFromVector(t, v),
			})
			if err != nil {
				t.Fatalf("BuildHeaders: %v", err)
			}

			ts := headers["X-GAIAN-TIMESTAMP"]
			want := signer.Sign(f.SecretKey, f.APIKey+v.Method+v.SignedPath+v.Query+v.CanonicalBody+ts)
			if got := headers["X-GAIAN-SIGNATURE"]; got != want {
				t.Errorf("signature = %q, want %q (canonical body %q, signed path %q)",
					got, want, v.CanonicalBody, v.SignedPath)
			}
		})
	}
}

// TestSignVectorsEscapedPath guards the client-side case: endpoint files
// build paths with url.PathEscape, so BuildHeaders can receive a path
// that is still percent-encoded. The gateway signs the decoded path.
func TestSignVectorsEscapedPath(t *testing.T) {
	f := loadVectors(t)

	cases := []struct {
		name        string
		path        string
		signedPath  string
		description string
	}{
		{"space", "/api/v2/users/wallet/addr%20with%20space", "/api/v2/users/wallet/addr with space", "percent-encoded space is decoded"},
		{"colon", "/api/v2/users/wallet/eip155%3A1%3A0xAbC", "/api/v2/users/wallet/eip15510xAbC", "encoded colon is decoded then dropped"},
		{"plain colon", "/api/v2/users/wallet/eip155:1:0xAbC", "/api/v2/users/wallet/eip15510xAbC", "colon is dropped"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			headers, err := signer.BuildHeaders(f.APIKey, f.SecretKey, signer.SignParams{
				Method: "GET",
				Path:   tc.path,
			})
			if err != nil {
				t.Fatalf("BuildHeaders: %v", err)
			}

			ts := headers["X-GAIAN-TIMESTAMP"]
			want := signer.Sign(f.SecretKey, f.APIKey+"GET"+tc.signedPath+ts)
			if got := headers["X-GAIAN-SIGNATURE"]; got != want {
				t.Errorf("%s: signature = %q, want %q", tc.description, got, want)
			}
			if strings.ContainsAny(headers["X-GAIAN-SIGNATURE"], "+/") {
				t.Errorf("signature is not URL-safe base64: %q", headers["X-GAIAN-SIGNATURE"])
			}
		})
	}
}
