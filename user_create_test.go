package gaian

import (
	"context"
	"net/http"
	"testing"
)

func TestCreateUser(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v2/users" {
			http.NotFound(w, r)
			return
		}
		body := readBody(t, r)
		verifySignature(t, r, body)

		writeJSON(t, w, http.StatusOK, map[string]any{
			"data":      map[string]any{"userId": "usr_abc123"},
			"requestId": "req_1",
		})
	})

	resp, err := client.CreateUser(context.Background(), &CreateUserRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.UserID != "usr_abc123" {
		t.Errorf("UserID = %q, want %q", resp.Data.UserID, "usr_abc123")
	}
}
