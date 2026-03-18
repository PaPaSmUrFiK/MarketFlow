package domain

import "github.com/google/uuid"

type Permission struct {
	ID          uuid.UUID
	AppID       uuid.UUID
	Code        string
	Description string
}
