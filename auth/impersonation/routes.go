package impersonation

import (
	"github.com/gin-gonic/gin"
	"github.com/krsoninikhil/go-rest-kit/request"
)

func RegisterRoutes(group *gin.RouterGroup, ctrl *Controller) {
	group.POST("/admin/impersonate", request.BindCreate(ctrl.Start))
	group.DELETE("/admin/impersonate", request.BindDelete(ctrl.Stop))
}
