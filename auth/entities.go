package auth

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
	return SigupInfo{
		Phone:    v.Phone,
		DialCode: v.DialCode,
		Country:  v.Country,
		Locale:   v.Locale,
	}
}
