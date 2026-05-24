package impersonation

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/krsoninikhil/go-rest-kit/apperrors"
	"github.com/krsoninikhil/go-rest-kit/auth"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	authConfig  auth.Config
	permissions PermissionChecker
	userLookup  UserLookup
	auditSink   AuditSink
	claimsSvc   tokenClaimsService
}

type tokenClaimsService interface {
	NewAccessTokenClaims(subject string) jwt.Claims
	NewRefreshTokenClaims(subject string) jwt.Claims
}

func NewService(conf auth.Config, permissions PermissionChecker, userLookup UserLookup, auditSink AuditSink) *Service {
	return &Service{
		authConfig:  conf,
		permissions: permissions,
		userLookup:  userLookup,
		auditSink:   auditSink,
		claimsSvc: auth.NewStdClaimsSvc(
			time.Duration(conf.AccessTokenValiditySeconds)*time.Second,
			time.Duration(conf.RefreshTokenValiditySeconds)*time.Second,
			conf.SecretKey,
		),
	}
}

func (s *Service) Start(ctx context.Context, actorUserID int, req StartRequest, ip string) (*StartResponse, error) {
	if actorUserID <= 0 {
		return nil, apperrors.NewPermissionError("user")
	}
	if req.TargetUserID <= 0 {
		return nil, apperrors.NewInvalidParamsError("target_user_id", errors.New("target_user_id is required"))
	}
	if strings.TrimSpace(req.AdminPassword) == "" {
		return nil, apperrors.NewInvalidParamsError("admin_password", errors.New("admin_password is required"))
	}

	actorEmail, actorPasswordHash, err := s.userLookup.GetByID(ctx, actorUserID)
	if err != nil {
		return nil, err
	}
	if !s.permissions.CanImpersonate(actorUserID, actorEmail) {
		return nil, apperrors.NewPermissionError("impersonation")
	}
	if strings.TrimSpace(actorPasswordHash) == "" {
		return nil, apperrors.NewInvalidParamsError("admin_password", errors.New("password is not set; set password first"))
	}
	if err := bcrypt.CompareHashAndPassword([]byte(actorPasswordHash), []byte(req.AdminPassword)); err != nil {
		return nil, apperrors.NewPermissionError("credentials")
	}
	if _, _, err := s.userLookup.GetByID(ctx, req.TargetUserID); err != nil {
		return nil, err
	}
	token, err := s.issueToken(req.TargetUserID)
	if err != nil {
		return nil, apperrors.NewServerError(err)
	}
	s.log(ctx, AuditEvent{
		ActorID:      actorUserID,
		Action:       "IMPERSONATE",
		Category:     "ADMIN",
		Comment:      strings.TrimSpace(req.Comment),
		TargetUserID: req.TargetUserID,
		IP:           ip,
	})
	return token, nil
}

func (s *Service) Stop(ctx context.Context, targetUserID int, ip string) error {
	// Symbolic stop: frontend discards impersonation tokens locally.
	_ = ctx
	_ = targetUserID
	_ = ip
	return nil
}

func (s *Service) issueToken(targetUserID int) (*StartResponse, error) {
	accessToken, err := signToken(s.claimsSvc.NewAccessTokenClaims(strconv.Itoa(targetUserID)), s.authConfig.SecretKey)
	if err != nil {
		return nil, err
	}
	refreshToken, err := signToken(s.claimsSvc.NewRefreshTokenClaims(strconv.Itoa(targetUserID)), s.authConfig.SecretKey)
	if err != nil {
		return nil, err
	}
	return &StartResponse{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		ExpiresIn:        int64(s.authConfig.AccessTokenValiditySeconds),
		RefreshExpiresIn: int64(s.authConfig.RefreshTokenValiditySeconds),
	}, nil
}

func signToken(claims jwt.Claims, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("unable to sign token: %w", err)
	}
	return tokenStr, nil
}

func (s *Service) log(ctx context.Context, event AuditEvent) {
	if s.auditSink == nil {
		return
	}
	_ = s.auditSink.LogImpersonationEvent(ctx, event)
}
