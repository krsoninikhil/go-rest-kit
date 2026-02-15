package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/krsoninikhil/go-rest-kit/apperrors"
	"github.com/pkg/errors"
)

const (
	googleTokenURL     = "https://oauth2.googleapis.com/token"
	googleUserInfoURL  = "https://www.googleapis.com/oauth2/v2/userinfo"
	googleProviderName = "google"
)

type googleOAuthProvider struct {
	config OAuthConfig
	client *http.Client
}

// NewGoogleOAuthProvider creates a new Google OAuth provider
func NewGoogleOAuthProvider(config OAuthConfig) OAuthProvider {
	return &googleOAuthProvider{
		config: config,
		client: &http.Client{},
	}
}

// ProviderName returns the provider name
func (p *googleOAuthProvider) ProviderName() string {
	return googleProviderName
}

// googleTokenResponse represents the response from Google's token endpoint
type googleTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
	IDToken      string `json:"id_token"`
}

// googleUserInfoResponse represents the response from Google's userinfo endpoint
type googleUserInfoResponse struct {
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Locale        string `json:"locale"`
}

// ExchangeCode exchanges the authorization code for user information
func (p *googleOAuthProvider) ExchangeCode(ctx context.Context, code string) (*OAuthUserInfo, error) {
	// Step 1: Exchange code for access token
	tokenResp, err := p.exchangeCodeForToken(ctx, code)
	if err != nil {
		return nil, errors.Wrap(err, "failed to exchange code for token")
	}

	// Step 2: Get user info using access token
	userInfo, err := p.getUserInfo(ctx, tokenResp.AccessToken)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user info")
	}

	return userInfo, nil
}

// exchangeCodeForToken exchanges the authorization code for an access token
func (p *googleOAuthProvider) exchangeCodeForToken(ctx context.Context, code string) (*googleTokenResponse, error) {
	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", p.config.ClientID)
	data.Set("client_secret", p.config.ClientSecret)
	data.Set("redirect_uri", p.config.RedirectURL)
	data.Set("grant_type", "authorization_code")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, googleTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, apperrors.NewServerError(errors.Wrap(err, "failed to create token request"))
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, apperrors.NewServerError(errors.Wrap(err, "failed to send token request"))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, apperrors.NewServerError(errors.Wrap(err, "failed to read token response"))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, apperrors.NewInvalidParamsError("code",
			fmt.Errorf("google token exchange failed with status %d: %s", resp.StatusCode, string(body)))
	}

	var tokenResp googleTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, apperrors.NewServerError(errors.Wrap(err, "failed to parse token response"))
	}

	return &tokenResp, nil
}

// getUserInfo retrieves user information from Google using the access token
func (p *googleOAuthProvider) getUserInfo(ctx context.Context, accessToken string) (*OAuthUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, googleUserInfoURL, nil)
	if err != nil {
		return nil, apperrors.NewServerError(errors.Wrap(err, "failed to create userinfo request"))
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, apperrors.NewServerError(errors.Wrap(err, "failed to send userinfo request"))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, apperrors.NewServerError(errors.Wrap(err, "failed to read userinfo response"))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, apperrors.NewServerError(
			fmt.Errorf("google userinfo request failed with status %d: %s", resp.StatusCode, string(body)))
	}

	var userInfoResp googleUserInfoResponse
	if err := json.Unmarshal(body, &userInfoResp); err != nil {
		return nil, apperrors.NewServerError(errors.Wrap(err, "failed to parse userinfo response"))
	}

	return &OAuthUserInfo{
		Email:      userInfoResp.Email,
		Name:       userInfoResp.Name,
		Picture:    userInfoResp.Picture,
		Locale:     userInfoResp.Locale,
		ProviderID: userInfoResp.Sub,
		Provider:   googleProviderName,
	}, nil
}
