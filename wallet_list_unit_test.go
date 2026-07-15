package gaian_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	gaian "github.com/gaiannetwork/gaian-go"
)

func TestListWallets_Success(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
			`{"data":[{"address":"0xA","chain":"base","isPrimary":true},{"address":"0xB","chain":"solana","isPrimary":false}],"requestId":"req_1"}`),
	}
	client := newMockClient(t, mock)

	resp, err := client.ListWallets(context.Background(), &gaian.ListWalletsRequest{UserID: "usr_1"})
	if err != nil {
		t.Fatalf("ListWallets: %v", err)
	}
	if resp.Data == nil || len(*resp.Data) != 2 {
		t.Fatalf("Data = %+v, want 2 wallets", resp.Data)
	}
	if (*resp.Data)[0].Address != "0xA" || !(*resp.Data)[0].IsPrimary {
		t.Errorf("first wallet = %+v, want primary 0xA", (*resp.Data)[0])
	}
	if !mock.calledPath("/api/v2/users/usr_1/wallets") {
		t.Fatalf("expected request to /api/v2/users/usr_1/wallets, got %q", mock.lastPath())
	}
}

func TestListWallets_HTTPErrorStatuses(t *testing.T) {
	t.Parallel()

	statuses := []int{http.StatusNotFound, http.StatusInternalServerError}

	for _, status := range statuses {
		t.Run(http.StatusText(status), func(t *testing.T) {
			t.Parallel()

			mock := &mockHTTPClient{
				response: mockResponse(status, `{"message":"user not found"}`), //nolint:bodyclose // response body closed by client
			}
			client := newMockClient(t, mock)

			_, err := client.ListWallets(context.Background(), &gaian.ListWalletsRequest{UserID: "usr_1"})
			if !errors.Is(err, gaian.ErrUnexpectedStatus) {
				t.Errorf("ListWallets() error = %v, want wrapping %v", err, gaian.ErrUnexpectedStatus)
			}
		})
	}
}

func TestListWallets_NetworkError(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: errConnectionRefused})

	_, err := client.ListWallets(context.Background(), &gaian.ListWalletsRequest{UserID: "usr_1"})
	if !errors.Is(err, gaian.ErrHTTPFailure) {
		t.Errorf("ListWallets() error = %v, want wrapping %v", err, gaian.ErrHTTPFailure)
	}
}

func TestListWallets_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.Canceled})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := client.ListWallets(ctx, &gaian.ListWalletsRequest{UserID: "usr_1"}); err == nil {
		t.Fatal("expected an error for a cancelled context, got nil")
	}
}

func TestListWallets_ContextTimeout(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.DeadlineExceeded})

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	if _, err := client.ListWallets(ctx, &gaian.ListWalletsRequest{UserID: "usr_1"}); err == nil {
		t.Fatal("expected an error for a timed-out context, got nil")
	}
}

func TestListWallets_InvalidJSON(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, "not json"), //nolint:bodyclose // response body closed by client
	}
	client := newMockClient(t, mock)

	if _, err := client.ListWallets(context.Background(), &gaian.ListWalletsRequest{UserID: "usr_1"}); err == nil {
		t.Fatal("expected a decode error for a malformed response body, got nil")
	}
}
