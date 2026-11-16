package gaian

import (
	"context"
	"encoding/json"
	"net/http"
)

// ChainID identifies the settlement chain for on-chain quotes/orders.
type ChainID int

const (
	ChainEthereum  ChainID = 1
	ChainOptimism  ChainID = 10
	ChainBSC       ChainID = 56
	ChainPolygon   ChainID = 137
	ChainBase      ChainID = 8453
	ChainArbitrum  ChainID = 42161
	ChainSolana    ChainID = 101
	ChainStellar   ChainID = 148
	ChainAvalanche ChainID = 43114
	ChainArc       ChainID = 5042002
)

// Route identifies the payment route chosen for a quote.
type Route struct {
	RouteID int `json:"routeId"`
}

// QuoteRequest carries the parameters for a QR-based on-chain quote. Use
// QuotePrefund instead if the tenant settles from a prefunded balance
// (no on-chain step), or QuoteDirect/QuoteDirectPrefund for a
// bank-account-only flow with no QR code at all.
//
// Exactly one of WalletAddress, UserID, or Email must be set to identify
// the payer; the API answers 400 otherwise.
type QuoteRequest struct {
	QRString           string  `json:"qrString"`
	Amount             float64 `json:"amount"`
	Country            string  `json:"country,omitempty"`
	ChainID            ChainID `json:"chainId"`
	SettlementCurrency string  `json:"settlementCurrency"`
	WalletAddress      string  `json:"walletAddress,omitempty"`
	UserID             string  `json:"userId,omitempty"`
	Email              string  `json:"email,omitempty"`
	// Network is the phone-payment rail for Peru QR codes: "YAPE" or
	// "PLIN". Leave empty for every other country.
	Network string `json:"network,omitempty"`
	// FromAddress is the on-chain wallet the deposit must come from. Set
	// it to pay from a wallet that is not linked to the user; it does
	// not replace the identity field.
	FromAddress string `json:"fromAddress,omitempty"`
}

// QuoteResponse is the result of a quote request. ChainID is the zero
// value on prefund quotes (no on-chain step involved). SettlementAmount
// equals SettlementBaseUSD + ProtocolFeeUSD + MinimumFeeUSD +
// TenantFeeUSD; all five are decimal strings.
type QuoteResponse struct {
	QuoteID            string  `json:"quoteId"`
	Route              Route   `json:"route"`
	FiatAmount         float64 `json:"fiatAmount"`
	FiatCurrency       string  `json:"fiatCurrency"`
	SettlementAmount   string  `json:"settlementAmount"`
	SettlementBaseUSD  string  `json:"settlementBaseUsd"`
	SettlementCurrency string  `json:"settlementCurrency"`
	ChainID            ChainID `json:"chainId,omitempty"`
	ExchangeRate       string  `json:"exchangeRate"`
	ProtocolFeeUSD     string  `json:"protocolFeeUsd"`
	MinimumFeeUSD      string  `json:"minimumFeeUsd"`
	TenantFeeUSD       string  `json:"tenantFeeUsd"`
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
