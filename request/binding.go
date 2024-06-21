package request

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/krsoninikhil/go-rest-kit/apperrors"
	"github.com/pkg/errors"
)

type bindedHandlerFunc[Req, Res any] func(*gin.Context, Req) (*Res, error)

// BindAll binds request body, uri, query params and headers to R type
// and respond with S type
func BindAll[R, S any](handler bindedHandlerFunc[R, S]) gin.HandlerFunc {
	var req R
	return func(c *gin.Context) {
		if err := bindRequestParams[R](c, &req, &req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		res, err := handler(c, req)
		Respond(c, res, err)
	}
}

// bindRequestParams bind the request based on tags, order matters as
// Uri params could be mentioned required and validation would fail if
// looked in query param or elsewhere
func bindRequestParams[P, R any](c *gin.Context, params *P, req *R) error {
	if params != nil {
		if err := c.ShouldBindUri(params); err != nil {
			return errors.WithStack(err)
		}
		if err := c.ShouldBindQuery(params); err != nil {
			return errors.WithStack(err)
		}
		if err := c.ShouldBindHeader(params); err != nil {
			return errors.WithStack(err)
		}
	}

	if req != nil {
		if err := c.ShouldBindJSON(req); err != nil {
			return errors.Wrap(err, "error binding request body")
		}
	}
	return nil
}

// Respond sets the http status code and response to gin context
func Respond(c *gin.Context, res any, err error) {
	if err == nil {
		status := http.StatusOK
		if res == nil {
			status = http.StatusNoContent
		} else if c.Request.Method == http.MethodGet {
			status = http.StatusOK
		} else if c.Request.Method == http.MethodPost {
			status = http.StatusCreated
		}

		c.JSON(status, res)
		return
	}

	// handle error
	appError, ok := err.(apperrors.AppError)
	if !ok {
		appError = apperrors.NewServerError(err)
	}

	// log causes of server errors
	if severErr, ok := appError.(apperrors.ServerError); ok && severErr.Cause != nil {
		log.Printf("server error cause: %T: %+v\n", severErr, severErr.Cause)
	}

	c.JSON(appError.HTTPCode(), appError.HTTPResponse())
}
