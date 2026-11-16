package gaian

import (
	"errors"
	"net/http"
	"testing"
)

func parseAPIError(status int, body []byte) *APIError {
	return newAPIError(&http.Response{StatusCode: status, Header: http.Header{}}, body)
}

func TestParseAPIError(t *testing.T) {
	e := parseAPIError(404, []byte(`{"status":"error","error":"not_found","message":"user not found"}`))
	if e.Code != "not_found" || e.ErrorCode != "not_found" {
		t.Errorf("Code = %q, ErrorCode = %q, want both %q", e.Code, e.ErrorCode, "not_found")
	}
	if e.Status != "error" {
		t.Errorf("Status = %q, want %q", e.Status, "error")
	}
	if e.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", e.StatusCode)
	}
	if e.Message != "user not found" {
		t.Errorf("Message = %q, want %q", e.Message, "user not found")
	}
}

func TestParseAPIError_UnparsableBodyIsKeptRaw(t *testing.T) {
	e := parseAPIError(500, []byte("internal server error"))
	if e.Body != "internal server error" {
		t.Errorf("Body = %q, want the raw body", e.Body)
	}
	if e.Message != "" {
		t.Errorf("Message = %q, want empty for a non-JSON body", e.Message)
	}
	if got, want := e.Error(), "gaian: api error: status=500"; got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestAPIError_ErrorString(t *testing.T) {
	resp := &http.Response{StatusCode: 400, Header: http.Header{}}
	body := `{"success":false,"requestId":"req_2","error":{"code":"QUOTE_EXPIRED","message":"quote expired"}}`

	got := newAPIError(resp, []byte(body)).Error()
	want := "gaian: api error: status=400 code=QUOTE_EXPIRED message=quote expired request_id=req_2"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestAPIError_NumericStatusFieldDoesNotDropMessage(t *testing.T) {
	e := parseAPIError(422, []byte(`{"status":422,"code":"VALIDATION","message":"bad field"}`))
	if e.Message != "bad field" || e.Code != "VALIDATION" {
		t.Errorf("Message = %q, Code = %q", e.Message, e.Code)
	}
}

func TestAPIError_Predicates(t *testing.T) {
	tests := []struct {
		name   string
		status int
		check  func(error) bool
	}{
		{"404 is not found", 404, IsNotFound},
		{"409 is conflict", 409, IsConflict},
		{"401 is unauthorized", 401, IsUnauthorized},
		{"403 is forbidden", 403, IsForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parseAPIError(tt.status, []byte(`{"message":"x"}`))
			if !tt.check(err) {
				t.Errorf("predicate returned false for status %d", tt.status)
			}
		})
	}
}

func TestAPIError_PredicatesFalseForOtherErrors(t *testing.T) {
	if IsNotFound(errors.New("some other error")) {
		t.Error("IsNotFound should be false for a non-*APIError")
	}
	if IsNotFound(parseAPIError(500, []byte(`{}`))) {
		t.Error("IsNotFound should be false for a 500")
	}
}
