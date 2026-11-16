package gaian_test

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	gaian "github.com/Gaian-Network/gaian-go/v2"
)

func TestGetMarketRequirements_ParsesRequiredAndMissingFieldsPerCountry(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
			`{"data":{"markets":{"VN":{"requiredFields":["nationalId","kycReport"],"missingFields":["kycReport"]},`+
				`"TH":{"requiredFields":[],"missingFields":[]}}},"requestId":"req_1"}`),
	}
	client := newMockClient(t, mock)

	resp, err := client.GetMarketRequirements(context.Background(), &gaian.GetMarketRequirementsRequest{UserID: "usr_1"})
	if err != nil {
		t.Fatalf("GetMarketRequirements: %v", err)
	}

	if mock.lastReq.Method != http.MethodGet {
		t.Errorf("method = %s, want GET", mock.lastReq.Method)
	}
	if !mock.calledPath("/api/v2/users/usr_1/markets/requirements") {
		t.Fatalf("path = %q", mock.lastPath())
	}
	if len(mock.lastBody()) != 0 || mock.lastReq.URL.RawQuery != "" {
		t.Errorf("GET must carry no body or query, got body=%q query=%q", mock.lastBody(), mock.lastReq.URL.RawQuery)
	}

	vn := resp.Data.Markets["VN"]
	if !reflect.DeepEqual(vn.RequiredFields, []string{"nationalId", "kycReport"}) ||
		!reflect.DeepEqual(vn.MissingFields, []string{"kycReport"}) {
		t.Errorf("VN = %+v", vn)
	}
	if _, ok := resp.Data.Markets["TH"]; !ok {
		t.Error("TH market missing from parsed response")
	}
}

func TestGetMarketRequirements_APIError(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusNotFound, `{"message":"user not found"}`), //nolint:bodyclose // response body closed by client
	}
	client := newMockClient(t, mock)

	_, err := client.GetMarketRequirements(context.Background(), &gaian.GetMarketRequirementsRequest{UserID: "usr_1"})
	if !gaian.IsNotFound(err) {
		t.Fatalf("error = %v, want a 404 APIError", err)
	}
}
