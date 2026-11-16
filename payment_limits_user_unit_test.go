package gaian_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	gaian "github.com/Gaian-Network/gaian-go/v2"
)

func TestGetUserLimits_SendsQueryAndParsesNestedLimits(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
			`{"success":true,"requestId":"req_1","data":{"userId":null,`+
				`"daily":{"spentUsd":12.5,"windowStart":"2030-01-01T00:00:00Z","windowEnd":"2030-01-02T00:00:00Z"},`+
				`"velocity":{"enabled":true,"maxPerWindow":5,"windowSeconds":3600,"usedInWindow":2,"remaining":3,"retryAfterSeconds":0},`+
				`"countries":[`+
				`{"country":"VN","supported":true,"kycVerified":true,"canTransact":true,"maxPerDayUsd":1000,"dailyRemainingUsd":987.5,`+
				`"weekly":{"capUsd":5000,"spentUsd":100,"remainingUsd":4900,"windowStart":"2030-01-01T00:00:00Z","windowEnd":"2030-01-08T00:00:00Z"},`+
				`"maxNextTransactionUsd":987.5},`+
				`{"country":"BR","supported":false,"kycVerified":false,"canTransact":false,"maxNextTransactionUsd":0}]}}`),
	}
	client := newMockClient(t, mock)

	resp, err := client.GetUserLimits(context.Background(), &gaian.GetUserLimitsRequest{UserID: "usr_1", Country: "VN"})
	if err != nil {
		t.Fatalf("GetUserLimits: %v", err)
	}

	if mock.lastReq.Method != http.MethodGet {
		t.Errorf("method = %s, want GET", mock.lastReq.Method)
	}
	if !mock.calledPath("/api/v2/limits/user") {
		t.Fatalf("path = %q", mock.lastPath())
	}
	if got := mock.lastReq.URL.RawQuery; got != "country=VN&userId=usr_1" {
		t.Errorf("query = %q, want country=VN&userId=usr_1", got)
	}

	d := resp.Data
	if d.UserID != nil {
		t.Errorf("UserID = %v, want nil", *d.UserID)
	}
	if d.Daily.SpentUSD != 12.5 || d.Daily.WindowEnd != "2030-01-02T00:00:00Z" {
		t.Errorf("Daily = %+v", d.Daily)
	}
	if !d.Velocity.Enabled || d.Velocity.MaxPerWindow != 5 || d.Velocity.WindowSeconds != 3600 ||
		d.Velocity.UsedInWindow != 2 || d.Velocity.Remaining != 3 || d.Velocity.RetryAfterSeconds != 0 {
		t.Errorf("Velocity = %+v", d.Velocity)
	}
	if len(d.Countries) != 2 {
		t.Fatalf("got %d countries, want 2", len(d.Countries))
	}

	vn := d.Countries[0]
	if vn.Country != "VN" || !vn.Supported || !vn.KYCVerified || !vn.CanTransact || vn.MaxNextTransactionUSD != 987.5 {
		t.Errorf("VN = %+v", vn)
	}
	if vn.MaxPerDayUSD == nil || *vn.MaxPerDayUSD != 1000 || vn.DailyRemainingUSD == nil || *vn.DailyRemainingUSD != 987.5 {
		t.Errorf("VN daily caps = %+v", vn)
	}
	if vn.Weekly == nil || vn.Weekly.CapUSD != 5000 || vn.Weekly.SpentUSD != 100 || vn.Weekly.RemainingUSD != 4900 ||
		vn.Weekly.WindowStart != "2030-01-01T00:00:00Z" {
		t.Errorf("VN weekly = %+v", vn.Weekly)
	}

	br := d.Countries[1]
	if br.MaxPerDayUSD != nil || br.DailyRemainingUSD != nil || br.Weekly != nil {
		t.Errorf("BR optional limits should be nil, got %+v", br)
	}
}

func TestGetUserLimits_APIError(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusBadRequest, //nolint:bodyclose // response body closed by client
			`{"success":false,"requestId":"req_9","error":{"code":"VALIDATION_ERROR","message":"bad query"}}`),
	}
	client := newMockClient(t, mock)

	_, err := client.GetUserLimits(context.Background(), &gaian.GetUserLimitsRequest{})

	var apiErr *gaian.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %v, want *gaian.APIError", err)
	}
	if apiErr.StatusCode != http.StatusBadRequest || apiErr.Code != "VALIDATION_ERROR" || apiErr.RequestID != "req_9" {
		t.Errorf("apiErr = %+v", apiErr)
	}
}

func TestGetUserLimits_OmitsEmptyCountry(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, `{"success":true,"requestId":"r","data":{"userId":"usr_1","countries":[]}}`), //nolint:bodyclose // response body closed by client
	}
	client := newMockClient(t, mock)

	resp, err := client.GetUserLimits(context.Background(), &gaian.GetUserLimitsRequest{UserID: "usr_1"})
	if err != nil {
		t.Fatalf("GetUserLimits: %v", err)
	}
	if got := mock.lastReq.URL.RawQuery; got != "userId=usr_1" {
		t.Errorf("query = %q, want userId=usr_1", got)
	}
	if resp.Data.UserID == nil || *resp.Data.UserID != "usr_1" {
		t.Errorf("UserID = %v", resp.Data.UserID)
	}
}
