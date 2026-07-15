package gaian_test

import (
	"context"
	"testing"
	"time"

	gaian "github.com/gaiannetwork/gaian-go"
)

// TestSandbox_VerifyOrder hits the real sandbox environment. Requires
// GAIAN_SANDBOX_ORDER_ID and GAIAN_SANDBOX_TX_HASH for a real order and
// a real broadcast on-chain transaction hash.
func TestSandbox_VerifyOrder(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)
	orderID := getTestOrderID(t)
	txHash := getTestTransactionHash(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := client.VerifyOrder(ctx, &gaian.VerifyOrderRequest{
		OrderID:         orderID,
		TransactionHash: txHash,
	})
	assertServerResponded(t, err)
}

func TestSandbox_VerifyOrder_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)
	orderID := getTestOrderID(t)
	txHash := getTestTransactionHash(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req := &gaian.VerifyOrderRequest{OrderID: orderID, TransactionHash: txHash}
	if _, err := client.VerifyOrder(ctx, req); err == nil {
		t.Error("expected an error for a cancelled context, got nil")
	}
}
