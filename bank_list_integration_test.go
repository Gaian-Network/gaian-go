package gaian_test

import (
	"context"
	"testing"
	"time"

	gaian "github.com/gaiannetwork/gaian-go"
)

// TestSandbox_ListBanks hits the real sandbox environment. ListBanks is
// documented as production-only, so a rejection here is an expected
// outcome — this test exists to prove the request/response wiring is
// correct end to end, not that sandbox actually returns bank data.
func TestSandbox_ListBanks(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := client.ListBanks(ctx, &gaian.ListBanksRequest{
		Country: "VN",
	})
	assertServerResponded(t, err)
}

func TestSandbox_ListBanks_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := client.ListBanks(ctx, &gaian.ListBanksRequest{Country: "VN"}); err == nil {
		t.Error("expected an error for a cancelled context, got nil")
	}
}
