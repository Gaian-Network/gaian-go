package gaian

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// GetMarketRequirementsRequest identifies the user whose KYC
// requirements to list.
type GetMarketRequirementsRequest struct {
	UserID string `param:"userId"`
}

// GetMarketRequirementsResponse maps ISO-3166-1 alpha-2 country codes to
// the KYC fields that market needs.
type GetMarketRequirementsResponse struct {
	Markets map[string]MarketRequirements `json:"markets"`
}

// MarketRequirements lists the KYC fields one market requires, by their
// JSON name in SubmitKYCRequest (for example "nationalId", "kycReport").
// MissingFields is the subset the user has not provided yet; it is empty
// when the user has everything the market needs.
type MarketRequirements struct {
	RequiredFields []string `json:"requiredFields"`
	MissingFields  []string `json:"missingFields"`
}

// GetMarketRequirements returns, per market, the KYC fields required and
// the ones req.UserID is still missing. Use it to decide what to collect
// before SubmitKYC or UpdateKYC.
//
// Example:
//
//	resp, err := client.GetMarketRequirements(ctx, &gaian.GetMarketRequirementsRequest{UserID: userID})
//	if err != nil {
//		return err
//	}
//	fmt.Println(resp.Data.Markets["VN"].MissingFields)
func (c *Client) GetMarketRequirements(ctx context.Context, req *GetMarketRequirementsRequest) (*UserResponse[GetMarketRequirementsResponse], error) {
	endpoint := fmt.Sprintf("/api/v2/users/%s/markets/requirements", url.PathEscape(req.UserID))

	apiRequest := request{
		Method:   http.MethodGet,
		Endpoint: endpoint,
	}

	rawResponse, err := c.excute(ctx, &apiRequest)
	if err != nil {
		return nil, err
	}

	response := new(UserResponse[GetMarketRequirementsResponse])
	if err := json.Unmarshal(rawResponse, response); err != nil {
		return nil, err
	}

	return response, nil
}
