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

func TestPlaceOrder_Success(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
			`{"success":true,"requestId":"req_1","data":{"orderId":"ord_1","status":1,"statusLabel":"awaiting_deposit",`+
				`"chainId":101,"depositAddress":"So1anaAddr","encodedTransaction":"base64tx","expiresAt":"2030-01-01T00:00:00Z"}}`),
	}
	client := newMockClient(t, mock)

	resp, err := client.PlaceOrder(context.Background(), &gaian.PlaceOrderRequest{QuoteID: "quote_1"})
	if err != nil {
		t.Fatalf("PlaceOrder: %v", err)
	}
	if resp.Data.OrderID != "ord_1" {
		t.Errorf("OrderID = %q, want %q", resp.Data.OrderID, "ord_1")
	}
	if resp.Data.EncodedTransaction != "base64tx" {
		t.Errorf("EncodedTransaction = %q, want %q", resp.Data.EncodedTransaction, "base64tx")
	}

	var body map[string]any
	if err := json.Unmarshal(mock.lastBody(), &body); err != nil {
		t.Fatalf("unmarshal request body: %v", err)
	}
	if body["quoteId"] != "quote_1" {
		t.Errorf(`body["quoteId"] = %v, want "quote_1"`, body["quoteId"])
	}
}

func TestPlaceOrder_HTTPErrorStatuses(t *testing.T) {
	t.Parallel()

	statuses := []int{
		http.StatusNotFound,
		http.StatusConflict, // quote already consumed, expired, or wrong flow
		http.StatusInternalServerError,
	}

	for _, status := range statuses {
		t.Run(http.StatusText(status), func(t *testing.T) {
			t.Parallel()

			mock := &mockHTTPClient{
				response: mockResponse(status, `{"message":"rejected"}`), //nolint:bodyclose // response body closed by client
			}
			client := newMockClient(t, mock)

			_, err := client.PlaceOrder(context.Background(), &gaian.PlaceOrderRequest{QuoteID: "quote_1"})
			if !errors.Is(err, gaian.ErrUnexpectedStatus) {
				t.Errorf("PlaceOrder() error = %v, want wrapping %v", err, gaian.ErrUnexpectedStatus)
			}
		})
	}
}

func TestPlaceOrder_NetworkError(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: errConnectionRefused})

	_, err := client.PlaceOrder(context.Background(), &gaian.PlaceOrderRequest{QuoteID: "quote_1"})
	if !errors.Is(err, gaian.ErrHTTPFailure) {
		t.Errorf("PlaceOrder() error = %v, want wrapping %v", err, gaian.ErrHTTPFailure)
	}
}

func TestPlaceOrder_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.Canceled})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := client.PlaceOrder(ctx, &gaian.PlaceOrderRequest{QuoteID: "quote_1"}); err == nil {
		t.Fatal("expected an error for a cancelled context, got nil")
	}
}

func TestPlaceOrder_ContextTimeout(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.DeadlineExceeded})

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	if _, err := client.PlaceOrder(ctx, &gaian.PlaceOrderRequest{QuoteID: "quote_1"}); err == nil {
		t.Fatal("expected an error for a timed-out context, got nil")
	}
}

func TestPlaceOrder_InvalidJSON(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, "not json"), //nolint:bodyclose // response body closed by client
	}
	client := newMockClient(t, mock)

	if _, err := client.PlaceOrder(context.Background(), &gaian.PlaceOrderRequest{QuoteID: "quote_1"}); err == nil {
		t.Fatal("expected a decode error for a malformed response body, got nil")
	}
}
