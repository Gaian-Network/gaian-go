package gaian_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	gaian "github.com/gaiannetwork/gaian-go"
)

func TestGetMarkets_Success(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
			`{"data":{"markets":{"VN":{"status":"APPROVED"},"PH":{"status":"REJECTED","rejectReason":"blurry ID"}}},"requestId":"req_1"}`),
	}
	client := newMockClient(t, mock)

	resp, err := client.GetMarkets(context.Background(), &gaian.GetMarketsRequest{UserID: "usr_1"})
	if err != nil {
		t.Fatalf("GetMarkets: %v", err)
	}
	if resp.Data.Markets["VN"].Status != gaian.MarketApproved {
		t.Errorf(`Markets["VN"].Status = %v, want %v`, resp.Data.Markets["VN"].Status, gaian.MarketApproved)
	}
	if resp.Data.Markets["PH"].Status != gaian.MarketRejected {
		t.Errorf(`Markets["PH"].Status = %v, want %v`, resp.Data.Markets["PH"].Status, gaian.MarketRejected)
	}
	if resp.Data.Markets["PH"].RejectReason == nil || *resp.Data.Markets["PH"].RejectReason != "blurry ID" {
		t.Errorf(`Markets["PH"].RejectReason = %v, want "blurry ID"`, resp.Data.Markets["PH"].RejectReason)
	}
	if !mock.calledPath("/api/v2/users/usr_1/markets") {
		t.Fatalf("expected request to /api/v2/users/usr_1/markets, got %q", mock.lastPath())
	}
}

func TestGetMarkets_AllStatusValues(t *testing.T) {
	t.Parallel()

	statuses := []gaian.MarketStatus{
		gaian.MarketApproved,
		gaian.MarketPending,
		gaian.MarketNotStarted,
		gaian.MarketRejected,
	}

	for _, status := range statuses {
		t.Run(string(status), func(t *testing.T) {
			t.Parallel()

			body := fmt.Sprintf(`{"data":{"markets":{"VN":{"status":%q}}},"requestId":"req_1"}`, status)
			mock := &mockHTTPClient{response: mockResponse(http.StatusOK, body)} //nolint:bodyclose // response body closed by client
			client := newMockClient(t, mock)

			resp, err := client.GetMarkets(context.Background(), &gaian.GetMarketsRequest{UserID: "usr_1"})
			if err != nil {
				t.Fatalf("GetMarkets: %v", err)
			}
			if resp.Data.Markets["VN"].Status != status {
				t.Errorf(`Markets["VN"].Status = %v, want %v`, resp.Data.Markets["VN"].Status, status)
			}
		})
	}
}

func TestGetMarkets_HTTPErrorStatuses(t *testing.T) {
	t.Parallel()

	statuses := []int{http.StatusNotFound, http.StatusInternalServerError}

	for _, status := range statuses {
		t.Run(http.StatusText(status), func(t *testing.T) {
			t.Parallel()

			mock := &mockHTTPClient{
				response: mockResponse(status, `{"message":"user not found"}`), //nolint:bodyclose // response body closed by client
			}
			client := newMockClient(t, mock)

			_, err := client.GetMarkets(context.Background(), &gaian.GetMarketsRequest{UserID: "usr_1"})
			if !errors.Is(err, gaian.ErrUnexpectedStatus) {
				t.Errorf("GetMarkets() error = %v, want wrapping %v", err, gaian.ErrUnexpectedStatus)
			}
		})
	}
}

func TestGetMarkets_NetworkError(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: errConnectionRefused})

	_, err := client.GetMarkets(context.Background(), &gaian.GetMarketsRequest{UserID: "usr_1"})
	if !errors.Is(err, gaian.ErrHTTPFailure) {
		t.Errorf("GetMarkets() error = %v, want wrapping %v", err, gaian.ErrHTTPFailure)
	}
}

func TestGetMarkets_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.Canceled})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := client.GetMarkets(ctx, &gaian.GetMarketsRequest{UserID: "usr_1"}); err == nil {
		t.Fatal("expected an error for a cancelled context, got nil")
	}
}

func TestGetMarkets_ContextTimeout(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.DeadlineExceeded})

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	if _, err := client.GetMarkets(ctx, &gaian.GetMarketsRequest{UserID: "usr_1"}); err == nil {
		t.Fatal("expected an error for a timed-out context, got nil")
	}
}

func TestGetMarkets_InvalidJSON(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, "not json"), //nolint:bodyclose // response body closed by client
	}
	client := newMockClient(t, mock)

	if _, err := client.GetMarkets(context.Background(), &gaian.GetMarketsRequest{UserID: "usr_1"}); err == nil {
		t.Fatal("expected a decode error for a malformed response body, got nil")
	}
}
