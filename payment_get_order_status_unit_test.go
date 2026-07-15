package gaian_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	gaian "github.com/gaiannetwork/gaian-go"
)

func TestGetOrderStatus_Success(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
			`{"success":true,"requestId":"req_1","data":{"orderId":"ord_1","status":10,"statusLabel":"completed",`+
				`"chainId":101,"fiatAmount":500000,"fiatCurrency":"VND","settlementAmount":"20.5",`+
				`"settlementCurrency":"USDC","exchangeRate":"24390","createdAt":"2030-01-01T00:00:00Z",`+
				`"updatedAt":"2030-01-01T00:05:00Z"}}`),
	}
	client := newMockClient(t, mock)

	resp, err := client.GetOrderStatus(context.Background(), &gaian.GetOrderStatusRequest{OrderID: "ord_1"})
	if err != nil {
		t.Fatalf("GetOrderStatus: %v", err)
	}
	if resp.Data.Status != gaian.OrderCompleted {
		t.Errorf("Status = %v, want %v", resp.Data.Status, gaian.OrderCompleted)
	}
	if !mock.calledPath("/api/v2/orders/ord_1") {
		t.Fatalf("expected request to /api/v2/orders/ord_1, got %q", mock.lastPath())
	}
}

// TestOrderStatus_IsTerminal locks down which statuses IsTerminal()
// reports as final — GetOrderStatus polling loops depend on this being
// correct for every enum value, not just the ones exercised elsewhere.
func TestOrderStatus_IsTerminal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status   gaian.OrderStatus
		terminal bool
	}{
		{gaian.OrderPending, false},
		{gaian.OrderAwaitingDeposit, false},
		{gaian.OrderPaymentReceived, false},
		{gaian.OrderQueued, false},
		{gaian.OrderProcessing, false},
		{gaian.OrderCompleted, true},
		{gaian.OrderFailed, true},
		{gaian.OrderCancelled, true},
		{gaian.OrderExpired, true},
		{gaian.OrderUnknown, false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("status_%d", tt.status), func(t *testing.T) {
			t.Parallel()

			if got := tt.status.IsTerminal(); got != tt.terminal {
				t.Errorf("OrderStatus(%d).IsTerminal() = %v, want %v", tt.status, got, tt.terminal)
			}
		})
	}
}

func TestGetOrderStatus_HTTPErrorStatuses(t *testing.T) {
	t.Parallel()

	statuses := []int{http.StatusNotFound, http.StatusInternalServerError}

	for _, status := range statuses {
		t.Run(http.StatusText(status), func(t *testing.T) {
			t.Parallel()

			mock := &mockHTTPClient{
				response: mockResponse(status, `{"message":"order not found"}`), //nolint:bodyclose // response body closed by client
			}
			client := newMockClient(t, mock)

			_, err := client.GetOrderStatus(context.Background(), &gaian.GetOrderStatusRequest{OrderID: "does-not-exist"})
			if !errors.Is(err, gaian.ErrUnexpectedStatus) {
				t.Errorf("GetOrderStatus() error = %v, want wrapping %v", err, gaian.ErrUnexpectedStatus)
			}
		})
	}
}

func TestGetOrderStatus_NetworkError(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: errConnectionRefused})

	_, err := client.GetOrderStatus(context.Background(), &gaian.GetOrderStatusRequest{OrderID: "ord_1"})
	if !errors.Is(err, gaian.ErrHTTPFailure) {
		t.Errorf("GetOrderStatus() error = %v, want wrapping %v", err, gaian.ErrHTTPFailure)
	}
}

func TestGetOrderStatus_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.Canceled})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := client.GetOrderStatus(ctx, &gaian.GetOrderStatusRequest{OrderID: "ord_1"}); err == nil {
		t.Fatal("expected an error for a cancelled context, got nil")
	}
}

func TestGetOrderStatus_ContextTimeout(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.DeadlineExceeded})

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	if _, err := client.GetOrderStatus(ctx, &gaian.GetOrderStatusRequest{OrderID: "ord_1"}); err == nil {
		t.Fatal("expected an error for a timed-out context, got nil")
	}
}

func TestGetOrderStatus_InvalidJSON(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, "not json"), //nolint:bodyclose // response body closed by client
	}
	client := newMockClient(t, mock)

	if _, err := client.GetOrderStatus(context.Background(), &gaian.GetOrderStatusRequest{OrderID: "ord_1"}); err == nil {
		t.Fatal("expected a decode error for a malformed response body, got nil")
	}
}
