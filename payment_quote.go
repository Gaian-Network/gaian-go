package gaian

import (
	"context"
	"encoding/json"
	"net/http"
)

// ChainID identifies the settlement chain for on-chain quotes/orders.
type ChainID int

const (
	ChainEthereum ChainID = 1
	ChainOptimism ChainID = 10
	ChainBSC      ChainID = 56
	ChainPolygon  ChainID = 137
	ChainBase     ChainID = 8453
	ChainArbitrum ChainID = 42161
	ChainSolana   ChainID = 101
)

// Route identifies the payment route chosen for a quote.
type Route struct {
	RouteID int `json:"routeId"`
}

// QuoteRequest carries the parameters for a QR-based on-chain quote. Use
// QuotePrefund instead if the tenant settles from a prefunded balance
// (no on-chain step), or QuoteDirect/QuoteDirectPrefund for a
// bank-account-only flow with no QR code at all.
type QuoteRequest struct {
	QRString           string  `json:"qrString"`
	Amount             float64 `json:"amount"`
	Country            string  `json:"country,omitempty"`
	ChainID            ChainID `json:"chainId"`
	SettlementCurrency string  `json:"settlementCurrency"`
	UserID             string  `json:"userId"`
}

// QuoteResponse is the result of a quote request. ChainID is the zero
// value on prefund quotes (no on-chain step involved).
type QuoteResponse struct {
	QuoteID            string  `json:"quoteId"`
	Route              Route   `json:"route"`
	FiatAmount         float64 `json:"fiatAmount"`
	FiatCurrency       string  `json:"fiatCurrency"`
	SettlementAmount   string  `json:"settlementAmount"`
	SettlementCurrency string  `json:"settlementCurrency"`
	ChainID            ChainID `json:"chainId,omitempty"`
	ExchangeRate       string  `json:"exchangeRate"`
	ProtocolFeeUSD     string  `json:"protocolFeeUsd"`
	MinimumFeeUSD      string  `json:"minimumFeeUsd"`
	QRInfo             QRInfo  `json:"qrInfo"`
	ExpiresAt          string  `json:"expiresAt"`
}

// Quote creates a single-use quote (~120s TTL) for a QR-based on-chain
// payment, locking in an exchange rate and route. Consume it with
// PlaceOrder before it expires; an expired or already-consumed QuoteID
// fails PlaceOrder with a 409.
//
// This is step 1 of the standard payment flow:
//
//	Quote → PlaceOrder → (sign & broadcast the returned transaction) → VerifyOrder → GetOrderStatus (poll)
//
// Example:
//
//	resp, err := client.Quote(ctx, &gaian.QuoteRequest{
//		QRString:           qrString,
//		Amount:             500000,
//		ChainID:            gaian.ChainSolana,
//		SettlementCurrency: "USDC",
//		UserID:             userID,
//	})
//	if err != nil {
//		return err
//	}
//	quoteID := resp.Data.QuoteID
func (c *Client) Quote(ctx context.Context, req *QuoteRequest) (*PaymentResponse[QuoteResponse], error) {
	apiRequest := request{
		Method:   http.MethodPost,
		Endpoint: "/api/v2/quote",
		Params:   req,
	}

	rawResponse, err := c.excute(ctx, &apiRequest)
	if err != nil {
		return nil, err
	}

	response := new(PaymentResponse[QuoteResponse])
	if err := json.Unmarshal(rawResponse, response); err != nil {
		return nil, err
	}

	return response, nil
}
