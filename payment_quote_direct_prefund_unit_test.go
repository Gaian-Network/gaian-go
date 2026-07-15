package gaian_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	gaian "github.com/gaiannetwork/gaian-go"
)

func TestQuoteDirectPrefund_Success(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
			`{"success":true,"requestId":"req_1","data":{"quoteId":"quote_1","fiatAmount":500000,`+
				`"fiatCurrency":"VND","settlementAmount":"20.5","settlementCurrency":"USDC",`+
				`"exchangeRate":"24390","qrInfo":{"accountNumber":"0123456789","bankBin":"970436","country":"VN"},`+
				`"expiresAt":"2030-01-01T00:00:00Z"}}`),
	}
	client := newMockClient(t, mock)

	resp, err := client.QuoteDirectPrefund(context.Background(), &gaian.QuoteDirectPrefundRequest{
		AccountNumber:      "0123456789",
		Code:               "970436",
		Amount:             500000,
		Country:            "VN",
		SettlementCurrency: "USDC",
		UserID:             "usr_1",
	})
	if err != nil {
		t.Fatalf("QuoteDirectPrefund: %v", err)
	}
	if resp.Data.QuoteID != "quote_1" {
		t.Errorf("QuoteID = %q, want %q", resp.Data.QuoteID, "quote_1")
	}
	if resp.Data.ChainID != 0 {
		t.Errorf("ChainID = %v, want 0 (no on-chain step for a prefunded quote)", resp.Data.ChainID)
	}
}

func TestQuoteDirectPrefund_HTTPErrorStatuses(t *testing.T) {
	t.Parallel()

	statuses := []int{http.StatusBadRequest, http.StatusInternalServerError}

	for _, status := range statuses {
		t.Run(http.StatusText(status), func(t *testing.T) {
			t.Parallel()

			mock := &mockHTTPClient{
				response: mockResponse(status, `{"message":"rejected"}`), //nolint:bodyclose // response body closed by client
			}
			client := newMockClient(t, mock)

			req := &gaian.QuoteDirectPrefundRequest{AccountNumber: "0123456789", Code: "970436", UserID: "usr_1"}
			_, err := client.QuoteDirectPrefund(context.Background(), req)
			if !errors.Is(err, gaian.ErrUnexpectedStatus) {
				t.Errorf("QuoteDirectPrefund() error = %v, want wrapping %v", err, gaian.ErrUnexpectedStatus)
			}
		})
	}
}

func TestQuoteDirectPrefund_NetworkError(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: errConnectionRefused})

	req := &gaian.QuoteDirectPrefundRequest{AccountNumber: "0123456789", Code: "970436", UserID: "usr_1"}
	_, err := client.QuoteDirectPrefund(context.Background(), req)
	if !errors.Is(err, gaian.ErrHTTPFailure) {
		t.Errorf("QuoteDirectPrefund() error = %v, want wrapping %v", err, gaian.ErrHTTPFailure)
	}
}

func TestQuoteDirectPrefund_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.Canceled})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req := &gaian.QuoteDirectPrefundRequest{AccountNumber: "0123456789", Code: "970436", UserID: "usr_1"}
	if _, err := client.QuoteDirectPrefund(ctx, req); err == nil {
		t.Fatal("expected an error for a cancelled context, got nil")
	}
}

func TestQuoteDirectPrefund_ContextTimeout(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.DeadlineExceeded})

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	req := &gaian.QuoteDirectPrefundRequest{AccountNumber: "0123456789", Code: "970436", UserID: "usr_1"}
	if _, err := client.QuoteDirectPrefund(ctx, req); err == nil {
		t.Fatal("expected an error for a timed-out context, got nil")
	}
}

func TestQuoteDirectPrefund_InvalidJSON(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, "not json"), //nolint:bodyclose // response body closed by client
	}
	client := newMockClient(t, mock)

	req := &gaian.QuoteDirectPrefundRequest{AccountNumber: "0123456789", Code: "970436", UserID: "usr_1"}
	if _, err := client.QuoteDirectPrefund(context.Background(), req); err == nil {
		t.Fatal("expected a decode error for a malformed response body, got nil")
	}
}
