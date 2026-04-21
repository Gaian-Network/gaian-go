package gaian

import "context"

// CalculateExchangeRequest specifies the parameters for a rate calculation.
type CalculateExchangeRequest struct {
	Amount  float64        `json:"amount"`  // min 0.01
	Country string         `json:"country"` // e.g. "VN", "PH", "BR"
	Chain   Chain          `json:"chain"`
	Token   CryptoCurrency `json:"token"`
}

type calculateExchangeResponse struct {
	Success      bool         `json:"success"`
	ExchangeInfo ExchangeInfo `json:"exchangeInfo"`
}

// CalculateExchange returns the crypto equivalent for a given fiat amount.
func (s *PaymentService) CalculateExchange(ctx context.Context, req CalculateExchangeRequest) (*ExchangeInfo, error) {
	var resp calculateExchangeResponse
	if err := s.c.post(ctx, s.c.pmtBaseURL, "/api/v1/calculateExchange", req, &resp); err != nil {
		return nil, err
	}
	return &resp.ExchangeInfo, nil
}
