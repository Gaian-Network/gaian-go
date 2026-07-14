package gaian

import (
	"context"
	"encoding/json"
	"net/http"
)

// ListBanksRequest specifies the market to list direct-transfer banks
// for. Country is an ISO-3166-1 alpha-2 code (e.g. "VN").
type ListBanksRequest struct {
	Country string `json:"country"`
}

// Bank is one bank supporting direct transfer.
type Bank struct {
	ID   int    `json:"id"`   // display row number only — not a stable identifier
	Code string `json:"code"` // transfer BIN — pass this to QuoteDirect/QuoteDirectPrefund/VerifyAccount
	Name string `json:"name"`
}

// ListBanksResponse is the set of banks available for direct transfer in
// the requested country.
type ListBanksResponse struct {
	Country string `json:"country"`
	Banks   []Bank `json:"banks"`
}

// ListBanks returns the banks supporting direct transfer in
// req.Country. Production only — not available in sandbox. Call this
// first in the direct-bank-transfer flow to get a bank's Code, which
// VerifyAccount and QuoteDirect/QuoteDirectPrefund both need.
//
// Example:
//
//	resp, err := client.ListBanks(ctx, &gaian.ListBanksRequest{Country: "VN"})
//	if err != nil {
//		return err
//	}
//	for _, b := range resp.Data.Banks {
//		fmt.Println(b.Code, b.Name)
//	}
func (c *Client) ListBanks(ctx context.Context, req *ListBanksRequest) (*PaymentResponse[ListBanksResponse], error) {
	apiRequest := request{
		Method:   http.MethodGet,
		Endpoint: "/api/v2/banks",
		Params:   req,
	}

	rawResponse, err := c.excute(ctx, &apiRequest)
	if err != nil {
		return nil, err
	}

	response := new(PaymentResponse[ListBanksResponse])
	if err := json.Unmarshal(rawResponse, response); err != nil {
		return nil, err
	}

	return response, nil
}
