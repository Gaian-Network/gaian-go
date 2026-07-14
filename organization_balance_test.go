package gaian

import (
	"context"
	"net/http"
	"testing"
)

func TestGetOrganizationBalance(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/organization/me/balance" {
			http.NotFound(w, r)
			return
		}
		writeJSON(t, w, http.StatusOK, map[string]any{
			"data": map[string]any{
				"currency":          "USDC",
				"availableBalance":  "100.5",
				"totalSpent":        "0",
				"pendingSettlement": "0",
			},
		})
	})

	resp, err := client.GetOrganizationBalance(context.Background(), &GetOrganizationBalanceRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.AvailableBalance != "100.5" {
		t.Errorf("AvailableBalance = %q, want 100.5", resp.Data.AvailableBalance)
	}
}
