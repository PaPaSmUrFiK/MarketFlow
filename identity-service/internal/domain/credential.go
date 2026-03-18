package domain

import (
	"github.com/google/uuid"
	"time"
)

type Credential struct {
	UserID              uuid.UUID
	Email               string
	PasswordHash        string
	EmailVerified       bool
	CreatedAt           time.Time
	LastPasswordChange  *time.Time
	FailedLoginAttempts int
	LockedUntil         *time.Time
}

func (c *Credential) IsLocked(now time.Time) bool {
	if c.LockedUntil == nil {
		return false
	}
	return now.Before(*c.LockedUntil)
}

func (c *Credential) RegisterFailedAttempt() {
	c.FailedLoginAttempts++
}

func (c *Credential) ResetFailedAttempts() {
	c.FailedLoginAttempts = 0
}
