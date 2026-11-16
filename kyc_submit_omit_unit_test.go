package gaian_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	gaian "github.com/Gaian-Network/gaian-go/v2"
)

func submitKYCBody(t *testing.T, req *gaian.SubmitKYCRequest) map[string]any {
	t.Helper()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
			`{"data":{"userId":"usr_1","email":"user@example.com","kycStatus":"PROCESSING"},"requestId":"req_1"}`),
	}
	client := newMockClient(t, mock)

	if _, err := client.SubmitKYC(context.Background(), req); err != nil {
		t.Fatalf("SubmitKYC: %v", err)
	}

	var body map[string]any
	if err := json.Unmarshal(mock.lastBody(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}

	return body
}

func TestSubmitKYCOmitsEmpty(t *testing.T) {
	t.Parallel()

	req := validSubmitKYCRequest()
	req.Gender = ""
	req.PhoneCountryCode = ""
	req.PhoneNumber = ""

	body := submitKYCBody(t, req)

	for _, key := range []string{"gender", "phoneCountryCode", "phoneNumber"} {
		if v, ok := body[key]; ok {
			t.Errorf("empty optional %s was sent as %v", key, v)
		}
	}
	// Required fields stay in the body.
	for _, key := range []string{"firstName", "nationalId", "type", "expiryDate"} {
		if _, ok := body[key]; !ok {
			t.Errorf("required field %s missing from body", key)
		}
	}
}

func TestSubmitKYC_SendsOptionalFieldsWhenSet(t *testing.T) {
	t.Parallel()

	body := submitKYCBody(t, validSubmitKYCRequest())

	if body["gender"] != "male" || body["phoneCountryCode"] != "+84" || body["phoneNumber"] != "912345678" {
		t.Errorf("gender/phone not sent: %v", body)
	}
}

func TestSubmitKYC_AcceptsNewDocumentTypesAndKYCReport(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		docType gaian.DocumentType
		want    string
	}{
		{gaian.DocumentTypeDriversLicense, "DRIVERS_LICENSE"},
		{gaian.DocumentTypeResidencePermit, "RESIDENCE_PERMIT"},
	} {
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()

			report := "data:application/pdf;base64,AAAA"
			req := validSubmitKYCRequest()
			req.Type = tc.docType
			req.KYCReport = &report

			body := submitKYCBody(t, req)

			if body["type"] != tc.want {
				t.Errorf("type = %v, want %s", body["type"], tc.want)
			}
			if body["kycReport"] != report {
				t.Errorf("kycReport = %v, want %s", body["kycReport"], report)
			}
		})
	}
}

func TestSubmitKYC_OmitsKYCReportWhenNil(t *testing.T) {
	t.Parallel()

	if _, ok := submitKYCBody(t, validSubmitKYCRequest())["kycReport"]; ok {
		t.Error("nil kycReport was sent")
	}
}
