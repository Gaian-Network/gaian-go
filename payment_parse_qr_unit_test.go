package gaian_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	gaian "github.com/gaiannetwork/gaian-go"
)

func TestParseQR_Success(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
			`{"success":true,"requestId":"req_1","data":{"isValid":true,"encodedString":"raw-qr","country":"VN",`+
				`"qrProvider":"napas","bankBin":"970436","accountNumber":"0123456789","beneficiaryName":"NGUYEN VAN A"}}`),
	}
	client := newMockClient(t, mock)

	resp, err := client.ParseQR(context.Background(), &gaian.ParseQRRequest{QRString: "raw-qr"})
	if err != nil {
		t.Fatalf("ParseQR: %v", err)
	}
	if !resp.Data.IsValid {
		t.Error("expected IsValid=true")
	}
	if resp.Data.BankBin != "970436" {
		t.Errorf("BankBin = %q, want %q", resp.Data.BankBin, "970436")
	}
}

// TestParseQR_InvalidQR_NotAnError guards the documented behavior: an
// unparseable QR string is not a client error, it's a normal 200 response
// with IsValid=false.
func TestParseQR_InvalidQR_NotAnError(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, `{"success":true,"requestId":"req_1","data":{"isValid":false}}`), //nolint:bodyclose // response body closed by client
	}
	client := newMockClient(t, mock)

	resp, err := client.ParseQR(context.Background(), &gaian.ParseQRRequest{QRString: "garbage"})
	if err != nil {
		t.Fatalf("ParseQR: %v (an invalid QR must not surface as an error)", err)
	}
	if resp.Data.IsValid {
		t.Error("expected IsValid=false for a garbage QR string")
	}
}

// TestParseQRRequest_WireFormat guards the omitempty Country field.
func TestParseQRRequest_WireFormat(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, `{"success":true,"requestId":"req_1","data":{"isValid":true}}`), //nolint:bodyclose // response body closed by client
	}
	client := newMockClient(t, mock)

	if _, err := client.ParseQR(context.Background(), &gaian.ParseQRRequest{QRString: "raw-qr"}); err != nil {
		t.Fatalf("ParseQR: %v", err)
	}

	var body map[string]any
	if err := json.Unmarshal(mock.lastBody(), &body); err != nil {
		t.Fatalf("unmarshal request body: %v", err)
	}
	if _, ok := body["country"]; ok {
		t.Error("empty optional country leaked into the request body")
	}
}

func TestParseQR_HTTPErrorStatuses(t *testing.T) {
	t.Parallel()

	statuses := []int{http.StatusBadRequest, http.StatusInternalServerError}

	for _, status := range statuses {
		t.Run(http.StatusText(status), func(t *testing.T) {
			t.Parallel()

			mock := &mockHTTPClient{
				response: mockResponse(status, `{"message":"rejected"}`), //nolint:bodyclose // response body closed by client
			}
			client := newMockClient(t, mock)

			_, err := client.ParseQR(context.Background(), &gaian.ParseQRRequest{QRString: "raw-qr"})
			if !errors.Is(err, gaian.ErrUnexpectedStatus) {
				t.Errorf("ParseQR() error = %v, want wrapping %v", err, gaian.ErrUnexpectedStatus)
			}
		})
	}
}

func TestParseQR_NetworkError(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: errConnectionRefused})

	_, err := client.ParseQR(context.Background(), &gaian.ParseQRRequest{QRString: "raw-qr"})
	if !errors.Is(err, gaian.ErrHTTPFailure) {
		t.Errorf("ParseQR() error = %v, want wrapping %v", err, gaian.ErrHTTPFailure)
	}
}

func TestParseQR_ContextCancellation(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.Canceled})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := client.ParseQR(ctx, &gaian.ParseQRRequest{QRString: "raw-qr"}); err == nil {
		t.Fatal("expected an error for a cancelled context, got nil")
	}
}

func TestParseQR_ContextTimeout(t *testing.T) {
	t.Parallel()

	client := newMockClient(t, &mockHTTPClient{err: context.DeadlineExceeded})

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	if _, err := client.ParseQR(ctx, &gaian.ParseQRRequest{QRString: "raw-qr"}); err == nil {
		t.Fatal("expected an error for a timed-out context, got nil")
	}
}

func TestParseQR_InvalidJSON(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, "not json"), //nolint:bodyclose // response body closed by client
	}
	client := newMockClient(t, mock)

	if _, err := client.ParseQR(context.Background(), &gaian.ParseQRRequest{QRString: "raw-qr"}); err == nil {
		t.Fatal("expected a decode error for a malformed response body, got nil")
	}
}
