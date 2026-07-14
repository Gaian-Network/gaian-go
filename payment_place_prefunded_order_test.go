package gaian

import (
	"context"
	"net/http"
	"testing"
)

func TestPlacePrefundedOrder(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v2/orders/prefund" {
			http.NotFound(w, r)
			return
		}
		writeJSON(t, w, http.StatusOK, map[string]any{
			"success": true,
			"data": map[string]any{
				"orderId":     "ord_2",
				"status":      3,
				"statusLabel": "queued",
			},
		})
	})

	resp, err := client.PlacePrefundedOrder(context.Background(), &PlacePrefundedOrderRequest{QuoteID: "quote_2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.StatusLabel != "queued" {
		t.Errorf("StatusLabel = %q, want queued", resp.Data.StatusLabel)
	}
}
