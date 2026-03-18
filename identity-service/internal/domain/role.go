package domain

import "github.com/google/uuid"

type Role struct {
	ID          uuid.UUID
	AppID       uuid.UUID
	Code        string
	Description string
	Permissions []Permission
}
