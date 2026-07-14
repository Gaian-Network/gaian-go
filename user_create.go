package gaian

import (
	"context"
	"encoding/json"
	"net/http"
)

// CreateUserRequest has no parameters — the user record starts empty and
// is filled in later via SubmitKYC/UpdateKYC and CreateWallet.
type CreateUserRequest struct{}

// CreateUserResponse holds the server-generated ID for the new user.
type CreateUserResponse struct {
	UserID string `json:"userId"`
}

// CreateUser creates a new user and returns its UserID. v2 no longer
// links a wallet or any KYC data at registration time — call
// CreateWallet and SubmitKYC/GenerateKYCLink afterward using the
// returned UserID.
//
// Example:
//
//	resp, err := client.CreateUser(ctx, &gaian.CreateUserRequest{})
//	if err != nil {
//		return err
//	}
//	userID := resp.Data.UserID
func (c *Client) CreateUser(ctx context.Context, req *CreateUserRequest) (*UserResposne[CreateUserResponse], error) {
	apiRequest := request{
		Method:   http.MethodPost,
		Endpoint: "/api/v2/users",
	}

	rawResponse, err := c.excute(ctx, &apiRequest)
	if err != nil {
		return nil, err
	}

	response := new(UserResposne[CreateUserResponse])
	if err := json.Unmarshal(rawResponse, response); err != nil {
		return nil, err
	}

	return response, nil
}
