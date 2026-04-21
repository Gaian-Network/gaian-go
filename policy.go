package gaian

import (
	"context"
	"net/url"
)

// CheckPolicy reports whether a wallet is allowed to make a payment in the
// given country and returns transaction limits and the KYC tier.
// countryCode is an ISO-3166-1 alpha-2 code (e.g. "VN", "PH").
func (s *PolicyService) CheckPolicy(ctx context.Context, walletAddress, countryCode string) (*PolicyResult, error) {
	q := url.Values{
		"walletAddress": {walletAddress},
		"countryCode":   {countryCode},
	}
	var resp PolicyResult
	if err := s.c.get(ctx, s.c.pmtBaseURL, "/api/v1/checkPolicy", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
