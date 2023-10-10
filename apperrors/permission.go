package apperrors

import (
	"errors"
	"net/http"
)

// PermissionError
type PermissionError struct {
	baseError
}

func NewPermissionError(resource string) PermissionError {
	return PermissionError{baseError{
		Resource: resource,
		Err:      errors.New("permission error"),
	}}
}

func (e PermissionError) HTTPCode() int                { return http.StatusUnauthorized }
func (e PermissionError) HTTPResponse() map[string]any { return e.httpResponse("PERMISSION_ERROR") }
