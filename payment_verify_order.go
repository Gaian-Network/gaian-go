package gaian

import "context"

type verifyOrderRequest struct {
	OrderID          string `json:"orderId"`
	TransactionProof string `json:"transactionProof"`
}

// VerifyOrderResult holds the outcome of transaction verification.
type VerifyOrderResult struct {
	OrderID            string      `json:"orderId"`
	Status             OrderStatus `json:"status"`
	TransactionHash    string      `json:"transactionHash"`
	Message            string      `json:"message"`
	BankTransferStatus *string     `json:"bankTransferStatus"` // "queued" | "failed" | nil
}

// VerifyOrder confirms the on-chain transaction for an order, triggering the
// bank transfer leg. transactionProof is typically the transaction signature.
func (s *PaymentService) VerifyOrder(ctx context.Context, orderID, transactionProof string) (*VerifyOrderResult, error) {
	var resp VerifyOrderResult
	if err := s.c.post(ctx, s.c.pmtBaseURL, "/api/v1/verifyOrder",
		verifyOrderRequest{OrderID: orderID, TransactionProof: transactionProof}, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
