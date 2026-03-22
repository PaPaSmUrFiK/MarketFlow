package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/domain"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/lib/sl"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RoleRepo struct {
	pool *pgxpool.Pool
}

func NewRoleRepo(pool *pgxpool.Pool) *RoleRepo {
	return &RoleRepo{pool: pool}
}

func (r *RoleRepo) CreateRole(ctx context.Context, role *domain.Role) error {
	const op = "storage.postgres.CreateRole"

	query := `
		INSERT INTO roles(app_id, code, description)
		VALUES ($1, $2, $3)
		RETURNING id`

	err := getDB(ctx, r.pool).QueryRow(ctx, query, role.AppID, role.Code, role.Description).
		Scan(&role.ID)

	if err != nil {
		return sl.Err(op, err)
	}

	return nil
}

func (r *RoleRepo) GetRoleByCode(ctx context.Context, appID uuid.UUID, code string) (*domain.Role, error) {
	const op = "storage.postgres.GetRoleByCode"

	query := `
		SELECT id, app_id, code, description
		FROM roles
		WHERE app_id = $1 AND code = $2`

	rows, err := getDB(ctx, r.pool).Query(ctx, query, appID, code)

	if err != nil {
		return nil, sl.Err(op, err)
	}
	defer rows.Close()

	role, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Role])

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sl.Err(op, domain.ErrRoleNotFound)
		}
		return nil, sl.Err(op, fmt.Errorf("collect row: %w", err))
	}

	return &role, nil
}

func (r *RoleRepo) GetRoleByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	const op = "storage.postgres.GetRoleByID"

	query := `
		SELECT id, app_id, code, description
		FROM roles
		WHERE id = $1`

	rows, err := getDB(ctx, r.pool).Query(ctx, query, id)
	if err != nil {
		return nil, sl.Err(op, err)
	}
	defer rows.Close()

	role, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Role])

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sl.Err(op, domain.ErrRoleNotFound)
		}
		return nil, sl.Err(op, fmt.Errorf("collect row: %w", err))
	}

	return &role, nil
}

func (r *RoleRepo) DeleteRole(ctx context.Context, roleID uuid.UUID) error {
	const op = "storage.postgres.DeleteRole"

	query := `
		DELETE FROM roles
		WHERE id = $1`

	tag, err := getDB(ctx, r.pool).Exec(ctx, query, roleID)
	if err != nil {
		return sl.Err(op, err)
	}

	if tag.RowsAffected() == 0 {
		return sl.Err(op, domain.ErrRoleNotFound)
	}

	return nil
}

func (r *RoleRepo) AssignPermission(ctx context.Context, roleID, permissionID uuid.UUID) error {
	const op = "storage.postgres.AssignPermission"

	query := `
		INSERT INTO role_permissions(role_id, permission_id)
		values ($1, $2)
		ON CONFLICT (role_id, permission_id) DO NOTHING`

	_, err := getDB(ctx, r.pool).Exec(ctx, query, roleID, permissionID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503": // Foreign key violation
				return sl.Err(op, domain.ErrRoleOrPermissionNotFound)
			case "23505": // Unique violation
				return sl.Err(op, domain.ErrPermissionAlreadyAssigned)
			}
		}
		return sl.Err(op, fmt.Errorf("insert failed: %w", err))
	}

	return nil
}

func (r *RoleRepo) RemovePermission(ctx context.Context, roleID, permissionID uuid.UUID) error {
	const op = "storage.postgres.RemovePermission"

	query := `
		DELETE FROM role_permissions
		WHERE role_id = $1 AND permission_id = $2`

	_, err := getDB(ctx, r.pool).Exec(ctx, query, roleID, permissionID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return sl.Err(op, domain.ErrRoleOrPermissionNotFound)
		}
		return sl.Err(op, fmt.Errorf("delete failed: %w", err))
	}
	return nil
}

func (r *RoleRepo) GetPermissionsByRole(ctx context.Context, roleID uuid.UUID) ([]domain.Permission, error) {
	const op = "storage.postgres.GetPermissionsByRole"

	query := `
		SELECT p.id, p.app_id, p.code, p.description
       FROM permissions p
       JOIN role_permissions rp ON p.id = rp.permission_id
       WHERE rp.role_id = $1`

	rows, err := getDB(ctx, r.pool).Query(ctx, query, roleID)
	if err != nil {
		return nil, sl.Err(op, err)
	}
	defer rows.Close()

	var permissions []domain.Permission
	for rows.Next() {
		var permission domain.Permission
		err = rows.Scan(&permission.ID, &permission.AppID, &permission.Code, &permission.Description)
		if err != nil {
			return nil, sl.Err(op, err)
		}
		permissions = append(permissions, permission)
	}

	if err := rows.Err(); err != nil {
		return nil, sl.Err(op, err)
	}

	return permissions, nil
}

func (r *RoleRepo) CreatePermission(ctx context.Context, permission *domain.Permission) error {
	const op = "storage.postgres.CreatePermission"

	query := `
		INSERT INTO permissions(app_id, code, description)
		VALUES ($1, $2, $3)
		RETURNING id`

	err := getDB(ctx, r.pool).QueryRow(ctx, query, permission.AppID, permission.Code, permission.Description).
		Scan(&permission.ID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return sl.Err(op, domain.ErrPermissionAlreadyExists)
		}
		return sl.Err(op, fmt.Errorf("create permission: %w", err))
	}

	return nil
}

func (r *RoleRepo) GetPermissionByID(ctx context.Context, id uuid.UUID) (*domain.Permission, error) {
	const op = "storage.postgres.GetPermissionByID"

	query := `
		SELECT id, app_id, code, description
		FROM permissions
		WHERE id = $1`

	rows, err := getDB(ctx, r.pool).Query(ctx, query, id)
	if err != nil {
		return nil, sl.Err(op, err)
	}
	defer rows.Close()

	perm, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Permission])

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sl.Err(op, domain.ErrPermissionNotFound)
		}
		return nil, sl.Err(op, fmt.Errorf("collect row: %w", err))
	}

	return &perm, nil
}

func (r *RoleRepo) DeletePermission(ctx context.Context, permissionID uuid.UUID) error {
	const op = "storage.postgres.DeletePermission"

	query := `
		DELETE FROM permissions
		WHERE id = $1`

	result, err := getDB(ctx, r.pool).Exec(ctx, query, permissionID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return sl.Err(op, domain.ErrPermissionNotFound)
		}
		return sl.Err(op, fmt.Errorf("delete failed: %w", err))
	}

	if result.RowsAffected() == 0 {
		return sl.Err(op, domain.ErrPermissionNotFound)
	}

	return nil
}
