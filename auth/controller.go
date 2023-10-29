package auth

import (
	"context"

	"github.com/gin-gonic/gin"
)

// dependencies
type (
	OAuthProvider interface {
		Initiate()
		Verify()
	}
	OTPSvcI interface {
		Send(ctx context.Context, phone string) (OTPStatus, error)
		Verify(ctx context.Context, phone, otp string) error
	}
	AuthGSI interface {
		VerifyToken()
	}
	AuthService interface {
		UpsertUser(ctx context.Context, phone string) (*Token, error)
		RefreshToken(ctx context.Context, refreshToken string) (*Token, error)
	}
)

// schema
type (
	SendOTPRequest struct {
		Phone string
	}
	SendOTPResponse struct {
		RetryAfter  int `json:"retry_after"`
		AttemptLeft int `json:"attempt_left"`
	}

	VerifyOTPRequest struct {
		Phone string `json:"phone" binding:"required"`
		OTP   string `json:"otp" binding:"required,numeric"`
	}
	VerifyOTPResponse struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresAt    int64  `json:"expires_in"`
	}

	RefreshTokenRequest struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
)

// dto
type (
	OTPStatus struct {
		RetryAfter  int
		AttemptLeft int
	}
	Token struct {
		AccessToken  string
		RefreshToken string
		ExpiresIn    int64
	}
)

type Controller struct {
	authSvc AuthService
	otpSvc  OTPSvcI
	// googleSvc OAuthProvider
	// appleSvc  OAuthProvider
	// googleGSI AuthGSI
}

func (a *Controller) SendOTP(c *gin.Context, r SendOTPRequest) (*SendOTPResponse, error) {
	res, err := a.otpSvc.Send(c, r.Phone)
	return &SendOTPResponse{
		RetryAfter:  res.RetryAfter,
		AttemptLeft: res.AttemptLeft,
	}, err
}

func (a *Controller) VerifyOTP(c *gin.Context, r VerifyOTPRequest) (*VerifyOTPResponse, error) {
	err := a.otpSvc.Verify(c, r.Phone, r.OTP)
	if err != nil {
		return nil, err
	}

	res, err := a.authSvc.UpsertUser(c, r.Phone)
	return &VerifyOTPResponse{
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
		ExpiresAt:    res.ExpiresIn,
	}, err
}

func (a *Controller) RefreshToken(c *gin.Context, r RefreshTokenRequest) (*VerifyOTPResponse, error) {
	res, err := a.authSvc.RefreshToken(c, r.RefreshToken)
	return &VerifyOTPResponse{
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
		ExpiresAt:    res.ExpiresIn,
	}, err
}
