package auth

import (
	"time"

	"github.com/krsoninikhil/go-rest-kit/integrations/twilio"
)

const (
	CtxKeyTokenClaims string = "tokenClaims"
	CtxKeyUserID      string = "userID"
	audienceLogin     string = "login"
	audienceRefresh   string = "refresh"
)

// Config holds the configuration for auth service
type Config struct {
	SecretKey                   string        `validate:"required" log:"-"`
	AccessTokenValiditySeconds  int           `validate:"required"`
	RefreshTokenValiditySeconds int           `validate:"required"`
	OTP                         otpConfig     `validate:"required"`
	Twilio                      twilio.Config `validate:"required"`
}

func (c Config) accessTokenValidity() time.Duration {
	return time.Duration(c.AccessTokenValiditySeconds) * time.Second
}

func (c Config) refreshTokenValidity() time.Duration {
	return time.Duration(c.RefreshTokenValiditySeconds) * time.Second
}

// otpConfig holds the configuration for OTP service
type otpConfig struct {
	ValiditySeconds   int `validate:"required"`
	MaxAttempts       int `validate:"required"`
	RetryAfterSeconds int `validate:"required"`
	Length            int `validate:"required"`
}

func (c otpConfig) validity() time.Duration {
	return time.Duration(c.ValiditySeconds) * time.Second
}
func (c otpConfig) retryAfter() time.Duration {
	return time.Duration(c.RetryAfterSeconds) * time.Second
}
