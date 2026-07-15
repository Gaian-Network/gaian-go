package gaian_test

import (
	"context"
	"testing"
	"time"

	gaian "github.com/gaiannetwork/gaian-go"
)

func validSandboxSubmitKYCRequest(userID string) *gaian.SubmitKYCRequest {
	return &gaian.SubmitKYCRequest{
		UserID:             userID,
		FirstName:          "Nguyen",
		LastName:           "Van A",
		Email:              "sandbox-test@example.com",
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

// TestSandbox_SubmitKYC hits the real sandbox environment. Requires
// GAIAN_SANDBOX_USER_ID for a real, pre-existing sandbox user — the KYC
// payload itself is otherwise well-formed sample data.
func TestSandbox_SubmitKYC(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)
	userID := getTestUserID(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := client.SubmitKYC(ctx, validSandboxSubmitKYCRequest(userID))
	assertServerResponded(t, err)
}

func TestSandbox_SubmitKYC_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newTestClient(t)
	userID := getTestUserID(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := client.SubmitKYC(ctx, validSandboxSubmitKYCRequest(userID)); err == nil {
		t.Error("expected an error for a cancelled context, got nil")
	}
}
