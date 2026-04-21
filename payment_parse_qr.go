package gaian

import "context"

type parseQRRequest struct {
	QRString string `json:"qrString"`
	Country  string `json:"country,omitempty"`
}

type parseQRResponse struct {
	Success   bool         `json:"success"`
	QRInfo    ParsedQRInfo `json:"qrInfo"`
	Timestamp string       `json:"timestamp"`
}

// ParseQROptions allows passing an optional country hint to the QR parser.
type ParseQROptions struct {
	// Country is an optional hint (e.g. "VN", "TH"). Auto-detected when empty.
	Country string
}

// ParseQR decodes an EMVCO-standard QR payment string.
// Supported countries: Vietnam (VN), Philippines (PH), Thailand (TH), Brazil (BR).
func (s *PaymentService) ParseQR(ctx context.Context, qrString string, opts ...ParseQROptions) (*ParsedQRInfo, error) {
	req := parseQRRequest{QRString: qrString}
	if len(opts) > 0 {
		req.Country = opts[0].Country
	}
	var resp parseQRResponse
	if err := s.c.post(ctx, s.c.pmtBaseURL, "/api/v1/parseQr", req, &resp); err != nil {
		return nil, err
	}
	return &resp.QRInfo, nil
}
