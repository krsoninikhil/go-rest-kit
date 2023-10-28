package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

type tokenSvc interface {
	VerifyToken(token string) (*jwt.Token, error)
	ValidateAccessTokenClaims(claims jwt.Claims) error
}

func GinMiddleware(tokenSvc tokenSvc) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			return
		}

		authHeaderParts := strings.Split(authHeader, " ")
		if len(authHeaderParts) != 2 || authHeaderParts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			return
		}

		parsedToken, err := tokenSvc.VerifyToken(authHeaderParts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		if err := tokenSvc.ValidateAccessTokenClaims(parsedToken.Claims); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		ctx := context.WithValue(c.Request.Context(), CtxKeyTokenClaims, parsedToken.Claims)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
