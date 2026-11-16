package gaian

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// ImportSharedKYCRequest carries a Didit share token issued by a partner
// application for a user who already passed KYC there.
type ImportSharedKYCRequest struct {
	UserID     string `param:"userId" json:"-"`
	ShareToken string `json:"shareToken"`
}

// ImportSharedKYCResponse identifies the imported verification session.
// Status is the session status as the KYC provider reports it (for
// example "Approved" or "In Review"), not a Gaian KYC status.
type ImportSharedKYCResponse struct {
	SessionID string `json:"sessionId"`
	Status    string `json:"status"`
}

// ImportSharedKYC imports a KYC session shared by a partner application
// (reusable KYC), so req.UserID does not have to verify again. It returns
// 400 for an invalid or expired share token, 404 for an unknown user and
// 409 when the user has already submitted KYC.
//
// Example:
//
//	resp, err := client.ImportSharedKYC(ctx, &gaian.ImportSharedKYCRequest{
//		UserID:     userID,
//		ShareToken: shareToken,
//	})
func (c *Client) ImportSharedKYC(ctx context.Context, req *ImportSharedKYCRequest) (*UserResponse[ImportSharedKYCResponse], error) {
	endpoint := fmt.Sprintf("/api/v2/users/%s/kyc/import-shared", url.PathEscape(req.UserID))

	apiRequest := request{
		Method:   http.MethodPost,
		Endpoint: endpoint,
		Params:   req,
	}

	rawResponse, err := c.excute(ctx, &apiRequest)
	if err != nil {
		return nil, err
	}

	response := new(UserResponse[ImportSharedKYCResponse])
	if err := json.Unmarshal(rawResponse, response); err != nil {
		return nil, err
	}

	return response, nil
}
