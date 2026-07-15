package gaian_test

import (
	"context"
	"testing"
	"time"

	gaian "github.com/gaiannetwork/gaian-go"
)

// TestSandbox_QuoteDirectPrefund hits the real sandbox environment.
// QuoteDirectPrefund is documented as production-only, so a rejection
// here is an expected outcome. Requires GAIAN_SANDBOX_USER_ID,
// GAIAN_SANDBOX_BANK_CODE, and GAIAN_SANDBOX_ACCOUNT_NUMBER.
func TestSandbox_QuoteDirectPrefund(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)
	userID := getTestUserID(t)
	code := getTestBankCode(t)
	accountNumber := getTestAccountNumber(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := client.QuoteDirectPrefund(ctx, &gaian.QuoteDirectPrefundRequest{
		AccountNumber:      accountNumber,
		Code:               code,
		Amount:             500000,
		Country:            "VN",
		SettlementCurrency: "USDC",
		UserID:             userID,
	})
	assertServerResponded(t, err)
}

func TestSandbox_QuoteDirectPrefund_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)
	userID := getTestUserID(t)
	code := getTestBankCode(t)
	accountNumber := getTestAccountNumber(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req := &gaian.QuoteDirectPrefundRequest{
		AccountNumber:      accountNumber,
		Code:               code,
		Amount:             500000,
		Country:            "VN",
		SettlementCurrency: "USDC",
		UserID:             userID,
	}
	if _, err := client.QuoteDirectPrefund(ctx, req); err == nil {
		t.Error("expected an error for a cancelled context, got nil")
	}
}
