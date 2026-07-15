package gaian_test

import (
	"context"
	"testing"
	"time"

	gaian "github.com/gaiannetwork/gaian-go"
)

// TestSandbox_ParseQR hits the real sandbox environment. Requires
// GAIAN_SANDBOX_QR_STRING for a real QR payment string — parsing
// garbage input isn't an error (the API returns isValid=false with
// HTTP 200), so this test wouldn't exercise anything meaningful without
// a real fixture.
func TestSandbox_ParseQR(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)
	qrString := getTestQRString(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := client.ParseQR(ctx, &gaian.ParseQRRequest{
		QRString: qrString,
		Country:  "VN",
	})
	assertServerResponded(t, err)
}

func TestSandbox_ParseQR_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)
	qrString := getTestQRString(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req := &gaian.ParseQRRequest{QRString: qrString, Country: "VN"}
	if _, err := client.ParseQR(ctx, req); err == nil {
		t.Error("expected an error for a cancelled context, got nil")
	}
}
