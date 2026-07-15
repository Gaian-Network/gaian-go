package gaian_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	gaian "github.com/gaiannetwork/gaian-go"
)

func TestVerifyAccount_Success(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
			`{"success":true,"requestId":"req_1","data":{"country":"VN","accountNumber":"0123456789",`+
				`"code":"970436","valid":true,"accountName":"NGUYEN VAN A"}}`),
	}
	client := newMockClient(t, mock)

	resp, err := client.VerifyAccount(context.Background(), &gaian.VerifyAccountRequest{
		AccountNumber: "0123456789",
		Code:          "970436",
		Country:       "VN",
	})
	if err != nil {
		t.Fatalf("VerifyAccount: %v", err)
	}
	if !resp.Data.Valid {
		t.Error("expected Valid=true")
	}
	if resp.Data.AccountName == nil || *resp.Data.AccountName != "NGUYEN VAN A" {
		t.Errorf("AccountName = %v, want \"NGUYEN VAN A\"", resp.Data.AccountName)
	}
}

// TestVerifyAccount_InvalidAccount_NotAnError guards the documented
// behavior: a bad account number is not a client error, it's a normal
// 200 response with Valid=false and a nil AccountName.
func TestVerifyAccount_InvalidAccount_NotAnError(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
			`{"success":true,"requestId":"req_1","data":{"country":"VN","accountNumber":"0000000000",`+
				`"code":"970436","valid":false,"accountName":null}}`),
	}
	client := newMockClient(t, mock)

	resp, err := client.VerifyAccount(context.Background(), &gaian.VerifyAccountRequest{
		AccountNumber: "0000000000",
		Code:          "970436",
		Country:       "VN",
	})
	if err != nil {
		t.Fatalf("VerifyAccount: %v (an invalid account must not surface as an error)", err)
	}
	if resp.Data.Valid {
		t.Error("expected Valid=false for a bad account number")
	}
	if resp.Data.AccountName != nil {
		t.Errorf("AccountName = %v, want nil", *resp.Data.AccountName)
	}
}

func TestVerifyAccount_HTTPErrorStatuses(t *testing.T) {
	t.Parallel()

	statuses := []int{http.StatusBadRequest, http.StatusForbidden, http.StatusInternalServerError}

	for _, status := range statuses {
		t.Run(http.StatusText(status), func(t *testing.T) {
			t.Parallel()

			mock := &mockHTTPClient{
				response: mockResponse(status, `{"message":"rejected"}`), //nolint:bodyclose // response body closed by client
			}
			client := newMockClient(t, mock)

			req := &gaian.VerifyAccountRequest{AccountNumber: "0123456789", Code: "970436", Country: "VN"}
			_, err := client.VerifyAccount(context.Background(), req)
			if !errors.Is(err, gaian.ErrUnexpectedStatus) {
				t.Errorf("VerifyAccount() error = %v, want wrapping %v", err, gaian.ErrUnexpectedStatus)
			}
		})
	}
}

func TestVerifyAccount_NetworkError(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: errConnectionRefused})

	req := &gaian.VerifyAccountRequest{AccountNumber: "0123456789", Code: "970436", Country: "VN"}
	_, err := client.VerifyAccount(context.Background(), req)
	if !errors.Is(err, gaian.ErrHTTPFailure) {
		t.Errorf("VerifyAccount() error = %v, want wrapping %v", err, gaian.ErrHTTPFailure)
	}
}

func TestVerifyAccount_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.Canceled})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req := &gaian.VerifyAccountRequest{AccountNumber: "0123456789", Code: "970436", Country: "VN"}
	if _, err := client.VerifyAccount(ctx, req); err == nil {
		t.Fatal("expected an error for a cancelled context, got nil")
	}
}

func TestVerifyAccount_ContextTimeout(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.DeadlineExceeded})

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	req := &gaian.VerifyAccountRequest{AccountNumber: "0123456789", Code: "970436", Country: "VN"}
	if _, err := client.VerifyAccount(ctx, req); err == nil {
		t.Fatal("expected an error for a timed-out context, got nil")
	}
}

func TestVerifyAccount_InvalidJSON(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, "not json"), //nolint:bodyclose // response body closed by client
	}
	client := newMockClient(t, mock)

	req := &gaian.VerifyAccountRequest{AccountNumber: "0123456789", Code: "970436", Country: "VN"}
	if _, err := client.VerifyAccount(context.Background(), req); err == nil {
		t.Fatal("expected a decode error for a malformed response body, got nil")
	}
}
