package admin

import (
	"context"
	"fmt"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/lib/sl"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/security"
	"log/slog"

	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/domain"
	"github.com/google/uuid"
)

type userAdmin interface {
	CreateUser(ctx context.Context, user *domain.User) error
	CreateCredentials(ctx context.Context, cred *domain.Credential) error
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
	GetPermissionByCode(ctx context.Context, appId uuid.UUID, permCode string) (*domain.Permission, error)
	DeletePermission(ctx context.Context, permissionID uuid.UUID) error
}

type appAdmin interface {
	GetAppByID(ctx context.Context, id uuid.UUID) (*domain.Application, error)
	GetAppByCode(ctx context.Context, code string) (*domain.Application, error)
	CreateApp(ctx context.Context, app *domain.Application) error
}

type TxManager interface {
	Transactional(ctx context.Context, f func(ctx context.Context) error) error
}

type Service struct {
	users     userAdmin
	roles     roleAdmin
	apps      appAdmin
	txManager TxManager
	log       *slog.Logger
}

func New(users userAdmin, roles roleAdmin, apps appAdmin, txManager TxManager, log *slog.Logger) *Service {
	return &Service{users: users, roles: roles, apps: apps, txManager: txManager, log: log}
}

// CreateApplication — создаёт новое приложение.
// Проверяем уникальность кода до создания через GetAppByCode.
func (s *Service) CreateApplication(ctx context.Context, in CreateApplicationInput) (*domain.Application, error) {
	const op = "admin.Service.CreateApplication"

	log := s.log.With(slog.String("op", op), slog.String("code", in.Code))

	if _, err := s.apps.GetAppByCode(ctx, in.Code); err == nil {
		log.Warn("application code already exists")
		return nil, domain.ErrAppAlreadyExists
	}

	app := &domain.Application{
		Code: in.Code,
		Name: in.Name,
	}

	if err := s.apps.CreateApp(ctx, app); err != nil {
		return nil, sl.Err(op, err)
	}

	log.Info("application created", slog.String("app_id", app.ID.String()))
	return app, nil
}

// CreateUser — создаёт пользователя с credentials и назначает роль USER.
func (s *Service) CreateUser(ctx context.Context, in CreateUserInput) (*domain.User, error) {
	const op = "admin.Service.CreateUser"

	log := s.log.With(
		slog.String("op", op),
		slog.String("email", in.Email),
		slog.String("app_id", in.AppID.String()),
	)

	if in.Email == "" || in.Password == "" {
		return nil, domain.ErrInvalidCredentials
	}

	passHash, err := security.HashPassword(in.Password)
	if err != nil {
		return nil, sl.Err(op, fmt.Errorf("hash password: %w", err))
	}

	// Проверяем роль USER для приложения до транзакции
	defaultRole, err := s.roles.GetRoleByCode(ctx, in.AppID, string(domain.DefaultUserRole))
	if err != nil {
		if isNotFound(err) {
			log.Error("default USER role not configured",
				slog.String("app_id", in.AppID.String()),
			)
			return nil, fmt.Errorf("%w: role USER not configured for this app", domain.ErrInternal)
		}
		return nil, sl.Err(op, err)
	}

	var newUser *domain.User

	err = s.txManager.Transactional(ctx, func(ctx context.Context) error {
		newUser = &domain.User{Status: domain.UserActive}

		if err := s.users.CreateUser(ctx, newUser); err != nil {
			return sl.Err(op, err)
		}

		cred := &domain.Credential{
			UserID:        newUser.ID,
			Email:         in.Email,
			PasswordHash:  passHash,
			EmailVerified: false,
		}
		if err := s.users.CreateCredentials(ctx, cred); err != nil {
			if isAlreadyExists(err) {
				log.Warn("email already taken")
				return domain.ErrEmailAlreadyExists
			}
			return sl.Err(op, err)
		}

		if err := s.users.AssignRole(ctx, newUser.ID, defaultRole.ID); err != nil {
			return sl.Err(op, fmt.Errorf("assign default role: %w", err))
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	log.Info("user created by admin", slog.String("user_id", newUser.ID.String()))
	return newUser, nil
}

func (s *Service) BlockUser(ctx context.Context, userID uuid.UUID) error {
	const op = "admin.Service.BlockUser"
	if err := s.users.UpdateStatus(ctx, userID, domain.UserBlocked); err != nil {
		return sl.Err(op, err)
	}
	s.log.Info("user blocked", slog.String("op", op), slog.String("user_id", userID.String()))
	return nil
}

func (s *Service) UnblockUser(ctx context.Context, userID uuid.UUID) error {
	const op = "admin.Service.UnblockUser"
	if err := s.users.UpdateStatus(ctx, userID, domain.UserActive); err != nil {
		return sl.Err(op, err)
	}
	s.log.Info("user unblocked", slog.String("op", op), slog.String("user_id", userID.String()))
	return nil
}

func (s *Service) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	const op = "admin.Service.DeleteUser"
	if err := s.users.UpdateStatus(ctx, userID, domain.UserDeleted); err != nil {
		return sl.Err(op, err)
	}
	s.log.Info("user deleted", slog.String("op", op), slog.String("user_id", userID.String()))
	return nil
}

// ListUsers — список пользователей приложения с пагинацией на стороне сервиса.
// ListByApp в репозитории не поддерживает пагинацию — делаем срез здесь.
// TODO: когда ListByApp получит пагинацию в БД — убрать срез отсюда.
func (s *Service) ListUsers(ctx context.Context, appID uuid.UUID, page, size int32) ([]domain.User, int64, error) {
	const op = "admin.Service.ListUsers"

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	all, err := s.users.ListByApp(ctx, appID)
	if err != nil {
		return nil, 0, sl.Err(op, err)
	}

	total := int64(len(all))
	start := int((page - 1) * size)
	end := int(page * size)

	if start >= len(all) {
		return []domain.User{}, total, nil
	}
	if end > len(all) {
		end = len(all)
	}

	return all[start:end], total, nil
}

func (s *Service) GetUserDetails(ctx context.Context, userID uuid.UUID, appID uuid.UUID) (*domain.User, error) {
	const op = "admin.Service.GetUserDetails"

	user, err := s.users.GetUserWithRoles(ctx, userID, appID)
	if err != nil {
		return nil, sl.Err(op, err)
	}

	return user, nil
}

// Roles

func (s *Service) CreateRole(ctx context.Context, in CreateRoleInput) (*domain.Role, error) {
	const op = "admin.Service.CreateRole"

	log := s.log.With(
		slog.String("op", op),
		slog.String("app_id", in.AppID.String()),
		slog.String("code", in.Code),
	)

	role := &domain.Role{
		AppID:       in.AppID,
		Code:        in.Code,
		Description: in.Description,
	}

	if err := s.roles.CreateRole(ctx, role); err != nil {
		return nil, sl.Err(op, err)
	}

	log.Info("role created", slog.String("role_id", role.ID.String()))
	return role, nil
}

func (s *Service) DeleteRole(ctx context.Context, appID uuid.UUID, roleCode string) error {
	const op = "admin.Service.DeleteRole"

	role, err := s.roles.GetRoleByCode(ctx, appID, roleCode)
	if err != nil {
		return sl.Err(op, err)
	}

	if err := s.roles.DeleteRole(ctx, role.ID); err != nil {
		return sl.Err(op, err)
	}

	s.log.Info("role deleted",
		slog.String("op", op),
		slog.String("role_code", roleCode),
		slog.String("app_id", appID.String()),
	)
	return nil
}

func (s *Service) AssignRoleToUser(ctx context.Context, in AssignRoleToUserInput) error {
	const op = "admin.Service.AssignRoleToUser"

	role, err := s.roles.GetRoleByCode(ctx, in.AppID, in.RoleCode)
	if err != nil {
		return sl.Err(op, err)
	}

	if err := s.users.AssignRole(ctx, in.UserID, role.ID); err != nil {
		return sl.Err(op, err)
	}

	s.log.Info("role assigned to user",
		slog.String("op", op),
		slog.String("user_id", in.UserID.String()),
		slog.String("role_code", in.RoleCode),
	)
	return nil
}

func (s *Service) RemoveRoleFromUser(ctx context.Context, in RemoveRoleFromUserInput) error {
	const op = "admin.Service.RemoveRoleFromUser"

	role, err := s.roles.GetRoleByCode(ctx, in.AppID, in.RoleCode)
	if err != nil {
		return sl.Err(op, err)
	}

	if err := s.users.RemoveRole(ctx, in.UserID, role.ID); err != nil {
		return sl.Err(op, err)
	}

	s.log.Info("role removed from user",
		slog.String("op", op),
		slog.String("user_id", in.UserID.String()),
		slog.String("role_code", in.RoleCode),
	)
	return nil
}

// Permissions

// CreatePermission — создаёт permission для приложения.
func (s *Service) CreatePermission(ctx context.Context, in CreatePermissionInput) (*domain.Permission, error) {
	const op = "admin.Service.CreatePermission"

	perm := &domain.Permission{
		AppID:       in.AppID,
		Code:        in.Code,
		Description: in.Description,
	}

	if err := s.roles.CreatePermission(ctx, perm); err != nil {
		return nil, sl.Err(op, err)
	}

	s.log.Info("permission created",
		slog.String("op", op),
		slog.String("perm_id", perm.ID.String()),
		slog.String("code", in.Code),
	)
	return perm, nil
}

// DeletePermission — удаляет permission по коду.
func (s *Service) DeletePermission(ctx context.Context, appID uuid.UUID, permCode string) error {
	const op = "admin.Service.DeletePermission"

	perm, err := s.roles.GetPermissionByCode(ctx, appID, permCode)
	if err != nil {
		return sl.Err(op, err)
	}

	if err := s.roles.DeletePermission(ctx, perm.ID); err != nil {
		return sl.Err(op, err)
	}

	s.log.Info("permission deleted",
		slog.String("op", op),
		slog.String("perm_code", permCode),
		slog.String("app_id", appID.String()),
	)
	return nil
}

// AssignPermissionToRole — назначает permission роли по кодам.
func (s *Service) AssignPermissionToRole(ctx context.Context, in AssignPermissionToRoleInput) error {
	const op = "admin.Service.AssignPermissionToRole"

	role, err := s.roles.GetRoleByCode(ctx, in.AppID, in.RoleCode)
	if err != nil {
		return sl.Err(op, err)
	}

	perm, err := s.roles.GetPermissionByCode(ctx, in.AppID, in.PermissionCode)
	if err != nil {
		return sl.Err(op, err)
	}

	if err := s.roles.AssignPermission(ctx, role.ID, perm.ID); err != nil {
		return sl.Err(op, err)
	}

	s.log.Info("permission assigned to role",
		slog.String("op", op),
		slog.String("role_code", in.RoleCode),
		slog.String("perm_code", in.PermissionCode),
	)
	return nil
}

// RemovePermissionFromRole — убирает permission из роли по кодам.
func (s *Service) RemovePermissionFromRole(ctx context.Context, in RemovePermissionFromRoleInput) error {
	const op = "admin.Service.RemovePermissionFromRole"

	role, err := s.roles.GetRoleByCode(ctx, in.AppID, in.RoleCode)
	if err != nil {
		return sl.Err(op, err)
	}

	perm, err := s.roles.GetPermissionByCode(ctx, in.AppID, in.PermissionCode)
	if err != nil {
		return sl.Err(op, err)
	}

	if err := s.roles.RemovePermission(ctx, role.ID, perm.ID); err != nil {
		return sl.Err(op, err)
	}

	s.log.Info("permission removed from role",
		slog.String("op", op),
		slog.String("role_code", in.RoleCode),
		slog.String("perm_code", in.PermissionCode),
	)
	return nil
}
