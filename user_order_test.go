package gaian

import (
	"context"
	"net/http"
	"testing"
)

func TestGetUserOrder(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/users/usr_1/orders/ord_1" {
			http.NotFound(w, r)
			return
		}
		writeJSON(t, w, http.StatusOK, map[string]any{
			"data": map[string]any{"orderId": "ord_1", "status": "completed"},
		})
	})

	resp, err := client.GetUserOrder(context.Background(), &GetUserOrderRequest{UserID: "usr_1", OrderID: "ord_1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.OrderID != "ord_1" {
		t.Errorf("OrderID = %q, want ord_1", resp.Data.OrderID)
	}
}
