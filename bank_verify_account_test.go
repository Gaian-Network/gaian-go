package gaian

import (
	"context"
	"net/http"
	"testing"
)

func TestVerifyAccount(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v2/accounts/verify" {
			http.NotFound(w, r)
			return
		}
		writeJSON(t, w, http.StatusOK, map[string]any{
			"success": true,
			"data": map[string]any{
				"country":       "VN",
				"accountNumber": "0123456789",
				"code":          "970418",
				"valid":         true,
				"accountName":   "NGUYEN VAN A",
			},
		})
	})

	resp, err := client.VerifyAccount(context.Background(), &VerifyAccountRequest{
		AccountNumber: "0123456789",
		Code:          "970418",
		Country:       "VN",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Data.Valid {
		t.Error("expected Valid=true")
	}
}

// TestVerifyAccount_InvalidIsNotAnError locks in the documented "a failed
// verification is NOT an error" behavior.
func TestVerifyAccount_InvalidIsNotAnError(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, http.StatusOK, map[string]any{
			"success": true,
			"data": map[string]any{
				"country":       "VN",
				"accountNumber": "0000000000",
				"code":          "970418",
				"valid":         false,
				"accountName":   nil,
			},
		})
	})

	resp, err := client.VerifyAccount(context.Background(), &VerifyAccountRequest{
		AccountNumber: "0000000000",
		Code:          "970418",
		Country:       "VN",
	})
	if err != nil {
		t.Fatalf("unexpected error (invalid account should not be a Go error): %v", err)
	}
	if resp.Data.Valid {
		t.Error("expected Valid=false")
	}
	if resp.Data.AccountName != nil {
		t.Errorf("AccountName = %v, want nil", resp.Data.AccountName)
	}
}
