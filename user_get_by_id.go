package gaian

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
)

// User represents a Gaian v2 platform user. Dates are left as strings
// (rather than time.Time) since the API's exact timestamp format isn't
// guaranteed — parse with time.Parse(time.RFC3339, ...) if you need them
// as time.Time.
type User struct {
	ID               string `json:"id"`
	FirstName        string `json:"firstName"`
	LastName         string `json:"lastName"`
	Email            string `json:"email"`
	Gender           string `json:"gender"`
	DateOfBirth      string `json:"dateOfBirth"`
	Nationality      string `json:"nationality"`
	NationalID       string `json:"nationalId"`
	Type             string `json:"type"`
	IssueDate        string `json:"issueDate"`
	ExpiryDate       string `json:"expiryDate"`
	AddressLine1     string `json:"addressLine1"`
	AddressLine2     string `json:"addressLine2"`
	City             string `json:"city"`
	State            string `json:"state"`
	ZipCode          string `json:"zipCode"`
	PhoneCountryCode string `json:"phoneCountryCode"`
	PhoneNumber      string `json:"phoneNumber"`
	KYCStatus        string `json:"kycStatus"`
	RejectReason     string `json:"rejectReason"`
}

// GetUserByIDRequest identifies a user by their server-generated ID.
type GetUserByIDRequest struct {
	UserID string `param:"userId"`
}

// GetUserByIDResponse is the full user record.
type GetUserByIDResponse = User

// GetUserByID retrieves a user by req.UserID (the ID returned by
// CreateUser). Use GetUserByWallet if you only have a wallet address.
//
// Example:
//
//	resp, err := client.GetUserByID(ctx, &gaian.GetUserByIDRequest{UserID: userID})
//	if err != nil {
//		return err
//	}
//	fmt.Println(resp.Data.KYCStatus)
func (c *Client) GetUserByID(ctx context.Context, req *GetUserByIDRequest) (*UserResposne[GetUserByIDResponse], error) {
	apiRequest := request{
		Method:   http.MethodGet,
		Endpoint: "/api/v2/users/" + url.PathEscape(req.UserID),
	}

	rawResponse, err := c.excute(ctx, &apiRequest)
	if err != nil {
		return nil, err
	}

	response := new(UserResposne[GetUserByIDResponse])
	if err := json.Unmarshal(rawResponse, response); err != nil {
		return nil, err
	}
	return response, nil
}
