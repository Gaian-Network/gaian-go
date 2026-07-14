package gaian

import (
	"context"
	"net/http"
	"testing"
)

func TestVerifyOrder(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v2/orders/ord_1/verify" {
			http.NotFound(w, r)
			return
		}
		writeJSON(t, w, http.StatusOK, map[string]any{
			"success": true,
			"data": map[string]any{
				"orderId":         "ord_1",
				"status":          1,
				"statusLabel":     "verified",
				"transactionHash": "0xTXHASH",
				"message":         "verified",
			},
		})
	})

	resp, err := client.VerifyOrder(context.Background(), &VerifyOrderRequest{
		OrderID:         "ord_1",
		TransactionHash: "0xTXHASH",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Status != VerifyVerified {
		t.Errorf("Status = %d, want VerifyVerified", resp.Data.Status)
	}
}

// TestVerifyOrder_PendingIsNotAnError locks in the documented "returns
// HTTP 200 even when verification did not succeed" behavior: a nil error
// does NOT mean the payment is confirmed, callers must branch on Status.
func TestVerifyOrder_PendingIsNotAnError(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusOK, map[string]any{
			"success": true,
			"data": map[string]any{
				"orderId":         "ord_1",
				"status":          0,
				"statusLabel":     "pending",
				"transactionHash": "0xTXHASH",
				"message":         "awaiting on-chain confirmation",
			},
		})
	})

	resp, err := client.VerifyOrder(context.Background(), &VerifyOrderRequest{
		OrderID:         "ord_1",
		TransactionHash: "0xTXHASH",
	})
	if err != nil {
		t.Fatalf("unexpected error (should be nil even though verification is pending): %v", err)
	}
	if resp.Data.Status != VerifyPending {
		t.Errorf("Status = %d, want VerifyPending", resp.Data.Status)
	}
}
