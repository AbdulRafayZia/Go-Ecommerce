package httputil

import (
	"errors"
	"net/http"
	"testing"
)

func TestToHTTPStatus(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
	}{
		{
			name:           "nil error",
			err:            nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "not found error",
			err:            ErrNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "already exists error",
			err:            ErrAlreadyExists,
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "invalid argument error",
			err:            ErrInvalidArgument,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "permission denied error",
			err:            ErrPermissionDenied,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "unauthenticated error",
			err:            ErrUnauthenticated,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "unavailable error",
			err:            ErrUnavailable,
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name:           "timeout error",
			err:            ErrTimeout,
			expectedStatus: http.StatusRequestTimeout,
		},
		{
			name:           "unknown error",
			err:            errors.New("unknown error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := ToHTTPStatus(tt.err)

			if status != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, status)
			}
		})
	}
}

func TestErrorCodeFromDomain(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode string
	}{
		{
			name:         "nil error",
			err:          nil,
			expectedCode: "",
		},
		{
			name:         "not found error",
			err:          ErrNotFound,
			expectedCode: "NOT_FOUND",
		},
		{
			name:         "already exists error",
			err:          ErrAlreadyExists,
			expectedCode: "ALREADY_EXISTS",
		},
		{
			name:         "invalid argument error",
			err:          ErrInvalidArgument,
			expectedCode: "INVALID_ARGUMENT",
		},
		{
			name:         "unknown error",
			err:          errors.New("unknown"),
			expectedCode: "INTERNAL_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := ErrorCodeFromDomain(tt.err)

			if code != tt.expectedCode {
				t.Errorf("expected code %s, got %s", tt.expectedCode, code)
			}
		})
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "not found error",
			err:      ErrNotFound,
			expected: true,
		},
		{
			name:     "wrapped not found error",
			err:      errors.Join(ErrNotFound, errors.New("additional context")),
			expected: true,
		},
		{
			name:     "other error",
			err:      ErrInternal,
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNotFound(tt.err)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestIsAuthError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "unauthenticated error",
			err:      ErrUnauthenticated,
			expected: true,
		},
		{
			name:     "permission denied error",
			err:      ErrPermissionDenied,
			expected: true,
		},
		{
			name:     "other error",
			err:      ErrNotFound,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAuthError(tt.err)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
