package auth

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"github.com/krsoninikhil/go-rest-kit/apperrors"
)

const (
	CtxKeyTokenClaims string = "tokenClaims"
	CtxKeyUserID      string = "userID"
	audienceLogin     string = "login"
	audienceRefresh   string = "refresh"

	// otp for config.TestPhone to allow app reviews
	testOTP = "000000"
)

// schema
type (
	SendOTPRequest struct {
		// Deprecated: use target with channel="sms".
		Phone    string `json:"phone"`
		Target   string `json:"target"`
		Channel  string `json:"channel"`
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

	OAuthAuthRequest struct {
		Code     string `json:"code" binding:"required"`
		Provider string `json:"provider" binding:"required,oneof=google twitter linkedin"` // Add more providers as needed
		Locale   string `json:"locale"`
	}
	OAuthAuthResponse struct {
		AccessToken      string `json:"access_token"`
		RefreshToken     string `json:"refresh_token"`
		ExpiresIn        int64  `json:"expires_in"`
		RefreshExpiresIn int64  `json:"refresh_expires_in"`
	}

	// UsernameCheckParam is the path param for GET /username/check/:username (protected)
	UsernameCheckParam struct {
		Username string `uri:"username" binding:"required,min=3,max=50,alphanum"`
	}
	UsernameCheckResponse struct {
		Available bool `json:"available"`
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
		Email    string
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
	// OAuthUserInfo is a generic structure for OAuth user information across all providers
	OAuthUserInfo struct {
		Email      string
		Name       string
		Picture    string
		Locale     string
		ProviderID string // Provider-specific user ID (Google Sub, Twitter ID, etc.)
		Provider   string // Provider name: "google", "twitter", "linkedin", etc.
	}
)

func (v *VerifyOTPRequest) toSigupInfo() SigupInfo {
	target := v.OTPDestination()
	channel := v.OTPChannel()
	phone := ""
	email := ""
	if channel == OTPChannelSMS {
		phone = target
	}
	if channel == OTPChannelEmail {
		email = target
	}
	return SigupInfo{
		Phone:    phone,
		Email:    email,
		DialCode: v.DialCode,
		Country:  v.Country,
		Locale:   v.Locale,
	}
}

func (r SendOTPRequest) OTPDestination() string {
	target := strings.TrimSpace(r.Target)
	if target != "" {
		return target
	}
	return strings.TrimSpace(r.Phone)
}

func (r SendOTPRequest) OTPChannel() string {
	channel := strings.TrimSpace(strings.ToLower(r.Channel))
	if channel == "" {
		return OTPChannelSMS
	}
	return channel
}

func (r SendOTPRequest) resolveOTPInputs() (string, string, error) {
	target := r.OTPDestination()
	channel := r.OTPChannel()
	if target == "" {
		return "", "", apperrors.NewInvalidParamsError("target", errors.New("phone or target is required"))
	}
	switch channel {
	case OTPChannelSMS:
		if err := validatePhone(target); err != nil {
			return "", "", apperrors.NewInvalidParamsError("phone", err)
		}
	case OTPChannelEmail:
		if _, err := mail.ParseAddress(target); err != nil {
			return "", "", apperrors.NewInvalidParamsError("email", errors.New("invalid email address"))
		}
	default:
		return "", "", apperrors.NewInvalidParamsError("channel", fmt.Errorf("unsupported channel: %s", channel))
	}
	return target, channel, nil
}
