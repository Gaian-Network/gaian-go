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

func TestUpdateKYC_Success(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
			`{"data":{"userId":"usr_1","kycStatus":"VERIFIED","email":"updated@example.com"},"requestId":"req_1"}`),
	}
	client := newMockClient(t, mock)

	email := "updated@example.com"
	resp, err := client.UpdateKYC(context.Background(), &gaian.UpdateKYCRequest{
		UserID: "usr_1",
		Email:  &email,
	})
	if err != nil {
		t.Fatalf("UpdateKYC: %v", err)
	}
	if resp.Data.Email != "updated@example.com" {
		t.Errorf("Email = %q, want %q", resp.Data.Email, "updated@example.com")
	}
	if resp.Data.KYCStatus != "VERIFIED" {
		t.Errorf("KYCStatus = %q, want %q", resp.Data.KYCStatus, "VERIFIED")
	}
}

// TestUpdateKYCRequest_WireFormat locks down UpdateKYCRequest's partial-
// update semantics: UserID stays in the path only, and every field left
// nil must be entirely absent from the JSON body — a present-but-empty
// field would tell the API to overwrite that field with "".
func TestUpdateKYCRequest_WireFormat(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
			`{"data":{"userId":"usr_1","kycStatus":"VERIFIED"},"requestId":"req_1"}`),
	}
	client := newMockClient(t, mock)

	email := "updated@example.com"
	if _, err := client.UpdateKYC(context.Background(), &gaian.UpdateKYCRequest{
		UserID: "usr_1",
		Email:  &email,
	}); err != nil {
		t.Fatalf("UpdateKYC: %v", err)
	}

	if !mock.calledPath("/api/v2/users/usr_1/kyc") {
		t.Fatalf("expected request to /api/v2/users/usr_1/kyc, got %q", mock.lastPath())
	}

	var body map[string]any
	if err := json.Unmarshal(mock.lastBody(), &body); err != nil {
		t.Fatalf("unmarshal request body: %v", err)
	}
	if _, ok := body["userId"]; ok {
		t.Error("userId leaked into the request body — it belongs in the URL path only")
	}
	if body["email"] != "updated@example.com" {
		t.Errorf(`body["email"] = %v, want "updated@example.com"`, body["email"])
	}
	for _, field := range []string{"firstName", "lastName", "nationalId", "dateOfBirth", "gender", "addressLine1"} {
		if _, ok := body[field]; ok {
			t.Errorf("unset field %q leaked into the request body: %v", field, body)
		}
	}
}

func TestUpdateKYC_HTTPErrorStatuses(t *testing.T) {
	t.Parallel()

	statuses := []int{
		http.StatusBadRequest,
		http.StatusNotFound,
		http.StatusConflict, // email/nationalId already used by another user
		http.StatusInternalServerError,
	}

	for _, status := range statuses {
		t.Run(http.StatusText(status), func(t *testing.T) {
			t.Parallel()

			mock := &mockHTTPClient{
				response: mockResponse(status, `{"message":"rejected"}`), //nolint:bodyclose // response body closed by client
			}
			client := newMockClient(t, mock)

			email := "updated@example.com"
			_, err := client.UpdateKYC(context.Background(), &gaian.UpdateKYCRequest{UserID: "usr_1", Email: &email})
			if !errors.Is(err, gaian.ErrUnexpectedStatus) {
				t.Errorf("UpdateKYC() error = %v, want wrapping %v", err, gaian.ErrUnexpectedStatus)
			}
		})
	}
}

func TestUpdateKYC_NetworkError(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: errConnectionRefused})

	email := "updated@example.com"
	_, err := client.UpdateKYC(context.Background(), &gaian.UpdateKYCRequest{UserID: "usr_1", Email: &email})
	if !errors.Is(err, gaian.ErrHTTPFailure) {
		t.Errorf("UpdateKYC() error = %v, want wrapping %v", err, gaian.ErrHTTPFailure)
	}
}

func TestUpdateKYC_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.Canceled})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	email := "updated@example.com"
	if _, err := client.UpdateKYC(ctx, &gaian.UpdateKYCRequest{UserID: "usr_1", Email: &email}); err == nil {
		t.Fatal("expected an error for a cancelled context, got nil")
	}
}

func TestUpdateKYC_ContextTimeout(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.DeadlineExceeded})

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	email := "updated@example.com"
	if _, err := client.UpdateKYC(ctx, &gaian.UpdateKYCRequest{UserID: "usr_1", Email: &email}); err == nil {
		t.Fatal("expected an error for a timed-out context, got nil")
	}
}

func TestUpdateKYC_InvalidJSON(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, "not json"), //nolint:bodyclose // response body closed by client
	}
	client := newMockClient(t, mock)

	email := "updated@example.com"
	if _, err := client.UpdateKYC(context.Background(), &gaian.UpdateKYCRequest{UserID: "usr_1", Email: &email}); err == nil {
		t.Fatal("expected a decode error for a malformed response body, got nil")
	}
}
