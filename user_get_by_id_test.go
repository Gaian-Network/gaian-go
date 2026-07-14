package gaian

import (
	"context"
	"net/http"
	"testing"
)

func TestGetUserByID(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/users/usr_abc123" {
			http.NotFound(w, r)
			return
		}
		assertSignedHeadersPresent(t, r)

		writeJSON(t, w, http.StatusOK, map[string]any{
			"data": map[string]any{
				"id":        "usr_abc123",
				"firstName": "Nguyen",
				"kycStatus": "APPROVED",
			},
			"requestId": "req_2",
		})
	})

	resp, err := client.GetUserByID(context.Background(), &GetUserByIDRequest{UserID: "usr_abc123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.KYCStatus != "APPROVED" {
		t.Errorf("KYCStatus = %q, want APPROVED", resp.Data.KYCStatus)
	}
}

func TestGetUserByID_PathEscaping(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/users/usr with spaces" && r.URL.EscapedPath() != "/api/v2/users/usr%20with%20spaces" {
			t.Errorf("unexpected path: %s (raw %s)", r.URL.Path, r.URL.EscapedPath())
		}
		writeJSON(t, w, http.StatusOK, map[string]any{"data": map[string]any{"id": "x"}})
	})

	_, err := client.GetUserByID(context.Background(), &GetUserByIDRequest{UserID: "usr with spaces"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
