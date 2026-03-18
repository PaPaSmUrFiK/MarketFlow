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
	ID        uuid.UUID
	Status    UserStatus
	CreatedAt time.Time
	UpdatedAt time.Time
	Roles     []Role
}
