package impersonation

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/krsoninikhil/go-rest-kit/auth"
)

type Controller struct {
	svc *Service
}

func NewController(svc *Service) *Controller {
	return &Controller{svc: svc}
}

func (c *Controller) Start(ctx *gin.Context, req StartRequest) (*StartResponse, error) {
	req.Comment = strings.TrimSpace(req.Comment)
	return c.svc.Start(ctx.Request.Context(), auth.UserID(ctx), req, ctx.ClientIP())
}

func (c *Controller) Stop(ctx *gin.Context, _ struct{}) error {
	return c.svc.Stop(ctx.Request.Context(), auth.UserID(ctx), ctx.ClientIP())
}
