package gaian_test

import (
	"context"
	"testing"
	"time"

	gaian "github.com/gaiannetwork/gaian-go"
)

// TestSandbox_UpdateKYC hits the real sandbox environment. Requires
// GAIAN_SANDBOX_USER_ID for a real, pre-existing sandbox user.
func TestSandbox_UpdateKYC(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)
	userID := getTestUserID(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	email := "sandbox-updated@example.com"
	_, err := client.UpdateKYC(ctx, &gaian.UpdateKYCRequest{
		UserID: userID,
		Email:  &email,
	})
	assertServerResponded(t, err)
}

func TestSandbox_UpdateKYC_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)
	userID := getTestUserID(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	email := "sandbox-updated@example.com"
	if _, err := client.UpdateKYC(ctx, &gaian.UpdateKYCRequest{UserID: userID, Email: &email}); err == nil {
		t.Error("expected an error for a cancelled context, got nil")
	}
}
