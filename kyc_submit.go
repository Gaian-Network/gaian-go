package gaian

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// SubmitKYCRequest carries all fields for direct KYC submission (no
// hosted flow — see GenerateKYCLink for that alternative). Gender, Type,
// Occupation, and Nationality are constrained to the enum types defined
// in kyc_enums.go, matching the values the API accepts.
type SubmitKYCRequest struct {
	UserID             string          `param:"userId" json:"-"`
	FirstName          string          `json:"firstName"`
	LastName           string          `json:"lastName"`
	Email              string          `json:"email"`
	Gender             Gender          `json:"gender"`
	DateOfBirth        string          `json:"dateOfBirth"`
	Nationality        NationalityCode `json:"nationality"`
	NationalID         string          `json:"nationalId"`
	Type               DocumentType    `json:"type"`
	IssueDate          string          `json:"issueDate,omitempty"`
	ExpiryDate         string          `json:"expiryDate"`
	AddressLine1       string          `json:"addressLine1"`
	AddressLine2       string          `json:"addressLine2,omitempty"`
	City               string          `json:"city"`
	State              string          `json:"state,omitempty"`
	ZipCode            string          `json:"zipCode,omitempty"`
	CountryOfResidence string          `json:"countryOfResidence"`
	Occupation         OccupationCode  `json:"occupation"`
	PhoneCountryCode   string          `json:"phoneCountryCode"`
	PhoneNumber        string          `json:"phoneNumber"`
	FrontIDImage       *string         `json:"frontIdImage,omitempty"`
	BackIDImage        *string         `json:"backIdImage,omitempty"`
	HoldIDImage        *string         `json:"holdIdImage,omitempty"`
}

// SubmitKYCResponse is returned immediately after submission — KYCStatus
// reflects the initial state (typically "PROCESSING"), not the final
// verification result. Poll GetUserByID or GetMarkets to track progress.
type SubmitKYCResponse struct {
	UserID    string `json:"userId"`
	Email     string `json:"email"`
	KYCStatus string `json:"kycStatus"`
}

// SubmitKYC submits KYC information directly for req.UserID, bypassing
// the hosted Didit redirect flow (see GenerateKYCLink for that
// alternative). Verification itself is processed asynchronously — this
// call only confirms the submission was accepted.
//
// Example:
//
//	resp, err := client.SubmitKYC(ctx, &gaian.SubmitKYCRequest{
//		UserID:             userID,
//		FirstName:          "Nguyen",
//		LastName:           "Van A",
//		Email:              "user@example.com",
//		DateOfBirth:        "1990-01-15",
//		Nationality:        gaian.NationalityCode("VN"),
//		NationalID:         "012345678901",
//		Type:               gaian.DocumentTypeIDCard,
//		ExpiryDate:         "2025-01-01",
//		AddressLine1:       "123 Le Loi",
//		City:               "Ho Chi Minh City",
//		CountryOfResidence: "VN",
//		Occupation:         gaian.OccupationProfessional,
//		PhoneCountryCode:   "+84",
//		PhoneNumber:        "912345678",
//	})
func (c *Client) SubmitKYC(ctx context.Context, req *SubmitKYCRequest) (*UserResposne[SubmitKYCResponse], error) {
	endpoint := fmt.Sprintf("/api/v2/users/%s/kyc", url.PathEscape(req.UserID))

	apiRequest := request{
		Method:   http.MethodPost,
		Endpoint: endpoint,
		Params:   req,
	}

	rawResponse, err := c.excute(ctx, &apiRequest)
	if err != nil {
		return nil, err
	}

	response := new(UserResposne[SubmitKYCResponse])
	if err := json.Unmarshal(rawResponse, response); err != nil {
		return nil, err
	}

	return response, nil
}
