package gaian_test

import (
	"context"
	"testing"
	"time"

	gaian "github.com/gaiannetwork/gaian-go"
)

// TestSandbox_QuotePrefund hits the real sandbox environment. Requires
// GAIAN_SANDBOX_USER_ID and GAIAN_SANDBOX_QR_STRING for a real,
// KYC-approved user and a real QR payment string.
func TestSandbox_QuotePrefund(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)
	userID := getTestUserID(t)
	qrString := getTestQRString(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := client.QuotePrefund(ctx, &gaian.QuotePrefundRequest{
		QRString:           qrString,
		Amount:             500000,
		Country:            "VN",
		SettlementCurrency: "USDC",
		UserID:             userID,
	})
	assertServerResponded(t, err)
}

func TestSandbox_QuotePrefund_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)
	userID := getTestUserID(t)
	qrString := getTestQRString(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req := &gaian.QuotePrefundRequest{
		QRString:           qrString,
		Amount:             500000,
		Country:            "VN",
		SettlementCurrency: "USDC",
		UserID:             userID,
	}
	if _, err := client.QuotePrefund(ctx, req); err == nil {
		t.Error("expected an error for a cancelled context, got nil")
	}
}
