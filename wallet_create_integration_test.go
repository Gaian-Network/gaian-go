package gaian_test

import (
	"context"
	"testing"
	"time"

	gaian "github.com/gaiannetwork/gaian-go"
)

// TestSandbox_CreateWallet hits the real sandbox environment. Requires
// GAIAN_SANDBOX_USER_ID for a real, pre-existing sandbox user; the
// wallet address itself is randomly generated per run so repeated runs
// don't collide with a wallet linked by a previous one.
func TestSandbox_CreateWallet(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)
	userID := getTestUserID(t)
	address := generateTestWalletAddress(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := client.CreateWallet(ctx, &gaian.CreateWalletRequest{
		UserID:  userID,
		Address: address,
		Chain:   "base",
	})
	assertServerResponded(t, err)
}

func TestSandbox_CreateWallet_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)
	userID := getTestUserID(t)
	address := generateTestWalletAddress(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req := &gaian.CreateWalletRequest{UserID: userID, Address: address, Chain: "base"}
	if _, err := client.CreateWallet(ctx, req); err == nil {
		t.Error("expected an error for a cancelled context, got nil")
	}
}
