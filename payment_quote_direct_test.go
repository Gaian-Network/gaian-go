package gaian

import (
	"context"
	"net/http"
	"testing"
)

func TestQuoteDirect(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/quote/direct" {
			http.NotFound(w, r)
			return
		}
		writeJSON(t, w, http.StatusOK, map[string]any{
			"success": true,
			"data": map[string]any{
				"quoteId":            "quote_3",
				"route":              map[string]any{"routeId": 2},
				"fiatAmount":         100000.0,
				"fiatCurrency":       "VND",
				"settlementAmount":   "4.0",
				"settlementCurrency": "USDC",
				"chainId":            101,
				"exchangeRate":       "25000",
				"qrInfo": map[string]any{
					"accountNumber":   "0123456789",
					"beneficiaryName": "NGUYEN VAN A",
					"bankBin":         "970418",
					"country":         "VN",
				},
				"expiresAt": "2024-01-01T00:02:00Z",
			},
		})
	})

	resp, err := client.QuoteDirect(context.Background(), &QuoteDirectRequest{
		AccountNumber:      "0123456789",
		Code:               "970418",
		Amount:             100000,
		Country:            "VN",
		ChainID:            ChainSolana,
		SettlementCurrency: "USDC",
		UserID:             "usr_1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.QRInfo.BankBin != "970418" {
		t.Errorf("QRInfo.BankBin = %q, want 970418", resp.Data.QRInfo.BankBin)
	}
}
