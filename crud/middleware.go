package crud

import (
	"context"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/krsoninikhil/go-rest-kit/apperrors"
	"github.com/krsoninikhil/go-rest-kit/auth"
)

type ParentDao interface {
	GetByUserID(ctx context.Context, userID int) ([]Model, error)
}

func GinParentVerifier(parentDao ParentDao) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := strconv.Atoi(c.GetString(auth.CtxKeyUserID))
		if err != nil {
			handleError(c, apperrors.NewServerError(err))
			return
		}

		parentID, err := strconv.Atoi(c.Param("parentID"))
		if err != nil {
			handleError(c, apperrors.NewInvalidParamsError("url", err))
			return
		}

		parents, err := parentDao.GetByUserID(c, userID)
		if err != nil {
			handleError(c, apperrors.NewServerError(err))
			return
		}
		print("parents", parents)
		for _, parent := range parents {
			if parent.PK() == parentID {
				c.Next()
				return
			}
		}
		handleError(c, apperrors.NewPermissionError("parent"))
	}
}

func handleError(c *gin.Context, err apperrors.AppError) {
	log.Printf("err=%s parentID=%s userID=%s", err.Error(), c.Param("parentID"), c.GetString(auth.CtxKeyUserID))
	c.AbortWithStatusJSON(err.HTTPCode(), err.HTTPResponse())
}
