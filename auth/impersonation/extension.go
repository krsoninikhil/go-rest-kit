package impersonation

import "github.com/gin-gonic/gin"

type Extension struct {
	Service    *Service
	Controller *Controller
}

func NewExtension(service *Service) *Extension {
	return &Extension{
		Service:    service,
		Controller: NewController(service),
	}
}

func (e *Extension) Middleware() gin.HandlerFunc {
	return ContextMiddleware(e.Service)
}
