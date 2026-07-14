package gaian

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// CreateWalletRequest links a new wallet to a user. Chain is optional —
// omit it if the wallet isn't chain-specific.
type CreateWalletRequest struct {
	UserID  string `param:"userId" json:"-"`
	Address string `json:"address"`
	Chain   string `json:"chain,omitempty"`
}

// CreateWalletResponse is the wallet that was just linked.
//
// Note: per the current openapi spec this response is a bare object, not
// wrapped in the usual envelope like every sibling user-service
// endpoint — worth re-verifying against a live v2 environment once
// available. If that holds, this method may need to skip the
// UserResposne wrapper and unmarshal CreateWalletResponse directly.
type CreateWalletResponse = WalletInfo

// CreateWallet links a new wallet to req.UserID. The first wallet added
// for a user becomes primary automatically; use ListWallets to see all
// linked wallets and which one is primary.
//
// Example:
//
//	resp, err := client.CreateWallet(ctx, &gaian.CreateWalletRequest{
//		UserID:  userID,
//		Address: "0xWalletAddress",
//		Chain:   "base",
//	})
func (c *Client) CreateWallet(ctx context.Context, req *CreateWalletRequest) (*UserResposne[CreateWalletResponse], error) {
	endpoint := fmt.Sprintf("/api/v2/users/%s/wallets", url.PathEscape(req.UserID))

	apiRequest := request{
		Method:   http.MethodPost,
		Endpoint: endpoint,
		Params:   req,
	}

	rawResponse, err := c.excute(ctx, &apiRequest)
	if err != nil {
		return nil, err
	}

	response := new(UserResposne[CreateWalletResponse])
	if err := json.Unmarshal(rawResponse, response); err != nil {
		return nil, err
	}

	return response, nil
}
