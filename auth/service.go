package auth

import (
	"context"
	"fmt"

	"github.com/golang-jwt/jwt"
	"github.com/krsoninikhil/go-rest-kit/apperrors"
)

type UserDao interface {
	Create(ctx context.Context, u SigupInfo) (userID int, err error)
	GetByPhone(ctx context.Context, phone string) (userID int, err error)
}

type TokenSvc interface {
	NewAccessTokenClaims(subject string) jwt.Claims
	NewRefreshTokenClaims(subject string) jwt.Claims
	VerifyToken(token string) (*jwt.Token, error)
	ValidateAccessTokenClaims(claims jwt.Claims) (subject string, err error)
	ValiateRefreshTokenClaims(claims jwt.Claims) (subject string, err error)
}

type Service struct {
	config   Config
	userDao  UserDao
	tokenSvc TokenSvc
}

func NewService(config Config, userDao UserDao) *Service {
	tokenSvc := NewStdClaimsSvc(
		config.accessTokenValidity(),
		config.refreshTokenValidity(),
		config.SecretKey,
	)
	return &Service{
		config:   config,
		userDao:  userDao,
		tokenSvc: tokenSvc,
	}
}

func (s *Service) UpsertUser(ctx context.Context, u SigupInfo) (*Token, error) {
	userID, err := s.userDao.GetByPhone(ctx, u.Phone)
	if err != nil {
		if _, ok := err.(apperrors.NotFoundError); ok {
			userID, err = s.userDao.Create(ctx, u)
			if err != nil {
				return nil, apperrors.NewServerError(fmt.Errorf("error creating user: %v", err))
			}
		} else {
			return nil, apperrors.NewServerError(fmt.Errorf("error getting user: %v", err))
		}
	}

	return s.generateToken(fmt.Sprintf("%d", userID))
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (*Token, error) {
	token, err := s.tokenSvc.VerifyToken(refreshToken)
	if err != nil {
		return nil, apperrors.NewInvalidParamsError("token", err)
	}

	subject, err := s.tokenSvc.ValiateRefreshTokenClaims(token.Claims)
	if err != nil {
		return nil, apperrors.NewInvalidParamsError("token", err)
	}

	return s.generateToken(subject)
}

func (s *Service) generateToken(subject string) (*Token, error) {
	accessClaims := s.tokenSvc.NewAccessTokenClaims(subject)
	accessToken, err := generateJWTToken(accessClaims, s.config.SecretKey)
	if err != nil {
		return nil, fmt.Errorf("unable to generate access token: %v", err)
	}

	refreshClaims := s.tokenSvc.NewRefreshTokenClaims(subject)
	refreshToken, err := generateJWTToken(refreshClaims, s.config.SecretKey)
	if err != nil {
		return nil, fmt.Errorf("unable to generate refresh token: %v", err)
	}

	return &Token{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		ExpiresIn:        int64(s.config.accessTokenValidity().Seconds()),
		RefreshExpiresIn: int64(s.config.refreshTokenValidity().Seconds()),
	}, nil
}
