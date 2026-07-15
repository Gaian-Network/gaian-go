package gaian_test

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	gaian "github.com/gaiannetwork/gaian-go"
)

func TestGetOrganizationBalance_Success(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
			`{"data":{"currency":"USDC","availableBalance":"1000.50","totalSpent":"500.00",`+
				`"pendingSettlement":"20.00"},"requestId":"req_1"}`),
	}
	client := newMockClient(t, mock)

	resp, err := client.GetOrganizationBalance(context.Background(), &gaian.GetOrganizationBalanceRequest{})
	if err != nil {
		t.Fatalf("GetOrganizationBalance: %v", err)
	}
	if resp.Data.AvailableBalance != "1000.50" {
		t.Errorf("AvailableBalance = %q, want %q", resp.Data.AvailableBalance, "1000.50")
	}
	if !mock.calledPath("/api/v2/organization/me/balance") {
		t.Fatalf("expected request to /api/v2/organization/me/balance, got %q", mock.lastPath())
	}
}

func TestGetOrganizationBalance_HTTPErrorStatuses(t *testing.T) {
	t.Parallel()

	// 400 is the documented outcome for a tenant not on the prefund plan.
	statuses := []int{http.StatusBadRequest, http.StatusInternalServerError}

	for _, status := range statuses {
		t.Run(http.StatusText(status), func(t *testing.T) {
			t.Parallel()

			mock := &mockHTTPClient{
				response: mockResponse(status, `{"message":"tenant is not prefund-enabled"}`), //nolint:bodyclose // response body closed by client
			}
			client := newMockClient(t, mock)

			_, err := client.GetOrganizationBalance(context.Background(), &gaian.GetOrganizationBalanceRequest{})
			if !errors.Is(err, gaian.ErrUnexpectedStatus) {
				t.Errorf("GetOrganizationBalance() error = %v, want wrapping %v", err, gaian.ErrUnexpectedStatus)
			}
		})
	}
}

func TestGetOrganizationBalance_NetworkError(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: errConnectionRefused})

	_, err := client.GetOrganizationBalance(context.Background(), &gaian.GetOrganizationBalanceRequest{})
	if !errors.Is(err, gaian.ErrHTTPFailure) {
		t.Errorf("GetOrganizationBalance() error = %v, want wrapping %v", err, gaian.ErrHTTPFailure)
	}
}

func TestGetOrganizationBalance_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.Canceled})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := client.GetOrganizationBalance(ctx, &gaian.GetOrganizationBalanceRequest{}); err == nil {
		t.Fatal("expected an error for a cancelled context, got nil")
	}
}

func TestGetOrganizationBalance_ContextTimeout(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.DeadlineExceeded})

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	if _, err := client.GetOrganizationBalance(ctx, &gaian.GetOrganizationBalanceRequest{}); err == nil {
		t.Fatal("expected an error for a timed-out context, got nil")
	}
}

func TestGetOrganizationBalance_InvalidJSON(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, "not json"), //nolint:bodyclose // response body closed by client
	}
	client := newMockClient(t, mock)

	if _, err := client.GetOrganizationBalance(context.Background(), &gaian.GetOrganizationBalanceRequest{}); err == nil {
		t.Fatal("expected a decode error for a malformed response body, got nil")
	}
}
