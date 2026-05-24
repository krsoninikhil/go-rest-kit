package auth

import (
	"context"
	"testing"
)

type testRefreshDao struct{}

func (d *testRefreshDao) Create(ctx context.Context, u SigupInfo) (int, error)      { return 0, nil }
func (d *testRefreshDao) GetByPhone(ctx context.Context, phone string) (int, error) { return 0, nil }
func (d *testRefreshDao) GetByEmail(ctx context.Context, email string) (int, error) { return 0, nil }
func (d *testRefreshDao) GetByID(ctx context.Context, userID int) (string, string, error) {
	return "", "", nil
}
func (d *testRefreshDao) GetByEmailWithPassword(ctx context.Context, email string) (int, string, error) {
	return 0, "", nil
}
func (d *testRefreshDao) CreateWithEmailAndPassword(ctx context.Context, email, name, passwordHash, username, locale string) (int, error) {
	return 0, nil
}
func (d *testRefreshDao) GetByUsername(ctx context.Context, username string) (int, error) {
	return 0, nil
}
func (d *testRefreshDao) UpsertByEmail(ctx context.Context, oauthInfo OAuthUserInfo) (int, error) {
	return 0, nil
}
func (d *testRefreshDao) SupportsUsername() bool { return false }
func (d *testRefreshDao) SupportsPassword() bool { return true }
func (d *testRefreshDao) SetPasswordHashByUserID(ctx context.Context, userID int, passwordHash string) error {
	return nil
}

func TestRefreshTokenPreservesSubject(t *testing.T) {
	dao := &testRefreshDao{}
	svc := NewService(Config{
		SecretKey:                   "secret",
		AccessTokenValiditySeconds:  3600,
		RefreshTokenValiditySeconds: 86400,
	}, dao)

	refreshToken, err := generateJWTToken(svc.tokenSvc.NewRefreshTokenClaims("77"), "secret")
	if err != nil {
		t.Fatalf("generate refresh token failed: %v", err)
	}

	refreshed, err := svc.RefreshToken(context.Background(), refreshToken)
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
	claimsSvc := NewStdClaimsSvc(3600, 86400, "secret")
	parsed, err := claimsSvc.VerifyToken(refreshed.AccessToken)
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	sub, err := claimsSvc.ValidateAccessTokenClaims(parsed.Claims)
	if err != nil {
		t.Fatalf("validate claims failed: %v", err)
	}
	if sub != "77" {
		t.Fatalf("expected subject 77, got %s", sub)
	}
}
