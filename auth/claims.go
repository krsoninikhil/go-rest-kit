package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

type claimsSvc struct {
	accessTokenValidity  time.Duration
	refreshTokenValidity time.Duration
	signingKey           string
}

func NewStdClaimsSvc(accessTokenValidity, refreshTokenValidity time.Duration, signingKey string) *claimsSvc {
	return &claimsSvc{
		accessTokenValidity:  accessTokenValidity,
		refreshTokenValidity: refreshTokenValidity,
		signingKey:           signingKey,
	}
}

func (s *claimsSvc) NewAccessTokenClaims(subject string) jwt.Claims {
	return jwt.StandardClaims{
		Audience:  audienceLogin,
		ExpiresAt: time.Now().Add(s.accessTokenValidity).Unix(),
		IssuedAt:  time.Now().Unix(),
		Subject:   subject,
	}
}

func (s *claimsSvc) NewRefreshTokenClaims(subject string) jwt.Claims {
	return jwt.StandardClaims{
		Audience:  audienceRefresh,
		ExpiresAt: time.Now().Add(s.refreshTokenValidity).Unix(),
		IssuedAt:  time.Now().Unix(),
		Subject:   subject,
	}
}

func (s *claimsSvc) ValidateAccessTokenClaims(claims jwt.Claims) (string, error) {
	if claims.Valid() != nil {
		return "", fmt.Errorf("expired token")
	}

	stdClaims, ok := claims.(*jwt.StandardClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	if stdClaims.Audience != audienceLogin {
		return "", fmt.Errorf("invalid token audience")
	}
	return stdClaims.Subject, nil
}

func (s *claimsSvc) ValiateRefreshTokenClaims(claims jwt.Claims) (string, error) {
	if claims.Valid() != nil {
		return "", fmt.Errorf("expired token")
	}
	stdClaims, ok := claims.(*jwt.StandardClaims)
	if !ok {
		return "", fmt.Errorf("invalid token claims")
	}

	if stdClaims.Audience != audienceRefresh {
		return "", fmt.Errorf("invalid token audience")
	}
	return stdClaims.Subject, nil
}

func (s *claimsSvc) VerifyToken(token string) (*jwt.Token, error) {
	parsedToken, err := jwt.ParseWithClaims(token, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.signingKey), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %v", err)
	}

	if !parsedToken.Valid {
		return nil, fmt.Errorf("expired token")
	}
	return parsedToken, nil
}
