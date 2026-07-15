package gaian_test

import (
	"context"
	"testing"
	"time"

	gaian "github.com/gaiannetwork/gaian-go"
)

// TestSandbox_PlacePrefundedOrder hits the real sandbox environment.
// Requires GAIAN_SANDBOX_PREFUND_QUOTE_ID for a real, unexpired prefund
// quote (from QuotePrefund, ~120s TTL) — generate one immediately before
// running this test.
func TestSandbox_PlacePrefundedOrder(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)
	quoteID := getTestPrefundQuoteID(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := client.PlacePrefundedOrder(ctx, &gaian.PlacePrefundedOrderRequest{
		QuoteID: quoteID,
	})
	assertServerResponded(t, err)
}

func TestSandbox_PlacePrefundedOrder_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)
	quoteID := getTestPrefundQuoteID(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := client.PlacePrefundedOrder(ctx, &gaian.PlacePrefundedOrderRequest{QuoteID: quoteID}); err == nil {
		t.Error("expected an error for a cancelled context, got nil")
	}
}
