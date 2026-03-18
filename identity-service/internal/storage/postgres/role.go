package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/domain"
	"github.com/google/uuid"
)

type RoleRepo struct {
	db *sql.DB
}

func NewRoleRepo(db *sql.DB) *RoleRepo {
	return &RoleRepo{db: db}
}

func (r *RoleRepo) CreateRole(ctx context.Context, role *domain.Role) error {
	return errors.New("not implemented")
}

func (r *RoleRepo) GetRoleByCode(ctx context.Context, appID uuid.UUID, code string) (*domain.Role, error) {
	return nil, errors.New("not implemented")
}

func (r *RoleRepo) GetRoleByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	return nil, errors.New("not implemented")
}

func (r *RoleRepo) DeleteRole(ctx context.Context, roleID uuid.UUID) error {
	return errors.New("not implemented")
}

func (r *RoleRepo) AssignPermission(ctx context.Context, roleID, permissionID uuid.UUID) error {
	return errors.New("not implemented")
}

func (r *RoleRepo) RemovePermission(ctx context.Context, roleID, permissionID uuid.UUID) error {
	return errors.New("not implemented")
}

func (r *RoleRepo) GetPermissionsByRole(ctx context.Context, roleID uuid.UUID) ([]domain.Permission, error) {
	return nil, errors.New("not implemented")
}

func (r *RoleRepo) CreatePermission(ctx context.Context, permission *domain.Permission) error {
	return errors.New("not implemented")
}

func (r *RoleRepo) GetPermissionByID(ctx context.Context, id uuid.UUID) (*domain.Permission, error) {
	return nil, errors.New("not implemented")
}

func (r *RoleRepo) DeletePermission(ctx context.Context, permissionID uuid.UUID) error {
	return errors.New("not implemented")
}
