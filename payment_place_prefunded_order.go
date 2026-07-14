package gaian

import (
	"context"
	"encoding/json"
	"net/http"
)

// PlacePrefundedOrderRequest consumes a quote from QuotePrefund or
// QuoteDirectPrefund. It fails with a 409 if the quote is already
// consumed, expired, or was created for the on-chain flow (use
// PlaceOrder for a Quote/QuoteDirect QuoteID instead).
type PlacePrefundedOrderRequest struct {
	QuoteID string `json:"quoteId"`
}

// PlacePrefundedOrderResponse has no on-chain fields (no
// DepositAddress/EncodedTransaction/ExpiresAt/ChainID), unlike
// PlaceOrderResponse — there's no on-chain step for a prefunded order.
type PlacePrefundedOrderResponse struct {
	OrderID     string `json:"orderId"`
	Status      int    `json:"status"`
	StatusLabel string `json:"statusLabel"`
}

// PlacePrefundedOrder consumes a prefund quote (from QuotePrefund or
// QuoteDirectPrefund) and creates an order settled immediately from the
// tenant's prefunded balance. No on-chain signing required — go straight
// to polling GetOrderStatus.
//
// This is step 2 of the prefunded payment flow:
//
//	QuotePrefund → PlacePrefundedOrder → GetOrderStatus (poll)
//
// Example:
//
//	resp, err := client.PlacePrefundedOrder(ctx, &gaian.PlacePrefundedOrderRequest{
//		QuoteID: quoteID,
//	})
func (c *Client) PlacePrefundedOrder(ctx context.Context, req *PlacePrefundedOrderRequest) (*PaymentResponse[PlacePrefundedOrderResponse], error) {
	apiRequest := request{
		Method:   http.MethodPost,
		Endpoint: "/api/v2/orders/prefund",
		Params:   req,
	}

	rawResponse, err := c.excute(ctx, &apiRequest)
	if err != nil {
		return nil, err
	}

	response := new(PaymentResponse[PlacePrefundedOrderResponse])
	if err := json.Unmarshal(rawResponse, response); err != nil {
		return nil, err
	}

	return response, nil
}
