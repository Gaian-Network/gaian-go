package gaian

import (
	"errors"
	"testing"
)

// NOTE: parseAPIError/IsXxx below are not currently wired into excute
// (see the NOTE at the top of errors.go) — these tests exercise the
// parsing/predicate logic in isolation so it's known-correct for
// whenever it does get wired up.

func TestParseAPIError(t *testing.T) {
	e := parseAPIError(404, []byte(`{"status":"error","error":"not_found","message":"user not found"}`))
	if e.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", e.StatusCode)
	}
	if e.Message != "user not found" {
		t.Errorf("Message = %q, want %q", e.Message, "user not found")
	}
}

func TestParseAPIError_UnparsableBodyFallsBackToRaw(t *testing.T) {
	e := parseAPIError(500, []byte("internal server error"))
	if e.Message != "internal server error" {
		t.Errorf("Message = %q, want the raw body", e.Message)
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
