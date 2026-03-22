package postgres

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OutboxRepo struct {
	pool *pgxpool.Pool
}

func (r *OutboxRepo) PublishEventTx(ctx context.Context, tx pgx.Tx, eventType string, payload any) error {
	payloadJSON, _ := json.Marshal(payload)

	query := `
        INSERT INTO outbox (event_type, payload, state) 
        VALUES ($1, $2, 'pending')
    `

	_, err := tx.Exec(ctx, query, eventType, payloadJSON)

	return err
}
