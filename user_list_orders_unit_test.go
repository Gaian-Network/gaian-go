package gaian_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	gaian "github.com/gaiannetwork/gaian-go"
)

func TestListUserOrders_Success(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
			`{"requestId":"req_1","data":[{"orderId":"ord_1","status":"completed"}],"pagination":{"page":1,"pageSize":20,"totalCount":1,"totalPages":1}}`),
	}
	client := newMockClient(t, mock)

	resp, err := client.ListUserOrders(context.Background(), &gaian.ListUserOrdersRequest{UserID: "usr_1", Page: 1})
	if err != nil {
		t.Fatalf("ListUserOrders: %v", err)
	}
	if len(resp.Items) != 1 || resp.Items[0].OrderID != "ord_1" {
		t.Errorf("Items = %+v, want a single ord_1 item", resp.Items)
	}
	if resp.Pagination.TotalCount != 1 {
		t.Errorf("Pagination.TotalCount = %d, want 1", resp.Pagination.TotalCount)
	}
}

// TestListUserOrdersRequest_WireFormat guards a real gotcha: UserID is
// tagged param:"userId" json:"-" — the "param" tag has no effect on its
// own (this SDK only recognizes json tags for the GET query encoding),
// so it's json:"-" alone keeping UserID out of the query string. If that
// tag were ever dropped, UserID would silently leak into the query
// alongside status/page/page_size.
func TestListUserOrdersRequest_WireFormat(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
			`{"requestId":"req_1","data":[],"pagination":{"page":1,"pageSize":20,"totalCount":0,"totalPages":0}}`),
	}
	client := newMockClient(t, mock)

	if _, err := client.ListUserOrders(context.Background(), &gaian.ListUserOrdersRequest{
		UserID: "usr_1",
		Status: "completed",
		Page:   2,
	}); err != nil {
		t.Fatalf("ListUserOrders: %v", err)
	}

	if !mock.calledPath("/api/v2/users/usr_1/orders") {
		t.Fatalf("expected request to /api/v2/users/usr_1/orders, got %q", mock.lastPath())
	}

	q := mock.lastReq.URL.Query()
	if q.Has("userId") || q.Has("UserID") {
		t.Errorf("UserID leaked into the query string: %s", mock.lastReq.URL.RawQuery)
	}
	if q.Get("status") != "completed" {
		t.Errorf(`query "status" = %q, want "completed"`, q.Get("status"))
	}
	if q.Get("page") != "2" {
		t.Errorf(`query "page" = %q, want "2"`, q.Get("page"))
	}
}

func TestListUserOrders_HTTPErrorStatuses(t *testing.T) {
	t.Parallel()

	statuses := []int{http.StatusNotFound, http.StatusInternalServerError}

	for _, status := range statuses {
		t.Run(http.StatusText(status), func(t *testing.T) {
			t.Parallel()

			mock := &mockHTTPClient{
				response: mockResponse(status, `{"message":"user not found"}`), //nolint:bodyclose // response body closed by client
			}
			client := newMockClient(t, mock)

			_, err := client.ListUserOrders(context.Background(), &gaian.ListUserOrdersRequest{UserID: "usr_1"})
			if !errors.Is(err, gaian.ErrUnexpectedStatus) {
				t.Errorf("ListUserOrders() error = %v, want wrapping %v", err, gaian.ErrUnexpectedStatus)
			}
		})
	}
}

func TestListUserOrders_NetworkError(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: errConnectionRefused})

	_, err := client.ListUserOrders(context.Background(), &gaian.ListUserOrdersRequest{UserID: "usr_1"})
	if !errors.Is(err, gaian.ErrHTTPFailure) {
		t.Errorf("ListUserOrders() error = %v, want wrapping %v", err, gaian.ErrHTTPFailure)
	}
}

func TestListUserOrders_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.Canceled})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := client.ListUserOrders(ctx, &gaian.ListUserOrdersRequest{UserID: "usr_1"}); err == nil {
		t.Fatal("expected an error for a cancelled context, got nil")
	}
}

func TestListUserOrders_ContextTimeout(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.DeadlineExceeded})

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	if _, err := client.ListUserOrders(ctx, &gaian.ListUserOrdersRequest{UserID: "usr_1"}); err == nil {
		t.Fatal("expected an error for a timed-out context, got nil")
	}
}

func TestListUserOrders_InvalidJSON(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, "not json"), //nolint:bodyclose // response body closed by client
	}
	client := newMockClient(t, mock)

	if _, err := client.ListUserOrders(context.Background(), &gaian.ListUserOrdersRequest{UserID: "usr_1"}); err == nil {
		t.Fatal("expected a decode error for a malformed response body, got nil")
	}
}
