package gaian

import (
	"context"
	"net/http"
	"testing"
)

func TestListWallets(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/users/usr_1/wallets" {
			http.NotFound(w, r)
			return
		}
		writeJSON(t, w, http.StatusOK, map[string]any{
			"data": []map[string]any{
				{"address": "0xABC", "chain": "base", "isPrimary": true},
			},
		})
	})

	resp, err := client.ListWallets(context.Background(), &ListWalletsRequest{UserID: "usr_1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(*resp.Data) != 1 || (*resp.Data)[0].Address != "0xABC" {
		t.Errorf("Data = %+v", resp.Data)
	}
}
