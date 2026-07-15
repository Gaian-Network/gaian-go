package gaian_test

import (
	"context"
	"testing"
	"time"

	gaian "github.com/gaiannetwork/gaian-go"
)

// TestSandbox_PlaceOrder hits the real sandbox environment. Requires
// GAIAN_SANDBOX_QUOTE_ID for a real, unexpired on-chain quote (from
// Quote, ~120s TTL) — generate one immediately before running this test.
func TestSandbox_PlaceOrder(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)
	quoteID := getTestQuoteID(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := client.PlaceOrder(ctx, &gaian.PlaceOrderRequest{
		QuoteID: quoteID,
	})
	assertServerResponded(t, err)
}

func TestSandbox_PlaceOrder_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)
	quoteID := getTestQuoteID(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := client.PlaceOrder(ctx, &gaian.PlaceOrderRequest{QuoteID: quoteID}); err == nil {
		t.Error("expected an error for a cancelled context, got nil")
	}
}
