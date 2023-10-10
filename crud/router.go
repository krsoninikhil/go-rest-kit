package crud

import (
	"fmt"
	"reflect"

	"github.com/gin-gonic/gin"
)

type HandlerFunc func(*gin.Context, any)

type router struct {
	gin *gin.Engine
}

// NewRouter returns a gin router with method that can
// take handler function with request parameters binded to the
// argument instead of only gin.Context
//
// Not being used currently
func NewRouter(r *gin.Engine) *router {
	return &router{gin: r}
}

func (r *router) GET(path string, handler any) {
	req := reflect.New(extractBindToType(handler))
	r.gin.GET(path, func(c *gin.Context) {
		if err := bindRequest(c, &req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		handler.(HandlerFunc)(c, req)
	})
}

func (r *router) POST(path string, handler any) {
	req := reflect.New(extractBindToType(handler))
	r.gin.POST(path, func(c *gin.Context) {
		if err := bindRequest(c, &req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		handler.(HandlerFunc)(c, req)
	})
}

func (r *router) PATCH(path string, handler any) {
	req := reflect.New(extractBindToType(handler))
	r.gin.PATCH(path, func(c *gin.Context) {
		if err := bindRequest(c, &req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		handler.(HandlerFunc)(c, req)
	})
}

func (r *router) DELETE(path string, handler any) {
	req := reflect.New(extractBindToType(handler))
	r.gin.DELETE(path, func(c *gin.Context) {
		if err := bindRequest(c, &req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		handler.(HandlerFunc)(c, req)
	})
}

func extractBindToType(handler any) reflect.Type {
	handlerType := reflect.TypeOf(handler)
	fmt.Println("handlerType", handlerType, handlerType.Kind())
	if handlerType.Kind() != reflect.Func {
		panic("handler must be a function with signature func(*gin.Context, any)")
	}

	if handlerType.NumIn() != 2 {
		panic("handler must be a function with signature func(*gin.Context, any)")
	}

	// match first param to be *gin.Context type
	if handlerType.In(0) != reflect.TypeOf(&gin.Context{}) {
		panic("handler must be first argument as *gin.Context")
	}
	return handlerType.In(1)
}

func bindRequest(c *gin.Context, req any) error {
	if err := c.ShouldBindJSON(&req); err != nil {
		return err
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		return err
	}
	if err := c.ShouldBindUri(&req); err != nil {
		return err
	}
	if err := c.ShouldBindHeader(&req); err != nil {
		return err
	}
	return nil
}
