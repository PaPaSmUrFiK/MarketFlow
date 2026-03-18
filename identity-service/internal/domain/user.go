package domain

import (
	"github.com/google/uuid"
	"time"
)

type UserStatus string

const (
	UserActive  UserStatus = "ACTIVE"
	UserBlocked UserStatus = "BLOCKED"
	UserDeleted UserStatus = "DELETED"
)

type User struct {
	ID        uuid.UUID  `db:"id"`
	Status    UserStatus `db:"status"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	Roles     []Role     `db:"-"`
}
