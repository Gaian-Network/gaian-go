package gaian

import (
	"context"
	"net/url"
)

// GetOrderStatus returns the full order object for the given order ID.
// Poll this endpoint until Status reaches OrderCompleted or OrderFailed.
func (s *PaymentService) GetOrderStatus(ctx context.Context, orderID string) (*Order, error) {
	var resp Order
	q := url.Values{"orderId": {orderID}}
	if err := s.c.get(ctx, s.c.pmtBaseURL, "/api/v1/status", q, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
