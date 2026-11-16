package gaian

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

// APIError is the error returned for every response with an HTTP status
// of 400 or above. Read it with errors.As; errors.Is(err,
// ErrUnexpectedStatus) is also true for it.
type APIError struct {
	// StatusCode is the HTTP status of the response.
	StatusCode int
	// Status is the "status" field of the error body, when present.
	//
	// Deprecated: v2 error bodies do not send it. Use StatusCode.
	Status string `json:"status"`
	// ErrorCode always equals Code.
	//
	// Deprecated: use Code.
	ErrorCode string `json:"error"`
	// Message is the message from the error body. It is empty when the
	// body is empty or is not JSON (for example a proxy's HTML page).
	Message string
	// Code is the API error code (for example "IDENTITY_REQUIRED"). Only
	// Payment endpoints send one; it is empty otherwise.
	Code string
	// RequestID is the request id from the error body, or from the
	// X-Request-ID response header when the body has none.
	RequestID string
	// Body is the raw response body, kept as received.
	Body string
}

// Error returns the status, and the code, message and request id when
// the response carried them. It never includes the raw body.
func (e *APIError) Error() string {
	var b strings.Builder

	b.WriteString("gaian: api error: status=")
	b.WriteString(strconv.Itoa(e.StatusCode))

	if e.Code != "" {
		b.WriteString(" code=")
		b.WriteString(e.Code)
	}

	if e.Message != "" {
		b.WriteString(" message=")
		b.WriteString(e.Message)
	}

	if e.RequestID != "" {
		b.WriteString(" request_id=")
		b.WriteString(e.RequestID)
	}

	return b.String()
}

// Unwrap returns ErrUnexpectedStatus, so errors.Is(err,
// ErrUnexpectedStatus) holds for every *APIError.
func (e *APIError) Unwrap() error {
	return ErrUnexpectedStatus
}

// IsNotFound reports whether err is, or wraps, an *APIError with status 404.
func IsNotFound(err error) bool {
	return hasStatus(err, http.StatusNotFound)
}

// IsConflict reports whether err is, or wraps, an *APIError with status
// 409 (resource already exists).
func IsConflict(err error) bool {
	return hasStatus(err, http.StatusConflict)
}

// IsUnauthorized reports whether err is, or wraps, an *APIError with
// status 401 (missing or invalid API key, signature or timestamp).
func IsUnauthorized(err error) bool {
	return hasStatus(err, http.StatusUnauthorized)
}

// IsForbidden reports whether err is, or wraps, an *APIError with status 403.
func IsForbidden(err error) bool {
	return hasStatus(err, http.StatusForbidden)
}

func hasStatus(err error, status int) bool {
	var apiErr *APIError

	return errors.As(err, &apiErr) && apiErr.StatusCode == status
}

type apiErrorBody struct {
	Message   string          `json:"message"`
	Status    json.RawMessage `json:"status"`
	Code      json.RawMessage `json:"code"`
	RequestID string          `json:"requestId"`
	Error     json.RawMessage `json:"error"`
}

type apiErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func newAPIError(resp *http.Response, body []byte) *APIError {
	apiErr := &APIError{
		StatusCode: resp.StatusCode,
		Body:       string(body),
	}

	var parsed apiErrorBody
	if err := json.Unmarshal(body, &parsed); err == nil {
		apiErr.Message = parsed.Message
		apiErr.RequestID = parsed.RequestID
		apiErr.Status = jsonString(parsed.Status)
		apiErr.Code = jsonString(parsed.Code)

		var detail apiErrorDetail
		if json.Unmarshal(parsed.Error, &detail) == nil {
			if detail.Code != "" {
				apiErr.Code = detail.Code
			}

			if detail.Message != "" {
				apiErr.Message = detail.Message
			}
		} else if apiErr.Code == "" {
			apiErr.Code = jsonString(parsed.Error)
		}
	}

	if apiErr.RequestID == "" {
		apiErr.RequestID = resp.Header.Get("X-Request-ID")
	}

	apiErr.ErrorCode = apiErr.Code

	return apiErr
}

func jsonString(raw json.RawMessage) string {
	var s string
	if json.Unmarshal(raw, &s) != nil {
		return ""
	}

	return s
}
