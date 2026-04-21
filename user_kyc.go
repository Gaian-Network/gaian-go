package gaian

import "context"

// ── Generate KYC link ─────────────────────────────────────────────────────────

type kycLinkRequest struct {
	WalletAddress string `json:"walletAddress"`
}

type kycLinkResponse struct {
	Status    bool   `json:"status"`
	Message   string `json:"message"`
	WebSDKURL string `json:"websdkUrl"`
}

// GenerateKYCLink creates a time-limited Sumsub KYC URL for the given wallet.
// The URL must be used immediately — do not cache it.
func (s *UserService) GenerateKYCLink(ctx context.Context, walletAddress string) (string, error) {
	var resp kycLinkResponse
	if err := s.c.post(ctx, s.c.userBaseURL, "/api/v1/kyc/link",
		kycLinkRequest{WalletAddress: walletAddress}, &resp); err != nil {
		return "", err
	}
	return resp.WebSDKURL, nil
}

// ── Submit KYC ────────────────────────────────────────────────────────────────

// SubmitKYCRequest carries all required fields for direct KYC submission.
type SubmitKYCRequest struct {
	UserWalletAddress string `json:"userWalletAddress"`
	UserEmail         string `json:"userEmail"`
	FirstName         string `json:"firstName"`
	LastName          string `json:"lastName"`
	DateOfBirth       string `json:"dateOfBirth"`
	Gender            string `json:"gender"`
	Nationality       string `json:"nationality"`
	Type              string `json:"type"`
	NationalID        string `json:"nationalId"`
	IssueDate         string `json:"issueDate"`
	ExpiryDate        string `json:"expiryDate"`
	AddressLine1      string `json:"addressLine1"`
	AddressLine2      string `json:"addressLine2"`
	City              string `json:"city"`
	State             string `json:"state"`
	ZipCode           string `json:"zipCode"`
	FrontIDImage      string `json:"frontIdImage"`
	BackIDImage       string `json:"backIdImage"`
	HoldIDImage       string `json:"holdIdImage"`
	PhoneNumber       string `json:"phoneNumber"`
	PhoneCountryCode  string `json:"phoneCountryCode"`
}

type submitKYCData struct {
	KYCStatus KYCStatus `json:"kycStatus"`
}

// SubmitKYC submits KYC information directly without the Sumsub web flow.
// Returns the resulting KYC status.
func (s *UserService) SubmitKYC(ctx context.Context, req SubmitKYCRequest) (KYCStatus, error) {
	var resp envelope[submitKYCData]
	if err := s.c.post(ctx, s.c.userBaseURL, "/api/v1/submit-kyc-information", req, &resp); err != nil {
		return "", err
	}
	if resp.Data == nil {
		return "", nil
	}
	return resp.Data.KYCStatus, nil
}
