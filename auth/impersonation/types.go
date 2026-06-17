package impersonation

import (
	"context"
)

type PermissionChecker interface {
	CanImpersonate(actorUserID int, actorEmail string) bool
}

type UserLookup interface {
	GetByID(ctx context.Context, userID int) (email string, err error)
	GetByEmail(ctx context.Context, email string) (userID int, err error)
	GetByEmailWithPassword(ctx context.Context, email string) (userID int, passwordHash string, err error)
}

type AuditEvent struct {
	ActorID      int
	Action       string
	Category     string
	Comment      string
	TargetUserID int
	IP           string
}

type AuditSink interface {
	LogImpersonationEvent(ctx context.Context, event AuditEvent) error
}

type StartRequest struct {
	TargetEmail   string `json:"target_email" binding:"required,email"`
	AdminPassword string `json:"admin_password" binding:"required,min=8,max=128"`
	Comment       string `json:"comment" binding:"required,min=3,max=500"`
}

type StartResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
}
