package gaian

import (
	"context"
	"net/http"
	"testing"
)

func TestCreateWallet(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v2/users/usr_1/wallets" {
			http.NotFound(w, r)
			return
		}
		body := readBody(t, r)
		if string(body) != `{"address":"0xABC","chain":"base"}` {
			t.Errorf("unexpected body: %s", body)
		}

		writeJSON(t, w, http.StatusOK, map[string]any{
			"data": map[string]any{"address": "0xABC", "chain": "base", "isPrimary": true},
		})
	})

	resp, err := client.CreateWallet(context.Background(), &CreateWalletRequest{
		UserID:  "usr_1",
		Address: "0xABC",
		Chain:   "base",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Data.IsPrimary {
		t.Error("expected first wallet to be primary")
	}
}
