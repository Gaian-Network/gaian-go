package gaian_test

import (
	"context"
	"net/http"
	"testing"

	gaian "github.com/Gaian-Network/gaian-go/v2"
)

func TestGetPaymentMethods_ParsesArrayWithOptionalFields(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusOK, //nolint:bodyclose // response body closed by client
			`{"data":[{"method":"QR_PAYMENT","name":"QR Payment Method","status":"APPROVED"},`+
				`{"method":"CARD_PAYMENT","name":"Card Payment Method","status":"REJECTED",`+
				`"rejectReason":"Document expired","action":"contact_support"}],"requestId":"req_1"}`),
	}
	client := newMockClient(t, mock)

	resp, err := client.GetPaymentMethods(context.Background(), &gaian.GetPaymentMethodsRequest{UserID: "usr_1"})
	if err != nil {
		t.Fatalf("GetPaymentMethods: %v", err)
	}

	if mock.lastReq.Method != http.MethodGet {
		t.Errorf("method = %s, want GET", mock.lastReq.Method)
	}
	if !mock.calledPath("/api/v2/users/usr_1/payment-methods") {
		t.Fatalf("path = %q", mock.lastPath())
	}

	methods := *resp.Data
	if len(methods) != 2 {
		t.Fatalf("got %d methods, want 2", len(methods))
	}
	if string(methods[0].Method) != "QR_PAYMENT" || string(methods[0].Status) != "APPROVED" ||
		methods[0].RejectReason != nil || methods[0].Action != nil {
		t.Errorf("methods[0] = %+v", methods[0])
	}
	if string(methods[1].Method) != "CARD_PAYMENT" || string(methods[1].Status) != "REJECTED" {
		t.Errorf("methods[1] = %+v", methods[1])
	}
	if methods[1].RejectReason == nil || *methods[1].RejectReason != "Document expired" ||
		methods[1].Action == nil || *methods[1].Action != "contact_support" {
		t.Errorf("methods[1] optional fields = %+v", methods[1])
	}
}

func TestGetPaymentMethods_APIError(t *testing.T) {
	t.Parallel()

	mock := &mockHTTPClient{
		response: mockResponse(http.StatusNotFound, `{"message":"user not found"}`), //nolint:bodyclose // response body closed by client
	}
	client := newMockClient(t, mock)

	_, err := client.GetPaymentMethods(context.Background(), &gaian.GetPaymentMethodsRequest{UserID: "usr_1"})
	if !gaian.IsNotFound(err) {
		t.Fatalf("error = %v, want a 404 APIError", err)
	}
}

func TestPaymentMethodConstantsMatchService(t *testing.T) {
	t.Parallel()

	if gaian.PaymentMethodQR != "QR_PAYMENT" || gaian.PaymentMethodCard != "CARD_PAYMENT" {
		t.Errorf("QR=%q Card=%q", gaian.PaymentMethodQR, gaian.PaymentMethodCard)
	}
}
