package gaian

import (
	"encoding/json"
	"fmt"
)

// NOTE: this file is not currently wired up. Client.excute (client.go)
// returns a plain fmt.Errorf-wrapped ErrUnexpectedStatus on a non-2xx
// response, not an *APIError — so APIError/parseAPIError/IsNotFound/
// IsConflict/IsUnauthorized/IsForbidden below never actually trigger
// today. v2 has two distinct error body shapes (flat {message} for
// User/KYC/Wallet/Organization endpoints, {success:false,
// error:{code,message,details}} for Payment endpoints) and this struct
// doesn't yet parse either correctly (it expects a v1-style flat
// {status,error,message} body). Left in place pending a decision on
// whether/how to restore typed error handling.

// APIError represents an error returned by the Gaian API.
type APIError struct {
	StatusCode int
	Status     string `json:"status"`
	ErrorCode  string `json:"error"`
	Message    string `json:"message"`
}

func (e *APIError) Error() string {
	msg := e.Message
	if msg == "" {
		msg = e.ErrorCode
	}
	return fmt.Sprintf("gaian: api error %d: %s", e.StatusCode, msg)
}

// IsNotFound reports whether the error is a 404.
func IsNotFound(err error) bool {
	e, ok := err.(*APIError)
	return ok && e.StatusCode == 404
}

// IsConflict reports whether the error is a 409 (resource already exists).
func IsConflict(err error) bool {
	e, ok := err.(*APIError)
	return ok && e.StatusCode == 409
}

// IsUnauthorized reports whether the error is a 401 (missing/invalid API key).
func IsUnauthorized(err error) bool {
	e, ok := err.(*APIError)
	return ok && e.StatusCode == 401
}

// IsForbidden reports whether the error is a 403.
func IsForbidden(err error) bool {
	e, ok := err.(*APIError)
	return ok && e.StatusCode == 403
}

func parseAPIError(status int, body []byte) *APIError {
	e := &APIError{StatusCode: status}
	// Best-effort decode; preserve raw body in Message on failure.
	if err := json.Unmarshal(body, e); err != nil {
		e.Message = string(body)
	}
	return e
}
