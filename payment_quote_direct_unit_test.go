package gaian_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	gaian "github.com/gaiannetwork/gaian-go"
)

func TestQuoteDirect_Success(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
			`{"success":true,"requestId":"req_1","data":{"quoteId":"quote_1","fiatAmount":500000,`+
				`"fiatCurrency":"VND","settlementAmount":"20.5","settlementCurrency":"USDC","chainId":101,`+
				`"exchangeRate":"24390","qrInfo":{"accountNumber":"0123456789","bankBin":"970436","country":"VN"},`+
				`"expiresAt":"2030-01-01T00:00:00Z"}}`),
	}
	client := newMockClient(t, mock)

	resp, err := client.QuoteDirect(context.Background(), &gaian.QuoteDirectRequest{
		AccountNumber:      "0123456789",
		Code:               "970436",
		Amount:             500000,
		Country:            "VN",
		ChainID:            gaian.ChainSolana,
		SettlementCurrency: "USDC",
		UserID:             "usr_1",
	})
	if err != nil {
		t.Fatalf("QuoteDirect: %v", err)
	}
	if resp.Data.QuoteID != "quote_1" {
		t.Errorf("QuoteID = %q, want %q", resp.Data.QuoteID, "quote_1")
	}
}

// TestQuoteDirectRequest_WireFormat guards the omitempty payer-identity
// fields (WalletAddress/UserID/Email/BeneficiaryName): only the one that
// was actually set should appear in the body.
func TestQuoteDirectRequest_WireFormat(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, `{"success":true,"requestId":"req_1","data":{"quoteId":"quote_1"}}`), //nolint:bodyclose // response body closed by client
	}
	client := newMockClient(t, mock)

	if _, err := client.QuoteDirect(context.Background(), &gaian.QuoteDirectRequest{
		AccountNumber:      "0123456789",
		Code:               "970436",
		Amount:             500000,
		Country:            "VN",
		ChainID:            gaian.ChainSolana,
		SettlementCurrency: "USDC",
		UserID:             "usr_1",
	}); err != nil {
		t.Fatalf("QuoteDirect: %v", err)
	}

	var body map[string]any
	if err := json.Unmarshal(mock.lastBody(), &body); err != nil {
		t.Fatalf("unmarshal request body: %v", err)
	}
	if _, ok := body["walletAddress"]; ok {
		t.Error("empty optional walletAddress leaked into the request body")
	}
	if _, ok := body["email"]; ok {
		t.Error("empty optional email leaked into the request body")
	}
	if body["userId"] != "usr_1" {
		t.Errorf(`body["userId"] = %v, want "usr_1"`, body["userId"])
	}
}

func TestQuoteDirect_HTTPErrorStatuses(t *testing.T) {
	t.Parallel()

	statuses := []int{http.StatusBadRequest, http.StatusNotFound, http.StatusInternalServerError}

	for _, status := range statuses {
		t.Run(http.StatusText(status), func(t *testing.T) {
			t.Parallel()

			mock := &mockHTTPClient{
				response: mockResponse(status, `{"message":"rejected"}`), //nolint:bodyclose // response body closed by client
			}
			client := newMockClient(t, mock)

			_, err := client.QuoteDirect(context.Background(), &gaian.QuoteDirectRequest{AccountNumber: "0123456789", Code: "970436", UserID: "usr_1"})
			if !errors.Is(err, gaian.ErrUnexpectedStatus) {
				t.Errorf("QuoteDirect() error = %v, want wrapping %v", err, gaian.ErrUnexpectedStatus)
			}
		})
	}
}

func TestQuoteDirect_NetworkError(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: errConnectionRefused})

	_, err := client.QuoteDirect(context.Background(), &gaian.QuoteDirectRequest{AccountNumber: "0123456789", Code: "970436", UserID: "usr_1"})
	if !errors.Is(err, gaian.ErrHTTPFailure) {
		t.Errorf("QuoteDirect() error = %v, want wrapping %v", err, gaian.ErrHTTPFailure)
	}
}

func TestQuoteDirect_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.Canceled})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req := &gaian.QuoteDirectRequest{AccountNumber: "0123456789", Code: "970436", UserID: "usr_1"}
	if _, err := client.QuoteDirect(ctx, req); err == nil {
		t.Fatal("expected an error for a cancelled context, got nil")
	}
}

func TestQuoteDirect_ContextTimeout(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.DeadlineExceeded})

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	req := &gaian.QuoteDirectRequest{AccountNumber: "0123456789", Code: "970436", UserID: "usr_1"}
	if _, err := client.QuoteDirect(ctx, req); err == nil {
		t.Fatal("expected an error for a timed-out context, got nil")
	}
}

func TestQuoteDirect_InvalidJSON(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, "not json"), //nolint:bodyclose // response body closed by client
	}
	client := newMockClient(t, mock)

	req := &gaian.QuoteDirectRequest{AccountNumber: "0123456789", Code: "970436", UserID: "usr_1"}
	if _, err := client.QuoteDirect(context.Background(), req); err == nil {
		t.Fatal("expected a decode error for a malformed response body, got nil")
	}
}
