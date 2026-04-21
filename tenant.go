package gaian

import (
	"context"
	"net/url"
)

// GetBalance returns the prefunded balance for the authenticated tenant.
// currency defaults to USDC when empty.
func (s *TenantService) GetBalance(ctx context.Context, currency CryptoCurrency) (*TenantBalance, error) {
	var resp TenantBalance
	q := url.Values{}
	if currency != "" {
		q.Set("currency", string(currency))
	}
	if err := s.c.get(ctx, s.c.userBaseURL, "/api/v1/tenant/balance", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetTotalSpent returns the total amount spent and the spend cap for the tenant.
func (s *TenantService) GetTotalSpent(ctx context.Context) (*TenantSpend, error) {
	var resp TenantSpend
	if err := s.c.get(ctx, s.c.userBaseURL, "/api/v1/tenant/total-spent", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
