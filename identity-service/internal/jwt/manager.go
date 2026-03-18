package jwt

import (
	"github.com/google/uuid"
	"time"
)

type Manager interface {
	GenerateAccessToken(userID uuid.UUID, appID uuid.UUID, roles []string) (string, error)
	ParseAccessToken(accessToken string) (*AccessClaims, error)
	AccessTTL() time.Duration
	RefreshTTL() time.Duration
}
