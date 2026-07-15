package gaian_test

import (
	"context"
	"testing"
	"time"

	gaian "github.com/gaiannetwork/gaian-go"
)

// TestSandbox_GetOrderStatus hits the real sandbox environment. Requires
// GAIAN_SANDBOX_ORDER_ID for a real, pre-existing order.
func TestSandbox_GetOrderStatus(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)
	orderID := getTestOrderID(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := client.GetOrderStatus(ctx, &gaian.GetOrderStatusRequest{
		OrderID: orderID,
	})
	assertServerResponded(t, err)
}

func TestSandbox_GetOrderStatus_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)
	orderID := getTestOrderID(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := client.GetOrderStatus(ctx, &gaian.GetOrderStatusRequest{OrderID: orderID}); err == nil {
		t.Error("expected an error for a cancelled context, got nil")
	}
}
