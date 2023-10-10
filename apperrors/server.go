package apperrors

import (
	"net/http"
)

// ServerError
type ServerError struct {
	baseError
}

func NewServerError(err error) ServerError {
	return ServerError{baseError{Err: err}}
}

func (e ServerError) HTTPCode() int                { return http.StatusInternalServerError }
func (e ServerError) HTTPResponse() map[string]any { return e.httpResponse("SERVER_ERROR") }
