package gaian_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	gaian "github.com/gaiannetwork/gaian-go"
)

func TestGetUserOrder_Success(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
			`{"data":{"orderId":"ord_1","status":"completed"},"requestId":"req_1"}`),
	}
	client := newMockClient(t, mock)

	resp, err := client.GetUserOrder(context.Background(), &gaian.GetUserOrderRequest{UserID: "usr_1", OrderID: "ord_1"})
	if err != nil {
		t.Fatalf("GetUserOrder: %v", err)
	}
	if resp.Data.OrderID != "ord_1" {
		t.Errorf("OrderID = %q, want %q", resp.Data.OrderID, "ord_1")
	}
	if !mock.calledPath("/api/v2/users/usr_1/orders/ord_1") {
		t.Fatalf("expected request to /api/v2/users/usr_1/orders/ord_1, got %q", mock.lastPath())
	}
}

func TestGetUserOrder_HTTPErrorStatuses(t *testing.T) {
	t.Parallel()

	statuses := []int{http.StatusNotFound, http.StatusInternalServerError}

	for _, status := range statuses {
		t.Run(http.StatusText(status), func(t *testing.T) {
			t.Parallel()

			mock := &mockHTTPClient{
				response: mockResponse(status, `{"message":"order not found"}`), //nolint:bodyclose // response body closed by client
			}
			client := newMockClient(t, mock)

			_, err := client.GetUserOrder(context.Background(), &gaian.GetUserOrderRequest{UserID: "usr_1", OrderID: "does-not-exist"})
			if !errors.Is(err, gaian.ErrUnexpectedStatus) {
				t.Errorf("GetUserOrder() error = %v, want wrapping %v", err, gaian.ErrUnexpectedStatus)
			}
		})
	}
}

func TestGetUserOrder_NetworkError(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: errConnectionRefused})

	_, err := client.GetUserOrder(context.Background(), &gaian.GetUserOrderRequest{UserID: "usr_1", OrderID: "ord_1"})
	if !errors.Is(err, gaian.ErrHTTPFailure) {
		t.Errorf("GetUserOrder() error = %v, want wrapping %v", err, gaian.ErrHTTPFailure)
	}
}

func TestGetUserOrder_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.Canceled})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := client.GetUserOrder(ctx, &gaian.GetUserOrderRequest{UserID: "usr_1", OrderID: "ord_1"}); err == nil {
		t.Fatal("expected an error for a cancelled context, got nil")
	}
}

func TestGetUserOrder_ContextTimeout(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.DeadlineExceeded})

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	if _, err := client.GetUserOrder(ctx, &gaian.GetUserOrderRequest{UserID: "usr_1", OrderID: "ord_1"}); err == nil {
		t.Fatal("expected an error for a timed-out context, got nil")
	}
}

func TestGetUserOrder_InvalidJSON(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, "not json"), //nolint:bodyclose // response body closed by client
	}
	client := newMockClient(t, mock)

	if _, err := client.GetUserOrder(context.Background(), &gaian.GetUserOrderRequest{UserID: "usr_1", OrderID: "ord_1"}); err == nil {
		t.Fatal("expected a decode error for a malformed response body, got nil")
	}
}
