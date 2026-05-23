package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strconv"
	"math/big"
	"net/mail"
	"net/url"
	"regexp"
	"strings"

	"github.com/golang-jwt/jwt"
	"github.com/krsoninikhil/go-rest-kit/apperrors"
	"golang.org/x/crypto/bcrypt"
)

var usernameSanitizer = regexp.MustCompile(`[^a-z0-9]+`)

type UserDao interface {
	Create(ctx context.Context, u SigupInfo) (userID int, err error)
	GetByPhone(ctx context.Context, phone string) (userID int, err error)
	GetByEmail(ctx context.Context, email string) (userID int, err error)
	GetByEmailWithPassword(ctx context.Context, email string) (userID int, passwordHash string, err error)
	CreateWithEmailAndPassword(ctx context.Context, email, name, passwordHash, username, locale string) (userID int, err error)
	GetByUsername(ctx context.Context, username string) (userID int, err error)
	UpsertByEmail(ctx context.Context, oauthInfo OAuthUserInfo) (userID int, err error)
	SupportsUsername() bool
	SupportsPassword() bool
	SetPasswordHashByUserID(ctx context.Context, userID int, passwordHash string) error
}

type resetEmailSender interface {
	SendEmail(to, subject, message string) error
}

type TokenSvc interface {
	NewAccessTokenClaims(subject string) jwt.Claims
	NewRefreshTokenClaims(subject string) jwt.Claims
	VerifyToken(token string) (*jwt.Token, error)
	ValidateAccessTokenClaims(claims jwt.Claims) (subject string, err error)
	ValiateRefreshTokenClaims(claims jwt.Claims) (subject string, err error)
}

type Service struct {
	config            Config
	userDao           UserDao
	tokenSvc          TokenSvc
	resetEmailSender  resetEmailSender
	passwordResetBase string
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

func (s *Service) WithPasswordReset(emailSender resetEmailSender, frontendBaseURL string) *Service {
	s.resetEmailSender = emailSender
	s.passwordResetBase = strings.TrimRight(strings.TrimSpace(frontendBaseURL), "/")
	return s
}

func (s *Service) UpsertUser(ctx context.Context, u SigupInfo) (*Token, error) {
	var (
		userID int
		err    error
	)
	switch {
	case u.Email != "":
		userID, err = s.userDao.GetByEmail(ctx, u.Email)
	case u.Phone != "":
		userID, err = s.userDao.GetByPhone(ctx, u.Phone)
	default:
		return nil, apperrors.NewInvalidParamsError("target", fmt.Errorf("phone or email is required"))
	}
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

func (s *Service) UpsertOAuthUser(ctx context.Context, oauthInfo OAuthUserInfo) (*Token, error) {
	oauthInfo.Email = strings.TrimSpace(strings.ToLower(oauthInfo.Email))
	userID, err := s.userDao.UpsertByEmail(ctx, oauthInfo)
	if err != nil {
		return nil, apperrors.NewServerError(fmt.Errorf("error upserting oauth user: %v", err))
	}

	return s.generateToken(fmt.Sprintf("%d", userID))
}

func (s *Service) PasswordSignupOrLogin(ctx context.Context, email, password, locale string) (*Token, error) {
	normalizedEmail := strings.TrimSpace(strings.ToLower(email))
	if normalizedEmail == "" {
		return nil, apperrors.NewInvalidParamsError("email", errors.New("email is required"))
	}
	if _, err := mail.ParseAddress(normalizedEmail); err != nil {
		return nil, apperrors.NewInvalidParamsError("email", errors.New("invalid email address"))
	}
	if len(password) < 8 {
		return nil, apperrors.NewInvalidParamsError("password", errors.New("password must be at least 8 characters"))
	}

	userID, passwordHash, err := s.userDao.GetByEmailWithPassword(ctx, normalizedEmail)
	if err != nil {
		if _, ok := err.(apperrors.NotFoundError); ok {
			hashedPassword, hashErr := hashPassword(password)
			if hashErr != nil {
				return nil, apperrors.NewServerError(hashErr)
			}
			name := deriveNameFromEmail(normalizedEmail)
			username := s.deriveAvailableUsername(ctx, normalizedEmail)
			userID, err = s.userDao.CreateWithEmailAndPassword(ctx, normalizedEmail, name, hashedPassword, username, locale)
			if err != nil {
				return nil, apperrors.NewServerError(fmt.Errorf("error creating password user: %v", err))
			}
			return s.generateToken(fmt.Sprintf("%d", userID))
		}
		return nil, apperrors.NewServerError(fmt.Errorf("error getting user by email: %v", err))
	}

	if strings.TrimSpace(passwordHash) == "" {
		return nil, apperrors.NewInvalidParamsError("password", errors.New("password is not set for this account; use Google sign-in or reset password"))
	}
	if verifyErr := verifyPassword(passwordHash, password); verifyErr != nil {
		return nil, apperrors.NewPermissionError("credentials")
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

func (s *Service) RequestPasswordReset(ctx context.Context, email string) error {
	if !s.userDao.SupportsPassword() {
		return nil
	}
	if s.resetEmailSender == nil || s.passwordResetBase == "" {
		return apperrors.NewServerError(errors.New("password reset is not configured"))
	}

	userID, passwordHash, err := s.userDao.GetByEmailWithPassword(ctx, email)
	if err != nil {
		if _, ok := err.(apperrors.NotFoundError); ok {
			return nil
		}
		return apperrors.NewServerError(fmt.Errorf("error looking up reset email: %v", err))
	}
	if strings.TrimSpace(passwordHash) == "" {
		// OAuth-only users: keep generic success, no reset email.
		return nil
	}

	token, err := s.generatePasswordResetToken(userID)
	if err != nil {
		return apperrors.NewServerError(err)
	}
	resetURL := fmt.Sprintf("%s/login?mode=reset&token=%s", s.passwordResetBase, url.QueryEscape(token))
	brand := s.config.brandName()
	message := fmt.Sprintf("Use this link to reset your %s password: %s\n\nThis link expires in %d minutes.", brand, resetURL, int(s.config.passwordResetTTL().Minutes()))
	if err := s.resetEmailSender.SendEmail(email, fmt.Sprintf("Reset your %s password", brand), message); err != nil {
		return apperrors.NewServerError(fmt.Errorf("error sending password reset email: %v", err))
	}
	return nil
}

func (s *Service) ConfirmPasswordReset(ctx context.Context, token, newPassword string) error {
	if !s.userDao.SupportsPassword() {
		return apperrors.NewServerError(errors.New("password reset is not supported"))
	}
	if len(strings.TrimSpace(newPassword)) < 8 {
		return apperrors.NewInvalidParamsError("new_password", errors.New("password must be at least 8 characters"))
	}
	userID, err := s.validatePasswordResetToken(strings.TrimSpace(token))
	if err != nil {
		return err
	}
	passwordHash, err := hashPassword(strings.TrimSpace(newPassword))
	if err != nil {
		return apperrors.NewServerError(err)
	}
	if err := s.userDao.SetPasswordHashByUserID(ctx, userID, passwordHash); err != nil {
		return apperrors.NewServerError(fmt.Errorf("error updating password: %v", err))
	}
	return nil
}

// CheckUsernameAvailable returns true if the username is available for the given user (not taken, or taken only by excludeUserID).
func (s *Service) CheckUsernameAvailable(ctx context.Context, username string, excludeUserID int) (bool, error) {
	if username == "" {
		return false, nil
	}
	userID, err := s.userDao.GetByUsername(ctx, username)
	if err != nil {
		if _, ok := err.(apperrors.NotFoundError); ok {
			return true, nil
		}
		return false, err
	}
	return userID == excludeUserID, nil
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

func generateJWTToken(claims jwt.Claims, secretKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", fmt.Errorf("unable to generate jwt token: %v", err)
	}
	return tokenStr, nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func verifyPassword(passwordHash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
}

func deriveNameFromEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		return "User"
	}
	localPart := strings.TrimSpace(parts[0])
	localPart = strings.ReplaceAll(localPart, ".", " ")
	localPart = strings.ReplaceAll(localPart, "_", " ")
	localPart = strings.ReplaceAll(localPart, "-", " ")
	segments := strings.Fields(localPart)
	for i := range segments {
		segment := segments[i]
		if segment == "" {
			continue
		}
		segments[i] = strings.ToUpper(segment[:1]) + strings.ToLower(segment[1:])
	}
	if len(segments) == 0 {
		return "User"
	}
	return strings.Join(segments, " ")
}

func (s *Service) deriveAvailableUsername(ctx context.Context, email string) string {
	if !s.userDao.SupportsUsername() {
		return ""
	}

	base := deriveUsernameBase(email)
	if isUsernameAvailable(ctx, s.userDao, base) {
		return base
	}

	for attempt := 0; attempt < 5; attempt++ {
		suffix, err := randomIntBetween(1, 1000)
		if err != nil {
			continue
		}
		candidate := withNumericSuffix(base, suffix, 50)
		if isUsernameAvailable(ctx, s.userDao, candidate) {
			return candidate
		}
	}

	// Best effort only: skip setting username when no available candidate is found.
	return ""
}

func deriveUsernameBase(email string) string {
	parts := strings.Split(strings.ToLower(strings.TrimSpace(email)), "@")
	local := ""
	if len(parts) > 0 {
		local = parts[0]
	}
	if local == "" {
		local = "user"
	}
	base := usernameSanitizer.ReplaceAllString(local, "")
	if base == "" {
		base = "user"
	}
	if len(base) > 50 {
		base = base[:50]
	}
	return base
}

func withNumericSuffix(base string, suffix int, maxLen int) string {
	suffixText := fmt.Sprintf("%d", suffix)
	if len(base)+len(suffixText) <= maxLen {
		return base + suffixText
	}
	if maxLen <= len(suffixText) {
		return suffixText[:maxLen]
	}
	return base[:maxLen-len(suffixText)] + suffixText
}

func randomIntBetween(min, max int) (int, error) {
	if max < min {
		return 0, errors.New("invalid random bounds")
	}
	span := max - min + 1
	n, err := rand.Int(rand.Reader, big.NewInt(int64(span)))
	if err != nil {
		return 0, err
	}
	return min + int(n.Int64()), nil
}

func isUsernameAvailable(ctx context.Context, dao UserDao, username string) bool {
	_, err := dao.GetByUsername(ctx, username)
	if err != nil {
		_, notFound := err.(apperrors.NotFoundError)
		return notFound
	}
	return false
}

func (s *Service) generatePasswordResetToken(userID int) (string, error) {
	claims := jwt.StandardClaims{
		Subject:   strconv.Itoa(userID),
		Audience:  audienceReset,
		IssuedAt:  jwt.TimeFunc().Unix(),
		ExpiresAt: jwt.TimeFunc().Add(s.config.passwordResetTTL()).Unix(),
	}
	return generateJWTToken(claims, s.config.SecretKey)
}

func (s *Service) validatePasswordResetToken(token string) (int, error) {
	parsedToken, err := s.tokenSvc.VerifyToken(strings.TrimSpace(token))
	if err != nil {
		return 0, apperrors.NewInvalidParamsError("token", errors.New("invalid or expired token"))
	}
	claims, ok := parsedToken.Claims.(*jwt.StandardClaims)
	if !ok {
		return 0, apperrors.NewInvalidParamsError("token", errors.New("invalid token claims"))
	}
	if claims.Audience != audienceReset {
		return 0, apperrors.NewInvalidParamsError("token", errors.New("invalid token audience"))
	}
	userID, convErr := strconv.Atoi(claims.Subject)
	if convErr != nil || userID <= 0 {
		return 0, apperrors.NewInvalidParamsError("token", errors.New("invalid token subject"))
	}
	return userID, nil
}
