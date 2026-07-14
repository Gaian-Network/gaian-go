package gaian

import (
	"context"
	"net/http"
	"testing"
)

func TestPlaceOrder(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v2/orders" {
			http.NotFound(w, r)
			return
		}
		body := readBody(t, r)
		if string(body) != `{"quoteId":"quote_1"}` {
			t.Errorf("unexpected body: %s", body)
		}

		writeJSON(t, w, http.StatusOK, map[string]any{
			"success": true,
			"data": map[string]any{
				"orderId":            "ord_1",
				"status":             1,
				"statusLabel":        "awaiting_deposit",
				"chainId":            101,
				"depositAddress":     "SoLDepositAddr",
				"encodedTransaction": "base64tx==",
				"expiresAt":          "2024-01-01T00:05:00Z",
			},
		})
	})

	resp, err := client.PlaceOrder(context.Background(), &PlaceOrderRequest{QuoteID: "quote_1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.OrderID != "ord_1" || resp.Data.EncodedTransaction != "base64tx==" {
		t.Errorf("Data = %+v", resp.Data)
	}
}
