package gaian

import (
	"context"
	"encoding/json"
	"net/http"
)

// GetOrganizationBalanceRequest has no parameters — balance is looked up
// for whichever tenant the API key belongs to.
type GetOrganizationBalanceRequest struct{}

// GetOrganizationBalanceResponse is the tenant's prefund balance. All
// amounts are decimal strings (not float64) to avoid precision loss.
type GetOrganizationBalanceResponse struct {
	Currency          string `json:"currency"`
	AvailableBalance  string `json:"availableBalance"`
	TotalSpent        string `json:"totalSpent"`
	PendingSettlement string `json:"pendingSettlement"`
}

// GetOrganizationBalance returns the prefund balance for the tenant that
// owns the API key used to sign the request. It's restricted to tenants
// on the prefund billing model — expect a 400 error if the tenant isn't
// prefund-enabled (see PlacePrefundedOrder/QuotePrefund for the flow this
// balance is drawn from).
//
// Note: unlike the other payments-group endpoints in this SDK
// (Quote/PlaceOrder/VerifyOrder/GetOrderStatus/ParseQR), this one returns
// the User-service-style envelope (UserResposne), not PaymentResponse.
//
// Example:
//
//	resp, err := client.GetOrganizationBalance(ctx, &gaian.GetOrganizationBalanceRequest{})
//	if err != nil {
//		return err
//	}
//	fmt.Printf("%s %s available\n", resp.Data.AvailableBalance, resp.Data.Currency)
func (c *Client) GetOrganizationBalance(ctx context.Context, req *GetOrganizationBalanceRequest) (*UserResposne[GetOrganizationBalanceResponse], error) {
	apiRequest := request{
		Method:   http.MethodGet,
		Endpoint: "/api/v2/organization/me/balance",
	}

	rawResponse, err := c.excute(ctx, &apiRequest)
	if err != nil {
		return nil, err
	}

	response := new(UserResposne[GetOrganizationBalanceResponse])
	if err := json.Unmarshal(rawResponse, response); err != nil {
		return nil, err
	}

	return response, nil
}
