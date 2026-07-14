package gaian

import (
	"context"
	"net/http"
	"testing"
)

func TestUpdateKYC(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/api/v2/users/usr_1/kyc" {
			http.NotFound(w, r)
			return
		}
		body := readBody(t, r)
		if string(body) != `{"email":"new@example.com"}` {
			t.Errorf("unexpected body: %s", body)
		}

		writeJSON(t, w, http.StatusOK, map[string]any{
			"data": map[string]any{"userId": "usr_1", "kycStatus": "APPROVED", "email": "new@example.com"},
		})
	})

	newEmail := "new@example.com"
	resp, err := client.UpdateKYC(context.Background(), &UpdateKYCRequest{UserID: "usr_1", Email: &newEmail})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Email != "new@example.com" {
		t.Errorf("Email = %q", resp.Data.Email)
	}
}
