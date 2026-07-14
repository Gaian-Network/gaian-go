package gaian

import (
	"context"
	"net/http"
	"testing"
)

func TestGenerateKYCLink(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v2/users/usr_1/kyc-url" {
			http.NotFound(w, r)
			return
		}
		writeJSON(t, w, http.StatusOK, map[string]any{
			"data": map[string]any{"url": "https://kyc.example.com/session/abc123"},
		})
	})

	resp, err := client.GenerateKYCLink(context.Background(), &GenerateKYCLinkRequest{UserID: "usr_1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.URL != "https://kyc.example.com/session/abc123" {
		t.Errorf("URL = %q", resp.Data.URL)
	}
}
