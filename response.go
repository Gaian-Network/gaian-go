package gaian

// UserResposne is the response envelope used by the User/KYC/Wallet/
// Organization endpoints ({data, requestId}, no success flag). Data is
// nil if the API omitted it, which shouldn't happen on a non-error
// response but is left as a pointer defensively — check it before
// dereferencing.
type UserResposne[T any] struct {
	RequestID string `json:"requestId"`
	Data      *T     `json:"data"`
}

// PaymentResponse is the response envelope used by the Payment endpoints
// (ParseQR, Quote*, PlaceOrder*, VerifyOrder, GetOrderStatus, ListBanks,
// VerifyAccount): {success, requestId, data}. Success is true whenever
// this envelope is populated at all — a request that fails returns a
// non-2xx status and an error from the call instead, so callers
// generally don't need to check Success themselves.
type PaymentResponse[T any] struct {
	RequestID string `json:"requestId"`
	Data      *T     `json:"data"`
	Success   bool   `json:"success"`
}
