package gaian_test

import (
	"context"
	"testing"
	"time"

	gaian "github.com/gaiannetwork/gaian-go"
)

// TestSandbox_ListWallets hits the real sandbox environment. Requires
// GAIAN_SANDBOX_USER_ID for a real, pre-existing sandbox user.
func TestSandbox_ListWallets(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)
	userID := getTestUserID(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := client.ListWallets(ctx, &gaian.ListWalletsRequest{
		UserID: userID,
	})
	assertServerResponded(t, err)
}

func TestSandbox_ListWallets_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)
	userID := getTestUserID(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := client.ListWallets(ctx, &gaian.ListWalletsRequest{UserID: userID}); err == nil {
		t.Error("expected an error for a cancelled context, got nil")
	}
}
