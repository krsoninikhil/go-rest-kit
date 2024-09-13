package auth

import (
	"time"

	"github.com/krsoninikhil/go-rest-kit/integrations/fast2sms"
	"github.com/krsoninikhil/go-rest-kit/integrations/twilio"
)

const (
	CtxKeyTokenClaims string = "tokenClaims"
	CtxKeyUserID      string = "userID"
	audienceLogin     string = "login"
	audienceRefresh   string = "refresh"

	// otp for config.TestPhone to allow app reviews
	testOTP = "000000"
)

// Config holds the configuration for auth service
type Config struct {
	SecretKey                   string          `validate:"required" log:"-"`
	AccessTokenValiditySeconds  int             `validate:"required"`
	RefreshTokenValiditySeconds int             `validate:"required"`
	OTP                         otpConfig       `validate:"required"`
	Twilio                      twilio.Config   `validate:"required"`
	Fast2SMS                    fast2sms.Config `validate:"required"`
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

// schema
type (
	SendOTPRequest struct {
		Phone    string `json:"phone" binding:"required"`
		DialCode string `json:"dial_code"`
		Country  string `json:"country"`
		Locale   string `json:"locale"`
	}
	SendOTPResponse struct {
		RetryAfter  int `json:"retry_after"`
		AttemptLeft int `json:"attempt_left"`
	}

	VerifyOTPRequest struct {
		SendOTPRequest
		OTP string `json:"otp" binding:"required,numeric"`
	}
	VerifyOTPResponse struct {
		AccessToken      string `json:"access_token"`
		RefreshToken     string `json:"refresh_token"`
		ExpiresIn        int64  `json:"expires_in"`
		RefreshExpiresIn int64  `json:"refresh_expires_in"`
	}

	RefreshTokenRequest struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	CountryInfoRequest struct {
		Apha2Code string `uri:"alpha2Code" binding:"required"`
	}
	CountryInfoResponse struct {
		Name        string `json:"name"`
		Nationality string `json:"nationality"`
		Code        string `json:"code"`
		DialCode    string `json:"dial_code"`
	}
)

// dto
type (
	OTPStatus struct {
		RetryAfter  int
		AttemptLeft int
	}
	Token struct {
		AccessToken      string
		RefreshToken     string
		ExpiresIn        int64
		RefreshExpiresIn int64
	}
	SigupInfo struct {
		Phone    string
		DialCode string
		Country  string
		Locale   string
	}
	CountryInfoSource struct {
		Name        string `json:"en_short_name"`
		Nationality string `json:"nationality"`
		Code        string `json:"alpha_2_code"`
		DialCode    string `json:"dial_code"`
	}
)

func (v *VerifyOTPRequest) toSigupInfo() SigupInfo {
	return SigupInfo{
		Phone:    v.Phone,
		DialCode: v.DialCode,
		Country:  v.Country,
		Locale:   v.Locale,
	}
}
