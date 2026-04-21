package gaian_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	gaian "github.com/gaiannetwork/gaian-go"
)

// testClient wires a Client against a local test server.
// userHandler handles User-service paths; pmtHandler handles Payment-service paths.
func testClient(t *testing.T, userHandler, pmtHandler http.HandlerFunc) *gaian.Client {
	t.Helper()
	userSrv := httptest.NewServer(userHandler)
	pmtSrv := httptest.NewServer(pmtHandler)
	t.Cleanup(func() { userSrv.Close(); pmtSrv.Close() })

	return gaian.NewWithURLs("test-key", userSrv.URL, pmtSrv.URL)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// ── User ──────────────────────────────────────────────────────────────────────

func TestRegister(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/user/register" {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, 200, map[string]any{
			"status":  "success",
			"message": "user created",
			"user": map[string]any{
				"id":            42,
				"walletAddress": "0xABC",
				"kycStatus":     "not_started",
				"createdAt":     "2024-01-01T00:00:00Z",
				"updatedAt":     "2024-01-01T00:00:00Z",
			},
		})
	}, nil)

	user, err := c.User.Register(context.Background(), "0xABC")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != 42 {
		t.Errorf("got id %d, want 42", user.ID)
	}
	if user.KYCStatus != gaian.KYCNotStarted {
		t.Errorf("got kyc status %q, want %q", user.KYCStatus, gaian.KYCNotStarted)
	}
}

func TestRegisterConflict(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 409, map[string]any{
			"status":  "error",
			"message": "user already exists",
		})
	}, nil)

	_, err := c.User.Register(context.Background(), "0xABC")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !gaian.IsConflict(err) {
		t.Errorf("expected conflict error, got %T: %v", err, err)
	}
}

func TestGetByWallet(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("walletAddress") == "" {
			http.Error(w, "missing walletAddress", 400)
			return
		}
		writeJSON(w, 200, map[string]any{
			"status": "success",
			"data": map[string]any{
				"id":            7,
				"walletAddress": "0xDEF",
				"kycStatus":     "approved",
				"createdAt":     "2024-01-01T00:00:00Z",
				"updatedAt":     "2024-01-01T00:00:00Z",
			},
		})
	}, nil)

	user, err := c.User.GetByWallet(context.Background(), "0xDEF")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.KYCStatus != gaian.KYCApproved {
		t.Errorf("got %q, want approved", user.KYCStatus)
	}
}

func TestGetMarkets(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{
			"status": "success",
			"data": map[string]any{
				"markets": map[string]any{
					"VN": map[string]any{"status": "approved"},
					"PH": map[string]any{"status": "pending"},
				},
			},
		})
	}, nil)

	markets, err := c.User.GetMarkets(context.Background(), "0xABC")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if markets["VN"].Status != gaian.MarketApproved {
		t.Errorf("VN: got %q, want approved", markets["VN"].Status)
	}
	if markets["PH"].Status != gaian.MarketPending {
		t.Errorf("PH: got %q, want pending", markets["PH"].Status)
	}
}

func TestListOrdersByWallet(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{
			"status": "success",
			"data": map[string]any{
				"user": map[string]any{
					"id": 1, "walletAddress": "0xABC",
					"kycStatus": "approved",
					"createdAt": "2024-01-01T00:00:00Z",
					"updatedAt": "2024-01-01T00:00:00Z",
				},
				"orders": map[string]any{
					"items": []map[string]any{
						{
							"id": 1, "orderId": "ord_001",
							"status": "completed", "fiatAmount": 100000.0,
							"fiatCurrency": "VND", "cryptoAmount": 4.0,
							"cryptoCurrency": "USDC", "exchangeRate": 25000.0,
							"expiresAt": "2024-01-01T01:00:00Z",
							"createdAt": "2024-01-01T00:00:00Z",
							"updatedAt": "2024-01-01T00:30:00Z",
						},
					},
					"pagination": map[string]any{
						"page": 1, "limit": 20, "total": 1,
						"totalPages": 1, "hasNext": false, "hasPrev": false,
					},
				},
			},
		})
	}, nil)

	list, err := c.User.ListOrdersByWallet(context.Background(), "0xABC", gaian.OrderListOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(list.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(list.Items))
	}
	if list.Items[0].OrderID != "ord_001" {
		t.Errorf("got order id %q, want ord_001", list.Items[0].OrderID)
	}
}

func TestGenerateKYCLink(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{
			"status":    true,
			"message":   "ok",
			"websdkUrl": "https://kyc.example.com/session/abc123",
		})
	}, nil)

	link, err := c.User.GenerateKYCLink(context.Background(), "0xABC")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if link != "https://kyc.example.com/session/abc123" {
		t.Errorf("unexpected link: %s", link)
	}
}

// ── Payment ───────────────────────────────────────────────────────────────────

func TestPlaceOrder(t *testing.T) {
	c := testClient(t, nil, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/placeOrder" {
			http.NotFound(w, r)
			return
		}
		writeJSON(w, 200, map[string]any{
			"orderId":        "ord_standard_001",
			"status":         "awaiting_crypto_transfer",
			"fiatAmount":     100000.0,
			"fiatCurrency":   "VND",
			"cryptoAmount":   4.0,
			"cryptoCurrency": "USDC",
			"exchangeRate":   25000.0,
			"routeId":        1,
		})
	})

	order, err := c.Payment.PlaceOrder(context.Background(), gaian.PlaceOrderRequest{
		QRString:       "00020101021238...",
		Amount:         100000,
		CryptoCurrency: gaian.USDC,
		FromAddress:    "0xABC",
		FiatCurrency:   gaian.VND,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.OrderID != "ord_standard_001" {
		t.Errorf("unexpected order id: %s", order.OrderID)
	}
	if order.Status != gaian.OrderAwaitingCryptoTransfer {
		t.Errorf("unexpected status: %s", order.Status)
	}
}

func TestVerifyOrder(t *testing.T) {
	c := testClient(t, nil, func(w http.ResponseWriter, r *http.Request) {
		status := "queued"
		writeJSON(w, 200, map[string]any{
			"orderId":            "ord_001",
			"status":             "verified",
			"transactionHash":    "0xTXHASH",
			"message":            "verified",
			"bankTransferStatus": &status,
		})
	})

	result, err := c.Payment.VerifyOrder(context.Background(), "ord_001", "0xTXHASH")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != gaian.OrderVerified {
		t.Errorf("got %q, want verified", result.Status)
	}
}

func TestGetOrderStatus(t *testing.T) {
	c := testClient(t, nil, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("orderId") == "" {
			http.Error(w, "missing orderId", 400)
			return
		}
		// /api/v1/status returns the Order object directly — no { data } wrapper
		writeJSON(w, 200, map[string]any{
			"id": 1, "orderId": "ord_001",
			"status": "completed", "fiatAmount": 100000.0,
			"fiatCurrency": "VND", "cryptoAmount": 4.0,
			"cryptoCurrency": "USDC", "exchangeRate": 25000.0,
			"expiresAt": "2024-01-01T01:00:00Z",
			"createdAt": "2024-01-01T00:00:00Z",
			"updatedAt": "2024-01-01T00:30:00Z",
		})
	})

	order, err := c.Payment.GetOrderStatus(context.Background(), "ord_001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if order.Status != gaian.OrderCompleted {
		t.Errorf("got %q, want completed", order.Status)
	}
}

func TestParseQR(t *testing.T) {
	c := testClient(t, nil, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{
			"success": true,
			"qrInfo": map[string]any{
				"isValid":         true,
				"country":         "VN",
				"beneficiaryName": "NGUYEN VAN A",
				"amount":          500000.0,
				"currency":        "VND",
			},
			"timestamp": "2024-01-01T00:00:00Z",
		})
	})

	info, err := c.Payment.ParseQR(context.Background(), "00020101...")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.IsValid {
		t.Error("expected isValid=true")
	}
	if info.Country != "VN" {
		t.Errorf("got country %q, want VN", info.Country)
	}
}

func TestCalculateExchange(t *testing.T) {
	c := testClient(t, nil, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{
			"success": true,
			"exchangeInfo": map[string]any{
				"fiatAmount":     100000.0,
				"fiatCurrency":   "VND",
				"cryptoAmount":   "4.0",
				"cryptoCurrency": "USDC",
				"exchangeRate":   "25000",
				"chain":          "Solana",
				"token":          "USDC",
				"feeAmount":      "0.05",
				"timestamp":      "2024-01-01T00:00:00Z",
			},
		})
	})

	info, err := c.Payment.CalculateExchange(context.Background(), gaian.CalculateExchangeRequest{
		Amount:  100000,
		Country: "VN",
		Chain:   gaian.Solana,
		Token:   gaian.USDC,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.CryptoAmount != "4.0" {
		t.Errorf("got crypto amount %q, want 4.0", info.CryptoAmount)
	}
}

// ── Tenant & Policy ───────────────────────────────────────────────────────────

func TestGetBalance(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{
			"currency":         "USDC",
			"availableBalance": 100.5,
			"walletAddress":    "0xTENANT",
			"chain":            "Solana",
		})
	}, nil)

	bal, err := c.Tenant.GetBalance(context.Background(), gaian.USDC)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bal.AvailableBalance != 100.5 {
		t.Errorf("got balance %v, want 100.5", bal.AvailableBalance)
	}
}

func TestCheckPolicy(t *testing.T) {
	c := testClient(t, nil, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{
			"isAllowed": true,
			"routeId":   1,
			"limits":    map[string]any{"perTransaction": 500.0, "perDay": 5000.0},
			"tier":      "KYC",
		})
	})

	policy, err := c.Policy.CheckPolicy(context.Background(), "0xABC", "VN")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !policy.IsAllowed {
		t.Error("expected isAllowed=true")
	}
	if policy.Tier != gaian.TierKYC {
		t.Errorf("got tier %q, want KYC", policy.Tier)
	}
	if policy.Limits.PerTransaction != 500 {
		t.Errorf("got per-tx limit %v, want 500", policy.Limits.PerTransaction)
	}
}

// ── Error helpers ─────────────────────────────────────────────────────────────

func TestIsNotFound(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 404, map[string]any{"status": "error", "message": "user not found"})
	}, nil)

	_, err := c.User.GetByID(context.Background(), 999)
	if !gaian.IsNotFound(err) {
		t.Errorf("expected IsNotFound=true, got %T: %v", err, err)
	}
}

func TestIsUnauthorized(t *testing.T) {
	c := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 401, map[string]any{"status": "error", "message": "invalid api key"})
	}, nil)

	_, err := c.User.GetByID(context.Background(), 1)
	if !gaian.IsUnauthorized(err) {
		t.Errorf("expected IsUnauthorized=true, got %T: %v", err, err)
	}
}
