package request

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type (
	createHandlerFunc[R, S any] func(ctx *gin.Context, req R) (*S, error)
	getHandlerFunc[P, S any]    func(ctx *gin.Context, params P) (*S, error)
	updateHandlerFunc[P, R any] func(ctx *gin.Context, params P, req R) error
	deleteHandlerFunc[P any]    func(ctx *gin.Context, params P) error

	// createNestedHandlerFunc represents create handler for nested resource as they
	// require URI params even for create method
	createNestedHandlerFunc[P, R, S any] func(ctx *gin.Context, params P, req R) (*S, error)
)

func BindCreate[R, S any](handler createHandlerFunc[R, S]) gin.HandlerFunc {
	var req R
	return func(c *gin.Context) {
		if err := bindRequestParams[any, R](c, nil, &req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		res, err := handler(c, req)
		Respond(c, res, err)
	}
}

func BindGet[P, S any](handler getHandlerFunc[P, S]) gin.HandlerFunc {
	var params P
	return func(c *gin.Context) {
		if err := bindRequestParams[P, any](c, &params, nil); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		res, err := handler(c, params)
		Respond(c, res, err)
	}
}

func BindUpdate[P, R any](handler updateHandlerFunc[P, R]) gin.HandlerFunc {
	var (
		req    R
		params P
	)
	return func(c *gin.Context) {
		if err := bindRequestParams(c, &params, &req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err := handler(c, params, req)
		Respond(c, nil, err)
	}
}

func BindDelete[P any](handler deleteHandlerFunc[P]) gin.HandlerFunc {
	var params P
	return func(c *gin.Context) {
		if err := bindRequestParams[P, any](c, &params, nil); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		err := handler(c, params)
		Respond(c, nil, err)
	}
}

func BindNestedCreate[P, R, S any](handler createNestedHandlerFunc[P, R, S]) gin.HandlerFunc {
	var (
		req    R
		params P
	)
	return func(c *gin.Context) {
		if err := bindRequestParams(c, &params, &req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		res, err := handler(c, params, req)
		Respond(c, res, err)
	}
}
