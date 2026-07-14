package gaian

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// MarketStatus is the per-market KYC/payment gating status. Values are
// uppercase per the API's enum — don't lowercase them when comparing.
type MarketStatus string

const (
	MarketApproved   MarketStatus = "APPROVED"
	MarketPending    MarketStatus = "PENDING"
	MarketNotStarted MarketStatus = "NOT_STARTED"
	MarketRejected   MarketStatus = "REJECTED"
)

// GetMarketsRequest identifies the user whose market access to check.
type GetMarketsRequest struct {
	UserID string `param:"userId"`
}

// GetMarketsResponse maps ISO-3166-1 alpha-2 country codes (e.g. "VN",
// "PH", "BR") to that market's gating status for the user.
type GetMarketsResponse struct {
	Markets map[string]MarketStatusResult `json:"markets"`
}

// MarketStatusResult describes payment/KYC gating for one market.
// RejectReason is only populated when Status is MarketRejected, and only
// for some markets (PH/BR) — a rejected VN verification instead reverts
// to MarketNotStarted with no reason attached.
type MarketStatusResult struct {
	Status       MarketStatus `json:"status"`
	RejectReason *string      `json:"rejectReason,omitempty"`
	Action       *string      `json:"action,omitempty"`
}

// GetMarkets returns the payment/KYC gating status per market for
// req.UserID. Check this (or the equivalent field on GetUserByID) before
// letting a user attempt a payment in a given country.
//
// Example:
//
//	resp, err := client.GetMarkets(ctx, &gaian.GetMarketsRequest{UserID: userID})
//	if err != nil {
//		return err
//	}
//	if resp.Data.Markets["VN"].Status != gaian.MarketApproved {
//		return fmt.Errorf("VN market not approved for this user")
//	}
func (c *Client) GetMarkets(ctx context.Context, req *GetMarketsRequest) (*UserResposne[GetMarketsResponse], error) {
	endpoint := fmt.Sprintf("/api/v2/users/%s/markets", url.PathEscape(req.UserID))

	apiRequest := request{
		Method:   http.MethodGet,
		Endpoint: endpoint,
	}

	rawResponse, err := c.excute(ctx, &apiRequest)
	if err != nil {
		return nil, err
	}

	response := new(UserResposne[GetMarketsResponse])
	if err := json.Unmarshal(rawResponse, response); err != nil {
		return nil, err
	}

	return response, nil
}
