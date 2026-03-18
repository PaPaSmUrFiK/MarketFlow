package admin

import (
	"context"
	"errors"
	"log/slog"

	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/domain"
	"github.com/google/uuid"
)

type userAdmin interface {
	CreateUser(ctx context.Context, user *domain.User) error
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.UserStatus) error
	ListByApp(ctx context.Context, appID uuid.UUID) ([]domain.User, error)
	AssignRole(ctx context.Context, userID, roleID uuid.UUID) error
	RemoveRole(ctx context.Context, userID, roleID uuid.UUID) error
	GetUserWithRoles(ctx context.Context, userID uuid.UUID, appID uuid.UUID) (*domain.User, error)
}

type roleAdmin interface {
	CreateRole(ctx context.Context, role *domain.Role) error
	GetRoleByCode(ctx context.Context, appID uuid.UUID, code string) (*domain.Role, error)
	GetRoleByID(ctx context.Context, id uuid.UUID) (*domain.Role, error)
	DeleteRole(ctx context.Context, roleID uuid.UUID) error
	AssignPermission(ctx context.Context, roleID, permissionID uuid.UUID) error
	RemovePermission(ctx context.Context, roleID, permissionID uuid.UUID) error
	GetPermissionsByRole(ctx context.Context, roleID uuid.UUID) ([]domain.Permission, error)
	CreatePermission(ctx context.Context, permission *domain.Permission) error
	GetPermissionByID(ctx context.Context, id uuid.UUID) (*domain.Permission, error)
	DeletePermission(ctx context.Context, permissionID uuid.UUID) error
}

type appAdmin interface {
	GetAppByID(ctx context.Context, id uuid.UUID) (*domain.Application, error)
	GetAppByCode(ctx context.Context, code string) (*domain.Application, error)
	CreateApp(ctx context.Context, app *domain.Application) error
}

type Service struct {
	users userAdmin
	roles roleAdmin
	apps  appAdmin
	log   *slog.Logger
}

func New(users userAdmin, roles roleAdmin, apps appAdmin, log *slog.Logger) *Service {
	return &Service{users: users, roles: roles, apps: apps, log: log}
}

// Applications

func (s *Service) CreateApplication(ctx context.Context, in CreateApplicationInput) (*domain.Application, error) {
	return nil, errors.New("not implemented")
}

// Users

func (s *Service) CreateUser(ctx context.Context, in CreateUserInput) (*domain.User, error) {
	return nil, errors.New("not implemented")
}

func (s *Service) BlockUser(ctx context.Context, userID uuid.UUID) error {
	return s.users.UpdateStatus(ctx, userID, domain.UserBlocked)
}

func (s *Service) UnblockUser(ctx context.Context, userID uuid.UUID) error {
	return s.users.UpdateStatus(ctx, userID, domain.UserActive)
}

func (s *Service) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	return s.users.UpdateStatus(ctx, userID, domain.UserDeleted)
}

func (s *Service) ListUsers(ctx context.Context, appID uuid.UUID, page, size int32) ([]domain.User, int64, error) {
	return nil, 0, errors.New("not implemented")
}

func (s *Service) GetUserDetails(ctx context.Context, userID uuid.UUID, appID uuid.UUID) (*domain.User, error) {
	return s.users.GetUserWithRoles(ctx, userID, appID)
}

// Roles

func (s *Service) CreateRole(ctx context.Context, in CreateRoleInput) (*domain.Role, error) {
	return nil, errors.New("not implemented")
}

func (s *Service) DeleteRole(ctx context.Context, appID uuid.UUID, roleCode string) error {
	role, err := s.roles.GetRoleByCode(ctx, appID, roleCode)
	if err != nil {
		return err
	}
	return s.roles.DeleteRole(ctx, role.ID)
}

func (s *Service) AssignRoleToUser(ctx context.Context, in AssignRoleToUserInput) error {
	role, err := s.roles.GetRoleByCode(ctx, in.AppID, in.RoleCode)
	if err != nil {
		return err
	}
	return s.users.AssignRole(ctx, in.UserID, role.ID)
}

func (s *Service) RemoveRoleFromUser(ctx context.Context, in RemoveRoleFromUserInput) error {
	role, err := s.roles.GetRoleByCode(ctx, in.AppID, in.RoleCode)
	if err != nil {
		return err
	}
	return s.users.RemoveRole(ctx, in.UserID, role.ID)
}

// Permissions

func (s *Service) CreatePermission(ctx context.Context, in CreatePermissionInput) (*domain.Permission, error) {
	return nil, errors.New("not implemented")
}

func (s *Service) DeletePermission(ctx context.Context, appID uuid.UUID, permCode string) error {
	return errors.New("not implemented")
}

func (s *Service) AssignPermissionToRole(ctx context.Context, in AssignPermissionToRoleInput) error {
	return errors.New("not implemented")
}

func (s *Service) RemovePermissionFromRole(ctx context.Context, in RemovePermissionFromRoleInput) error {
	return errors.New("not implemented")
}
