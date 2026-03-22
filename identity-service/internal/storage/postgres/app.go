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

type AppRepo struct {
	pool *pgxpool.Pool
}

func NewAppRepo(pool *pgxpool.Pool) *AppRepo {
	return &AppRepo{pool: pool}
}

func (r *AppRepo) GetAppByCode(ctx context.Context, code string) (*domain.Application, error) {
	const op = "storage.postgres.GetByCode"

	query := `
       SELECT id, code, created_at, name, active 
       FROM applications 
       WHERE code = $1 AND active = TRUE`

	rows, err := getDB(ctx, r.pool).Query(ctx, query, code)
	if err != nil {
		return nil, sl.Err(op, fmt.Errorf("query execution: %w", err))
	}
	defer rows.Close()

	app, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Application])

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sl.Err(op, domain.ErrApplicationNotFound)
		}
		return nil, sl.Err(op, fmt.Errorf("collect row: %w", err))
	}

	return &app, nil
}

func (r *AppRepo) GetAppByID(ctx context.Context, id uuid.UUID) (*domain.Application, error) {
	const op = "storage.postgres.GetByID"

	query := `
       SELECT id, code, created_at, name, active 
       FROM applications 
       WHERE id = $1 AND active = TRUE`

	rows, err := getDB(ctx, r.pool).Query(ctx, query, id)
	if err != nil {
		return nil, sl.Err(op, fmt.Errorf("query execution: %w", err))
	}
	defer rows.Close()

	app, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Application])

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, sl.Err(op, domain.ErrApplicationNotFound)
		}
		return nil, sl.Err(op, fmt.Errorf("collect row: %w", err))
	}

	return &app, nil
}

func (r *AppRepo) CreateApp(ctx context.Context, app *domain.Application) error {
	const op = "storage.postgres.CreateApp"

	query := `
        INSERT INTO applications (code, name)
        VALUES ($1, $2)
        RETURNING id, active, created_at`

	err := getDB(ctx, r.pool).QueryRow(ctx, query, app.Code, app.Name).Scan(
		&app.ID,
		&app.Active,
		&app.CreatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return sl.Err(op, domain.ErrAppAlreadyExists)
		}
		return sl.Err(op, fmt.Errorf("create app: %w", err))
	}

	return nil

}
