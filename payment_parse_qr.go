package gaian

import (
	"context"
	"encoding/json"
	"net/http"
)

// ParseQRRequest carries a raw QR payment string to decode. Country is
// an optional ISO-3166-1 alpha-2 hint; leave it empty to auto-detect
// from the QR's bank BIN.
type ParseQRRequest struct {
	QRString string `json:"qrString"`
	Country  string `json:"country,omitempty"`
}

// QRInfo is the parsed content of a QR payment code, shared by ParseQR
// and the QR-based Quote/QuotePrefund responses.
type QRInfo struct {
	IsValid         bool           `json:"isValid"`
	EncodedString   string         `json:"encodedString"`
	Country         string         `json:"country"`
	QRProvider      string         `json:"qrProvider"`
	BankBin         string         `json:"bankBin"`
	AccountNumber   string         `json:"accountNumber"`
	Amount          *float64       `json:"amount,omitempty"`
	Currency        *string        `json:"currency,omitempty"`
	Purpose         *string        `json:"purpose,omitempty"`
	BeneficiaryName string         `json:"beneficiaryName"`
	DetailedQRInfo  map[string]any `json:"detailedQrInfo,omitempty"`
}

// ParseQRResponse is the decoded QR content.
type ParseQRResponse = QRInfo

// ParseQR decodes a QR payment code (EMVCO-standard QR from Vietnam,
// Thailand, and other supported markets) without creating a quote or
// reserving routing. Call Quote/QuoteDirect afterward to actually price
// and reserve the payment.
//
// Example:
//
//	resp, err := client.ParseQR(ctx, &gaian.ParseQRRequest{QRString: qrString})
//	if err != nil {
//		return err
//	}
//	fmt.Println(resp.Data.BankBin, resp.Data.AccountNumber)
func (c *Client) ParseQR(ctx context.Context, req *ParseQRRequest) (*PaymentResponse[ParseQRResponse], error) {
	apiRequest := request{
		Method:   http.MethodPost,
		Endpoint: "/api/v2/parseQr",
		Params:   req,
	}

	rawResponse, err := c.excute(ctx, &apiRequest)
	if err != nil {
		return nil, err
	}

	response := new(PaymentResponse[ParseQRResponse])
	if err := json.Unmarshal(rawResponse, response); err != nil {
		return nil, err
	}

	return response, nil
}
