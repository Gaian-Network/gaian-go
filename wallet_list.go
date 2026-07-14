package gaian

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// ListWalletsRequest identifies the user whose linked wallets to list.
type ListWalletsRequest struct {
	UserID string `param:"userId"`
}

// WalletInfo describes one wallet linked to a user.
type WalletInfo struct {
	Address   string `json:"address"`
	Chain     string `json:"chain"`
	IsPrimary bool   `json:"isPrimary"`
}

// ListWalletsResponse is the list of wallets linked to a user. There's
// no pagination on this endpoint — it always returns the full list.
type ListWalletsResponse = []WalletInfo

// ListWallets returns all wallets linked to req.UserID. Use CreateWallet
// to link a new one — the first wallet added becomes primary
// automatically.
//
// Example:
//
//	resp, err := client.ListWallets(ctx, &gaian.ListWalletsRequest{UserID: userID})
//	if err != nil {
//		return err
//	}
//	for _, w := range *resp.Data {
//		fmt.Println(w.Address, w.Chain, w.IsPrimary)
//	}
func (c *Client) ListWallets(ctx context.Context, req *ListWalletsRequest) (*UserResposne[ListWalletsResponse], error) {
	endpoint := fmt.Sprintf("/api/v2/users/%s/wallets", url.PathEscape(req.UserID))

	apiRequest := request{
		Method:   http.MethodGet,
		Endpoint: endpoint,
	}

	rawResponse, err := c.excute(ctx, &apiRequest)
	if err != nil {
		return nil, err
	}

	response := new(UserResposne[ListWalletsResponse])
	if err := json.Unmarshal(rawResponse, response); err != nil {
		return nil, err
	}

	return response, nil
}
