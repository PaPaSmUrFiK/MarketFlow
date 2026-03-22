package user

import (
	"context"
	"fmt"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/lib/sl"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/oauth"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/security"
	"log/slog"
	"time"

	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/domain"
	"github.com/google/uuid"
)

const MinPasswordLength = 8

type userStore interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.UserStatus) error
	GetCredentials(ctx context.Context, userID uuid.UUID) (*domain.Credential, error)
	UpdateCredentials(ctx context.Context, cred *domain.Credential) error
	GetUserWithRoles(ctx context.Context, userID uuid.UUID, appID uuid.UUID) (*domain.User, error)
	CreateIdentity(ctx context.Context, identity *domain.UserIdentity) error
	GetIdentity(ctx context.Context, provider domain.OAuthProvider, providerUserID string) (*domain.UserIdentity, error)
	DeleteIdentity(ctx context.Context, userID uuid.UUID, provider domain.OAuthProvider) error
}

type sessionStore interface {
	GetSessionByID(ctx context.Context, id uuid.UUID) (*domain.Session, error)
	ListByUser(ctx context.Context, userID uuid.UUID, appID uuid.UUID) ([]domain.Session, error)
	RevokeSession(ctx context.Context, sessionID uuid.UUID) error
}

type Service struct {
	users          userStore
	sessions       sessionStore
	oauthProviders map[domain.OAuthProvider]oauth.Provider
	log            *slog.Logger
}

func New(
	users userStore,
	sessions sessionStore,
	oauthProviders map[domain.OAuthProvider]oauth.Provider,
	logger *slog.Logger,
) *Service {
	return &Service{
		users:          users,
		sessions:       sessions,
		oauthProviders: oauthProviders,
		log:            logger,
	}
}

// GetMe — профиль текущего пользователя с ролями и permissions.
func (s *Service) GetMe(ctx context.Context, userID uuid.UUID, appID uuid.UUID) (*domain.User, error) {
	const op = "user.Service.GetMe"

	user, err := s.users.GetUserWithRoles(ctx, userID, appID)
	if err != nil {
		return nil, sl.Err(op, err)
	}

	return user, nil
}

// GetUserById — профиль другого пользователя.
// Возвращает только базовые данные и роли — без credentials.
func (s *Service) GetUserById(ctx context.Context, userID uuid.UUID, appID uuid.UUID) (*domain.User, error) {
	const op = "user.Service.GetUserById"

	user, err := s.users.GetUserWithRoles(ctx, userID, appID)
	if err != nil {
		return nil, sl.Err(op, err)
	}

	return user, nil
}

// ChangePassword — смена пароля.
// Проверяем старый пароль, хэшируем новый, сбрасываем счётчик попыток и блокировку.
func (s *Service) ChangePassword(ctx context.Context, userID uuid.UUID, in ChangePasswordInput) error {
	const op = "user.Service.ChangePassword"

	log := s.log.With(slog.String("op", op), slog.String("user_id", userID.String()))

	if len(in.NewPassword) < MinPasswordLength {
		return domain.ErrPasswordTooWeak
	}

	// Загружаем текущие credentials
	cred, err := s.users.GetCredentials(ctx, userID)
	if err != nil {
		return sl.Err(op, err)
	}

	// OAuth пользователь может не иметь пароля — запрещаем смену через этот метод
	if cred.PasswordHash == "" {
		return fmt.Errorf("%w: oauth users cannot change password directly", domain.ErrInvalidCredentials)
	}

	if !security.VerifyPassword(cred.PasswordHash, in.OldPassword) {
		return domain.ErrInvalidCredentials
	}

	newHash, err := security.HashPassword(in.NewPassword)
	if err != nil {
		return sl.Err(op, fmt.Errorf("hash password: %w", err))
	}

	// Сохраняем — сбрасываем блокировку и счётчик попыток
	now := time.Now()
	cred.PasswordHash = newHash
	cred.LastPasswordChange = &now
	cred.ResetFailedAttempts()
	cred.LockedUntil = nil

	if err := s.users.UpdateCredentials(ctx, cred); err != nil {
		return sl.Err(op, err)
	}

	log.Info("password changed successfully")
	return nil
}

// ListSessions — список активных сессий пользователя в приложении.
func (s *Service) ListSessions(ctx context.Context, userID uuid.UUID, appID uuid.UUID) ([]domain.Session, error) {
	const op = "user.Service.ListSessions"

	sessions, err := s.sessions.ListByUser(ctx, userID, appID)
	if err != nil {
		return nil, sl.Err(op, err)
	}

	return sessions, nil
}

// RevokeSession — пользователь завершает конкретную сессию.
// Проверяем что сессия принадлежит этому пользователю — защита от IDOR.
func (s *Service) RevokeSession(ctx context.Context, userID uuid.UUID, sessionID uuid.UUID) error {
	const op = "user.Service.RevokeSession"

	// Загружаем сессию чтобы проверить владельца
	session, err := s.sessions.GetSessionByID(ctx, sessionID)
	if err != nil {
		return sl.Err(op, err)
	}

	// IDOR защита — нельзя отозвать чужую сессию
	if session.UserID != userID {
		s.log.Warn("attempt to revoke foreign session",
			slog.String("op", op),
			slog.String("user_id", userID.String()),
			slog.String("session_owner", session.UserID.String()),
			slog.String("session_id", sessionID.String()),
		)
		return domain.ErrPermissionDenied
	}

	if err := s.sessions.RevokeSession(ctx, sessionID); err != nil {
		return sl.Err(op, err)
	}

	return nil
}

// LinkIdentity — привязывает OAuth провайдера к существующему аккаунту.
// Обменивает code на профиль провайдера и создаёт запись в user_identities.
func (s *Service) LinkIdentity(ctx context.Context, userID uuid.UUID, in LinkIdentityInput) error {
	const op = "user.Service.LinkIdentity"

	log := s.log.With(
		slog.String("op", op),
		slog.String("user_id", userID.String()),
		slog.String("provider", in.Provider),
	)

	// Проверяем что провайдер поддерживается
	provider := domain.OAuthProvider(in.Provider)
	oauthClient, ok := s.oauthProviders[provider]
	if !ok {
		return fmt.Errorf("unsupported oauth provider: %s", in.Provider)
	}

	// Получаем профиль от провайдера — HTTP запрос до записи в БД
	userInfo, err := oauthClient.GetUserInfo(ctx, in.Code)
	if err != nil {
		log.Error("failed to get user info from provider", sl.ErrAttr(err))
		return fmt.Errorf("oauth provider error: %w", err)
	}

	// Проверяем что этот провайдерский аккаунт не привязан к другому пользователю
	existing, err := s.users.GetIdentity(ctx, provider, userInfo.ID)
	if err != nil && err != domain.ErrIdentityNotFound {
		return sl.Err(op, err)
	}
	if existing != nil {
		if existing.UserID == userID {
			// Уже привязан к этому же пользователю
			return nil
		}
		// Привязан к другому пользователю
		return domain.ErrIdentityAlreadyExists
	}

	identity := &domain.UserIdentity{
		UserID:         userID,
		Provider:       provider,
		ProviderUserID: userInfo.ID,
	}
	if err := s.users.CreateIdentity(ctx, identity); err != nil {
		return sl.Err(op, err)
	}

	log.Info("identity linked successfully",
		slog.String("provider_user_id", userInfo.ID),
	)
	return nil
}

// UnlinkIdentity — отвязывает OAuth провайдера от аккаунта.
// Проверяем что у пользователя есть способ войти без этого провайдера.
func (s *Service) UnlinkIdentity(ctx context.Context, userID uuid.UUID, provider string) error {
	const op = "user.Service.UnlinkIdentity"

	log := s.log.With(
		slog.String("op", op),
		slog.String("user_id", userID.String()),
		slog.String("provider", provider),
	)

	oauthProvider := domain.OAuthProvider(provider)

	cred, err := s.users.GetCredentials(ctx, userID)
	if err != nil {
		return sl.Err(op, err)
	}

	// Защита от потери доступа — нельзя отвязать провайдера если нет пароля
	if cred.PasswordHash == "" {
		log.Warn("attempt to unlink last auth method")
		return fmt.Errorf("%w: cannot unlink provider — no password set", domain.ErrPermissionDenied)
	}

	if err := s.users.DeleteIdentity(ctx, userID, oauthProvider); err != nil {
		return sl.Err(op, err)
	}

	log.Info("identity unlinked successfully")
	return nil
}
