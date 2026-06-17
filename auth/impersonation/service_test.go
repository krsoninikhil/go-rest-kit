package impersonation

import (
	"context"
	"strings"
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

func (f fakeUserLookup) GetByID(ctx context.Context, userID int) (string, error) {
	if userID == 10 && f.existsByUser[10] {
		return "admin@example.com", nil
	}
	return "", context.Canceled
}

func (f fakeUserLookup) GetByEmailWithPassword(ctx context.Context, email string) (int, string, error) {
	if strings.EqualFold(email, "admin@example.com") && f.existsByUser[10] {
		return 10, f.password, nil
	}
	return 0, "", context.Canceled
}

func (f fakeUserLookup) GetByEmail(ctx context.Context, email string) (int, error) {
	if strings.EqualFold(email, "target@example.com") && f.existsByUser[77] {
		return 77, nil
	}
	return 0, context.Canceled
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
		TargetEmail:   "target@example.com",
		AdminPassword: "password123",
		Comment:       "Verify an issue",
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
