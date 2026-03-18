package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/domain"
	"github.com/google/uuid"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	return errors.New("not implemented")
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return nil, errors.New("not implemented")
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return nil, errors.New("not implemented")
}

func (r *UserRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.UserStatus) error {
	return errors.New("not implemented")
}

func (r *UserRepo) GetCredentials(ctx context.Context, userID uuid.UUID) (*domain.Credential, error) {
	return nil, errors.New("not implemented")
}

func (r *UserRepo) UpdateCredentials(ctx context.Context, cred *domain.Credential) error {
	return errors.New("not implemented")
}

func (r *UserRepo) GetUserWithRoles(ctx context.Context, userID uuid.UUID, appID uuid.UUID) (*domain.User, error) {
	return nil, errors.New("not implemented")
}

func (r *UserRepo) CreateIdentity(ctx context.Context, identity *domain.UserIdentity) error {
	return errors.New("not implemented")
}

func (r *UserRepo) GetIdentity(ctx context.Context, provider domain.OAuthProvider, providerUserID string) (*domain.UserIdentity, error) {
	return nil, errors.New("not implemented")
}

// нужен для admin.userAdmin
func (r *UserRepo) ListByApp(ctx context.Context, appID uuid.UUID) ([]domain.User, error) {
	return nil, errors.New("not implemented")
}

func (r *UserRepo) AssignRole(ctx context.Context, userID, roleID uuid.UUID) error {
	return errors.New("not implemented")
}

func (r *UserRepo) RemoveRole(ctx context.Context, userID, roleID uuid.UUID) error {
	return errors.New("not implemented")
}
