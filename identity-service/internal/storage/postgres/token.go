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

type TokenRepo struct {
	pool *pgxpool.Pool
}

func NewTokenRepo(pool *pgxpool.Pool) *TokenRepo {
	return &TokenRepo{pool: pool}
}

func (r *TokenRepo) CreateToken(ctx context.Context, token *domain.RefreshToken) error {
	const op = "storage.postgres.CreateToken"

	query := `
		INSERT INTO refresh_tokens (user_id, app_id, session_id, token_hash, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`

	err := getDB(ctx, r.pool).QueryRow(ctx, query, token.UserID, token.AppID, token.SessionID, token.TokenHash, token.ExpiresAt).Scan(&token.ID, &token.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505": // UNIQUE violation — token_hash уже существует
				return sl.Err(op, domain.ErrTokenAlreadyExists)
			case "23503": // FK violation — user_id, app_id или session_id не существует
				return sl.Err(op, domain.ErrUserOrSessionNotFound)
			}
		}
		return sl.Err(op, fmt.Errorf("create token: %w", err))
	}

	return nil
}

func (r *TokenRepo) GetTokenByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	const op = "storage.postgres.GetTokenByHash"

	query := `
		SELECT id, user_id, app_id, session_id,
		       token_hash, expires_at, revoked_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1`

	rows, err := getDB(ctx, r.pool).Query(ctx, query, hash)
	if err != nil {
		return nil, sl.Err(op, fmt.Errorf("query: %w", err))
	}
	defer rows.Close()

	token, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.RefreshToken])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sl.Err(op, domain.ErrInvalidCredentials)
		}
		return nil, sl.Err(op, fmt.Errorf("collect row: %w", err))
	}

	return &token, nil
}

func (r *TokenRepo) RevokeToken(ctx context.Context, tokenID uuid.UUID) error {
	const op = "storage.postgres.RevokeToken"

	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE id = $1 AND revoked_at IS NULL`

	result, err := getDB(ctx, r.pool).Exec(ctx, query, tokenID)
	if err != nil {
		return sl.Err(op, fmt.Errorf("update: %w", err))
	}

	if result.RowsAffected() == 0 {
		return sl.Err(op, domain.ErrTokenNotFoundOrAlreadyRevoked)
	}

	return nil
}

func (r *TokenRepo) RevokeAllBySession(ctx context.Context, sessionID uuid.UUID) error {
	const op = "storage.postgres.RevokeAllBySession"

	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE session_id = $1 AND revoked_at IS NULL`

	_, err := getDB(ctx, r.pool).Exec(ctx, query, sessionID)
	if err != nil {
		return sl.Err(op, fmt.Errorf("update: %w", err))
	}

	return nil
}

func (r *TokenRepo) RevokeAllByUser(ctx context.Context, userID uuid.UUID, appID uuid.UUID) error {
	const op = "storage.postgres.RevokeAllByUser"

	query := `
		UPDATE refresh_tokens
		SET revoked_at = NOW()
		WHERE user_id = $1
		  AND app_id = $2
		  AND revoked_at IS NULL`

	_, err := getDB(ctx, r.pool).Exec(ctx, query, userID, appID)
	if err != nil {
		return sl.Err(op, fmt.Errorf("update: %w", err))
	}

	return nil
}

// DeleteExpired — удаляет протухшие токены.
// Вызывается из фонового воркера по расписанию, не из горячего пути.
// Удаляем только отозванные или истёкшие — активные токены не трогаем.
func (r *TokenRepo) DeleteExpired(ctx context.Context) (int64, error) {
	const op = "storage.postgres.DeleteExpired"

	query := `
		DELETE FROM refresh_tokens
		WHERE expires_at < NOW()
		   OR revoked_at IS NOT NULL`

	result, err := getDB(ctx, r.pool).Exec(ctx, query)
	if err != nil {
		return 0, sl.Err(op, fmt.Errorf("delete: %w", err))
	}

	return result.RowsAffected(), nil

}
