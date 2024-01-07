package request

import (
	"context"

	"github.com/gin-gonic/gin"
)

type Context struct {
	userID int
	Gin    *gin.Context

	context.Context
}

func (ctx *Context) SetUserID(userID int) {
	ctx.userID = userID
}

func (ctx *Context) UserID() int {
	return ctx.userID
}
