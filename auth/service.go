package auth

import (
	"context"
	"fmt"

	"github.com/golang-jwt/jwt"
)

type UserDao interface {
	Create(ctx context.Context, phone string) (user, error)
}

type TokenSvc interface {
	NewAccessTokenClaims(subject string) jwt.Claims
	NewRefreshTokenClaims(subject string) jwt.Claims
	VerifyToken(token string, signingKey string) (*jwt.Token, error)
	ValidateAccessTokenClaims(claims jwt.Claims) (subject string, err error)
	ValiateRefreshTokenClaims(claims jwt.Claims) (subject string, err error)
}

type Service struct {
	config   Config
	userDao  UserDao
	tokenSvc TokenSvc
}

func NewService(config Config, userDao UserDao, tokenSvc TokenSvc) *Service {
	if tokenSvc == nil {
		tokenSvc = newClaimsSvc(config.accessTokenValidity(), config.refreshTokenValidity())
	}
	return &Service{
		config:   config,
		userDao:  userDao,
		tokenSvc: tokenSvc,
	}
}

func (s *Service) UpsertUser(ctx context.Context, phone string) (*Token, error) {
	u, err := s.userDao.Create(ctx, phone)
	if err != nil {
		return nil, fmt.Errorf("unable to create user: %v", err)
	}

	return s.generateToken(ctx, fmt.Sprintf("%d", u.ID))
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*Token, error) {
	token, err := s.tokenSvc.VerifyToken(refreshToken, s.config.SecretKey)
	if err != nil {
		return nil, fmt.Errorf("unable to verify refresh token: %v", err)
	}

	subject, err := s.tokenSvc.ValiateRefreshTokenClaims(token.Claims)
	if err != nil {
		return nil, fmt.Errorf("unable to validate refresh token claims: %v", err)
	}

	return s.generateToken(ctx, subject)
}

func (s *Service) generateToken(ctx context.Context, subject string) (*Token, error) {
	accessClaims := s.tokenSvc.NewAccessTokenClaims(subject)
	accessToken, err := generateJWTToken(ctx, accessClaims, s.config.SecretKey)
	if err != nil {
		return nil, fmt.Errorf("unable to generate access token: %v", err)
	}

	refreshClaims := s.tokenSvc.NewRefreshTokenClaims(subject)
	refreshToken, err := generateJWTToken(ctx, refreshClaims, s.config.SecretKey)
	if err != nil {
		return nil, fmt.Errorf("unable to generate refresh token: %v", err)
	}

	return &Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.config.accessTokenValidity().Seconds()),
	}, nil
}
