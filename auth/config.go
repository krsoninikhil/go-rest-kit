package auth

import (
	"time"

	"github.com/krsoninikhil/go-rest-kit/integrations/fast2sms"
	"github.com/krsoninikhil/go-rest-kit/integrations/twilio"
)

// Config holds the configuration for auth service
type Config struct {
	SecretKey                   string          `validate:"required" log:"-"`
	AccessTokenValiditySeconds  int             `validate:"required"`
	RefreshTokenValiditySeconds int             `validate:"required"`
	OTP                         otpConfig       `validate:"required"`
	Twilio                      twilio.Config   `validate:"required"`
	Fast2SMS                    fast2sms.Config `validate:"required"`
	OAuthGoogle                 OAuthConfig
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
	TestPhone         string
}

func (c otpConfig) validity() time.Duration {
	return time.Duration(c.ValiditySeconds) * time.Second
}
func (c otpConfig) retryAfter() time.Duration {
	return time.Duration(c.RetryAfterSeconds) * time.Second
}

// OAuthConfig holds the configuration for OAuth providers (Google, Twitter, LinkedIn, etc.)
type OAuthConfig struct {
	ClientID     string `validate:"required"`
	ClientSecret string `validate:"required" log:"-"`
	RedirectURL  string `validate:"required"`
}
