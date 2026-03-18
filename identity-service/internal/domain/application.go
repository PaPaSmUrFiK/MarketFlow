package domain

import (
	"github.com/google/uuid"
	"time"
)

type Application struct {
	ID        uuid.UUID
	Code      string
	Name      string
	Active    bool
	CreatedAt time.Time
}
