package auth

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

type tokenSvc interface {
	VerifyToken(token string) (*jwt.Token, error)
	ValidateAccessTokenClaims(claims jwt.Claims) (string, error)
}

func GinStdMiddleware(conf Config) gin.HandlerFunc {
	return GinMiddleware(NewStdClaimsSvc(
		time.Duration(conf.AccessTokenValiditySeconds),
		time.Duration(conf.RefreshTokenValiditySeconds),
		conf.SecretKey,
	))
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

		sub, err := tokenSvc.ValidateAccessTokenClaims(parsedToken.Claims)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.Set(CtxKeyTokenClaims, parsedToken.Claims)
		c.Set(CtxKeyUserID, sub)
		c.Next()
	}
}

func UserID(c *gin.Context) int {
	userID, _ := strconv.Atoi(c.GetString(CtxKeyUserID))
	return userID
}
