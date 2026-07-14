package gaian

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

// GetUserByWalletRequest identifies a user by one of their linked wallet
// addresses.
type GetUserByWalletRequest struct {
	WalletAddress string `param:"walletAddress"`
}

// GetUserByWalletResponse is the full user record.
type GetUserByWalletResponse = User

// GetUserByWallet retrieves a user by req.WalletAddress. Use GetUserByID
// if you already have the user's ID — it's a more direct lookup.
//
// Example:
//
//	resp, err := client.GetUserByWallet(ctx, &gaian.GetUserByWalletRequest{
//		WalletAddress: walletAddress,
//	})
func (c *Client) GetUserByWallet(ctx context.Context, req *GetUserByWalletRequest) (*UserResposne[GetUserByWalletResponse], error) {
	apiRequest := request{
		Method:   http.MethodGet,
		Endpoint: "/api/v2/users/wallet/" + url.PathEscape(req.WalletAddress),
	}

	rawResponse, err := c.excute(ctx, &apiRequest)
	if err != nil {
		return nil, err
	}

	response := new(UserResposne[GetUserByWalletResponse])
	if err := json.Unmarshal(rawResponse, response); err != nil {
		return nil, err
	}
	return response, nil
}
