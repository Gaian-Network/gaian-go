package gaian_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	gaian "github.com/gaiannetwork/gaian-go"
)

func TestGetUserByWallet_Success(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
			`{"data":{"id":"usr_1","kycStatus":"VERIFIED"},"requestId":"req_1"}`),
	}
	client := newMockClient(t, mock)

	resp, err := client.GetUserByWallet(context.Background(), &gaian.GetUserByWalletRequest{
		WalletAddress: "0xWalletAddress",
	})
	if err != nil {
		t.Fatalf("GetUserByWallet: %v", err)
	}
	if resp.Data.ID != "usr_1" {
		t.Errorf("ID = %q, want %q", resp.Data.ID, "usr_1")
	}
	if !mock.calledPath("/api/v2/users/wallet/0xWalletAddress") {
		t.Fatalf("expected request to /api/v2/users/wallet/0xWalletAddress, got %q", mock.lastPath())
	}
}

func TestGetUserByWallet_HTTPErrorStatuses(t *testing.T) {
	t.Parallel()

	statuses := []int{http.StatusNotFound, http.StatusInternalServerError}

	for _, status := range statuses {
		t.Run(http.StatusText(status), func(t *testing.T) {
			t.Parallel()

			mock := &mockHTTPClient{
				response: mockResponse(status, `{"message":"user not found"}`), //nolint:bodyclose // response body closed by client
			}
			client := newMockClient(t, mock)

			_, err := client.GetUserByWallet(context.Background(), &gaian.GetUserByWalletRequest{WalletAddress: "0xUnknown"})
			if !errors.Is(err, gaian.ErrUnexpectedStatus) {
				t.Errorf("GetUserByWallet() error = %v, want wrapping %v", err, gaian.ErrUnexpectedStatus)
			}
		})
	}
}

func TestGetUserByWallet_NetworkError(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: errConnectionRefused})

	_, err := client.GetUserByWallet(context.Background(), &gaian.GetUserByWalletRequest{WalletAddress: "0xWalletAddress"})
	if !errors.Is(err, gaian.ErrHTTPFailure) {
		t.Errorf("GetUserByWallet() error = %v, want wrapping %v", err, gaian.ErrHTTPFailure)
	}
}

func TestGetUserByWallet_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.Canceled})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := client.GetUserByWallet(ctx, &gaian.GetUserByWalletRequest{WalletAddress: "0xWalletAddress"}); err == nil {
		t.Fatal("expected an error for a cancelled context, got nil")
	}
}

func TestGetUserByWallet_ContextTimeout(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.DeadlineExceeded})

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	if _, err := client.GetUserByWallet(ctx, &gaian.GetUserByWalletRequest{WalletAddress: "0xWalletAddress"}); err == nil {
		t.Fatal("expected an error for a timed-out context, got nil")
	}
}

func TestGetUserByWallet_InvalidJSON(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, "not json"), //nolint:bodyclose // response body closed by client
	}
	client := newMockClient(t, mock)

	if _, err := client.GetUserByWallet(context.Background(), &gaian.GetUserByWalletRequest{WalletAddress: "0xWalletAddress"}); err == nil {
		t.Fatal("expected a decode error for a malformed response body, got nil")
	}
}
