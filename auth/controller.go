package auth

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/krsoninikhil/go-rest-kit/apperrors"
)

// dependencies
type (
	OTPSvcI interface {
		Send(ctx context.Context, phone string) (*OTPStatus, error)
		Verify(ctx context.Context, phone, otp string) error
	}
	AuthService interface {
		UpsertUser(ctx context.Context, u SigupInfo) (*Token, error)
		UpsertOAuthUser(ctx context.Context, oauthInfo OAuthUserInfo) (*Token, error)
		RefreshToken(ctx context.Context, refreshToken string) (*Token, error)
	}
	LocalSvc interface {
		GetCountryInfo(ctx context.Context, locale string) (*CountryInfoSource, error)
	}
)

type Controller struct {
	authSvc        AuthService
	otpSvc         OTPSvcI
	oauthProviders map[string]OAuthProvider // Map of provider name to provider implementation
	localeSvc      LocalSvc
}

func NewController(authSvc AuthService, otpSvc OTPSvcI, cacheClient cacheClient) *Controller {
	return &Controller{
		authSvc:        authSvc,
		otpSvc:         otpSvc,
		oauthProviders: make(map[string]OAuthProvider),
		localeSvc:      NewLocaleSvc(cacheClient),
	}
}

// WithOAuthProvider adds an OAuth provider to the controller
func (c *Controller) WithOAuthProvider(provider OAuthProvider) *Controller {
	c.oauthProviders[provider.ProviderName()] = provider
	return c
}

// WithOAuthProviders adds multiple OAuth providers to the controller
func (c *Controller) WithOAuthProviders(providers ...OAuthProvider) *Controller {
	for _, provider := range providers {
		c.oauthProviders[provider.ProviderName()] = provider
	}
	return c
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

func (a *Controller) OAuthAuth(c *gin.Context, r OAuthAuthRequest) (*OAuthAuthResponse, error) {
	// Get the appropriate OAuth provider
	provider, exists := a.oauthProviders[r.Provider]
	if !exists {
		log.Printf("auth: oauth provider '%s' not configured", r.Provider)
		return nil, apperrors.NewInvalidParamsError("provider",
			fmt.Errorf("provider '%s' not configured or not supported", r.Provider))
	}

	log.Printf("auth: exchanging %s auth code", r.Provider)
	oauthUserInfo, err := provider.ExchangeCode(c, r.Code)
	if err != nil {
		return nil, err
	}

	// Set locale from request if provided
	if r.Locale != "" {
		oauthUserInfo.Locale = r.Locale
	}

	log.Printf("auth: upserting %s user: %s", r.Provider, oauthUserInfo.Email)
	res, err := a.authSvc.UpsertOAuthUser(c, *oauthUserInfo)
	if err != nil {
		return nil, err
	}

	return &OAuthAuthResponse{
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
