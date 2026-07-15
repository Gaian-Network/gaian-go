package gaian

import (
	"context"
	"net/http"
	"testing"
)

func TestSubmitKYC(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v2/users/usr_1/kyc" {
			http.NotFound(w, r)
			return
		}
		body := readBody(t, r)
		verifySignature(t, r, body)

		writeJSON(t, w, http.StatusOK, map[string]any{
			"data": map[string]any{"userId": "usr_1", "email": "user@example.com", "kycStatus": "PROCESSING"},
		})
	})

	resp, err := client.SubmitKYC(context.Background(), &SubmitKYCRequest{
		UserID:             "usr_1",
		FirstName:          "Nguyen",
		LastName:           "Van A",
		Email:              "user@example.com",
		Gender:             GenderMale,
		DateOfBirth:        "1990-01-15",
		Nationality:        "VN",
		NationalID:         "012345678901",
		Type:               DocumentTypeIDCard,
		ExpiryDate:         "2025-01-01",
		AddressLine1:       "123 Le Loi",
		City:               "Ho Chi Minh City",
		CountryOfResidence: "VN",
		Occupation:         OccupationProfessional,
		PhoneCountryCode:   "+84",
		PhoneNumber:        "912345678",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.KYCStatus != "PROCESSING" {
		t.Errorf("KYCStatus = %q, want PROCESSING", resp.Data.KYCStatus)
	}
}
