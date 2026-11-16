package gaian_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	gaian "github.com/Gaian-Network/gaian-go/v2"
)

func TestGetExchangeRates_SendsQueryAndKeepsRateAsString(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
			`{"success":true,"requestId":"req_1","data":{"indicativeOnly":true,"rates":[`+
				`{"country":"VN","fiatCurrency":"VND","settlementCurrency":"USDC","routeId":7,`+
				`"exchangeRate":"25300.0000","capturedAt":"2030-01-01T00:00:00Z"}]}}`),
	}
	client := newMockClient(t, mock)

	resp, err := client.GetExchangeRates(context.Background(), &gaian.GetExchangeRatesRequest{
		Country:            "VN",
		SettlementCurrency: "USDC",
	})
	if err != nil {
		t.Fatalf("GetExchangeRates: %v", err)
	}

	if mock.lastReq.Method != http.MethodGet {
		t.Errorf("method = %s, want GET", mock.lastReq.Method)
	}
	if !mock.calledPath("/api/v2/exchange-rate") {
		t.Fatalf("path = %q", mock.lastPath())
	}
	if got := mock.lastReq.URL.RawQuery; got != "country=VN&settlementCurrency=USDC" {
		t.Errorf("query = %q", got)
	}

	if !resp.Data.IndicativeOnly || len(resp.Data.Rates) != 1 {
		t.Fatalf("data = %+v", resp.Data)
	}
	r := resp.Data.Rates[0]
	if r.Country != "VN" || r.FiatCurrency != "VND" || r.SettlementCurrency != "USDC" || r.RouteID != 7 ||
		r.ExchangeRate != "25300.0000" || r.CapturedAt != "2030-01-01T00:00:00Z" {
		t.Errorf("rate = %+v", r)
	}
}

func TestGetExchangeRates_APIError(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusBadRequest, //nolint:bodyclose // response body closed by client
			`{"success":false,"requestId":"req_9","error":{"code":"VALIDATION_ERROR","message":"bad query"}}`),
	}
	client := newMockClient(t, mock)

	_, err := client.GetExchangeRates(context.Background(), &gaian.GetExchangeRatesRequest{})

	var apiErr *gaian.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %v, want *gaian.APIError", err)
	}
	if apiErr.StatusCode != http.StatusBadRequest || apiErr.Code != "VALIDATION_ERROR" || apiErr.RequestID != "req_9" {
		t.Errorf("apiErr = %+v", apiErr)
	}
}

func TestGetExchangeRates_NoFiltersSendsNoQuery(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, `{"success":true,"requestId":"r","data":{"indicativeOnly":true,"rates":[]}}`), //nolint:bodyclose // response body closed by client
	}
	client := newMockClient(t, mock)

	if _, err := client.GetExchangeRates(context.Background(), &gaian.GetExchangeRatesRequest{}); err != nil {
		t.Fatalf("GetExchangeRates: %v", err)
	}
	if got := mock.lastReq.URL.RawQuery; got != "" {
		t.Errorf("query = %q, want empty", got)
	}
}
