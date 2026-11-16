package gaian_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	gaian "github.com/Gaian-Network/gaian-go/v2"
)

func TestAPIErrorShapes(t *testing.T) {
	t.Parallel()

	const htmlBody = "<html><body><h1>503 Service Unavailable</h1></body></html>"

	tests := []struct {
		name          string
		status        int
		body          string
		headerReqID   string
		wantMessage   string
		wantCode      string
		wantRequestID string
	}{
		{
			name:          "flat message body with request id header",
			status:        http.StatusNotFound,
			body:          `{"message":"user not found"}`,
			headerReqID:   "req-header-1",
			wantMessage:   "user not found",
			wantRequestID: "req-header-1",
		},
		{
			name:          "payments body carries code and request id",
			status:        http.StatusBadRequest,
			body:          `{"success":false,"requestId":"req-body-2","error":{"code":"IDENTITY_REQUIRED","message":"userId, email or walletAddress is required","details":{"field":"userId"}}}`,
			wantMessage:   "userId, email or walletAddress is required",
			wantCode:      "IDENTITY_REQUIRED",
			wantRequestID: "req-body-2",
		},
		{
			name:   "502 with empty body",
			status: http.StatusBadGateway,
			body:   "",
		},
		{
			name:   "503 with html body",
			status: http.StatusServiceUnavailable,
			body:   htmlBody,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resp := mockResponse(tt.status, tt.body) //nolint:bodyclose // response body closed by client
			if tt.headerReqID != "" {
				resp.Header.Set("X-Request-ID", tt.headerReqID)
			}
			client := newMockClient(t, &mockHTTPClient{response: resp})

			_, err := client.GetUserByID(context.Background(), &gaian.GetUserByIDRequest{UserID: "usr_1"})
			if err == nil {
				t.Fatal("expected an error")
			}

			var apiErr *gaian.APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("errors.As(*APIError) failed for %T: %v", err, err)
			}
			if apiErr.StatusCode != tt.status {
				t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, tt.status)
			}
			if apiErr.Body != tt.body {
				t.Errorf("Body = %q, want the raw response body %q", apiErr.Body, tt.body)
			}
			if tt.wantMessage != "" && apiErr.Message != tt.wantMessage {
				t.Errorf("Message = %q, want %q", apiErr.Message, tt.wantMessage)
			}
			if tt.wantCode != "" && apiErr.Code != tt.wantCode {
				t.Errorf("Code = %q, want %q", apiErr.Code, tt.wantCode)
			}
			if tt.wantRequestID != "" && apiErr.RequestID != tt.wantRequestID {
				t.Errorf("RequestID = %q, want %q", apiErr.RequestID, tt.wantRequestID)
			}
			if !errors.Is(err, gaian.ErrUnexpectedStatus) {
				t.Errorf("errors.Is(err, ErrUnexpectedStatus) = false; existing callers rely on it")
			}
		})
	}
}

func TestAPIErrorPredicatesSeeWrappedErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status int
		check  func(error) bool
		want   bool
	}{
		{"404 is not found", http.StatusNotFound, gaian.IsNotFound, true},
		{"500 is not not-found", http.StatusInternalServerError, gaian.IsNotFound, false},
		{"409 is conflict", http.StatusConflict, gaian.IsConflict, true},
		{"401 is unauthorized", http.StatusUnauthorized, gaian.IsUnauthorized, true},
		{"403 is forbidden", http.StatusForbidden, gaian.IsForbidden, true},
		{"404 is not forbidden", http.StatusNotFound, gaian.IsForbidden, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mock := &mockHTTPClient{response: mockResponse(tt.status, `{"message":"x"}`)} //nolint:bodyclose // response body closed by client
			client := newMockClient(t, mock)

			_, err := client.GetUserByID(context.Background(), &gaian.GetUserByIDRequest{UserID: "usr_1"})
			wrapped := fmt.Errorf("loading profile: %w", err)

			if got := tt.check(wrapped); got != tt.want {
				t.Errorf("predicate on wrapped error = %v, want %v", got, tt.want)
			}
		})
	}
}
