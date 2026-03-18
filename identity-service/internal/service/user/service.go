package user

import (
	"context"
	"errors"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/domain"
	"github.com/google/uuid"
	"log/slog"
)

type userStore interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.UserStatus) error
	// смена пароля
	GetCredentials(ctx context.Context, userID uuid.UUID) (*domain.Credential, error)
	UpdateCredentials(ctx context.Context, cred *domain.Credential) error
	GetUserWithRoles(ctx context.Context, userID uuid.UUID, appID uuid.UUID) (*domain.User, error)
	// управление OAuth-провайдерами
	CreateIdentity(ctx context.Context, identity *domain.UserIdentity) error
	GetIdentity(ctx context.Context, provider domain.OAuthProvider, providerUserID string) (*domain.UserIdentity, error)
}

type sessionStore interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Session, error)
	ListByUser(ctx context.Context, userID uuid.UUID, appID uuid.UUID) ([]domain.Session, error)
	Revoke(ctx context.Context, sessionID uuid.UUID) error
}

type Service struct {
	users    userStore
	sessions sessionStore
	log      *slog.Logger
}

func New(users userStore, sessions sessionStore, logger *slog.Logger) *Service {
	return &Service{users: users, sessions: sessions, log: logger}
}

func (s *Service) GetMe(ctx context.Context, userID uuid.UUID, appID uuid.UUID) (*domain.User, error) {
	// TODO: реализовать — вернуть пользователя с ролями и permissions
	return nil, errors.New("not implemented")
}

func (s *Service) GetUserById(ctx context.Context, userID uuid.UUID, appID uuid.UUID) (*domain.User, error) {
	// TODO: реализовать
	return nil, errors.New("not implemented")
}

func (s *Service) ChangePassword(ctx context.Context, userID uuid.UUID, in ChangePasswordInput) error {
	// TODO: реализовать — проверить старый пароль, сохранить новый хэш
	return errors.New("not implemented")
}

func (s *Service) ListSessions(ctx context.Context, userID uuid.UUID, appID uuid.UUID) ([]domain.Session, error) {
	// TODO: реализовать
	return nil, errors.New("not implemented")
}

func (s *Service) RevokeSession(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) error {
	// TODO: реализовать — проверить что сессия принадлежит userID, затем отозвать
	return errors.New("not implemented")
}

func (s *Service) LinkIdentity(ctx context.Context, userID uuid.UUID, in LinkIdentityInput) error {
	// TODO: реализовать — обмен OAuth code на провайдерский user_id
	return errors.New("not implemented")
}

func (s *Service) UnlinkIdentity(ctx context.Context, userID uuid.UUID, provider string) error {
	// TODO: реализовать
	return errors.New("not implemented")
}
