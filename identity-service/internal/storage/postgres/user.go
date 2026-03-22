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

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) CreateUser(ctx context.Context, user *domain.User) error {
	const op = "storage.postgres.CreateUser"

	query := `
		INSERT INTO users (status)
		VALUES ($1)
		RETURNING id, created_at, updated_at`

	err := getDB(ctx, r.pool).QueryRow(ctx, query, user.Status).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505": // Unique violation
				return sl.Err(op, domain.ErrUserAlreadyExists)
			case "23514": // Check violation
				return sl.Err(op, domain.ErrInvalidUserStatus)
			case "23502": // Not null violation
				return sl.Err(op, domain.ErrUserStatusRequired)
			}
		}
		return sl.Err(op, fmt.Errorf("insert user: %w", err))
	}

	return nil
}

func (r *UserRepo) CreateCredentials(ctx context.Context, cred *domain.Credential) error {
	const op = "storage.postgres.CreateCredentials"

	query := `
		INSERT INTO user_credentials (user_id, email, password_hash, email_verified)
		VALUES ($1, $2, $3, $4)`

	_, err := getDB(ctx, r.pool).Exec(ctx, query,
		cred.UserID, cred.Email, cred.PasswordHash, cred.EmailVerified,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return sl.Err(op, domain.ErrEmailAlreadyExists)
		}
		return sl.Err(op, fmt.Errorf("insert credentials: %w", err))
	}

	return nil
}

func (r *UserRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	const op = "storage.postgres.GetUserByID"

	query := `
		SELECT id, status, created_at, updated_at
		FROM users
		WHERE id = $1`

	rows, err := getDB(ctx, r.pool).Query(ctx, query, id)
	if err != nil {
		return nil, sl.Err(op, fmt.Errorf("query: %w", err))
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sl.Err(op, domain.ErrUserNotFound)
		}
		return nil, sl.Err(op, fmt.Errorf("collect row: %w", err))
	}

	return &user, nil
}

func (r *UserRepo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	const op = "storage.postgres.GetUserByEmail"

	query := `
		SELECT u.id, u.status, u.created_at, u.updated_at
		FROM users u
		JOIN user_credentials uc ON uc.user_id = u.id
		WHERE uc.email = $1`

	rows, err := getDB(ctx, r.pool).Query(ctx, query, email)
	if err != nil {
		return nil, sl.Err(op, fmt.Errorf("query: %w", err))
	}
	defer rows.Close()

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.User])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sl.Err(op, domain.ErrInvalidCredentials)
		}
		return nil, sl.Err(op, fmt.Errorf("collect row: %w", err))
	}

	return &user, nil
}

func (r *UserRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.UserStatus) error {
	const op = "storage.postgres.UpdateStatus"

	query := `
		UPDATE users
		SET status = $1, updated_at = NOW()
		WHERE id = $2`

	result, err := getDB(ctx, r.pool).Exec(ctx, query, status, id)
	if err != nil {
		return sl.Err(op, fmt.Errorf("update: %w", err))
	}

	if result.RowsAffected() == 0 {
		return sl.Err(op, domain.ErrUserNotFound)
	}

	return nil
}

func (r *UserRepo) GetCredentials(ctx context.Context, userID uuid.UUID) (*domain.Credential, error) {
	const op = "storage.postgres.GetCredentials"

	query := `
		SELECT user_id, email, password_hash, email_verified,
		       created_at, last_password_change,
		       failed_login_attempts, locked_until
		FROM user_credentials
		WHERE user_id = $1`

	rows, err := getDB(ctx, r.pool).Query(ctx, query, userID)
	if err != nil {
		return nil, sl.Err(op, fmt.Errorf("query: %w", err))
	}
	defer rows.Close()

	cred, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Credential])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sl.Err(op, domain.ErrUserNotFound)
		}
		return nil, sl.Err(op, fmt.Errorf("collect row: %w", err))
	}

	return &cred, nil
}

func (r *UserRepo) UpdateCredentials(ctx context.Context, cred *domain.Credential) error {
	const op = "storage.postgres.UpdateCredentials"

	query := `
		UPDATE user_credentials
		SET password_hash         = $1,
		    failed_login_attempts = $2,
		    locked_until          = $3,
		    last_password_change  = $4
		WHERE user_id = $5`

	result, err := getDB(ctx, r.pool).Exec(ctx, query,
		cred.PasswordHash,
		cred.FailedLoginAttempts,
		cred.LockedUntil,
		cred.LastPasswordChange,
		cred.UserID,
	)
	if err != nil {
		return sl.Err(op, fmt.Errorf("update: %w", err))
	}

	if result.RowsAffected() == 0 {
		return sl.Err(op, domain.ErrUserNotFound)
	}

	return nil
}

func (r *UserRepo) GetUserWithRoles(ctx context.Context, userID uuid.UUID, appID uuid.UUID) (*domain.User, error) {
	const op = "storage.postgres.GetUserWithRoles"

	query := `
		SELECT u.id, u.status, u.created_at, u.updated_at,
			   COALESCE(uc.email, '') AS email,   -- ← добавить
			   r.id      AS role_id,
			   r.app_id  AS role_app_id,
			   r.code    AS role_code,
			   r.description AS role_description
		FROM users u
		LEFT JOIN user_credentials uc ON uc.user_id = u.id   -- ← добавить
		LEFT JOIN user_roles ur ON ur.user_id = u.id
		LEFT JOIN roles r ON r.id = ur.role_id AND r.app_id = $2
		WHERE u.id = $1`

	rows, err := getDB(ctx, r.pool).Query(ctx, query, userID, appID)
	if err != nil {
		return nil, sl.Err(op, fmt.Errorf("query: %w", err))
	}
	defer rows.Close()

	var user *domain.User

	for rows.Next() {
		var (
			roleID          *uuid.UUID
			roleAppID       *uuid.UUID
			roleCode        *string
			roleDescription *string
		)

		if user == nil {
			user = &domain.User{}
		}

		if err := rows.Scan(
			&user.ID, &user.Status, &user.CreatedAt, &user.UpdatedAt, &user.Email,
			&roleID, &roleAppID, &roleCode, &roleDescription,
		); err != nil {
			return nil, sl.Err(op, fmt.Errorf("scan: %w", err))
		}

		// Добавляем роль только если она не NULL
		if roleID != nil {
			user.Roles = append(user.Roles, domain.Role{
				ID:          *roleID,
				AppID:       *roleAppID,
				Code:        *roleCode,
				Description: *roleDescription,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, sl.Err(op, fmt.Errorf("rows: %w", err))
	}

	if user == nil {
		return nil, sl.Err(op, domain.ErrUserNotFound)
	}

	return user, nil
}

func (r *UserRepo) CreateIdentity(ctx context.Context, identity *domain.UserIdentity) error {
	const op = "storage.postgres.CreateIdentity"

	query := `
		INSERT INTO user_identities (user_id, provider, provider_user_id)
		VALUES ($1, $2, $3)
		RETURNING id, created_at`

	err := getDB(ctx, r.pool).QueryRow(ctx, query,
		identity.UserID, identity.Provider, identity.ProviderUserID,
	).Scan(&identity.ID, &identity.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return sl.Err(op, domain.ErrIdentityAlreadyExists)
		}
		return sl.Err(op, fmt.Errorf("insert identity: %w", err))
	}

	return nil
}

func (r *UserRepo) GetIdentity(ctx context.Context, provider domain.OAuthProvider, providerUserID string) (*domain.UserIdentity, error) {
	const op = "storage.postgres.GetIdentity"

	query := `
		SELECT id, user_id, provider, provider_user_id, created_at
		FROM user_identities
		WHERE provider = $1 AND provider_user_id = $2`

	rows, err := getDB(ctx, r.pool).Query(ctx, query, provider, providerUserID)
	if err != nil {
		return nil, sl.Err(op, fmt.Errorf("query: %w", err))
	}
	defer rows.Close()

	identity, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.UserIdentity])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sl.Err(op, domain.ErrIdentityNotFound)
		}
		return nil, sl.Err(op, fmt.Errorf("collect row: %w", err))
	}

	return &identity, nil
}

// нужен для admin.userAdmin
func (r *UserRepo) ListByApp(ctx context.Context, appID uuid.UUID) ([]domain.User, error) {
	const op = "storage.postgres.ListByApp"

	// Пользователи связаны с приложением через роли — user_roles → roles → app_id
	// DISTINCT чтобы не дублировать пользователей у которых несколько ролей
	query := `
		SELECT DISTINCT u.id, u.status, u.created_at, u.updated_at
		FROM users u
		JOIN user_roles ur ON ur.user_id = u.id
		JOIN roles r       ON r.id = ur.role_id AND r.app_id = $1
		ORDER BY u.created_at DESC`

	rows, err := getDB(ctx, r.pool).Query(ctx, query, appID)
	if err != nil {
		return nil, sl.Err(op, fmt.Errorf("query: %w", err))
	}
	defer rows.Close()

	users, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.User])
	if err != nil {
		return nil, sl.Err(op, fmt.Errorf("collect rows: %w", err))
	}

	if users == nil {
		return []domain.User{}, nil
	}

	return users, nil
}

func (r *UserRepo) AssignRole(ctx context.Context, userID, roleID uuid.UUID) error {
	const op = "storage.postgres.AssignRole"

	query := `
		INSERT INTO user_roles (user_id, role_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING`

	_, err := getDB(ctx, r.pool).Exec(ctx, query, userID, roleID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return sl.Err(op, domain.ErrUserOrRoleNotFound)
		}
		return sl.Err(op, fmt.Errorf("assign role: %w", err))
	}

	return nil
}

func (r *UserRepo) RemoveRole(ctx context.Context, userID, roleID uuid.UUID) error {
	const op = "storage.postgres.RemoveRole"

	query := `
		DELETE FROM user_roles
		WHERE user_id = $1 AND role_id = $2`

	result, err := getDB(ctx, r.pool).Exec(ctx, query, userID, roleID)
	if err != nil {
		return sl.Err(op, fmt.Errorf("remove role: %w", err))
	}

	if result.RowsAffected() == 0 {
		return sl.Err(op, domain.ErrUserOrRoleNotFound)
	}

	return nil
}
