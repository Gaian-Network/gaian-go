package gaian

import "context"

type registerRequest struct {
	WalletAddress string `json:"walletAddress"`
}

type registerEnvelope struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	User    User   `json:"user"`
}

// Register creates a new user for the given wallet address.
// Idempotent — returns the existing user if the wallet is already registered.
func (s *UserService) Register(ctx context.Context, walletAddress string) (*User, error) {
	var resp registerEnvelope
	if err := s.c.post(ctx, s.c.userBaseURL, "/api/v1/user/register",
		registerRequest{WalletAddress: walletAddress}, &resp); err != nil {
		return nil, err
	}
	return &resp.User, nil
}
