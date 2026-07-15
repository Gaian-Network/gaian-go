package gaian_test

import (
	"context"
	"testing"
	"time"

	gaian "github.com/gaiannetwork/gaian-go"
)

// TestSandbox_GetOrganizationBalance hits the real sandbox environment.
// It's restricted to tenants on the prefund billing model, so a
// rejection here is an expected outcome for a non-prefund tenant.
func TestSandbox_GetOrganizationBalance(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := client.GetOrganizationBalance(ctx, &gaian.GetOrganizationBalanceRequest{})
	assertServerResponded(t, err)
}

func TestSandbox_GetOrganizationBalance_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := client.GetOrganizationBalance(ctx, &gaian.GetOrganizationBalanceRequest{}); err == nil {
		t.Error("expected an error for a cancelled context, got nil")
	}
}
