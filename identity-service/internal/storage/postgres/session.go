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

type SessionRepo struct {
	pool *pgxpool.Pool
}

func NewSessionRepo(pool *pgxpool.Pool) *SessionRepo {
	return &SessionRepo{pool: pool}
}

func (r *SessionRepo) CreateSession(ctx context.Context, session *domain.Session) error {
	const op = "storage.postgres.CreateSession"

	query := `
		INSERT INTO sessions(user_id, app_id, user_agent, ip_address, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`

	err := getDB(ctx, r.pool).QueryRow(ctx, query, session.UserID, session.AppID, session.UserAgent,
		session.IPAddress, session.ExpiresAt).
		Scan(&session.ID, &session.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505": // UNIQUE violation (user_id+app_id?)
				return sl.Err(op, domain.ErrSessionAlreadyExists)
			case "23503": // Foreign key (user_id, app_id)
				return sl.Err(op, domain.ErrUserOrAppNotFound)
			}
		}
		return sl.Err(op, fmt.Errorf("create session: %w", err))
	}

	return nil
}

func (r *SessionRepo) GetSessionByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	const op = "storage.postgres.GetSessionByID"

	query := `
       SELECT id, user_id, app_id, user_agent, ip_address, created_at, expires_at, revoked_at
       FROM sessions 
       WHERE id = $1`

	rows, err := getDB(ctx, r.pool).Query(ctx, query, id)
	if err != nil {
		return nil, sl.Err(op, fmt.Errorf("query execution: %w", err))
	}
	defer rows.Close()

	session, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Session])

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sl.Err(op, domain.ErrSessionNotFound)
		}
		return nil, sl.Err(op, fmt.Errorf("collect row: %w", err))
	}

	return &session, nil
}

func (r *SessionRepo) RevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	const op = "storage.postgres.RevokeSession"

	query := `
       UPDATE sessions 
       SET revoked_at = NOW()
       WHERE id = $1 AND revoked_at IS NULL`

	result, err := getDB(ctx, r.pool).Exec(ctx, query, sessionID)
	if err != nil {
		return sl.Err(op, fmt.Errorf("update failed: %w", err))
	}

	if result.RowsAffected() == 0 {
		return sl.Err(op, domain.ErrSessionNotFoundOrAlreadyRevoked)
	}

	return nil
}

func (r *SessionRepo) RevokeAllByUser(ctx context.Context, userID uuid.UUID, appID uuid.UUID) error {
	const op = "storage.postgres.RevokeAllByUser"

	query := `
       UPDATE sessions 
       SET revoked_at = NOW()
       WHERE user_id = $1 
         AND app_id = $2 
         AND revoked_at IS NULL
         AND expires_at > NOW()`

	result, err := getDB(ctx, r.pool).Exec(ctx, query, userID, appID)
	if err != nil {
		return sl.Err(op, fmt.Errorf("update failed: %w", err))
	}

	if result.RowsAffected() == 0 {
		return nil
	}

	return nil
}

// нужен для user.sessionReader
func (r *SessionRepo) ListByUser(ctx context.Context, userID uuid.UUID, appID uuid.UUID) ([]domain.Session, error) {
	const op = "storage.postgres.ListByUser"

	query := `
       SELECT id, user_id, app_id, user_agent, ip_address, created_at, expires_at, revoked_at
       FROM sessions 
       WHERE user_id = $1 AND app_id = $2
       ORDER BY created_at DESC`

	rows, err := getDB(ctx, r.pool).Query(ctx, query, userID, appID)
	if err != nil {
		return nil, sl.Err(op, fmt.Errorf("query execution: %w", err))
	}
	defer rows.Close()

	sessions, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Session])
	if err != nil {
		return nil, sl.Err(op, fmt.Errorf("collect sessions: %w", err))
	}

	// Опционально: если sessions == nil (хотя CollectRows обычно возвращает пустой слайс),
	// можно гарантировать возврат инициализированного слайса для JSON-ответов.
	if sessions == nil {
		return []domain.Session{}, nil
	}

	return sessions, nil
}
