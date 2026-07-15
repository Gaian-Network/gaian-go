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

// GenerateKYCLink is intentionally never called against the real
// sandbox: each call provisions a real hosted KYC session with Didit,
// which costs money per link generated. This file is its only test
// coverage, so it carries the full suite (wire format, HTTP errors,
// network/context failures, invalid JSON) that other endpoints split
// across a *_unit_test.go and a *_integration_test.go.

func TestGenerateKYCLink_Success(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
			`{"data":{"url":"https://verify.didit.me/session/abc123"},"requestId":"req_1"}`),
	}
	client := newMockClient(t, mock)

	resp, err := client.GenerateKYCLink(context.Background(), &gaian.GenerateKYCLinkRequest{
		UserID: "usr_1",
	})
	if err != nil {
		t.Fatalf("GenerateKYCLink: %v", err)
	}
	if resp.Data.URL != "https://verify.didit.me/session/abc123" {
		t.Errorf("URL = %q, want %q", resp.Data.URL, "https://verify.didit.me/session/abc123")
	}

	if !mock.calledPath("/api/v2/users/usr_1/kyc-url") {
		t.Fatalf("expected request to /api/v2/users/usr_1/kyc-url, got %q", mock.lastPath())
	}

	var body map[string]any
	if err := json.Unmarshal(mock.lastBody(), &body); err != nil {
		t.Fatalf("unmarshal request body: %v", err)
	}
	if _, ok := body["userId"]; ok {
		t.Error("userId leaked into the request body — it belongs in the URL path only")
	}
	if _, ok := body["callbackUrl"]; ok {
		t.Error("empty optional callbackUrl leaked into the request body")
	}
}

// TestGenerateKYCLink_CallbackURLIncludedWhenSet guards the omitempty
// behavior the other way: when CallbackURL is set, it must be sent.
func TestGenerateKYCLink_CallbackURLIncludedWhenSet(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
			`{"data":{"url":"https://verify.didit.me/session/abc123"},"requestId":"req_1"}`),
	}
	client := newMockClient(t, mock)

	_, err := client.GenerateKYCLink(context.Background(), &gaian.GenerateKYCLinkRequest{
		UserID:      "usr_1",
		CallbackURL: "https://myapp.com/kyc/callback",
	})
	if err != nil {
		t.Fatalf("GenerateKYCLink: %v", err)
	}

	var body map[string]any
	if err := json.Unmarshal(mock.lastBody(), &body); err != nil {
		t.Fatalf("unmarshal request body: %v", err)
	}
	if body["callbackUrl"] != "https://myapp.com/kyc/callback" {
		t.Errorf(`body["callbackUrl"] = %v, want "https://myapp.com/kyc/callback"`, body["callbackUrl"])
	}
}

func TestGenerateKYCLink_HTTPErrorStatuses(t *testing.T) {
	t.Parallel()

	statuses := []int{
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusInternalServerError,
	}

	for _, status := range statuses {
		t.Run(http.StatusText(status), func(t *testing.T) {
			t.Parallel()

			mock := &mockHTTPClient{
				response: mockResponse(status, `{"message":"rejected"}`), //nolint:bodyclose // response body closed by client
			}
			client := newMockClient(t, mock)

			_, err := client.GenerateKYCLink(context.Background(), &gaian.GenerateKYCLinkRequest{UserID: "usr_1"})
			if !errors.Is(err, gaian.ErrUnexpectedStatus) {
				t.Errorf("GenerateKYCLink() error = %v, want wrapping %v", err, gaian.ErrUnexpectedStatus)
			}
		})
	}
}

func TestGenerateKYCLink_NetworkError(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: errConnectionRefused})

	_, err := client.GenerateKYCLink(context.Background(), &gaian.GenerateKYCLinkRequest{UserID: "usr_1"})
	if !errors.Is(err, gaian.ErrHTTPFailure) {
		t.Errorf("GenerateKYCLink() error = %v, want wrapping %v", err, gaian.ErrHTTPFailure)
	}
}

func TestGenerateKYCLink_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.Canceled})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := client.GenerateKYCLink(ctx, &gaian.GenerateKYCLinkRequest{UserID: "usr_1"}); err == nil {
		t.Fatal("expected an error for a cancelled context, got nil")
	}
}

func TestGenerateKYCLink_ContextTimeout(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.DeadlineExceeded})

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	if _, err := client.GenerateKYCLink(ctx, &gaian.GenerateKYCLinkRequest{UserID: "usr_1"}); err == nil {
		t.Fatal("expected an error for a timed-out context, got nil")
	}
}

func TestGenerateKYCLink_InvalidJSON(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, "not json"), //nolint:bodyclose // response body closed by client
	}
	client := newMockClient(t, mock)

	if _, err := client.GenerateKYCLink(context.Background(), &gaian.GenerateKYCLinkRequest{UserID: "usr_1"}); err == nil {
		t.Fatal("expected a decode error for a malformed response body, got nil")
	}
}
