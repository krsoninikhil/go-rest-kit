package request

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"
)

const ctxKeyUserID = "UserID"

func UserID(ctx context.Context) int {
	val, ok := ctx.Value(ctxKeyUserID).(int)
	if !ok {
		return 0
	}
	return val
}

func PathParam(c *gin.Context, param string) (int, error) {
	val, err := strconv.ParseInt(c.Param(param), 10, 8)
	if err != nil {
		return 0, err
	}
	return int(val), err
}
