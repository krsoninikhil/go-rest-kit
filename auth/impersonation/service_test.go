package impersonation

import (
	"context"
	"testing"

	"github.com/krsoninikhil/go-rest-kit/auth"
	"golang.org/x/crypto/bcrypt"
)

type allowChecker struct{}

func (a allowChecker) CanImpersonate(actorUserID int, actorEmail string) bool { return true }

type fakeUserLookup struct {
	email        string
	password     string
	existsByUser map[int]bool
}

func (f fakeUserLookup) GetByID(ctx context.Context, userID int) (string, string, error) {
	if f.existsByUser[userID] {
		return f.email, f.password, nil
	}
	return "", "", context.Canceled
}

func TestStartAndStopImpersonation(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash failed: %v", err)
	}
	svc := NewService(auth.Config{
		SecretKey:                   "secret",
		AccessTokenValiditySeconds:  3600,
		RefreshTokenValiditySeconds: 86400,
	}, allowChecker{}, fakeUserLookup{
		email:        "admin@example.com",
		password:     string(hash),
		existsByUser: map[int]bool{10: true, 77: true},
	}, nil)

	token, err := svc.Start(context.Background(), 10, StartRequest{
		TargetUserID:  77,
		AdminPassword: "password123",
	}, "127.0.0.1")
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}
	if token.AccessToken == "" || token.RefreshToken == "" {
		t.Fatalf("expected token pair")
	}
	if err := svc.Stop(context.Background(), 77, "127.0.0.1"); err != nil {
		t.Fatalf("stop failed: %v", err)
	}
}
