package domain

import (
	"github.com/google/uuid"
	"time"
)

type Session struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	AppID     uuid.UUID
	UserAgent string
	IPAddress string
	CreatedAt time.Time
	RevokedAt *time.Time
	ExpiresAt time.Time
}

func (s *Session) IsRevoked() bool {
	return s.RevokedAt != nil
}

func (s *Session) Revoke(now time.Time) {
	s.RevokedAt = &now
}
