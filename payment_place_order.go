package gaian

import "context"

// PlaceOrderRequest holds the parameters for a standard (on-chain) order.
type PlaceOrderRequest struct {
	// Required
	QRString       string         `json:"qrString"`
	Amount         float64        `json:"amount"`
	CryptoCurrency CryptoCurrency `json:"cryptoCurrency"`
	FromAddress    string         `json:"fromAddress"`

	// Optional
	FiatCurrency         FiatCurrency `json:"fiatCurrency,omitempty"`
	Chain                Chain        `json:"chain,omitempty"`
	TransactionReference string       `json:"transactionReference,omitempty"`
	RouteID              Route        `json:"routeId,omitempty"`
}

// PlaceOrderResponse is returned by both standard and prefunded order placement.
type PlaceOrderResponse struct {
	OrderID              string         `json:"orderId"`
	Status               OrderStatus    `json:"status"`
	FiatAmount           float64        `json:"fiatAmount"`
	FiatCurrency         FiatCurrency   `json:"fiatCurrency"`
	CryptoAmount         float64        `json:"cryptoAmount"`
	CryptoCurrency       CryptoCurrency `json:"cryptoCurrency"`
	ExchangeRate         float64        `json:"exchangeRate"`
	QRInfo               QRInfo         `json:"qrInfo"`
	CryptoTransferInfo   map[string]any `json:"cryptoTransferInfo"`
	Timestamp            string         `json:"timestamp"`
	TransactionReference string         `json:"transactionReference"`
	RouteID              Route          `json:"routeId"`
	IsPrefunded          bool           `json:"isPrefunded"`
}

// PlaceOrder creates a standard payment order. The caller must then build,
// sign, and broadcast the on-chain transaction, then call VerifyOrder.
func (s *PaymentService) PlaceOrder(ctx context.Context, req PlaceOrderRequest) (*PlaceOrderResponse, error) {
	var resp PlaceOrderResponse
	if err := s.c.post(ctx, s.c.pmtBaseURL, "/api/v1/placeOrder", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
