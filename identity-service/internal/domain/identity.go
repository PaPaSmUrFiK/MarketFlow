package domain

import (
	"time"

	"github.com/google/uuid"
)

type OAuthProvider string

const (
	ProviderGoogle OAuthProvider = "google"
	ProviderGitHub OAuthProvider = "github"
)

type UserIdentity struct {
	ID             uuid.UUID     `db:"id"`
	UserID         uuid.UUID     `db:"user_id"`
	Provider       OAuthProvider `db:"provider"`
	ProviderUserID string        `db:"provider_user_id"`
	CreatedAt      time.Time     `db:"created_at"`
}
