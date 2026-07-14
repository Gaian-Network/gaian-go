package gaian

import (
	"context"
	"net/http"
	"testing"
)

func TestQuotePrefund(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/quote/prefund" {
			http.NotFound(w, r)
			return
		}
		writeJSON(t, w, http.StatusOK, map[string]any{
			"success": true,
			"data": map[string]any{
				"quoteId":            "quote_2",
				"route":              map[string]any{"routeId": 1},
				"fiatAmount":         100000.0,
				"fiatCurrency":       "VND",
				"settlementAmount":   "4.0",
				"settlementCurrency": "USDC",
				"exchangeRate":       "25000",
				"expiresAt":          "2024-01-01T00:02:00Z",
			},
		})
	})

	resp, err := client.QuotePrefund(context.Background(), &QuotePrefundRequest{
		QRString:           "00020101...",
		Amount:             100000,
		SettlementCurrency: "USDC",
		UserID:             "usr_1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.QuoteID != "quote_2" {
		t.Errorf("QuoteID = %q, want quote_2", resp.Data.QuoteID)
	}
}
