package gaian

import (
	"context"
	"net/http"
	"testing"
)

func TestListUserOrders(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/users/usr_1/orders" {
			http.NotFound(w, r)
			return
		}
		q := r.URL.Query()
		if q.Get("status") != "completed" {
			t.Errorf(`query "status" = %q, want "completed"`, q.Get("status"))
		}
		if q.Get("page") != "2" {
			t.Errorf(`query "page" = %q, want "2"`, q.Get("page"))
		}
		if q.Has("userId") {
			t.Error(`"userId" leaked into the query string`)
		}

		// requestId/data/pagination are siblings at the top level for
		// this endpoint — no outer {data: ...} envelope.
		writeJSON(t, w, http.StatusOK, map[string]any{
			"requestId": "req_4",
			"data": []map[string]any{
				{"orderId": "ord_1", "status": "completed", "fiatAmount": "100000"},
			},
			"pagination": map[string]any{"page": 2, "pageSize": 20, "totalCount": 1, "totalPages": 1},
		})
	})

	resp, err := client.ListUserOrders(context.Background(), &ListUserOrdersRequest{
		UserID: "usr_1",
		Status: "completed",
		Page:   2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Items) != 1 || resp.Items[0].OrderID != "ord_1" {
		t.Errorf("Items = %+v", resp.Items)
	}
	if resp.Pagination.TotalCount != 1 {
		t.Errorf("Pagination.TotalCount = %d, want 1", resp.Pagination.TotalCount)
	}
}
