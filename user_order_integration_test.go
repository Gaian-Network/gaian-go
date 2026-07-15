package gaian_test

import (
	"context"
	"testing"
	"time"

	gaian "github.com/gaiannetwork/gaian-go"
)

// TestSandbox_GetUserOrder hits the real sandbox environment. Requires
// GAIAN_SANDBOX_USER_ID and GAIAN_SANDBOX_ORDER_ID for a real order
// belonging to a real user.
func TestSandbox_GetUserOrder(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)
	userID := getTestUserID(t)
	orderID := getTestOrderID(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := client.GetUserOrder(ctx, &gaian.GetUserOrderRequest{
		UserID:  userID,
		OrderID: orderID,
	})
	assertServerResponded(t, err)
}

func TestSandbox_GetUserOrder_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)
	userID := getTestUserID(t)
	orderID := getTestOrderID(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := client.GetUserOrder(ctx, &gaian.GetUserOrderRequest{UserID: userID, OrderID: orderID}); err == nil {
		t.Error("expected an error for a cancelled context, got nil")
	}
}
