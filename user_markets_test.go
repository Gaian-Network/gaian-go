package gaian

import (
	"context"
	"net/http"
	"testing"
)

func TestGetMarkets(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/users/usr_1/markets" {
			http.NotFound(w, r)
			return
		}
		writeJSON(t, w, http.StatusOK, map[string]any{
			"data": map[string]any{
				"markets": map[string]any{
					"VN": map[string]any{"status": "APPROVED"},
					"PH": map[string]any{"status": "PENDING"},
				},
			},
		})
	})

	resp, err := client.GetMarkets(context.Background(), &GetMarketsRequest{UserID: "usr_1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data.Markets["VN"].Status != MarketApproved {
		t.Errorf(`Markets["VN"].Status = %q, want %q`, resp.Data.Markets["VN"].Status, MarketApproved)
	}
	if resp.Data.Markets["PH"].Status != MarketPending {
		t.Errorf(`Markets["PH"].Status = %q, want %q`, resp.Data.Markets["PH"].Status, MarketPending)
	}
}
