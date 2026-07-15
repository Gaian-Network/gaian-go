package gaian_test

import (
	"context"
	"testing"
	"time"

	gaian "github.com/gaiannetwork/gaian-go"
)

// TestSandbox_QuoteDirect hits the real sandbox environment. QuoteDirect
// is documented as production-only, so a rejection here is an expected
// outcome. Requires GAIAN_SANDBOX_USER_ID, GAIAN_SANDBOX_BANK_CODE, and
// GAIAN_SANDBOX_ACCOUNT_NUMBER.
func TestSandbox_QuoteDirect(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)
	userID := getTestUserID(t)
	code := getTestBankCode(t)
	accountNumber := getTestAccountNumber(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := client.QuoteDirect(ctx, &gaian.QuoteDirectRequest{
		AccountNumber:      accountNumber,
		Code:               code,
		Amount:             500000,
		Country:            "VN",
		ChainID:            gaian.ChainSolana,
		SettlementCurrency: "USDC",
		UserID:             userID,
	})
	assertServerResponded(t, err)
}

func TestSandbox_QuoteDirect_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)
	userID := getTestUserID(t)
	code := getTestBankCode(t)
	accountNumber := getTestAccountNumber(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req := &gaian.QuoteDirectRequest{
		AccountNumber:      accountNumber,
		Code:               code,
		Amount:             500000,
		Country:            "VN",
		ChainID:            gaian.ChainSolana,
		SettlementCurrency: "USDC",
		UserID:             userID,
	}
	if _, err := client.QuoteDirect(ctx, req); err == nil {
		t.Error("expected an error for a cancelled context, got nil")
	}
}
