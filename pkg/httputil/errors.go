package httputil

import (
	"errors"
	"net/http"
)

// Common domain errors
var (
	ErrNotFound         = errors.New("resource not found")
	ErrAlreadyExists    = errors.New("resource already exists")
	ErrInvalidArgument  = errors.New("invalid argument")
	ErrPermissionDenied = errors.New("permission denied")
	ErrUnauthenticated  = errors.New("unauthenticated")
	ErrInternal         = errors.New("internal error")
	ErrUnavailable      = errors.New("service unavailable")
	ErrTimeout          = errors.New("request timeout")
)

// ToHTTPStatus converts a domain error to an HTTP status code
func ToHTTPStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}

	switch {
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrAlreadyExists):
		return http.StatusConflict
	case errors.Is(err, ErrInvalidArgument):
		return http.StatusBadRequest
	case errors.Is(err, ErrPermissionDenied):
		return http.StatusForbidden
	case errors.Is(err, ErrUnauthenticated):
		return http.StatusUnauthorized
	case errors.Is(err, ErrUnavailable):
		return http.StatusServiceUnavailable
	case errors.Is(err, ErrTimeout):
		return http.StatusRequestTimeout
	default:
		return http.StatusInternalServerError
	}
}

// ErrorCodeFromDomain returns an error code string from domain error
func ErrorCodeFromDomain(err error) string {
	if err == nil {
		return ""
	}

	switch {
	case errors.Is(err, ErrNotFound):
		return "NOT_FOUND"
	case errors.Is(err, ErrAlreadyExists):
		return "ALREADY_EXISTS"
	case errors.Is(err, ErrInvalidArgument):
		return "INVALID_ARGUMENT"
	case errors.Is(err, ErrPermissionDenied):
		return "PERMISSION_DENIED"
	case errors.Is(err, ErrUnauthenticated):
		return "UNAUTHENTICATED"
	case errors.Is(err, ErrUnavailable):
		return "SERVICE_UNAVAILABLE"
	case errors.Is(err, ErrTimeout):
		return "TIMEOUT"
	default:
		return "INTERNAL_ERROR"
	}
}

// IsNotFound checks if the error is a NotFound error
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsAlreadyExists checks if the error is an AlreadyExists error
func IsAlreadyExists(err error) bool {
	return errors.Is(err, ErrAlreadyExists)
}

// IsInvalidArgument checks if the error is an InvalidArgument error
func IsInvalidArgument(err error) bool {
	return errors.Is(err, ErrInvalidArgument)
}

// IsUnavailable checks if the error is an Unavailable error
func IsUnavailable(err error) bool {
	return errors.Is(err, ErrUnavailable)
}

// IsAuthError checks if the error is an authentication/authorization error
func IsAuthError(err error) bool {
	return errors.Is(err, ErrUnauthenticated) || errors.Is(err, ErrPermissionDenied)
}
