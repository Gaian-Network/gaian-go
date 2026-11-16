package gaian

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// PaymentMethodType identifies a way a user can pay.
type PaymentMethodType string

const (
	PaymentMethodQR   PaymentMethodType = "QR_PAYMENT"
	PaymentMethodCard PaymentMethodType = "CARD_PAYMENT"
)

// GetPaymentMethodsRequest identifies the user whose payment methods to
// list.
type GetPaymentMethodsRequest struct {
	UserID string `param:"userId"`
}

// PaymentMethod is the availability of one payment method for a user,
// derived from the user's KYC status. RejectReason and Action are only
// set when Status is MarketRejected.
type PaymentMethod struct {
	Method       PaymentMethodType `json:"method"`
	Name         string            `json:"name"`
	Status       MarketStatus      `json:"status"`
	RejectReason *string           `json:"rejectReason,omitempty"`
	Action       *string           `json:"action,omitempty"`
}

// GetPaymentMethodsResponse is the list of payment methods for a user.
type GetPaymentMethodsResponse = []PaymentMethod

// GetPaymentMethods returns every payment method and whether req.UserID
// may use it.
//
// Example:
//
//	resp, err := client.GetPaymentMethods(ctx, &gaian.GetPaymentMethodsRequest{UserID: userID})
//	if err != nil {
//		return err
//	}
//	for _, m := range *resp.Data {
//		fmt.Println(m.Method, m.Status)
//	}
func (c *Client) GetPaymentMethods(ctx context.Context, req *GetPaymentMethodsRequest) (*UserResponse[GetPaymentMethodsResponse], error) {
	endpoint := fmt.Sprintf("/api/v2/users/%s/payment-methods", url.PathEscape(req.UserID))

	apiRequest := request{
		Method:   http.MethodGet,
		Endpoint: endpoint,
	}

	rawResponse, err := c.excute(ctx, &apiRequest)
	if err != nil {
		return nil, err
	}

	response := new(UserResponse[GetPaymentMethodsResponse])
	if err := json.Unmarshal(rawResponse, response); err != nil {
		return nil, err
	}

	return response, nil
}
