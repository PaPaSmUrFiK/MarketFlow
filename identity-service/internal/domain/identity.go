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
	ID             uuid.UUID
	UserID         uuid.UUID
	Provider       OAuthProvider
	ProviderUserID string
	CreatedAt      time.Time
}
