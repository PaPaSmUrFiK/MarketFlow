package domain

import "errors"

var (
	// User
	ErrUserNotFound          = errors.New("user not found")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrUserLocked            = errors.New("user account locked")
	ErrUserBlocked           = errors.New("user blocked")
	ErrEmailAlreadyExists    = errors.New("email already exists")
	ErrIdentityAlreadyExists = errors.New("identity already exists")
	ErrIdentityNotFound      = errors.New("identity not found")
	ErrUserOrRoleNotFound    = errors.New("user or role not found")

	// Session
	ErrSessionRevoked                  = errors.New("session revoked")       // используется в RevokeSession
	ErrSessionNotFound                 = errors.New("session not found")     // используется в GetSessionByID
	ErrUserOrAppNotFound               = errors.New("user or app not found") // FK при создании сессии
	ErrSessionAlreadyExists            = errors.New("session already exists")
	ErrSessionNotFoundOrAlreadyRevoked = errors.New("session not found")

	// Token
	ErrRefreshTokenExpired           = errors.New("refresh token expired or revoked")
	ErrTokenAlreadyExists            = errors.New("token already exists")
	ErrUserOrSessionNotFound         = errors.New("user or session not found")
	ErrTokenNotFoundOrAlreadyRevoked = errors.New("token not found or already revoked")

	// Application
	ErrApplicationNotFound = errors.New("application not found")
	ErrAppAlreadyExists    = errors.New("application already exists")

	// Role
	ErrRoleNotFound             = errors.New("role not found")
	ErrRoleOrPermissionNotFound = errors.New("role or permission not found")

	// Permission
	ErrPermissionNotFound        = errors.New("permission not found")
	ErrPermissionAlreadyExists   = errors.New("permission already exists")
	ErrPermissionAlreadyAssigned = errors.New("permission already assigned to role")

	// General
	ErrPermissionDenied = errors.New("permission denied")
	ErrNotImplemented   = errors.New("not implemented")
)
