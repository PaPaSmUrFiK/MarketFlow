package admin

import "github.com/google/uuid"

type CreateUserInput struct {
	AppID    uuid.UUID
	Email    string
	Password string
}

type CreateApplicationInput struct {
	Code string
	Name string
}

type CreateRoleInput struct {
	AppID       uuid.UUID
	Code        string
	Description string
}

type AssignRoleToUserInput struct {
	AppID    uuid.UUID
	UserID   uuid.UUID
	RoleCode string
}

type RemoveRoleFromUserInput struct {
	AppID    uuid.UUID
	UserID   uuid.UUID
	RoleCode string
}

type CreatePermissionInput struct {
	AppID       uuid.UUID
	Code        string
	Description string
}

type AssignPermissionToRoleInput struct {
	AppID          uuid.UUID
	RoleCode       string
	PermissionCode string
}

type RemovePermissionFromRoleInput struct {
	AppID          uuid.UUID
	RoleCode       string
	PermissionCode string
}
