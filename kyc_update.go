package gaian

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// UpdateKYCRequest partially updates previously submitted KYC
// information for a user. Every body field is an optional pointer —
// only the non-nil fields are sent and updated; omitted fields keep
// their current value. Use SubmitKYC for the initial submission instead.
type UpdateKYCRequest struct {
	UserID             string  `param:"userId" json:"-"`
	FirstName          *string `json:"firstName,omitempty"`
	LastName           *string `json:"lastName,omitempty"`
	Email              *string `json:"email,omitempty"`
	NationalID         *string `json:"nationalId,omitempty"`
	DateOfBirth        *string `json:"dateOfBirth,omitempty"`
	Gender             *string `json:"gender,omitempty"` // "male" | "female"
	Nationality        *string `json:"nationality,omitempty"`
	Type               *string `json:"type,omitempty"` // "ID_CARD" | "PASSPORT"
	IssueDate          *string `json:"issueDate,omitempty"`
	ExpiryDate         *string `json:"expiryDate,omitempty"`
	AddressLine1       *string `json:"addressLine1,omitempty"`
	AddressLine2       *string `json:"addressLine2,omitempty"`
	City               *string `json:"city,omitempty"`
	State              *string `json:"state,omitempty"`
	ZipCode            *string `json:"zipCode,omitempty"`
	CountryOfResidence *string `json:"countryOfResidence,omitempty"`
	Occupation         *string `json:"occupation,omitempty"`
	PhoneCountryCode   *string `json:"phoneCountryCode,omitempty"`
	PhoneNumber        *string `json:"phoneNumber,omitempty"`
	FrontIDImage       *string `json:"frontIdImage,omitempty"`
	BackIDImage        *string `json:"backIdImage,omitempty"`
	HoldIDImage        *string `json:"holdIdImage,omitempty"`
}

// UpdateKYCResponse mirrors the user's full KYC record back after the
// update is applied, alongside the resulting KYCStatus.
type UpdateKYCResponse struct {
	UserID             string `json:"userId"`
	KYCStatus          string `json:"kycStatus"`
	FirstName          string `json:"firstName"`
	LastName           string `json:"lastName"`
	Email              string `json:"email"`
	DateOfBirth        string `json:"dateOfBirth"`
	Gender             string `json:"gender"`
	Nationality        string `json:"nationality"`
	Type               string `json:"type"`
	NationalID         string `json:"nationalId"`
	IssueDate          string `json:"issueDate"`
	ExpiryDate         string `json:"expiryDate"`
	AddressLine1       string `json:"addressLine1"`
	AddressLine2       string `json:"addressLine2"`
	City               string `json:"city"`
	State              string `json:"state"`
	ZipCode            string `json:"zipCode"`
	CountryOfResidence string `json:"countryOfResidence"`
	Occupation         string `json:"occupation"`
	PhoneNumber        string `json:"phoneNumber"`
	PhoneCountryCode   string `json:"phoneCountryCode"`
}

// UpdateKYC applies a partial update to req.UserID's previously
// submitted KYC information. Returns 409 if the new Email or NationalID
// is already in use by another user.
//
// Example:
//
//	newEmail := "updated@example.com"
//	resp, err := client.UpdateKYC(ctx, &gaian.UpdateKYCRequest{
//		UserID: userID,
//		Email:  &newEmail,
//	})
func (c *Client) UpdateKYC(ctx context.Context, req *UpdateKYCRequest) (*UserResposne[UpdateKYCResponse], error) {
	endpoint := fmt.Sprintf("/api/v2/users/%s/kyc", url.PathEscape(req.UserID))

	apiRequest := request{
		Method:   http.MethodPatch,
		Endpoint: endpoint,
		Params:   req,
	}

	rawResponse, err := c.excute(ctx, &apiRequest)
	if err != nil {
		return nil, err
	}

	response := new(UserResposne[UpdateKYCResponse])
	if err := json.Unmarshal(rawResponse, response); err != nil {
		return nil, err
	}

	return response, nil
}
