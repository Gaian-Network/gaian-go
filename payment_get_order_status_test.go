package gaian

import (
	"context"
	"net/http"
	"testing"
)

func TestGetOrderStatus(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/orders/ord_1" {
			http.NotFound(w, r)
			return
		}
		writeJSON(t, w, http.StatusOK, map[string]any{
			"success": true,
			"data": map[string]any{
				"orderId":            "ord_1",
				"status":             10,
				"statusLabel":        "completed",
				"chainId":            101,
				"fiatAmount":         100000.0,
				"fiatCurrency":       "VND",
				"settlementAmount":   "4.0",
				"settlementCurrency": "USDC",
				"exchangeRate":       "25000",
				"createdAt":          "2024-01-01T00:00:00Z",
				"updatedAt":          "2024-01-01T00:03:00Z",
			},
		})
	})

	resp, err := client.GetOrderStatus(context.Background(), &GetOrderStatusRequest{OrderID: "ord_1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Data.Status.IsTerminal() {
		t.Errorf("expected OrderCompleted to be terminal, got status=%d", resp.Data.Status)
	}
}

// TestOrderStatus_IsTerminal is pure business logic that every polling
// loop (GetOrderStatus) depends on — worth a dedicated table-driven unit
// test independent of the HTTP round trip above.
func TestOrderStatus_IsTerminal(t *testing.T) {
	tests := []struct {
		status OrderStatus
		want   bool
	}{
		{OrderPending, false},
		{OrderAwaitingDeposit, false},
		{OrderPaymentReceived, false},
		{OrderQueued, false},
		{OrderProcessing, false},
		{OrderCompleted, true},
		{OrderFailed, true},
		{OrderCancelled, true},
		{OrderExpired, true},
		{OrderUnknown, false},
		{OrderStatus(-1), false}, // unrecognized value must not be treated as terminal
	}

	for _, tt := range tests {
		if got := tt.status.IsTerminal(); got != tt.want {
			t.Errorf("OrderStatus(%d).IsTerminal() = %v, want %v", tt.status, got, tt.want)
		}
	}
}
