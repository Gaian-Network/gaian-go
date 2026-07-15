package gaian_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	gaian "github.com/gaiannetwork/gaian-go"
)

func TestGetUserByID_Success(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
			`{"data":{"id":"usr_1","firstName":"Nguyen","kycStatus":"VERIFIED"},"requestId":"req_1"}`),
	}
	client := newMockClient(t, mock)

	resp, err := client.GetUserByID(context.Background(), &gaian.GetUserByIDRequest{UserID: "usr_1"})
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if resp.Data.ID != "usr_1" {
		t.Errorf("ID = %q, want %q", resp.Data.ID, "usr_1")
	}
	if resp.Data.KYCStatus != "VERIFIED" {
		t.Errorf("KYCStatus = %q, want %q", resp.Data.KYCStatus, "VERIFIED")
	}
}

// TestGetUserByID_PathEscaping guards against a UserID containing
// characters that are meaningful in a URL path (e.g. "/") corrupting the
// request path.
func TestGetUserByID_PathEscaping(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, `{"data":{"id":"usr/1"},"requestId":"req_1"}`), //nolint:bodyclose // response body closed by client
	}
	client := newMockClient(t, mock)

	if _, err := client.GetUserByID(context.Background(), &gaian.GetUserByIDRequest{UserID: "usr/1"}); err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}

	// URL.Path holds the decoded path ("usr/1" round-trips back to a
	// literal slash), so the escaped form actually sent on the wire must
	// be read from EscapedPath() instead.
	if got := mock.lastReq.URL.EscapedPath(); got != "/api/v2/users/usr%2F1" {
		t.Fatalf("expected path-escaped UserID in request path, got %q", got)
	}
}

func TestGetUserByID_HTTPErrorStatuses(t *testing.T) {
	t.Parallel()

	statuses := []int{
		http.StatusNotFound,
		http.StatusUnauthorized,
		http.StatusInternalServerError,
	}

	for _, status := range statuses {
		t.Run(http.StatusText(status), func(t *testing.T) {
			t.Parallel()

			mock := &mockHTTPClient{
				response: mockResponse(status, `{"message":"user not found"}`), //nolint:bodyclose // response body closed by client
			}
			client := newMockClient(t, mock)

			_, err := client.GetUserByID(context.Background(), &gaian.GetUserByIDRequest{UserID: "does-not-exist"})
			if !errors.Is(err, gaian.ErrUnexpectedStatus) {
				t.Errorf("GetUserByID() error = %v, want wrapping %v", err, gaian.ErrUnexpectedStatus)
			}
		})
	}
}

func TestGetUserByID_NetworkError(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: errConnectionRefused})

	_, err := client.GetUserByID(context.Background(), &gaian.GetUserByIDRequest{UserID: "usr_1"})
	if !errors.Is(err, gaian.ErrHTTPFailure) {
		t.Errorf("GetUserByID() error = %v, want wrapping %v", err, gaian.ErrHTTPFailure)
	}
}

func TestGetUserByID_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.Canceled})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := client.GetUserByID(ctx, &gaian.GetUserByIDRequest{UserID: "usr_1"}); err == nil {
		t.Fatal("expected an error for a cancelled context, got nil")
	}
}

func TestGetUserByID_ContextTimeout(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.DeadlineExceeded})

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	if _, err := client.GetUserByID(ctx, &gaian.GetUserByIDRequest{UserID: "usr_1"}); err == nil {
		t.Fatal("expected an error for a timed-out context, got nil")
	}
}

func TestGetUserByID_InvalidJSON(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, "not json"), //nolint:bodyclose // response body closed by client
	}
	client := newMockClient(t, mock)

	if _, err := client.GetUserByID(context.Background(), &gaian.GetUserByIDRequest{UserID: "usr_1"}); err == nil {
		t.Fatal("expected a decode error for a malformed response body, got nil")
	}
}
