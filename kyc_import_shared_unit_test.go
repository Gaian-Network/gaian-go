package gaian_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	gaian "github.com/Gaian-Network/gaian-go/v2"
)

func TestImportSharedKYC_SendsShareTokenAndParsesSession(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
			`{"data":{"sessionId":"sess_1","status":"In Review"},"requestId":"req_1"}`),
	}
	client := newMockClient(t, mock)

	resp, err := client.ImportSharedKYC(context.Background(), &gaian.ImportSharedKYCRequest{
		UserID:     "usr_1",
		ShareToken: "share-token-abc",
	})
	if err != nil {
		t.Fatalf("ImportSharedKYC: %v", err)
	}

	if mock.lastReq.Method != http.MethodPost {
		t.Errorf("method = %s, want POST", mock.lastReq.Method)
	}
	if !mock.calledPath("/api/v2/users/usr_1/kyc/import-shared") {
		t.Fatalf("path = %q", mock.lastPath())
	}

	var body map[string]any
	if err := json.Unmarshal(mock.lastBody(), &body); err != nil {
		t.Fatalf("unmarshal body: %v", err)
	}
	if len(body) != 1 || body["shareToken"] != "share-token-abc" {
		t.Errorf("body = %v, want only shareToken", body)
	}

	if resp.RequestID != "req_1" || resp.Data.SessionID != "sess_1" || resp.Data.Status != "In Review" {
		t.Errorf("response = %+v / %+v", resp, resp.Data)
	}
}

func TestImportSharedKYC_APIError(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusConflict, `{"message":"user has already submitted KYC"}`), //nolint:bodyclose // response body closed by client
	}
	client := newMockClient(t, mock)

	_, err := client.ImportSharedKYC(context.Background(), &gaian.ImportSharedKYCRequest{UserID: "usr_1", ShareToken: "t"})
	if !gaian.IsConflict(err) {
		t.Fatalf("error = %v, want a 409 APIError", err)
	}
}

func TestImportSharedKYC_EscapesUserID(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, `{"data":{"sessionId":"s","status":"Approved"},"requestId":"r"}`), //nolint:bodyclose // response body closed by client
	}
	client := newMockClient(t, mock)

	if _, err := client.ImportSharedKYC(context.Background(), &gaian.ImportSharedKYCRequest{UserID: "a/b", ShareToken: "t"}); err != nil {
		t.Fatalf("ImportSharedKYC: %v", err)
	}
	if got := mock.lastReq.URL.EscapedPath(); got != "/api/v2/users/a%2Fb/kyc/import-shared" {
		t.Errorf("escaped path = %q", got)
	}
}
