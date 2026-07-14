package gaian

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// VerifyStatus is the outcome of VerifyOrder. Its integer values overlap
// with OrderStatus (payment_get_order_status.go) but mean something
// different — never compare a VerifyStatus to an OrderStatus constant.
type VerifyStatus int

const (
	VerifyPending  VerifyStatus = 0
	VerifyVerified VerifyStatus = 1
	VerifyFailed   VerifyStatus = 2
)

// VerifyOrderRequest submits the on-chain transaction hash for OrderID
// (the order returned by PlaceOrder) so the gateway can confirm the
// deposit and trigger settlement.
type VerifyOrderRequest struct {
	OrderID         string `param:"orderId" json:"-"`
	TransactionHash string `json:"transactionHash"`
}

// VerifyOrderResponse is the outcome of verifying an order's on-chain
// transaction.
//
// IMPORTANT: the API returns HTTP 200 even when verification did not
// succeed — a nil error from VerifyOrder does NOT mean the payment is
// confirmed. Callers MUST branch on Status:
//   - VerifyPending: the transaction hasn't propagated on-chain yet.
//     Wait a few seconds and resubmit the same TransactionHash.
//   - VerifyVerified: confirmed — poll GetOrderStatus for settlement.
//   - VerifyFailed: verification failed outright; see Message.
type VerifyOrderResponse struct {
	OrderID         string       `json:"orderId"`
	Status          VerifyStatus `json:"status"`
	StatusLabel     string       `json:"statusLabel"`
	TransactionHash string       `json:"transactionHash"`
	Message         string       `json:"message"`
}

// VerifyOrder submits the on-chain transaction hash for req.OrderID so
// the gateway can confirm the deposit and trigger settlement.
//
// This is step 3 of the standard payment flow:
//
//	Quote → PlaceOrder → (sign & broadcast) → VerifyOrder → GetOrderStatus (poll)
//
// See the VerifyOrderResponse doc comment for the "200 on failure"
// gotcha before wiring up error handling around this call.
//
// Example:
//
//	resp, err := client.VerifyOrder(ctx, &gaian.VerifyOrderRequest{
//		OrderID:         orderID,
//		TransactionHash: txHash,
//	})
//	if err != nil {
//		return err
//	}
//	switch resp.Data.Status {
//	case gaian.VerifyPending:
//		// resubmit the same TransactionHash after a short delay
//	case gaian.VerifyFailed:
//		return fmt.Errorf("verification failed: %s", resp.Data.Message)
//	}
func (c *Client) VerifyOrder(ctx context.Context, req *VerifyOrderRequest) (*PaymentResponse[VerifyOrderResponse], error) {
	endpoint := fmt.Sprintf("/api/v2/orders/%s/verify", url.PathEscape(req.OrderID))

	apiRequest := request{
		Method:   http.MethodPost,
		Endpoint: endpoint,
		Params:   req,
	}

	rawResponse, err := c.excute(ctx, &apiRequest)
	if err != nil {
		return nil, err
	}

	response := new(PaymentResponse[VerifyOrderResponse])
	if err := json.Unmarshal(rawResponse, response); err != nil {
		return nil, err
	}

	return response, nil
}
