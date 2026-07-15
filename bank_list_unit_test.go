package gaian_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	gaian "github.com/gaiannetwork/gaian-go"
)

func TestListBanks_Success(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
			`{"success":true,"requestId":"req_1","data":{"country":"VN","banks":[{"id":1,"code":"970436","name":"Vietcombank"}]}}`),
	}
	client := newMockClient(t, mock)

	resp, err := client.ListBanks(context.Background(), &gaian.ListBanksRequest{Country: "VN"})
	if err != nil {
		t.Fatalf("ListBanks: %v", err)
	}
	if len(resp.Data.Banks) != 1 || resp.Data.Banks[0].Code != "970436" {
		t.Errorf("Banks = %+v, want a single bank with code 970436", resp.Data.Banks)
	}

	q := mock.lastReq.URL.Query()
	if q.Get("country") != "VN" {
		t.Errorf(`query "country" = %q, want "VN"`, q.Get("country"))
	}
}

func TestListBanks_HTTPErrorStatuses(t *testing.T) {
	t.Parallel()

	// ListBanks is production-only — sandbox rejects it, and this covers
	// the same error-wrapping contract regardless of environment.
	statuses := []int{http.StatusBadRequest, http.StatusForbidden, http.StatusInternalServerError}

	for _, status := range statuses {
		t.Run(http.StatusText(status), func(t *testing.T) {
			t.Parallel()

			mock := &mockHTTPClient{
				response: mockResponse(status, `{"message":"rejected"}`), //nolint:bodyclose // response body closed by client
			}
			client := newMockClient(t, mock)

			_, err := client.ListBanks(context.Background(), &gaian.ListBanksRequest{Country: "VN"})
			if !errors.Is(err, gaian.ErrUnexpectedStatus) {
				t.Errorf("ListBanks() error = %v, want wrapping %v", err, gaian.ErrUnexpectedStatus)
			}
		})
	}
}

func TestListBanks_NetworkError(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: errConnectionRefused})

	_, err := client.ListBanks(context.Background(), &gaian.ListBanksRequest{Country: "VN"})
	if !errors.Is(err, gaian.ErrHTTPFailure) {
		t.Errorf("ListBanks() error = %v, want wrapping %v", err, gaian.ErrHTTPFailure)
	}
}

func TestListBanks_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.Canceled})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := client.ListBanks(ctx, &gaian.ListBanksRequest{Country: "VN"}); err == nil {
		t.Fatal("expected an error for a cancelled context, got nil")
	}
}

func TestListBanks_ContextTimeout(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.DeadlineExceeded})

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	if _, err := client.ListBanks(ctx, &gaian.ListBanksRequest{Country: "VN"}); err == nil {
		t.Fatal("expected an error for a timed-out context, got nil")
	}
}

func TestListBanks_InvalidJSON(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, "not json"), //nolint:bodyclose // response body closed by client
	}
	client := newMockClient(t, mock)

	if _, err := client.ListBanks(context.Background(), &gaian.ListBanksRequest{Country: "VN"}); err == nil {
		t.Fatal("expected a decode error for a malformed response body, got nil")
	}
}
