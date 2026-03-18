package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type AccessClaims struct {
	UserID uuid.UUID `json:"uid"`
	AppID  uuid.UUID `json:"aid"`
	Roles  []string  `json:"roles"`

	jwt.RegisteredClaims
}
