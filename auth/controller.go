package auth

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
)

// dependencies
type (
	OAuthProvider interface {
		Initiate()
		Verify()
	}
	OTPSvcI interface {
		Send(ctx context.Context, phone string) (*OTPStatus, error)
		Verify(ctx context.Context, phone, otp string) error
	}
	AuthGSI interface {
		VerifyToken()
	}
	AuthService interface {
		UpsertUser(ctx context.Context, u SigupInfo) (*Token, error)
		RefreshToken(ctx context.Context, refreshToken string) (*Token, error)
	}
	LocalSvc interface {
		GetCountryInfo(ctx context.Context, locale string) (*CountryInfoSource, error)
	}
)

type Controller struct {
	authSvc   AuthService
	otpSvc    OTPSvcI
	localeSvc LocalSvc
	// googleSvc OAuthProvider
	// appleSvc  OAuthProvider
	// googleGSI AuthGSI
}

func NewController(authSvc AuthService, otpSvc OTPSvcI, cacheClient cacheClient) *Controller {
	return &Controller{
		authSvc:   authSvc,
		otpSvc:    otpSvc,
		localeSvc: NewLocaleSvc(cacheClient),
	}
}

func (a *Controller) SendOTP(c *gin.Context, r SendOTPRequest) (*SendOTPResponse, error) {
	log.Printf("auth: sending otp request=%+v", r)
	res, err := a.otpSvc.Send(c, r.Phone)
	if err != nil {
		return nil, err
	}
	log.Printf("auth: otp sent successfully")

	return &SendOTPResponse{
		RetryAfter:  res.RetryAfter,
		AttemptLeft: res.AttemptLeft,
	}, nil
}

func (a *Controller) VerifyOTP(c *gin.Context, r VerifyOTPRequest) (*VerifyOTPResponse, error) {
	err := a.otpSvc.Verify(c, r.Phone, r.OTP)
	if err != nil {
		return nil, err
	}

	res, err := a.authSvc.UpsertUser(c, r.toSigupInfo())
	if err != nil {
		return nil, err
	}

	return &VerifyOTPResponse{
		AccessToken:      res.AccessToken,
		RefreshToken:     res.RefreshToken,
		ExpiresIn:        res.ExpiresIn,
		RefreshExpiresIn: res.RefreshExpiresIn,
	}, nil
}

func (a *Controller) RefreshToken(c *gin.Context, r RefreshTokenRequest) (*VerifyOTPResponse, error) {
	res, err := a.authSvc.RefreshToken(c, r.RefreshToken)
	if err != nil {
		return nil, err
	}

	return &VerifyOTPResponse{
		AccessToken:      res.AccessToken,
		RefreshToken:     res.RefreshToken,
		ExpiresIn:        res.ExpiresIn,
		RefreshExpiresIn: res.RefreshExpiresIn,
	}, nil
}

func (a *Controller) CountryInfo(c *gin.Context, r CountryInfoRequest) (*CountryInfoResponse, error) {
	country, err := a.localeSvc.GetCountryInfo(c, r.Apha2Code)
	if err != nil {
		return nil, err
	}

	res := CountryInfoResponse(*country)
	return &res, nil
}
