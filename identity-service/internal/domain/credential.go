package domain

import (
	"github.com/google/uuid"
	"time"
)

type Credential struct {
	UserID              uuid.UUID  `db:"user_id"`
	Email               string     `db:"email"`
	PasswordHash        string     `db:"password_hash"`
	EmailVerified       bool       `db:"email_verified"`
	CreatedAt           time.Time  `db:"created_at"`
	LastPasswordChange  *time.Time `db:"last_password_change"`
	FailedLoginAttempts int        `db:"failed_login_attempts"`
	LockedUntil         *time.Time `db:"locked_until"`
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
