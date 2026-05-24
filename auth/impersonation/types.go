package impersonation

import (
	"context"
)

type PermissionChecker interface {
	CanImpersonate(actorUserID int, actorEmail string) bool
}

type UserLookup interface {
	GetByID(ctx context.Context, userID int) (email string, passwordHash string, err error)
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
	TargetUserID  int    `json:"target_user_id" binding:"required,min=1"`
	AdminPassword string `json:"admin_password" binding:"required,min=8,max=128"`
	Comment       string `json:"comment"`
}

type StartResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
}
