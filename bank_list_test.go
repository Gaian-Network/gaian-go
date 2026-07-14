package gaian

import (
	"context"
	"net/http"
	"testing"
)

func TestListBanks(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/banks" {
			http.NotFound(w, r)
			return
		}
		if r.URL.Query().Get("country") != "VN" {
			t.Errorf(`query "country" = %q, want "VN"`, r.URL.Query().Get("country"))
		}

		writeJSON(t, w, http.StatusOK, map[string]any{
			"success": true,
			"data": map[string]any{
				"country": "VN",
				"banks": []map[string]any{
					{"id": 1, "code": "970418", "name": "BIDV"},
				},
			},
		})
	})

	resp, err := client.ListBanks(context.Background(), &ListBanksRequest{Country: "VN"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data.Banks) != 1 || resp.Data.Banks[0].Code != "970418" {
		t.Errorf("Banks = %+v", resp.Data.Banks)
	}
}
