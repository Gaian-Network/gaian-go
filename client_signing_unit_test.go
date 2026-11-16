package gaian_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"

	gaian "github.com/Gaian-Network/gaian-go/v2"
)

// gatewayBody mirrors how the gateway canonicalizes a request body before
// verifying (gaian-api-gateway signature_auth.go prepareRequestBody): empty
// body or empty object -> "", otherwise top-level keys sorted and
// re-marshaled. It deliberately does not call anything from the SDK.
func gatewayBody(t *testing.T, raw []byte) string {
	t.Helper()

	if len(raw) == 0 {
		return ""
	}

	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("gateway: body is not a JSON object: %v", err)
	}
	if len(m) == 0 {
		return ""
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	sorted := make(map[string]any, len(m))
	for _, k := range keys {
		sorted[k] = m[k]
	}

	out, err := json.Marshal(sorted)
	if err != nil {
		t.Fatalf("gateway: marshal: %v", err)
	}

	return string(out)
}

// gatewayVerifier is a fake gateway: for each request it recomputes the
// signature from the raw request exactly as the gateway does and records
// whether it matched the X-GAIAN-SIGNATURE header.
type gatewayVerifier struct {
	t *testing.T

	method    string
	rawQuery  string
	rawBody   string
	matched   bool
	requestID string
	gotPath   string
}

func (g *gatewayVerifier) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)

	g.method = r.Method
	g.rawQuery = r.URL.RawQuery
	g.rawBody = string(body)
	g.requestID = r.Header.Get("X-Request-ID")
	g.gotPath = r.URL.Path

	signedPath := strings.ReplaceAll(r.URL.Path, ":", "")
	message := r.Header.Get("X-GAIAN-KEY") + r.Method + signedPath + r.URL.RawQuery +
		gatewayBody(g.t, body) + r.Header.Get("X-GAIAN-TIMESTAMP")

	mac := hmac.New(sha256.New, []byte(testSecretKey))
	mac.Write([]byte(message))
	want := base64.URLEncoding.EncodeToString(mac.Sum(nil))

	g.matched = r.Header.Get("X-GAIAN-SIGNATURE") == want

	if g.matched {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/orders") {
			_, _ = w.Write([]byte(`{"data":[],"requestId":"r1"}`))
			return
		}

		_, _ = w.Write([]byte(`{"data":{},"requestId":"r1"}`))
		return
	}

	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"message":"invalid signature"}`))
}

func newGatewayClient(t *testing.T) (*gaian.Client, *gatewayVerifier) {
	t.Helper()

	g := &gatewayVerifier{t: t}
	srv := httptest.NewServer(g)
	t.Cleanup(srv.Close)

	client, err := gaian.NewClient(srv.URL, testAPIKey, testSecretKey)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	return client, g
}

func assertAccepted(t *testing.T, g *gatewayVerifier, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("request rejected by fake gateway: %v", err)
	}
	if !g.matched {
		t.Fatal("signature does not match the one the gateway recomputes")
	}
	if _, perr := uuid.Parse(g.requestID); perr != nil {
		t.Errorf("X-Request-ID = %q, want a UUID", g.requestID)
	}
}

func TestGETSignsEmptyBody(t *testing.T) {
	t.Parallel()

	t.Run("list orders with query params", func(t *testing.T) {
		t.Parallel()

		client, g := newGatewayClient(t)
		_, err := client.ListUserOrders(context.Background(), &gaian.ListUserOrdersRequest{
			UserID:   "u1",
			Status:   "completed",
			Page:     2,
			PageSize: 50,
		})

		assertAccepted(t, g, err)
		if g.method != http.MethodGet {
			t.Errorf("method = %s, want GET", g.method)
		}
		if g.rawBody != "" {
			t.Errorf("GET body = %q, want empty", g.rawBody)
		}
		if g.rawQuery != "page=2&page_size=50&status=completed" {
			t.Errorf("query = %q", g.rawQuery)
		}
	})

	t.Run("list banks with query param", func(t *testing.T) {
		t.Parallel()

		client, g := newGatewayClient(t)
		_, err := client.ListBanks(context.Background(), &gaian.ListBanksRequest{Country: "VN"})

		assertAccepted(t, g, err)
		if g.rawBody != "" {
			t.Errorf("GET body = %q, want empty", g.rawBody)
		}
		if g.rawQuery != "country=VN" {
			t.Errorf("query = %q", g.rawQuery)
		}
	})

	t.Run("wallet address with colons and a space", func(t *testing.T) {
		t.Parallel()

		client, g := newGatewayClient(t)
		_, err := client.GetUserByWallet(context.Background(), &gaian.GetUserByWalletRequest{
			WalletAddress: "eip155:1:0xAbC d",
		})

		assertAccepted(t, g, err)
		if g.gotPath != "/api/v2/users/wallet/eip155:1:0xAbC d" {
			t.Errorf("decoded path = %q", g.gotPath)
		}
	})
}

// TestGETLargeNumberQueryIsNotScientific guards fmt "%v" turning 1000000
// into "1e+06" in the query string, which the service would not parse.
func TestGETLargeNumberQueryIsNotScientific(t *testing.T) {
	t.Parallel()

	client, g := newGatewayClient(t)
	_, err := client.ListUserOrders(context.Background(), &gaian.ListUserOrdersRequest{
		UserID: "u1",
		Page:   1000000,
	})

	assertAccepted(t, g, err)
	if g.rawQuery != "page=1000000" {
		t.Errorf("query = %q, want page=1000000", g.rawQuery)
	}
}

func TestPOSTSignsBodyLikeGateway(t *testing.T) {
	t.Parallel()

	t.Run("no body", func(t *testing.T) {
		t.Parallel()

		client, g := newGatewayClient(t)
		_, err := client.CreateUser(context.Background(), &gaian.CreateUserRequest{})

		assertAccepted(t, g, err)
		if g.rawBody != "" {
			t.Errorf("body = %q, want empty", g.rawBody)
		}
	})

	t.Run("body with html characters", func(t *testing.T) {
		t.Parallel()

		client, g := newGatewayClient(t)
		_, err := client.CreateWallet(context.Background(), &gaian.CreateWalletRequest{
			UserID:  "u1",
			Address: "A & B <Co> Nguyễn",
			Chain:   "solana",
		})

		assertAccepted(t, g, err)
		if !strings.Contains(g.rawBody, "Nguyễn") {
			t.Errorf("body = %q, want it to carry the address", g.rawBody)
		}
	})
}
