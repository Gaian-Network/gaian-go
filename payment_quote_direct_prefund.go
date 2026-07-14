package gaian

import (
	"context"
	"encoding/json"
	"net/http"
)

// QuoteDirectPrefundRequest is QuoteDirectRequest without ChainID, for
// the prefunded direct-bank-transfer flow (settles from the tenant's
// prefunded balance, no on-chain step). Production only.
type QuoteDirectPrefundRequest struct {
	AccountNumber      string  `json:"accountNumber"`
	Code               string  `json:"code"`
	Amount             float64 `json:"amount"`
	Country            string  `json:"country"`
	SettlementCurrency string  `json:"settlementCurrency"`
	WalletAddress      string  `json:"walletAddress,omitempty"`
	UserID             string  `json:"userId,omitempty"`
	Email              string  `json:"email,omitempty"`
	BeneficiaryName    string  `json:"beneficiaryName,omitempty"`
}

// QuoteDirectPrefundResponse is the same shape as QuoteDirectResponse
// (ChainID will be its zero value).
type QuoteDirectPrefundResponse = QuoteDirectResponse

// QuoteDirectPrefund creates a single-use prefunded direct-bank-transfer
// quote. Production only — not available in sandbox. Requires the
// tenant to be on the prefund billing model. The resulting QuoteID must
// be consumed by PlacePrefundedOrder (not PlaceOrder).
//
// Recommended sequence: ListBanks → VerifyAccount → QuoteDirectPrefund →
// PlacePrefundedOrder → GetOrderStatus (poll). No VerifyOrder step.
//
// Example:
//
//	resp, err := client.QuoteDirectPrefund(ctx, &gaian.QuoteDirectPrefundRequest{
//		AccountNumber:      "0123456789",
//		Code:               bankCode, // from ListBanks
//		Amount:             500000,
//		Country:            "VN",
//		SettlementCurrency: "USDC",
//		UserID:             userID,
//	})
func (c *Client) QuoteDirectPrefund(ctx context.Context, req *QuoteDirectPrefundRequest) (*PaymentResponse[QuoteDirectPrefundResponse], error) {
	apiRequest := request{
		Method:   http.MethodPost,
		Endpoint: "/api/v2/quote/direct/prefund",
		Params:   req,
	}

	rawResponse, err := c.excute(ctx, &apiRequest)
	if err != nil {
		return nil, err
	}

	response := new(PaymentResponse[QuoteDirectPrefundResponse])
	if err := json.Unmarshal(rawResponse, response); err != nil {
		return nil, err
	}

	return response, nil
}
