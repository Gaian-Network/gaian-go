package gaian

import (
	"context"
	"encoding/json"
	"net/http"
)

// QuotePrefundRequest is QuoteRequest without ChainID — the payment
// settles from the tenant's prefunded balance, so there's no on-chain
// step. Requires the tenant to be on the prefund billing model.
type QuotePrefundRequest struct {
	QRString           string  `json:"qrString"`
	Amount             float64 `json:"amount"`
	Country            string  `json:"country,omitempty"`
	SettlementCurrency string  `json:"settlementCurrency"`
	UserID             string  `json:"userId"`
}

// QuotePrefundResponse is the same shape as QuoteResponse (ChainID will
// be its zero value since no chain is involved).
type QuotePrefundResponse = QuoteResponse

// QuotePrefund creates a single-use prefunded quote. Consume it with
// PlacePrefundedOrder before it expires.
//
// This is step 1 of the prefunded payment flow:
//
//	QuotePrefund → PlacePrefundedOrder → GetOrderStatus (poll)
//
// No VerifyOrder step — there's no on-chain transaction to confirm.
//
// Example:
//
//	resp, err := client.QuotePrefund(ctx, &gaian.QuotePrefundRequest{
//		QRString:           qrString,
//		Amount:             500000,
//		SettlementCurrency: "USDC",
//		UserID:             userID,
//	})
func (c *Client) QuotePrefund(ctx context.Context, req *QuotePrefundRequest) (*PaymentResponse[QuotePrefundResponse], error) {
	apiRequest := request{
		Method:   http.MethodPost,
		Endpoint: "/api/v2/quote/prefund",
		Params:   req,
	}

	rawResponse, err := c.excute(ctx, &apiRequest)
	if err != nil {
		return nil, err
	}

	response := new(PaymentResponse[QuotePrefundResponse])
	if err := json.Unmarshal(rawResponse, response); err != nil {
		return nil, err
	}

	return response, nil
}
