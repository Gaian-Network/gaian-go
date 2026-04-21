package gaian

import (
	"context"
	"net/url"
)

type marketsData struct {
	Markets map[string]MarketInfo `json:"markets"`
}

// GetMarkets returns the payment market capabilities for the given wallet.
// The map key is the ISO-3166-1 alpha-2 country code (e.g. "VN", "PH").
func (s *UserService) GetMarkets(ctx context.Context, walletAddress string) (map[string]MarketInfo, error) {
	var resp envelope[marketsData]
	q := url.Values{"walletAddress": {walletAddress}}
	if err := s.c.get(ctx, s.c.userBaseURL, "/api/v1/users/markets", q, &resp); err != nil {
		return nil, err
	}
	if resp.Data == nil {
		return nil, nil
	}
	return resp.Data.Markets, nil
}
