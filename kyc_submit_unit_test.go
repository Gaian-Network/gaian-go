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

func validSubmitKYCRequest() *gaian.SubmitKYCRequest {
	return &gaian.SubmitKYCRequest{
		UserID:             "usr_1",
		FirstName:          "Nguyen",
		LastName:           "Van A",
		Email:              "user@example.com",
		Gender:             gaian.GenderMale,
		DateOfBirth:        "1990-01-15",
		Nationality:        gaian.NationalityCode("VN"),
		NationalID:         "012345678901",
		Type:               gaian.DocumentTypeIDCard,
		ExpiryDate:         "2030-01-01",
		AddressLine1:       "123 Le Loi",
		City:               "Ho Chi Minh City",
		CountryOfResidence: "VN",
		Occupation:         gaian.OccupationProfessional,
		PhoneCountryCode:   "+84",
		PhoneNumber:        "912345678",
	}
}

func TestSubmitKYC_Success(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
			`{"data":{"userId":"usr_1","email":"user@example.com","kycStatus":"PROCESSING"},"requestId":"req_1"}`),
	}
	client := newMockClient(t, mock)

	resp, err := client.SubmitKYC(context.Background(), validSubmitKYCRequest())
	if err != nil {
		t.Fatalf("SubmitKYC: %v", err)
	}
	if resp.Data.UserID != "usr_1" {
		t.Errorf("UserID = %q, want %q", resp.Data.UserID, "usr_1")
	}
	if resp.Data.Email != "user@example.com" {
		t.Errorf("Email = %q, want %q", resp.Data.Email, "user@example.com")
	}
	if resp.Data.KYCStatus != "PROCESSING" {
		t.Errorf("KYCStatus = %q, want %q", resp.Data.KYCStatus, "PROCESSING")
	}
}

// TestSubmitKYCRequest_WireFormat locks down SubmitKYCRequest's wire
// format: UserID must go in the path only (json:"-", never in the body),
// and empty optional fields (IssueDate, AddressLine2, ...) must not leak
// into the JSON body since the API distinguishes "omitted" from "empty
// string" for several of them.
func TestSubmitKYCRequest_WireFormat(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
			`{"data":{"userId":"usr_1","email":"user@example.com","kycStatus":"PROCESSING"},"requestId":"req_1"}`),
	}
	client := newMockClient(t, mock)

	if _, err := client.SubmitKYC(context.Background(), validSubmitKYCRequest()); err != nil {
		t.Fatalf("SubmitKYC: %v", err)
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
	if _, ok := body["issueDate"]; ok {
		t.Error("empty optional issueDate leaked into the request body")
	}
	if _, ok := body["addressLine2"]; ok {
		t.Error("empty optional addressLine2 leaked into the request body")
	}
}

func TestSubmitKYC_HTTPErrorStatuses(t *testing.T) {
	t.Parallel()

	statuses := []int{
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusConflict,
		http.StatusInternalServerError,
	}

	for _, status := range statuses {
		t.Run(http.StatusText(status), func(t *testing.T) {
			t.Parallel()

			mock := &mockHTTPClient{
				response: mockResponse(status, `{"message":"rejected"}`), //nolint:bodyclose // response body closed by client
			}
			client := newMockClient(t, mock)

			_, err := client.SubmitKYC(context.Background(), validSubmitKYCRequest())
			if !errors.Is(err, gaian.ErrUnexpectedStatus) {
				t.Errorf("SubmitKYC() error = %v, want wrapping %v", err, gaian.ErrUnexpectedStatus)
			}
		})
	}
}

func TestSubmitKYC_NetworkError(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: errConnectionRefused})

	_, err := client.SubmitKYC(context.Background(), validSubmitKYCRequest())
	if !errors.Is(err, gaian.ErrHTTPFailure) {
		t.Errorf("SubmitKYC() error = %v, want wrapping %v", err, gaian.ErrHTTPFailure)
	}
}

func TestSubmitKYC_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.Canceled})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := client.SubmitKYC(ctx, validSubmitKYCRequest()); err == nil {
		t.Fatal("expected an error for a cancelled context, got nil")
	}
}

func TestSubmitKYC_ContextTimeout(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.DeadlineExceeded})

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	if _, err := client.SubmitKYC(ctx, validSubmitKYCRequest()); err == nil {
		t.Fatal("expected an error for a timed-out context, got nil")
	}
}

func TestSubmitKYC_InvalidJSON(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, "not json"), //nolint:bodyclose // response body closed by client
	}
	client := newMockClient(t, mock)

	if _, err := client.SubmitKYC(context.Background(), validSubmitKYCRequest()); err == nil {
		t.Fatal("expected a decode error for a malformed response body, got nil")
	}
}

func TestSubmitKYC_GenderValues(t *testing.T) {
	t.Parallel()

	for _, gender := range []gaian.Gender{gaian.GenderMale, gaian.GenderFemale} {
		t.Run(string(gender), func(t *testing.T) {
			t.Parallel()

			mock := &mockHTTPClient{
				response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
					`{"data":{"userId":"usr_1","email":"user@example.com","kycStatus":"PROCESSING"},"requestId":"req_1"}`),
			}
			client := newMockClient(t, mock)

			req := validSubmitKYCRequest()
			req.Gender = gender

			if _, err := client.SubmitKYC(context.Background(), req); err != nil {
				t.Fatalf("SubmitKYC: %v", err)
			}

			var body map[string]any
			if err := json.Unmarshal(mock.lastBody(), &body); err != nil {
				t.Fatalf("unmarshal request body: %v", err)
			}
			if body["gender"] != string(gender) {
				t.Errorf(`body["gender"] = %v, want %q`, body["gender"], gender)
			}
		})
	}
}

func TestSubmitKYC_DocumentTypes(t *testing.T) {
	t.Parallel()

	for _, docType := range []gaian.DocumentType{gaian.DocumentTypeIDCard, gaian.DocumentTypePassport} {
		t.Run(string(docType), func(t *testing.T) {
			t.Parallel()

			mock := &mockHTTPClient{
				response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
					`{"data":{"userId":"usr_1","email":"user@example.com","kycStatus":"PROCESSING"},"requestId":"req_1"}`),
			}
			client := newMockClient(t, mock)

			req := validSubmitKYCRequest()
			req.Type = docType

			if _, err := client.SubmitKYC(context.Background(), req); err != nil {
				t.Fatalf("SubmitKYC: %v", err)
			}

			var body map[string]any
			if err := json.Unmarshal(mock.lastBody(), &body); err != nil {
				t.Fatalf("unmarshal request body: %v", err)
			}
			if body["type"] != string(docType) {
				t.Errorf(`body["type"] = %v, want %q`, body["type"], docType)
			}
		})
	}
}
