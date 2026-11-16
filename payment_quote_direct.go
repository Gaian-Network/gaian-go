package gaian

import (
	"context"
	"encoding/json"
	"net/http"
)

// DirectQRInfo is the qrInfo shape returned by the direct-bank-transfer
// quote endpoints — distinct from QRInfo (no isValid/provider/amount,
// since there's no QR code involved).
type DirectQRInfo struct {
	AccountNumber   string `json:"accountNumber"`
	BeneficiaryName string `json:"beneficiaryName"`
	BankBin         string `json:"bankBin"`
	Country         string `json:"country"`
}

// QuoteDirectRequest carries the parameters for a direct-bank-transfer
// on-chain quote — no QR code, the beneficiary is identified purely by
// bank account (Code comes from ListBanks). Exactly one of
// WalletAddress, UserID, or Email must be set to identify the payer.
//
// Production only — this endpoint isn't available in sandbox.
type QuoteDirectRequest struct {
	AccountNumber      string  `json:"accountNumber"`
	Code               string  `json:"code"`
	Amount             float64 `json:"amount"`
	Country            string  `json:"country"`
	ChainID            ChainID `json:"chainId"`
	SettlementCurrency string  `json:"settlementCurrency"`
	WalletAddress      string  `json:"walletAddress,omitempty"`
	UserID             string  `json:"userId,omitempty"`
	Email              string  `json:"email,omitempty"`
	// BeneficiaryName is accepted but ignored by the API — the
	// bank-verified account name (see VerifyAccount) always takes
	// precedence.
	BeneficiaryName string `json:"beneficiaryName,omitempty"`
	// FromAddress is the on-chain wallet the deposit must come from. Set
	// it to pay from a wallet that is not linked to the user; it does
	// not replace the identity field.
	FromAddress string `json:"fromAddress,omitempty"`
}

// QuoteDirectResponse is the result of a direct-bank-transfer quote
// request. SettlementAmount equals SettlementBaseUSD + ProtocolFeeUSD +
// MinimumFeeUSD + TenantFeeUSD; all five are decimal strings.
type QuoteDirectResponse struct {
	QuoteID            string       `json:"quoteId"`
	Route              Route        `json:"route"`
	FiatAmount         float64      `json:"fiatAmount"`
	FiatCurrency       string       `json:"fiatCurrency"`
	SettlementAmount   string       `json:"settlementAmount"`
	SettlementBaseUSD  string       `json:"settlementBaseUsd"`
	SettlementCurrency string       `json:"settlementCurrency"`
	ChainID            ChainID      `json:"chainId,omitempty"`
	ExchangeRate       string       `json:"exchangeRate"`
	ProtocolFeeUSD     string       `json:"protocolFeeUsd"`
	MinimumFeeUSD      string       `json:"minimumFeeUsd"`
	TenantFeeUSD       string       `json:"tenantFeeUsd"`
	QRInfo             DirectQRInfo `json:"qrInfo"`
	ExpiresAt          string       `json:"expiresAt"`
}

// QuoteDirect creates a single-use direct-bank-transfer quote. Production
// only — not available in sandbox. The resulting QuoteID must be
// consumed by PlaceOrder (not PlacePrefundedOrder) — cross-use is
// rejected.
//
// Recommended sequence: ListBanks → VerifyAccount → QuoteDirect →
// PlaceOrder → VerifyOrder → GetOrderStatus (poll).
//
// Example:
//
//	resp, err := client.QuoteDirect(ctx, &gaian.QuoteDirectRequest{
//		AccountNumber:      "0123456789",
//		Code:               bankCode, // from ListBanks
//		Amount:             500000,
//		Country:            "VN",
//		ChainID:            gaian.ChainSolana,
//		SettlementCurrency: "USDC",
//		UserID:             userID,
//	})
func (c *Client) QuoteDirect(ctx context.Context, req *QuoteDirectRequest) (*PaymentResponse[QuoteDirectResponse], error) {
	apiRequest := request{
		Method:   http.MethodPost,
		Endpoint: "/api/v2/quote/direct",
		Params:   req,
	}

	rawResponse, err := c.excute(ctx, &apiRequest)
	if err != nil {
		return nil, err
	}

	response := new(PaymentResponse[QuoteDirectResponse])
	if err := json.Unmarshal(rawResponse, response); err != nil {
		return nil, err
	}

	return response, nil
}
