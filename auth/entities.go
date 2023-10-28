package auth

import "time"

type (
	CtxKey string
)

type user struct {
	ID    int
	Phone string
}

const (
	CtxKeyTokenClaims CtxKey = "tokenClaims"
	audienceLogin     string = "login"
	audienceRefresh   string = "refresh"
)

type Config struct {
	SecretKey                   string `validate:"required" log:"-"`
	AccessTokenValiditySeconds  int    `validate:"required"`
	RefreshTokenValiditySeconds int    `validate:"required"`
}

func (c Config) accessTokenValidity() time.Duration {
	return time.Duration(c.AccessTokenValiditySeconds) * time.Second
}

func (c Config) refreshTokenValidity() time.Duration {
	return time.Duration(c.RefreshTokenValiditySeconds) * time.Second
}
