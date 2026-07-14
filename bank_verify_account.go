package gaian

import (
	"context"
	"encoding/json"
	"net/http"
)

// VerifyAccountRequest identifies a beneficiary bank account to verify
// before creating a direct quote. Code is a bank BIN from ListBanks.
type VerifyAccountRequest struct {
	AccountNumber string `json:"accountNumber"`
	Code          string `json:"code"`
	Country       string `json:"country"`
}

// VerifyAccountResponse is the result of verifying a bank account.
//
// A failed verification (bad account number) is NOT an error: it comes
// back as HTTP 200 with Valid=false and AccountName=nil. Check Valid,
// not just the returned error, before proceeding to QuoteDirect.
type VerifyAccountResponse struct {
	Country       string  `json:"country"`
	AccountNumber string  `json:"accountNumber"`
	Code          string  `json:"code"`
	Valid         bool    `json:"valid"`
	AccountName   *string `json:"accountName"`
}

// VerifyAccount checks a beneficiary bank account before creating a
// direct quote. Production only — not available in sandbox. Recommended
// (not required) before QuoteDirect/QuoteDirectPrefund. Any
// BeneficiaryName passed later in a quote request is ignored — this
// call's AccountName always wins.
//
// Example:
//
//	resp, err := client.VerifyAccount(ctx, &gaian.VerifyAccountRequest{
//		AccountNumber: "0123456789",
//		Code:          bankCode, // from ListBanks
//		Country:       "VN",
//	})
//	if err != nil {
//		return err
//	}
//	if !resp.Data.Valid {
//		return fmt.Errorf("invalid bank account")
//	}
func (c *Client) VerifyAccount(ctx context.Context, req *VerifyAccountRequest) (*PaymentResponse[VerifyAccountResponse], error) {
	apiRequest := request{
		Method:   http.MethodPost,
		Endpoint: "/api/v2/accounts/verify",
		Params:   req,
	}

	rawResponse, err := c.excute(ctx, &apiRequest)
	if err != nil {
		return nil, err
	}

	response := new(PaymentResponse[VerifyAccountResponse])
	if err := json.Unmarshal(rawResponse, response); err != nil {
		return nil, err
	}

	return response, nil
}
