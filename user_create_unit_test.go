package gaian_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	gaian "github.com/gaiannetwork/gaian-go"
)

func TestCreateUser_Success(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, `{"data":{"userId":"usr_abc123"},"requestId":"req_1"}`), //nolint:bodyclose // response body closed by client
	}
	client := newMockClient(t, mock)

	resp, err := client.CreateUser(context.Background(), &gaian.CreateUserRequest{})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if resp.Data.UserID != "usr_abc123" {
		t.Errorf("UserID = %q, want %q", resp.Data.UserID, "usr_abc123")
	}
	if !mock.calledPath("/api/v2/users") {
		t.Fatalf("expected request to /api/v2/users, got %q", mock.lastPath())
	}
	if len(mock.lastBody()) != 0 {
		t.Errorf("request body = %q, want empty (CreateUserRequest has no fields)", mock.lastBody())
	}
}

func TestCreateUser_HTTPErrorStatuses(t *testing.T) {
	t.Parallel()

	statuses := []int{
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusInternalServerError,
		http.StatusServiceUnavailable,
	}

	for _, status := range statuses {
		t.Run(http.StatusText(status), func(t *testing.T) {
			t.Parallel()

			mock := &mockHTTPClient{
				response: mockResponse(status, `{"message":"rejected"}`), //nolint:bodyclose // response body closed by client
			}
			client := newMockClient(t, mock)

			_, err := client.CreateUser(context.Background(), &gaian.CreateUserRequest{})
			if !errors.Is(err, gaian.ErrUnexpectedStatus) {
				t.Errorf("CreateUser() error = %v, want wrapping %v", err, gaian.ErrUnexpectedStatus)
			}
		})
	}
}

func TestCreateUser_NetworkError(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: errConnectionRefused})

	_, err := client.CreateUser(context.Background(), &gaian.CreateUserRequest{})
	if !errors.Is(err, gaian.ErrHTTPFailure) {
		t.Errorf("CreateUser() error = %v, want wrapping %v", err, gaian.ErrHTTPFailure)
	}
}

func TestCreateUser_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.Canceled})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := client.CreateUser(ctx, &gaian.CreateUserRequest{}); err == nil {
		t.Fatal("expected an error for a cancelled context, got nil")
	}
}

func TestCreateUser_ContextTimeout(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.DeadlineExceeded})

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	if _, err := client.CreateUser(ctx, &gaian.CreateUserRequest{}); err == nil {
		t.Fatal("expected an error for a timed-out context, got nil")
	}
}

func TestCreateUser_InvalidJSON(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, "not json"), //nolint:bodyclose // response body closed by client
	}
	client := newMockClient(t, mock)

	if _, err := client.CreateUser(context.Background(), &gaian.CreateUserRequest{}); err == nil {
		t.Fatal("expected a decode error for a malformed response body, got nil")
	}
}
