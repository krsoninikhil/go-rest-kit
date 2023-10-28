package apperrors

import (
	"net/http"
)

// ServerError
type ServerError struct {
	baseError
}

func NewServerError(err error) ServerError {
	return ServerError{baseError{Cause: err}}
}

func (e ServerError) HTTPCode() int { return http.StatusInternalServerError }
func (e ServerError) HTTPResponse() map[string]any {
	res := e.httpResponse("SERVER_ERROR")
	res["detail"] = "internal server errror"
	return res
}
