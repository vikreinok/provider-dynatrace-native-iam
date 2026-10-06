package dynatrace

import (
	"errors"
	"fmt"
	"testing"
)

func TestIsNotFound(t *testing.T) {
	if IsNotFound(nil) {
		t.Error("IsNotFound(nil) = true, want false")
	}
	if IsNotFound(errors.New("generic error")) {
		t.Error("IsNotFound(generic) = true, want false")
	}
	if !IsNotFound(&APIError{StatusCode: 404, Message: "Not found"}) {
		t.Error("IsNotFound(404) = false, want true")
	}
	if IsNotFound(&APIError{StatusCode: 400, Message: "Bad Request"}) {
		t.Error("IsNotFound(400) = true, want false")
	}
}

func TestIsAlreadyExists(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "NilError",
			err:      nil,
			expected: false,
		},
		{
			name:     "GenericError",
			err:      errors.New("something went wrong"),
			expected: false,
		},
		{
			name:     "Status409Conflict",
			err:      &APIError{StatusCode: 409, Message: "Conflict"},
			expected: true,
		},
		{
			name: "Status400AlreadyBeenStored",
			err: &APIError{
				StatusCode: 400,
				Message:    "Dynatrace API error (HTTP 400): {\"error\":true,\"message\":\"At least one of the values has already been stored for this account and dimension.\",\"payload\":null}",
				RawBody:    "{\"error\":true,\"message\":\"At least one of the values has already been stored for this account and dimension.\",\"payload\":null}",
			},
			expected: true,
		},
		{
			name: "Status400AlreadyExistsInRawBody",
			err: &APIError{
				StatusCode: 400,
				Message:    "Bad request",
				RawBody:    "{\"error\":true,\"message\":\"Entity already exists\"}",
			},
			expected: true,
		},
		{
			name:     "WrappedGenericAlreadyExists",
			err:      fmt.Errorf("operation failed: %w", errors.New("group with name Platform Engineers already exists")),
			expected: true,
		},
		{
			name:     "Status400OtherMessage",
			err:      &APIError{StatusCode: 400, Message: "Invalid query syntax"},
			expected: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsAlreadyExists(tc.err)
			if got != tc.expected {
				t.Errorf("IsAlreadyExists(%v) = %v, want %v", tc.err, got, tc.expected)
			}
		})
	}
}
