package auth

import (
	"log"
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
		time.Duration(conf.accessTokenValidity()),
		time.Duration(conf.refreshTokenValidity()),
		conf.SecretKey,
	))
}

func GinMiddleware(tokenSvc tokenSvc) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Printf("[auth] 401 %s %s: Authorization header is missing", c.Request.Method, c.Request.URL.Path)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			return
		}

		authHeaderParts := strings.Split(authHeader, " ")
		if len(authHeaderParts) != 2 || authHeaderParts[0] != "Bearer" {
			log.Printf("[auth] 401 %s %s: Invalid Authorization header format (expected 'Bearer <token>')", c.Request.Method, c.Request.URL.Path)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			return
		}

		tokenStr := authHeaderParts[1]
		parsedToken, err := tokenSvc.VerifyToken(tokenStr)
		if err != nil {
			log.Printf("[auth] 401 %s %s: Invalid token (verify failed): %v", c.Request.Method, c.Request.URL.Path, err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		sub, err := tokenSvc.ValidateAccessTokenClaims(parsedToken.Claims)
		if err != nil {
			log.Printf("[auth] 401 %s %s: Invalid token (claims): %v", c.Request.Method, c.Request.URL.Path, err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.Set(CtxKeyTokenClaims, parsedToken.Claims)
		c.Set(CtxKeyUserID, sub)
		c.Next()
	}
}

// OptionalGinStdMiddleware returns a middleware that parses JWT when present and sets user in context;
// it never aborts, so routes can be public but still know the viewer's user_id when a valid token is sent.
func OptionalGinStdMiddleware(conf Config) gin.HandlerFunc {
	tokenSvc := NewStdClaimsSvc(
		time.Duration(conf.accessTokenValidity()),
		time.Duration(conf.refreshTokenValidity()),
		conf.SecretKey,
	)
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}
		parsedToken, err := tokenSvc.VerifyToken(parts[1])
		if err != nil {
			c.Next()
			return
		}
		sub, err := tokenSvc.ValidateAccessTokenClaims(parsedToken.Claims)
		if err != nil {
			c.Next()
			return
		}
		c.Set(CtxKeyTokenClaims, parsedToken.Claims)
		c.Set(CtxKeyUserID, sub)
		c.Next()
	}
}

func UserID(c *gin.Context) int {
	val, exists := c.Get(CtxKeyUserID)
	if !exists || val == nil {

		return 0
	}
	switch v := val.(type) {
	case string:
		id, _ := strconv.Atoi(v)
		return id
	case int:
		return v
	case int64:
		return int(v)
	default:
		return 0
	}
}
