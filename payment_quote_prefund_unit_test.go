package gaian_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	gaian "github.com/gaiannetwork/gaian-go"
)

func TestQuotePrefund_Success(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
			`{"success":true,"requestId":"req_1","data":{"quoteId":"quote_1","fiatAmount":500000,`+
				`"fiatCurrency":"VND","settlementAmount":"20.5","settlementCurrency":"USDC",`+
				`"exchangeRate":"24390","qrInfo":{"isValid":true},"expiresAt":"2030-01-01T00:00:00Z"}}`),
	}
	client := newMockClient(t, mock)

	resp, err := client.QuotePrefund(context.Background(), &gaian.QuotePrefundRequest{
		QRString:           "raw-qr",
		Amount:             500000,
		SettlementCurrency: "USDC",
		UserID:             "usr_1",
	})
	if err != nil {
		t.Fatalf("QuotePrefund: %v", err)
	}
	if resp.Data.QuoteID != "quote_1" {
		t.Errorf("QuoteID = %q, want %q", resp.Data.QuoteID, "quote_1")
	}
	// No on-chain step: ChainID must be the zero value on a prefund quote.
	if resp.Data.ChainID != 0 {
		t.Errorf("ChainID = %v, want 0 (no on-chain step for a prefunded quote)", resp.Data.ChainID)
	}
}

func TestQuotePrefund_HTTPErrorStatuses(t *testing.T) {
	t.Parallel()

	statuses := []int{http.StatusBadRequest, http.StatusInternalServerError}

	for _, status := range statuses {
		t.Run(http.StatusText(status), func(t *testing.T) {
			t.Parallel()

			mock := &mockHTTPClient{
				response: mockResponse(status, `{"message":"rejected"}`), //nolint:bodyclose // response body closed by client
			}
			client := newMockClient(t, mock)

			_, err := client.QuotePrefund(context.Background(), &gaian.QuotePrefundRequest{QRString: "raw-qr", UserID: "usr_1"})
			if !errors.Is(err, gaian.ErrUnexpectedStatus) {
				t.Errorf("QuotePrefund() error = %v, want wrapping %v", err, gaian.ErrUnexpectedStatus)
			}
		})
	}
}

func TestQuotePrefund_NetworkError(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: errConnectionRefused})

	_, err := client.QuotePrefund(context.Background(), &gaian.QuotePrefundRequest{QRString: "raw-qr", UserID: "usr_1"})
	if !errors.Is(err, gaian.ErrHTTPFailure) {
		t.Errorf("QuotePrefund() error = %v, want wrapping %v", err, gaian.ErrHTTPFailure)
	}
}

func TestQuotePrefund_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.Canceled})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := client.QuotePrefund(ctx, &gaian.QuotePrefundRequest{QRString: "raw-qr", UserID: "usr_1"}); err == nil {
		t.Fatal("expected an error for a cancelled context, got nil")
	}
}

func TestQuotePrefund_ContextTimeout(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.DeadlineExceeded})

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	if _, err := client.QuotePrefund(ctx, &gaian.QuotePrefundRequest{QRString: "raw-qr", UserID: "usr_1"}); err == nil {
		t.Fatal("expected an error for a timed-out context, got nil")
	}
}

func TestQuotePrefund_InvalidJSON(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, "not json"), //nolint:bodyclose // response body closed by client
	}
	client := newMockClient(t, mock)

	if _, err := client.QuotePrefund(context.Background(), &gaian.QuotePrefundRequest{QRString: "raw-qr", UserID: "usr_1"}); err == nil {
		t.Fatal("expected a decode error for a malformed response body, got nil")
	}
}
