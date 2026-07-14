package gaian

import (
	"context"
	"net/http"
	"testing"
)

func TestQuote(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v2/quote" {
			http.NotFound(w, r)
			return
		}
		body := readBody(t, r)
		verifySignature(t, r, body)

		writeJSON(t, w, http.StatusOK, map[string]any{
			"success": true,
			"data": map[string]any{
				"quoteId":            "quote_1",
				"route":              map[string]any{"routeId": 1},
				"fiatAmount":         100000.0,
				"fiatCurrency":       "VND",
				"settlementAmount":   "4.0",
				"settlementCurrency": "USDC",
				"chainId":            101,
				"exchangeRate":       "25000",
				"expiresAt":          "2024-01-01T00:02:00Z",
			},
			"requestId": "req_6",
		})
	})

	resp, err := client.Quote(context.Background(), &QuoteRequest{
		QRString:           "00020101...",
		Amount:             100000,
		ChainID:            ChainSolana,
		SettlementCurrency: "USDC",
		UserID:             "usr_1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.QuoteID != "quote_1" {
		t.Errorf("QuoteID = %q, want quote_1", resp.Data.QuoteID)
	}
	if resp.Data.ChainID != ChainSolana {
		t.Errorf("ChainID = %d, want %d", resp.Data.ChainID, ChainSolana)
	}
}
