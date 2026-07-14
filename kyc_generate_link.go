package gaian

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// GenerateKYCLinkRequest identifies the user to start a hosted KYC
// session for. CallbackURL is optional — the Didit flow redirects there
// on completion; leave it empty to skip the redirect.
type GenerateKYCLinkRequest struct {
	UserID      string `param:"userId" json:"-"`
	CallbackURL string `json:"callbackUrl,omitempty"`
}

// GenerateKYCLinkResponse carries the hosted session URL to redirect the
// user to.
type GenerateKYCLinkResponse struct {
	URL string `json:"url"`
}

// GenerateKYCLink creates a hosted Didit KYC session URL for req.UserID.
// CallbackURL is optional. The URL is time-limited — use it immediately,
// do not cache it. Use SubmitKYC instead if you'd rather collect and
// submit KYC information directly, without the hosted redirect flow.
//
// Example:
//
//	resp, err := client.GenerateKYCLink(ctx, &gaian.GenerateKYCLinkRequest{
//		UserID: userID,
//	})
//	if err != nil {
//		return err
//	}
//	redirectTo(resp.Data.URL)
func (c *Client) GenerateKYCLink(ctx context.Context, req *GenerateKYCLinkRequest) (*UserResposne[GenerateKYCLinkResponse], error) {
	endpoint := fmt.Sprintf("/api/v2/users/%s/kyc-url", url.PathEscape(req.UserID))

	apiRequest := request{
		Method:   http.MethodPost,
		Endpoint: endpoint,
		Params:   req,
	}

	rawResponse, err := c.excute(ctx, &apiRequest)
	if err != nil {
		return nil, err
	}

	response := new(UserResposne[GenerateKYCLinkResponse])
	if err := json.Unmarshal(rawResponse, response); err != nil {
		return nil, err
	}

	return response, nil
}
