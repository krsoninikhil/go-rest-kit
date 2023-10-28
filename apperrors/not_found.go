package apperrors

import (
	"errors"
	"net/http"
)

// NotFoundError
type NotFoundError struct {
	baseError
}

func NewNotFoundError(resource string) NotFoundError {
	return NotFoundError{baseError{
		Resource: resource,
		Cause:    errors.New("not found"),
	}}
}

func (e NotFoundError) HTTPCode() int                { return http.StatusNotFound }
func (e NotFoundError) HTTPResponse() map[string]any { return e.httpResponse("NOT_FOUND") }
