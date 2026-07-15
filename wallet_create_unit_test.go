package gaian_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	gaian "github.com/gaiannetwork/gaian-go"
)

func TestCreateWallet_Success(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
			`{"data":{"address":"0xWalletAddress","chain":"base","isPrimary":true},"requestId":"req_1"}`),
	}
	client := newMockClient(t, mock)

	resp, err := client.CreateWallet(context.Background(), &gaian.CreateWalletRequest{
		UserID:  "usr_1",
		Address: "0xWalletAddress",
		Chain:   "base",
	})
	if err != nil {
		t.Fatalf("CreateWallet: %v", err)
	}
	if resp.Data.Address != "0xWalletAddress" {
		t.Errorf("Address = %q, want %q", resp.Data.Address, "0xWalletAddress")
	}
	if !resp.Data.IsPrimary {
		t.Error("expected IsPrimary=true for the first wallet")
	}
}

// TestCreateWalletRequest_WireFormat locks down the wire format: UserID
// goes in the path only, and an empty optional Chain must not leak into
// the JSON body.
func TestCreateWalletRequest_WireFormat(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
			`{"data":{"address":"0xWalletAddress","isPrimary":true},"requestId":"req_1"}`),
	}
	client := newMockClient(t, mock)

	if _, err := client.CreateWallet(context.Background(), &gaian.CreateWalletRequest{
		UserID:  "usr_1",
		Address: "0xWalletAddress",
	}); err != nil {
		t.Fatalf("CreateWallet: %v", err)
	}

	if !mock.calledPath("/api/v2/users/usr_1/wallets") {
		t.Fatalf("expected request to /api/v2/users/usr_1/wallets, got %q", mock.lastPath())
	}

	var body map[string]any
	if err := json.Unmarshal(mock.lastBody(), &body); err != nil {
		t.Fatalf("unmarshal request body: %v", err)
	}
	if _, ok := body["userId"]; ok {
		t.Error("userId leaked into the request body — it belongs in the URL path only")
	}
	if _, ok := body["chain"]; ok {
		t.Error("empty optional chain leaked into the request body")
	}
}

func TestCreateWallet_HTTPErrorStatuses(t *testing.T) {
	t.Parallel()

	statuses := []int{
		http.StatusBadRequest,
		http.StatusNotFound,
		http.StatusConflict, // wallet already linked
		http.StatusInternalServerError,
	}

	for _, status := range statuses {
		t.Run(http.StatusText(status), func(t *testing.T) {
			t.Parallel()

			mock := &mockHTTPClient{
				response: mockResponse(status, `{"message":"rejected"}`), //nolint:bodyclose // response body closed by client
			}
			client := newMockClient(t, mock)

			_, err := client.CreateWallet(context.Background(), &gaian.CreateWalletRequest{UserID: "usr_1", Address: "0xWalletAddress"})
			if !errors.Is(err, gaian.ErrUnexpectedStatus) {
				t.Errorf("CreateWallet() error = %v, want wrapping %v", err, gaian.ErrUnexpectedStatus)
			}
		})
	}
}

func TestCreateWallet_NetworkError(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: errConnectionRefused})

	_, err := client.CreateWallet(context.Background(), &gaian.CreateWalletRequest{UserID: "usr_1", Address: "0xWalletAddress"})
	if !errors.Is(err, gaian.ErrHTTPFailure) {
		t.Errorf("CreateWallet() error = %v, want wrapping %v", err, gaian.ErrHTTPFailure)
	}
}

func TestCreateWallet_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.Canceled})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := client.CreateWallet(ctx, &gaian.CreateWalletRequest{UserID: "usr_1", Address: "0xWalletAddress"}); err == nil {
		t.Fatal("expected an error for a cancelled context, got nil")
	}
}

func TestCreateWallet_ContextTimeout(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.DeadlineExceeded})

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	if _, err := client.CreateWallet(ctx, &gaian.CreateWalletRequest{UserID: "usr_1", Address: "0xWalletAddress"}); err == nil {
		t.Fatal("expected an error for a timed-out context, got nil")
	}
}

func TestCreateWallet_InvalidJSON(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, "not json"), //nolint:bodyclose // response body closed by client
	}
	client := newMockClient(t, mock)

	if _, err := client.CreateWallet(context.Background(), &gaian.CreateWalletRequest{UserID: "usr_1", Address: "0xWalletAddress"}); err == nil {
		t.Fatal("expected a decode error for a malformed response body, got nil")
	}
}
