package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/domain"
	"github.com/google/uuid"
)

type TokenRepo struct {
	db *sql.DB
}

func NewTokenRepo(db *sql.DB) *TokenRepo {
	return &TokenRepo{db: db}
}

func (r *TokenRepo) Create(ctx context.Context, token *domain.RefreshToken) error {
	return errors.New("not implemented")
}

func (r *TokenRepo) GetByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	return nil, errors.New("not implemented")
}

func (r *TokenRepo) Revoke(ctx context.Context, tokenID uuid.UUID) error {
	return errors.New("not implemented")
}

func (r *TokenRepo) RevokeAllBySession(ctx context.Context, sessionID uuid.UUID) error {
	return errors.New("not implemented")
}

func (r *TokenRepo) RevokeAllByUser(ctx context.Context, userID uuid.UUID, appID uuid.UUID) error {
	return errors.New("not implemented")
}

func (r *TokenRepo) DeleteExpired(ctx context.Context) error {
	return errors.New("not implemented")
}
