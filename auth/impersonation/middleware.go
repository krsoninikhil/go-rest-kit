package impersonation

import (
	"github.com/gin-gonic/gin"
)

func ContextMiddleware(svc *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		_ = svc
		c.Next()
	}
}
