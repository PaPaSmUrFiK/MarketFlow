package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/domain"
	"github.com/google/uuid"
)

type SessionRepo struct {
	db *sql.DB
}

func NewSessionRepo(db *sql.DB) *SessionRepo {
	return &SessionRepo{db: db}
}

func (r *SessionRepo) Create(ctx context.Context, session *domain.Session) error {
	return errors.New("not implemented")
}

func (r *SessionRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	return nil, errors.New("not implemented")
}

func (r *SessionRepo) Revoke(ctx context.Context, sessionID uuid.UUID) error {
	return errors.New("not implemented")
}

func (r *SessionRepo) RevokeAllByUser(ctx context.Context, userID uuid.UUID, appID uuid.UUID) error {
	return errors.New("not implemented")
}

// нужен для user.sessionReader
func (r *SessionRepo) ListByUser(ctx context.Context, userID uuid.UUID, appID uuid.UUID) ([]domain.Session, error) {
	return nil, errors.New("not implemented")
}
