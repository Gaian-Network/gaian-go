package gaian

import (
	"context"
	"fmt"
	"net/url"
)

// GetByID retrieves a user by their numeric database ID.
func (s *UserService) GetByID(ctx context.Context, id int) (*User, error) {
	var resp envelope[User]
	path := fmt.Sprintf("/api/v1/users/%d", id)
	if err := s.c.get(ctx, s.c.userBaseURL, path, nil, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// GetByWallet retrieves a user by their wallet address.
func (s *UserService) GetByWallet(ctx context.Context, walletAddress string) (*User, error) {
	var resp envelope[User]
	q := url.Values{"walletAddress": {walletAddress}}
	if err := s.c.get(ctx, s.c.userBaseURL, "/api/v1/users", q, &resp); err != nil {
		return nil, err
	}
	return resp.Data, nil
}
