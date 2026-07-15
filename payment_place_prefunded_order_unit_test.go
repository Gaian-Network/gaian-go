package gaian_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	gaian "github.com/gaiannetwork/gaian-go"
)

func TestPlacePrefundedOrder_Success(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
			`{"success":true,"requestId":"req_1","data":{"orderId":"ord_1","status":2,"statusLabel":"payment_received"}}`),
	}
	client := newMockClient(t, mock)

	resp, err := client.PlacePrefundedOrder(context.Background(), &gaian.PlacePrefundedOrderRequest{QuoteID: "quote_1"})
	if err != nil {
		t.Fatalf("PlacePrefundedOrder: %v", err)
	}
	if resp.Data.OrderID != "ord_1" {
		t.Errorf("OrderID = %q, want %q", resp.Data.OrderID, "ord_1")
	}
}

func TestPlacePrefundedOrder_HTTPErrorStatuses(t *testing.T) {
	t.Parallel()

	statuses := []int{http.StatusConflict, http.StatusInternalServerError}

	for _, status := range statuses {
		t.Run(http.StatusText(status), func(t *testing.T) {
			t.Parallel()

			mock := &mockHTTPClient{
				response: mockResponse(status, `{"message":"rejected"}`), //nolint:bodyclose // response body closed by client
			}
			client := newMockClient(t, mock)

			_, err := client.PlacePrefundedOrder(context.Background(), &gaian.PlacePrefundedOrderRequest{QuoteID: "quote_1"})
			if !errors.Is(err, gaian.ErrUnexpectedStatus) {
				t.Errorf("PlacePrefundedOrder() error = %v, want wrapping %v", err, gaian.ErrUnexpectedStatus)
			}
		})
	}
}

func TestPlacePrefundedOrder_NetworkError(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: errConnectionRefused})

	_, err := client.PlacePrefundedOrder(context.Background(), &gaian.PlacePrefundedOrderRequest{QuoteID: "quote_1"})
	if !errors.Is(err, gaian.ErrHTTPFailure) {
		t.Errorf("PlacePrefundedOrder() error = %v, want wrapping %v", err, gaian.ErrHTTPFailure)
	}
}

func TestPlacePrefundedOrder_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.Canceled})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := client.PlacePrefundedOrder(ctx, &gaian.PlacePrefundedOrderRequest{QuoteID: "quote_1"}); err == nil {
		t.Fatal("expected an error for a cancelled context, got nil")
	}
}

func TestPlacePrefundedOrder_ContextTimeout(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.DeadlineExceeded})

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	if _, err := client.PlacePrefundedOrder(ctx, &gaian.PlacePrefundedOrderRequest{QuoteID: "quote_1"}); err == nil {
		t.Fatal("expected an error for a timed-out context, got nil")
	}
}

func TestPlacePrefundedOrder_InvalidJSON(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, "not json"), //nolint:bodyclose // response body closed by client
	}
	client := newMockClient(t, mock)

	if _, err := client.PlacePrefundedOrder(context.Background(), &gaian.PlacePrefundedOrderRequest{QuoteID: "quote_1"}); err == nil {
		t.Fatal("expected a decode error for a malformed response body, got nil")
	}
}
