package gaian

import (
	"context"
	"encoding/json"
	"net/http"
)

// PlaceOrderRequest consumes a quote from Quote or QuoteDirect. It fails
// with a 409 if the quote is already consumed, expired, or was created
// for the prefunded flow (use PlacePrefundedOrder for a QuotePrefund/
// QuoteDirectPrefund QuoteID instead).
type PlaceOrderRequest struct {
	QuoteID string `json:"quoteId"`
}

// PlaceOrderResponse carries the on-chain transaction the caller must
// sign and broadcast to fund the order. SettlementAmount is a decimal
// string in token units; SettlementAmountBaseUnits is the same amount as
// an integer string in the token's smallest unit.
type PlaceOrderResponse struct {
	OrderID        string  `json:"orderId"`
	Status         int     `json:"status"`
	StatusLabel    string  `json:"statusLabel"`
	ChainID        ChainID `json:"chainId"`
	DepositAddress string  `json:"depositAddress"`
	// TokenAddress is the contract (or mint) address of the settlement
	// token on ChainID.
	TokenAddress              string `json:"tokenAddress"`
	SettlementAmount          string `json:"settlementAmount"`
	SettlementAmountBaseUnits string `json:"settlementAmountBaseUnits"`
	SettlementCurrency        string `json:"settlementCurrency"`
	EncodedTransaction        string `json:"encodedTransaction"`
	ExpiresAt                 string `json:"expiresAt"`
	// Memo must be attached to the deposit on chains that need one (for
	// example Stellar). It is empty on every other chain.
	Memo string `json:"memo,omitempty"`
}

// PlaceOrder consumes a quote (from Quote or QuoteDirect) and creates an
// order. After this call, sign and broadcast EncodedTransaction
// on-chain yourself, then call VerifyOrder with the resulting
// transaction hash.
//
// This is step 2 of the standard payment flow:
//
//	Quote → PlaceOrder → (sign & broadcast) → VerifyOrder → GetOrderStatus (poll)
//
// Example:
//
//	resp, err := client.PlaceOrder(ctx, &gaian.PlaceOrderRequest{QuoteID: quoteID})
//	if err != nil {
//		return err
//	}
//	// sign and broadcast resp.Data.EncodedTransaction on-chain, then:
//	txHash := broadcastTransaction(resp.Data.EncodedTransaction)
func (c *Client) PlaceOrder(ctx context.Context, req *PlaceOrderRequest) (*PaymentResponse[PlaceOrderResponse], error) {
	apiRequest := request{
		Method:   http.MethodPost,
		Endpoint: "/api/v2/orders",
		Params:   req,
	}

	rawResponse, err := c.excute(ctx, &apiRequest)
	if err != nil {
		return nil, err
	}

	response := new(PaymentResponse[PlaceOrderResponse])
	if err := json.Unmarshal(rawResponse, response); err != nil {
		return nil, err
	}

	return response, nil
}
