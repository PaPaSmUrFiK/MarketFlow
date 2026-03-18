package auth

import (
	"context"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/lib/sl"
	"log/slog"
	"time"

	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/domain"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/jwt"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/security"
	"github.com/google/uuid"
)

const (
	MaxFailedLoginAttempts = 5
	LockDuration           = 15 * time.Minute
)

// Интерфейсы обновлены под имена методов репозиториев (с суффиксами)

type userReader interface {
	CreateUser(ctx context.Context, user *domain.User) error
	CreateCredentials(ctx context.Context, cred *domain.Credential) error
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetCredentials(ctx context.Context, userID uuid.UUID) (*domain.Credential, error)
	UpdateCredentials(ctx context.Context, cred *domain.Credential) error
	GetUserWithRoles(ctx context.Context, userID uuid.UUID, appID uuid.UUID) (*domain.User, error)
	CreateIdentity(ctx context.Context, identity *domain.UserIdentity) error
	GetIdentity(ctx context.Context, provider domain.OAuthProvider, providerUserID string) (*domain.UserIdentity, error)
}

type appReader interface {
	GetAppByCode(ctx context.Context, code string) (*domain.Application, error)
	GetAppByID(ctx context.Context, id uuid.UUID) (*domain.Application, error)
}

type sessionStore interface {
	CreateSession(ctx context.Context, session *domain.Session) error
	GetSessionByID(ctx context.Context, id uuid.UUID) (*domain.Session, error)
	RevokeSession(ctx context.Context, sessionID uuid.UUID) error
	RevokeAllByUser(ctx context.Context, userID uuid.UUID, appID uuid.UUID) error
}

type tokenStore interface {
	CreateToken(ctx context.Context, token *domain.RefreshToken) error
	GetTokenByHash(ctx context.Context, hash string) (*domain.RefreshToken, error)
	RevokeToken(ctx context.Context, tokenID uuid.UUID) error
	RevokeAllBySession(ctx context.Context, sessionID uuid.UUID) error
	RevokeAllByUser(ctx context.Context, userID uuid.UUID, appID uuid.UUID) error
	DeleteExpired(ctx context.Context) (int64, error) // возвращает кол-во удалённых
}

type Service struct {
	users    userReader
	apps     appReader
	sessions sessionStore
	tokens   tokenStore
	log      *slog.Logger
	jwt      jwt.Manager
}

func New(
	users userReader,
	apps appReader,
	sessions sessionStore,
	tokens tokenStore,
	logger *slog.Logger,
	jwtManager jwt.Manager,
) *Service {
	return &Service{
		users:    users,
		apps:     apps,
		sessions: sessions,
		tokens:   tokens,
		log:      logger,
		jwt:      jwtManager,
	}
}

func (s *Service) RegisterNewUser(ctx context.Context, in RegisterInput) (*TokenPair, error) {
	const op = "auth.Service.Register"
	// TODO: реализовать
	return nil, sl.Err(op, domain.ErrNotImplemented)
}

func (s *Service) Login(ctx context.Context, in LoginInput) (*TokenPair, error) {
	const op = "auth.Service.Login"

	now := time.Now()

	user, err := s.users.GetUserByEmail(ctx, in.Email)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if user.Status != domain.UserActive {
		return nil, domain.ErrUserBlocked
	}

	cred, err := s.users.GetCredentials(ctx, user.ID)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if cred.IsLocked(now) {
		return nil, domain.ErrUserLocked
	}

	if !security.VerifyPassword(cred.PasswordHash, in.Password) {
		cred.RegisterFailedAttempt()

		if cred.FailedLoginAttempts >= MaxFailedLoginAttempts {
			lockUntil := now.Add(LockDuration)
			cred.LockedUntil = &lockUntil
		}

		_ = s.users.UpdateCredentials(ctx, cred)
		return nil, domain.ErrInvalidCredentials
	}

	cred.ResetFailedAttempts()
	cred.LockedUntil = nil
	if err := s.users.UpdateCredentials(ctx, cred); err != nil {
		return nil, sl.Err(op, err)
	}

	app, err := s.apps.GetAppByCode(ctx, in.AppCode)
	if err != nil {
		return nil, sl.Err(op, err)
	}

	userWithRoles, err := s.users.GetUserWithRoles(ctx, user.ID, app.ID)
	if err != nil {
		return nil, sl.Err(op, err)
	}

	roleCodes := extractRoleCodes(userWithRoles)

	return s.createSessionAndTokens(ctx, user.ID, app.ID, roleCodes, in, now)
}

func (s *Service) LoginWithOAuth(ctx context.Context, in OAuthLoginInput) (*TokenPair, error) {
	const op = "auth.Service.LoginWithOAuth"
	// TODO: реализовать
	return nil, sl.Err(op, domain.ErrNotImplemented)
}

func (s *Service) Refresh(ctx context.Context, rawToken string) (*TokenPair, error) {
	const op = "auth.Service.Refresh"

	now := time.Now()

	hash, err := security.HashRefreshToken(rawToken)
	if err != nil {
		return nil, sl.Err(op, err)
	}

	oldToken, err := s.tokens.GetTokenByHash(ctx, hash)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	if oldToken.RevokedAt != nil || now.After(oldToken.ExpiresAt) {
		return nil, domain.ErrInvalidCredentials
	}

	userWithRoles, err := s.users.GetUserWithRoles(ctx, oldToken.UserID, oldToken.AppID)
	if err != nil {
		return nil, sl.Err(op, err)
	}

	roleCodes := extractRoleCodes(userWithRoles)

	accessToken, err := s.jwt.GenerateAccessToken(oldToken.UserID, oldToken.AppID, roleCodes)
	if err != nil {
		return nil, sl.Err(op, err)
	}

	rawRefresh, newHash, err := security.GenerateRefreshToken()
	if err != nil {
		return nil, sl.Err(op, err)
	}

	newToken := domain.RefreshToken{
		UserID:    oldToken.UserID,
		AppID:     oldToken.AppID,
		SessionID: oldToken.SessionID,
		TokenHash: newHash,
		ExpiresAt: now.Add(s.jwt.RefreshTTL()),
	}

	// TODO: обернуть в транзакцию через TokenRepo.RotateToken
	if err := s.tokens.RevokeToken(ctx, oldToken.ID); err != nil {
		return nil, sl.Err(op, err)
	}

	if err := s.tokens.CreateToken(ctx, &newToken); err != nil {
		return nil, sl.Err(op, err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
	}, nil
}

func (s *Service) Logout(ctx context.Context, sessionID uuid.UUID) error {
	const op = "auth.Service.Logout"

	if err := s.sessions.RevokeSession(ctx, sessionID); err != nil {
		return sl.Err(op, err)
	}

	if err := s.tokens.RevokeAllBySession(ctx, sessionID); err != nil {
		return sl.Err(op, err)
	}

	return nil
}

func (s *Service) LogoutAll(ctx context.Context, userID, appID uuid.UUID) error {
	const op = "auth.Service.LogoutAll"

	if err := s.sessions.RevokeAllByUser(ctx, userID, appID); err != nil {
		return sl.Err(op, err)
	}

	if err := s.tokens.RevokeAllByUser(ctx, userID, appID); err != nil {
		return sl.Err(op, err)
	}

	return nil
}

func (s *Service) ValidateAccessToken(ctx context.Context, token string) (*jwt.AccessClaims, error) {
	const op = "auth.Service.ValidateAccessToken"
	// TODO: реализовать
	return nil, sl.Err(op, domain.ErrNotImplemented)
}

func (s *Service) createSessionAndTokens(
	ctx context.Context,
	userID uuid.UUID,
	appID uuid.UUID,
	roleCodes []string,
	in LoginInput,
	now time.Time,
) (*TokenPair, error) {
	const op = "auth.Service.createSessionAndTokens"

	session := domain.Session{
		UserID:    userID,
		AppID:     appID,
		UserAgent: in.UserAgent,
		IPAddress: in.IPAddress,
		ExpiresAt: now.Add(s.jwt.RefreshTTL()), // сессия живёт столько же сколько refresh токен
	}

	if err := s.sessions.CreateSession(ctx, &session); err != nil {
		return nil, sl.Err(op, err)
	}

	accessToken, err := s.jwt.GenerateAccessToken(userID, appID, roleCodes)
	if err != nil {
		return nil, sl.Err(op, err)
	}

	rawRefresh, hashRefresh, err := security.GenerateRefreshToken()
	if err != nil {
		return nil, sl.Err(op, err)
	}

	refresh := domain.RefreshToken{
		UserID:    userID,
		AppID:     appID,
		SessionID: session.ID,
		TokenHash: hashRefresh,
		ExpiresAt: now.Add(s.jwt.RefreshTTL()),
	}

	if err := s.tokens.CreateToken(ctx, &refresh); err != nil {
		return nil, sl.Err(op, err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
	}, nil
}
