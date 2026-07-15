package gaian_test

import (
	"context"
	"testing"
	"time"

	gaian "github.com/gaiannetwork/gaian-go"
)

// TestSandbox_VerifyAccount hits the real sandbox environment.
// VerifyAccount is documented as production-only, so a rejection here is
// an expected outcome. Requires GAIAN_SANDBOX_BANK_CODE and
// GAIAN_SANDBOX_ACCOUNT_NUMBER to identify a real beneficiary account.
func TestSandbox_VerifyAccount(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)
	code := getTestBankCode(t)
	accountNumber := getTestAccountNumber(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := client.VerifyAccount(ctx, &gaian.VerifyAccountRequest{
		AccountNumber: accountNumber,
		Code:          code,
		Country:       "VN",
	})
	assertServerResponded(t, err)
}

func TestSandbox_VerifyAccount_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)
	code := getTestBankCode(t)
	accountNumber := getTestAccountNumber(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req := &gaian.VerifyAccountRequest{AccountNumber: accountNumber, Code: code, Country: "VN"}
	if _, err := client.VerifyAccount(ctx, req); err == nil {
		t.Error("expected an error for a cancelled context, got nil")
	}
}
