package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/domain"
	"github.com/google/uuid"
)

type AppRepo struct {
	db *sql.DB
}

func NewAppRepo(db *sql.DB) *AppRepo {
	return &AppRepo{db: db}
}

func (r *AppRepo) GetByCode(ctx context.Context, code string) (*domain.Application, error) {
	return nil, errors.New("not implemented")
}

func (r *AppRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Application, error) {
	return nil, errors.New("not implemented")
}

func (r *AppRepo) Create(ctx context.Context, app *domain.Application) error {
	return errors.New("not implemented")
}
