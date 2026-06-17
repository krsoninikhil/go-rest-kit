package auth

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/krsoninikhil/go-rest-kit/apperrors"
)

type mockResetDao struct {
	userIDByEmail map[string]int
	passwordByID  map[int]string
	updated       map[int]string
}

func (m *mockResetDao) Create(ctx context.Context, u SigupInfo) (int, error) {
	return 0, errors.New("not implemented")
}
func (m *mockResetDao) GetByPhone(ctx context.Context, phone string) (int, error) {
	return 0, apperrors.NewNotFoundError("user")
}
func (m *mockResetDao) GetByEmail(ctx context.Context, email string) (int, error) {
	return 0, apperrors.NewNotFoundError("user")
}
func (m *mockResetDao) GetByEmailWithPassword(ctx context.Context, email string) (int, string, error) {
	userID, ok := m.userIDByEmail[email]
	if !ok {
		return 0, "", apperrors.NewNotFoundError("user")
	}
	return userID, m.passwordByID[userID], nil
}
func (m *mockResetDao) CreateWithEmailAndPassword(ctx context.Context, email, name, passwordHash, username, locale string) (int, error) {
	return 0, errors.New("not implemented")
}
func (m *mockResetDao) GetByUsername(ctx context.Context, username string) (int, error) {
	return 0, apperrors.NewNotFoundError("user")
}
func (m *mockResetDao) UpsertByEmail(ctx context.Context, oauthInfo OAuthUserInfo) (int, error) {
	return 0, errors.New("not implemented")
}
func (m *mockResetDao) SupportsUsername() bool { return false }
func (m *mockResetDao) SupportsPassword() bool { return true }
func (m *mockResetDao) SetPasswordHashByUserID(ctx context.Context, userID int, passwordHash string) error {
	if m.updated == nil {
		m.updated = map[int]string{}
	}
	m.updated[userID] = passwordHash
	return nil
}

type mockResetEmailSender struct {
	lastTo      string
	lastSubject string
	lastMessage string
}

func (m *mockResetEmailSender) SendEmail(to, subject, message string) error {
	m.lastTo = to
	m.lastSubject = subject
	m.lastMessage = message
	return nil
}

func newResetServiceForTest(dao UserDao, sender resetEmailSender) *Service {
	cfg := Config{
		SecretKey:                   "test-secret-key-123456",
		AccessTokenValiditySeconds:  3600,
		RefreshTokenValiditySeconds: 86400,
		Organisation:                "TestApp",
		Password: passwordConfig{
			ResetTTLSeconds: 600,
		},
	}
	svc := NewService(cfg, dao)
	return svc.WithPasswordReset(sender, "http://localhost:8081")
}

func TestPasswordResetRequestAndConfirm(t *testing.T) {
	dao := &mockResetDao{
		userIDByEmail: map[string]int{
			"user@example.com": 42,
		},
		passwordByID: map[int]string{
			42: "$2a$10$abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789",
		},
	}
	emailSender := &mockResetEmailSender{}
	svc := newResetServiceForTest(dao, emailSender)

	err := svc.RequestPasswordReset(context.Background(), "user@example.com")
	if err != nil {
		t.Fatalf("request reset returned error: %v", err)
	}
	if emailSender.lastTo != "user@example.com" {
		t.Fatalf("expected reset email recipient user@example.com, got %s", emailSender.lastTo)
	}
	if !strings.Contains(emailSender.lastMessage, "mode=reset&token=") {
		t.Fatalf("expected reset message to contain reset token link, got: %s", emailSender.lastMessage)
	}

	linkToken := extractTokenFromMessage(t, emailSender.lastMessage)
	if linkToken == "" {
		t.Fatalf("expected non-empty token")
	}

	err = svc.ConfirmPasswordReset(context.Background(), linkToken, "NewPassw0rd!")
	if err != nil {
		t.Fatalf("confirm reset returned error: %v", err)
	}
	if _, ok := dao.updated[42]; !ok {
		t.Fatalf("expected password update for user 42")
	}
}

func TestPasswordResetRequestIsGenericForUnknownEmail(t *testing.T) {
	dao := &mockResetDao{
		userIDByEmail: map[string]int{},
		passwordByID:  map[int]string{},
	}
	emailSender := &mockResetEmailSender{}
	svc := newResetServiceForTest(dao, emailSender)

	err := svc.RequestPasswordReset(context.Background(), "missing@example.com")
	if err != nil {
		t.Fatalf("expected generic success for unknown email, got error: %v", err)
	}
	if emailSender.lastTo != "" {
		t.Fatalf("expected no email sent for unknown account")
	}
}

func TestPasswordResetRequestSendsForOAuthOnlyUser(t *testing.T) {
	dao := &mockResetDao{
		userIDByEmail: map[string]int{
			"oauth@example.com": 99,
		},
		passwordByID: map[int]string{
			99: "",
		},
	}
	emailSender := &mockResetEmailSender{}
	svc := newResetServiceForTest(dao, emailSender)

	err := svc.RequestPasswordReset(context.Background(), "oauth@example.com")
	if err != nil {
		t.Fatalf("expected reset request success for oauth-only user, got error: %v", err)
	}
	if emailSender.lastTo != "oauth@example.com" {
		t.Fatalf("expected reset email recipient oauth@example.com, got %s", emailSender.lastTo)
	}
	linkToken := extractTokenFromMessage(t, emailSender.lastMessage)
	if linkToken == "" {
		t.Fatalf("expected non-empty token for oauth-only user")
	}
	err = svc.ConfirmPasswordReset(context.Background(), linkToken, "NewPassw0rd!")
	if err != nil {
		t.Fatalf("confirm reset returned error for oauth-only user: %v", err)
	}
	if _, ok := dao.updated[99]; !ok {
		t.Fatalf("expected password update for oauth-only user")
	}
}

func TestPasswordResetRejectsAccessTokenAudience(t *testing.T) {
	dao := &mockResetDao{
		userIDByEmail: map[string]int{
			"user@example.com": 42,
		},
		passwordByID: map[int]string{
			42: "existing",
		},
	}
	emailSender := &mockResetEmailSender{}
	svc := newResetServiceForTest(dao, emailSender)

	accessToken, err := generateJWTToken(svc.tokenSvc.NewAccessTokenClaims("42"), svc.config.SecretKey)
	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}

	err = svc.ConfirmPasswordReset(context.Background(), accessToken, "NewPassw0rd!")
	if err == nil {
		t.Fatalf("expected error when using non-reset token")
	}
}

func extractTokenFromMessage(t *testing.T, message string) string {
	t.Helper()
	parts := strings.Split(message, "token=")
	if len(parts) < 2 {
		return ""
	}
	tokenPart := strings.Split(parts[1], "\n")[0]
	token, err := url.QueryUnescape(strings.TrimSpace(tokenPart))
	if err != nil {
		t.Fatalf("failed to parse token from message: %v", err)
	}
	return token
}

func TestPasswordResetRejectsMalformedToken(t *testing.T) {
	dao := &mockResetDao{userIDByEmail: map[string]int{}, passwordByID: map[int]string{}}
	emailSender := &mockResetEmailSender{}
	svc := newResetServiceForTest(dao, emailSender)

	err := svc.ConfirmPasswordReset(context.Background(), "not.a.valid.token", "NewPassw0rd!")
	if err == nil {
		t.Fatalf("expected error for malformed token")
	}
	if !strings.Contains(fmt.Sprintf("%v", err), "token") {
		t.Fatalf("expected token-related error, got: %v", err)
	}
}
