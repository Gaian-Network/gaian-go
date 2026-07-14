package gaian

import (
	"context"
	"net/http"
	"testing"
)

func TestParseQR(t *testing.T) {
	client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v2/parseQr" {
			http.NotFound(w, r)
			return
		}
		writeJSON(t, w, http.StatusOK, map[string]any{
			"success": true,
			"data": map[string]any{
				"isValid":         true,
				"country":         "VN",
				"beneficiaryName": "NGUYEN VAN A",
				"bankBin":         "970407",
				"accountNumber":   "0123456789",
			},
			"requestId": "req_5",
		})
	})

	resp, err := client.ParseQR(context.Background(), &ParseQRRequest{QRString: "00020101..."})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Error("expected success=true")
	}
	if !resp.Data.IsValid || resp.Data.Country != "VN" {
		t.Errorf("Data = %+v", resp.Data)
	}
}
