package domain

import "github.com/google/uuid"

type Permission struct {
	ID          uuid.UUID `db:"id"`
	AppID       uuid.UUID `db:"app_id"`
	Code        string    `db:"code"`
	Description string    `db:"description"`
}
