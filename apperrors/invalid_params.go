package apperrors

import (
	"net/http"
)

type InvalidParamsError struct {
	baseError
}

func NewInvalidParamsError(resource string, err error) InvalidParamsError {
	return InvalidParamsError{baseError{
		Resource: resource,
		Cause:    err,
	}}
}

func (e InvalidParamsError) HTTPCode() int                { return http.StatusBadRequest }
func (e InvalidParamsError) HTTPResponse() map[string]any { return e.httpResponse("INVALID_PARAM") }
