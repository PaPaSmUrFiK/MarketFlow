package domain

import "errors"

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrUserLocked          = errors.New("user account locked")
	ErrUserBlocked         = errors.New("user blocked")
	ErrSessionRevoked      = errors.New("session revoked")
	ErrRefreshTokenExpired = errors.New("refresh token expired")
	ErrPermissionDenied    = errors.New("permission denied")
	ErrNotImplemented      = errors.New("not implemented")
)
