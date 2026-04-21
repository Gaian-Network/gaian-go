package gaian

import "context"

// PlacePrefundedOrderRequest is a subset of PlaceOrderRequest — chain is always
// Solana and no on-chain transaction is required by the caller.
type PlacePrefundedOrderRequest struct {
	// Required
	QRString       string         `json:"qrString"`
	Amount         float64        `json:"amount"`
	CryptoCurrency CryptoCurrency `json:"cryptoCurrency"`
	FromAddress    string         `json:"fromAddress"`

	// Optional
	FiatCurrency         FiatCurrency `json:"fiatCurrency,omitempty"`
	TransactionReference string       `json:"transactionReference,omitempty"`
}

// PlacePrefundedOrder creates an order backed by the tenant's prefunded balance.
// No on-chain signing is required; poll GetOrderStatus until terminal state.
func (s *PaymentService) PlacePrefundedOrder(ctx context.Context, req PlacePrefundedOrderRequest) (*PlaceOrderResponse, error) {
	var resp PlaceOrderResponse
	if err := s.c.post(ctx, s.c.pmtBaseURL, "/api/v1/placeOrder/prefund", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
