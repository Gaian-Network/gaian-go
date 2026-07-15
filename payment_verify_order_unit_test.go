package gaian_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	gaian "github.com/gaiannetwork/gaian-go"
)

func TestVerifyOrder_Success(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
			`{"success":true,"requestId":"req_1","data":{"orderId":"ord_1","status":1,"statusLabel":"verified",`+
				`"transactionHash":"0xabc","message":""}}`),
	}
	client := newMockClient(t, mock)

	resp, err := client.VerifyOrder(context.Background(), &gaian.VerifyOrderRequest{
		OrderID:         "ord_1",
		TransactionHash: "0xabc",
	})
	if err != nil {
		t.Fatalf("VerifyOrder: %v", err)
	}
	if resp.Data.Status != gaian.VerifyVerified {
		t.Errorf("Status = %v, want %v", resp.Data.Status, gaian.VerifyVerified)
	}
}

// TestVerifyOrder_StatusValues covers the "HTTP 200 even on failure"
// gotcha documented on VerifyOrderResponse — a nil error must not be
// mistaken for a confirmed payment; callers have to branch on Status.
func TestVerifyOrder_StatusValues(t *testing.T) {
	t.Parallel()

	statuses := []gaian.VerifyStatus{gaian.VerifyPending, gaian.VerifyVerified, gaian.VerifyFailed}

	for _, status := range statuses {
		t.Run(fmt.Sprintf("status_%d", status), func(t *testing.T) {
			t.Parallel()

			body := fmt.Sprintf(`{"success":true,"requestId":"req_1","data":{"orderId":"ord_1","status":%d,"message":"x"}}`, status)
			mock := &mockHTTPClient{response: mockResponse(http.StatusOK, body)} //nolint:bodyclose // response body closed by client
			client := newMockClient(t, mock)

			resp, err := client.VerifyOrder(context.Background(), &gaian.VerifyOrderRequest{OrderID: "ord_1", TransactionHash: "0xabc"})
			if err != nil {
				t.Fatalf("VerifyOrder: %v (a %v result must still be a nil-error 200 response)", err, status)
			}
			if resp.Data.Status != status {
				t.Errorf("Status = %v, want %v", resp.Data.Status, status)
			}
		})
	}
}

// TestVerifyOrderRequest_WireFormat locks down the wire format: OrderID
// goes in the path only (json:"-"), never in the body.
func TestVerifyOrderRequest_WireFormat(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, `{"success":true,"requestId":"req_1","data":{"orderId":"ord_1"}}`), //nolint:bodyclose // response body closed by client
	}
	client := newMockClient(t, mock)

	if _, err := client.VerifyOrder(context.Background(), &gaian.VerifyOrderRequest{
		OrderID:         "ord_1",
		TransactionHash: "0xabc",
	}); err != nil {
		t.Fatalf("VerifyOrder: %v", err)
	}

	if !mock.calledPath("/api/v2/orders/ord_1/verify") {
		t.Fatalf("expected request to /api/v2/orders/ord_1/verify, got %q", mock.lastPath())
	}

	var body map[string]any
	if err := json.Unmarshal(mock.lastBody(), &body); err != nil {
		t.Fatalf("unmarshal request body: %v", err)
	}
	if _, ok := body["orderId"]; ok {
		t.Error("orderId leaked into the request body — it belongs in the URL path only")
	}
	if body["transactionHash"] != "0xabc" {
		t.Errorf(`body["transactionHash"] = %v, want "0xabc"`, body["transactionHash"])
	}
}

func TestVerifyOrder_HTTPErrorStatuses(t *testing.T) {
	t.Parallel()

	statuses := []int{http.StatusNotFound, http.StatusInternalServerError}

	for _, status := range statuses {
		t.Run(http.StatusText(status), func(t *testing.T) {
			t.Parallel()

			mock := &mockHTTPClient{
				response: mockResponse(status, `{"message":"order not found"}`), //nolint:bodyclose // response body closed by client
			}
			client := newMockClient(t, mock)

			_, err := client.VerifyOrder(context.Background(), &gaian.VerifyOrderRequest{OrderID: "ord_1", TransactionHash: "0xabc"})
			if !errors.Is(err, gaian.ErrUnexpectedStatus) {
				t.Errorf("VerifyOrder() error = %v, want wrapping %v", err, gaian.ErrUnexpectedStatus)
			}
		})
	}
}

func TestVerifyOrder_NetworkError(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: errConnectionRefused})

	_, err := client.VerifyOrder(context.Background(), &gaian.VerifyOrderRequest{OrderID: "ord_1", TransactionHash: "0xabc"})
	if !errors.Is(err, gaian.ErrHTTPFailure) {
		t.Errorf("VerifyOrder() error = %v, want wrapping %v", err, gaian.ErrHTTPFailure)
	}
}

func TestVerifyOrder_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.Canceled})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req := &gaian.VerifyOrderRequest{OrderID: "ord_1", TransactionHash: "0xabc"}
	if _, err := client.VerifyOrder(ctx, req); err == nil {
		t.Fatal("expected an error for a cancelled context, got nil")
	}
}

func TestVerifyOrder_ContextTimeout(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.DeadlineExceeded})

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	req := &gaian.VerifyOrderRequest{OrderID: "ord_1", TransactionHash: "0xabc"}
	if _, err := client.VerifyOrder(ctx, req); err == nil {
		t.Fatal("expected an error for a timed-out context, got nil")
	}
}

func TestVerifyOrder_InvalidJSON(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, "not json"), //nolint:bodyclose // response body closed by client
	}
	client := newMockClient(t, mock)

	req := &gaian.VerifyOrderRequest{OrderID: "ord_1", TransactionHash: "0xabc"}
	if _, err := client.VerifyOrder(context.Background(), req); err == nil {
		t.Fatal("expected a decode error for a malformed response body, got nil")
	}
}
