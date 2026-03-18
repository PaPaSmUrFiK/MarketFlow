package domain

import (
	"github.com/google/uuid"
	"time"
)

type Session struct {
	ID        uuid.UUID  `db:"id"`
	UserID    uuid.UUID  `db:"user_id"`
	AppID     uuid.UUID  `db:"app_id"`
	UserAgent string     `db:"user_agent"`
	IPAddress string     `db:"ip_address"`
	CreatedAt time.Time  `db:"created_at"`
	ExpiresAt time.Time  `db:"expires_at"`
	RevokedAt *time.Time `db:"revoked_at"`
}

func (s *Session) IsRevoked() bool {
	return s.RevokedAt != nil
}

func (s *Session) Revoke(now time.Time) {
	s.RevokedAt = &now
}
