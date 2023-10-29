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

type OTPConfig struct {
	ValiditySeconds int `validate:"required"`
	MaxAttempts     int `validate:"required"`
	RetryAfter      int `validate:"required"`
	Length          int `validate:"required"`
}

func (c OTPConfig) validity() time.Duration {
	return time.Duration(c.ValiditySeconds) * time.Second
}
func (c OTPConfig) retryAfter() time.Duration {
	return time.Duration(c.RetryAfter) * time.Second
}

type Config struct {
	SecretKey                   string `validate:"required" log:"-"`
	AccessTokenValiditySeconds  int    `validate:"required"`
	RefreshTokenValiditySeconds int    `validate:"required"`
	OTPConfig                   OTPConfig
}

func (c Config) accessTokenValidity() time.Duration {
	return time.Duration(c.AccessTokenValiditySeconds) * time.Second
}

func (c Config) refreshTokenValidity() time.Duration {
	return time.Duration(c.RefreshTokenValiditySeconds) * time.Second
}
