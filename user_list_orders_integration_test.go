package gaian_test

import (
	"context"
	"testing"
	"time"

	gaian "github.com/gaiannetwork/gaian-go"
)

// TestSandbox_ListUserOrders hits the real sandbox environment. Requires
// GAIAN_SANDBOX_USER_ID for a real, pre-existing sandbox user.
func TestSandbox_ListUserOrders(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)
	userID := getTestUserID(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := client.ListUserOrders(ctx, &gaian.ListUserOrdersRequest{
		UserID: userID,
		Page:   1,
	})
	assertServerResponded(t, err)
}

func TestSandbox_ListUserOrders_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)
	userID := getTestUserID(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := client.ListUserOrders(ctx, &gaian.ListUserOrdersRequest{UserID: userID, Page: 1}); err == nil {
		t.Error("expected an error for a cancelled context, got nil")
	}
}
