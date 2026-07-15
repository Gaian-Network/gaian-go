package gaian_test

import (
	"context"
	"testing"
	"time"

	gaian "github.com/gaiannetwork/gaian-go"
)

// TestSandbox_GetUserByWallet hits the real sandbox environment. Requires
// GAIAN_SANDBOX_WALLET_ADDRESS for a wallet linked to a real sandbox
// user.
func TestSandbox_GetUserByWallet(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)
	walletAddress := getTestWalletAddress(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := client.GetUserByWallet(ctx, &gaian.GetUserByWalletRequest{
		WalletAddress: walletAddress,
	})
	assertServerResponded(t, err)
}

func TestSandbox_GetUserByWallet_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)
	walletAddress := getTestWalletAddress(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := client.GetUserByWallet(ctx, &gaian.GetUserByWalletRequest{WalletAddress: walletAddress}); err == nil {
		t.Error("expected an error for a cancelled context, got nil")
	}
}
