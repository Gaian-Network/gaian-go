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

func TestQuote_Success(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
			`{"success":true,"requestId":"req_1","data":{"quoteId":"quote_1","route":{"routeId":1},"fiatAmount":500000,`+
				`"fiatCurrency":"VND","settlementAmount":"20.5","settlementCurrency":"USDC","chainId":101,`+
				`"exchangeRate":"24390","protocolFeeUsd":"0.1","minimumFeeUsd":"0.05","qrInfo":{"isValid":true,`+
				`"encodedString":"raw-qr","country":"VN","qrProvider":"napas","bankBin":"970436","accountNumber":"0123456789",`+
				`"beneficiaryName":"NGUYEN VAN A"},"expiresAt":"2030-01-01T00:00:00Z"}}`),
	}
	client := newMockClient(t, mock)

	resp, err := client.Quote(context.Background(), &gaian.QuoteRequest{
		QRString:           "raw-qr",
		Amount:             500000,
		ChainID:            gaian.ChainSolana,
		SettlementCurrency: "USDC",
		UserID:             "usr_1",
	})
	if err != nil {
		t.Fatalf("Quote: %v", err)
	}
	if resp.Data.QuoteID != "quote_1" {
		t.Errorf("QuoteID = %q, want %q", resp.Data.QuoteID, "quote_1")
	}
	if resp.Data.ChainID != gaian.ChainSolana {
		t.Errorf("ChainID = %v, want %v", resp.Data.ChainID, gaian.ChainSolana)
	}
	if !resp.Data.QRInfo.IsValid {
		t.Error("expected QRInfo.IsValid=true")
	}
}

// TestQuoteRequest_WireFormat guards the omitempty Country field: it
// must not appear in the body when left unset.
func TestQuoteRequest_WireFormat(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, `{"success":true,"requestId":"req_1","data":{"quoteId":"quote_1"}}`), //nolint:bodyclose // response body closed by client
	}
	client := newMockClient(t, mock)

	if _, err := client.Quote(context.Background(), &gaian.QuoteRequest{
		QRString:           "raw-qr",
		Amount:             500000,
		ChainID:            gaian.ChainSolana,
		SettlementCurrency: "USDC",
		UserID:             "usr_1",
	}); err != nil {
		t.Fatalf("Quote: %v", err)
	}

	var body map[string]any
	if err := json.Unmarshal(mock.lastBody(), &body); err != nil {
		t.Fatalf("unmarshal request body: %v", err)
	}
	if _, ok := body["country"]; ok {
		t.Error("empty optional country leaked into the request body")
	}
}

func TestQuote_HTTPErrorStatuses(t *testing.T) {
	t.Parallel()

	statuses := []int{
		http.StatusBadRequest,
		http.StatusNotFound, // e.g. no route available for this QR/amount
		http.StatusInternalServerError,
	}

	for _, status := range statuses {
		t.Run(http.StatusText(status), func(t *testing.T) {
			t.Parallel()

			mock := &mockHTTPClient{
				response: mockResponse(status, `{"message":"rejected"}`), //nolint:bodyclose // response body closed by client
			}
			client := newMockClient(t, mock)

			_, err := client.Quote(context.Background(), &gaian.QuoteRequest{QRString: "raw-qr", UserID: "usr_1"})
			if !errors.Is(err, gaian.ErrUnexpectedStatus) {
				t.Errorf("Quote() error = %v, want wrapping %v", err, gaian.ErrUnexpectedStatus)
			}
		})
	}
}

func TestQuote_NetworkError(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: errConnectionRefused})

	_, err := client.Quote(context.Background(), &gaian.QuoteRequest{QRString: "raw-qr", UserID: "usr_1"})
	if !errors.Is(err, gaian.ErrHTTPFailure) {
		t.Errorf("Quote() error = %v, want wrapping %v", err, gaian.ErrHTTPFailure)
	}
}

func TestQuote_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.Canceled})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := client.Quote(ctx, &gaian.QuoteRequest{QRString: "raw-qr", UserID: "usr_1"}); err == nil {
		t.Fatal("expected an error for a cancelled context, got nil")
	}
}

func TestQuote_ContextTimeout(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.DeadlineExceeded})

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	if _, err := client.Quote(ctx, &gaian.QuoteRequest{QRString: "raw-qr", UserID: "usr_1"}); err == nil {
		t.Fatal("expected an error for a timed-out context, got nil")
	}
}

func TestQuote_InvalidJSON(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, "not json"), //nolint:bodyclose // response body closed by client
	}
	client := newMockClient(t, mock)

	if _, err := client.Quote(context.Background(), &gaian.QuoteRequest{QRString: "raw-qr", UserID: "usr_1"}); err == nil {
		t.Fatal("expected a decode error for a malformed response body, got nil")
	}
}
