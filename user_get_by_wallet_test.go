package gaian

import (
	"context"
	"net/http"
	"testing"
)

func TestGetUserByWallet(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/users/wallet/0xDEF" {
			http.NotFound(w, r)
			return
		}
		writeJSON(t, w, http.StatusOK, map[string]any{
			"data":      map[string]any{"id": "usr_7", "kycStatus": "APPROVED"},
			"requestId": "req_3",
		})
	})

	resp, err := client.GetUserByWallet(context.Background(), &GetUserByWalletRequest{WalletAddress: "0xDEF"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.ID != "usr_7" {
		t.Errorf("ID = %q, want usr_7", resp.Data.ID)
	}
}
