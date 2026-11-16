package gaian

import (
	"context"
	"encoding/json"
	"net/http"
)

// GetUserLimitsRequest selects the user whose limits to read. UserID is
// required. Country is an optional ISO-3166-1 alpha-2 filter; leave it
// empty to get every configured country.
type GetUserLimitsRequest struct {
	UserID  string `json:"userId"`
	Country string `json:"country,omitempty"`
}

// GetUserLimitsResponse is a user's remaining payment allowance. Every
// USD figure is a JSON number on the wire, so it is a float64 here: use
// it for display and pre-checks, not for accounting.
type GetUserLimitsResponse struct {
	UserID    *string            `json:"userId"`
	Daily     UserDailySpend     `json:"daily"`
	Velocity  UserVelocityLimit  `json:"velocity"`
	Countries []UserCountryLimit `json:"countries"`
}

// UserDailySpend is the user's spend over the rolling 24-hour window,
// summed across all countries.
type UserDailySpend struct {
	SpentUSD    float64 `json:"spentUsd"`
	WindowStart string  `json:"windowStart"`
	WindowEnd   string  `json:"windowEnd"`
}

// UserVelocityLimit is the cap on how many transactions a user may start
// per window, independent of amounts. When Enabled is false the cap is
// off. RetryAfterSeconds is 0 while a slot is free.
type UserVelocityLimit struct {
	Enabled           bool `json:"enabled"`
	MaxPerWindow      int  `json:"maxPerWindow"`
	WindowSeconds     int  `json:"windowSeconds"`
	UsedInWindow      int  `json:"usedInWindow"`
	Remaining         int  `json:"remaining"`
	RetryAfterSeconds int  `json:"retryAfterSeconds"`
}

// UserCountryLimit is the user's allowance in one country.
// MaxNextTransactionUSD is the largest order the user can place right
// now (0 when the allowance is used up). MaxPerDayUSD, DailyRemainingUSD
// and Weekly are nil when that limit does not apply.
type UserCountryLimit struct {
	Country               string           `json:"country"`
	Supported             bool             `json:"supported"`
	KYCVerified           bool             `json:"kycVerified"`
	CanTransact           bool             `json:"canTransact"`
	MaxPerDayUSD          *float64         `json:"maxPerDayUsd"`
	DailyRemainingUSD     *float64         `json:"dailyRemainingUsd"`
	Weekly                *UserWeeklyLimit `json:"weekly"`
	MaxNextTransactionUSD float64          `json:"maxNextTransactionUsd"`
}

// UserWeeklyLimit is the per-user weekly cap in one country. The window
// runs Monday 00:00 UTC to the next Monday 00:00 UTC.
type UserWeeklyLimit struct {
	CapUSD       float64 `json:"capUsd"`
	SpentUSD     float64 `json:"spentUsd"`
	RemainingUSD float64 `json:"remainingUsd"`
	WindowStart  string  `json:"windowStart"`
	WindowEnd    string  `json:"windowEnd"`
}

// GetUserLimits returns how much req.UserID can still pay: the rolling
// daily spend, the transaction-count limit, and the per-country caps.
//
// Example:
//
//	resp, err := client.GetUserLimits(ctx, &gaian.GetUserLimitsRequest{UserID: userID, Country: "VN"})
//	if err != nil {
//		return err
//	}
//	for _, c := range resp.Data.Countries {
//		fmt.Println(c.Country, c.CanTransact, c.MaxNextTransactionUSD)
//	}
func (c *Client) GetUserLimits(ctx context.Context, req *GetUserLimitsRequest) (*PaymentResponse[GetUserLimitsResponse], error) {
	apiRequest := request{
		Method:   http.MethodGet,
		Endpoint: "/api/v2/limits/user",
		Params:   req,
	}

	rawResponse, err := c.excute(ctx, &apiRequest)
	if err != nil {
		return nil, err
	}

	response := new(PaymentResponse[GetUserLimitsResponse])
	if err := json.Unmarshal(rawResponse, response); err != nil {
		return nil, err
	}

	return response, nil
}
