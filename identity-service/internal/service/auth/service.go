package auth

import (
	"context"
	"errors"
	"fmt"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/oauth"
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/domain"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/jwt"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/lib/sl"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/security"
	"github.com/google/uuid"
)

const (
	MaxFailedLoginAttempts = 5
	LockDuration           = 15 * time.Minute
	MinPasswordLength      = 8
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
	AssignRole(ctx context.Context, userID, roleID uuid.UUID) error
}

type appReader interface {
	GetAppByCode(ctx context.Context, code string) (*domain.Application, error)
	GetAppByID(ctx context.Context, id uuid.UUID) (*domain.Application, error)
}

type roleReader interface {
	GetRoleByCode(ctx context.Context, appID uuid.UUID, code string) (*domain.Role, error)
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

type TxManager interface {
	Transactional(ctx context.Context, f func(ctx context.Context) error) error
}

type Service struct {
	users          userReader
	roles          roleReader
	apps           appReader
	sessions       sessionStore
	tokens         tokenStore
	txManager      TxManager
	oauthProviders map[domain.OAuthProvider]oauth.Provider
	log            *slog.Logger
	jwt            jwt.Manager
}

func New(
	users userReader,
	roles roleReader,
	apps appReader,
	sessions sessionStore,
	tokens tokenStore,
	txManager TxManager,
	oauthProviders map[domain.OAuthProvider]oauth.Provider,
	logger *slog.Logger,
	jwtManager jwt.Manager,
) *Service {
	return &Service{
		users:          users,
		roles:          roles,
		apps:           apps,
		sessions:       sessions,
		tokens:         tokens,
		txManager:      txManager,
		oauthProviders: oauthProviders,
		log:            logger,
		jwt:            jwtManager,
	}
}

func (s *Service) RegisterNewUser(ctx context.Context, in RegisterInput) (*TokenPair, error) {
	const op = "auth.Service.Register"

	log := s.log.With(
		slog.String("op", op),
		slog.String("app_code", in.AppCode),
	)

	if err := validateRegisterInput(in); err != nil {
		log.Warn("validation failed", slog.String("reason", err.Error()))
		return nil, err
	}

	passHash, err := security.HashPassword(in.Password)
	if err != nil {
		return nil, sl.Err(op, fmt.Errorf("hash password: %w", err))
	}

	app, err := s.apps.GetAppByCode(ctx, in.AppCode)
	if err != nil {
		if errors.Is(err, domain.ErrApplicationNotFound) {
			log.Warn("application not found", slog.String("app_code", in.AppCode))
			return nil, domain.ErrApplicationNotFound
		}
		return nil, sl.Err(op, err)
	}

	defaultRole, err := s.roles.GetRoleByCode(ctx, app.ID, string(domain.DefaultUserRole))
	if err != nil {
		if errors.Is(err, domain.ErrRoleNotFound) {
			log.Error("default USER role not found for app",
				slog.String("app_id", app.ID.String()),
			)
			return nil, fmt.Errorf("%s: default role USER not configured for app %s", op, in.AppCode)
		}
		return nil, sl.Err(op, err)
	}

	var finalTokens *TokenPair

	err = s.txManager.Transactional(ctx, func(ctx context.Context) error {
		newUser := &domain.User{
			Status: domain.UserActive,
		}

		if err := s.users.CreateUser(ctx, newUser); err != nil {
			switch {
			case errors.Is(err, domain.ErrUserAlreadyExists):
				log.Warn("user already exists", slog.String("email", in.Email))
				return domain.ErrUserAlreadyExists
			case errors.Is(err, domain.ErrInvalidUserStatus):
				log.Error("invalid user status", slog.String("status", string(newUser.Status)))
				return domain.ErrInvalidUserStatus
			case errors.Is(err, domain.ErrUserStatusRequired):
				log.Error("user status required")
				return domain.ErrUserStatusRequired
			default:
				return sl.Err(op, err)
			}
		}

		cred := &domain.Credential{
			UserID:        newUser.ID,
			Email:         in.Email,
			PasswordHash:  passHash,
			EmailVerified: false,
		}

		if err := s.users.CreateCredentials(ctx, cred); err != nil {
			if errors.Is(err, domain.ErrEmailAlreadyExists) {
				log.Warn("email already exists", slog.String("email", in.Email))
				return domain.ErrEmailAlreadyExists
			}
			return sl.Err(op, err)
		}

		if err := s.users.AssignRole(ctx, newUser.ID, defaultRole.ID); err != nil {
			return sl.Err(op, fmt.Errorf("assign default role: %w", err))
		}

		roleCodes := []string{string(domain.DefaultUserRole)}

		tokens, err := s.createSessionAndTokens(ctx, newUser.ID, app.ID, roleCodes, LoginInput{
			AppCode:   in.AppCode,
			Email:     in.Email,
			UserAgent: in.UserAgent,
			IPAddress: in.IPAddress,
		}, time.Now())

		if err != nil {
			return sl.Err(op, err)
		}
		finalTokens = tokens

		return nil
	})

	if err != nil {
		return nil, err
	}

	return finalTokens, nil
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

		if err := s.users.UpdateCredentials(ctx, cred); err != nil {
			s.log.Error("failed to update failed attempts counter",
				slog.String("user_id", user.ID.String()),
				sl.ErrAttr(err),
			)
		}
		return nil, domain.ErrInvalidCredentials
	}

	cred.ResetFailedAttempts()
	cred.LockedUntil = nil
	if err := s.users.UpdateCredentials(ctx, cred); err != nil {
		s.log.Error("failed to reset failed attempts counter",
			slog.String("user_id", user.ID.String()),
			sl.ErrAttr(err),
		)
	}

	app, err := s.apps.GetAppByCode(ctx, in.AppCode)
	if err != nil {
		if errors.Is(err, domain.ErrApplicationNotFound) {
			return nil, domain.ErrApplicationNotFound
		}
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

	log := s.log.With(
		slog.String("op", op),
		slog.String("provider", in.Provider),
		slog.String("app_code", in.AppCode),
	)

	if in.AppCode == "" || in.Provider == "" || in.Code == "" {
		return nil, domain.ErrInvalidCredentials
	}

	//Проверяем что провайдер поддерживается и включён
	provider := domain.OAuthProvider(in.Provider)
	oauthClient, ok := s.oauthProviders[provider]
	if !ok {
		log.Warn("unsupported oauth provider", slog.String("provider", in.Provider))
		return nil, fmt.Errorf("unsupported oauth provider: %s", in.Provider)
	}

	// Обмениваем code на профиль пользователя у провайдера
	// Это HTTP запрос к Google/GitHub — делаем до открытия транзакции
	userInfo, err := oauthClient.GetUserInfo(ctx, in.Code)
	if err != nil {
		log.Error("failed to get user info from provider",
			slog.String("provider", in.Provider),
			sl.ErrAttr(err),
		)
		return nil, fmt.Errorf("oauth provider error: %w", err)
	}

	log.Info("got user info from provider",
		slog.String("provider_user_id", userInfo.ID),
	)

	// Проверяем приложение
	app, err := s.apps.GetAppByCode(ctx, in.AppCode)
	if err != nil {
		if errors.Is(err, domain.ErrApplicationNotFound) {
			return nil, domain.ErrApplicationNotFound
		}
		return nil, sl.Err(op, err)
	}

	now := time.Now()

	// Ищем существующую привязку provider → наш пользователь
	identity, err := s.users.GetIdentity(ctx, provider, userInfo.ID)
	if err != nil && !errors.Is(err, domain.ErrIdentityNotFound) {
		return nil, sl.Err(op, err)
	}

	// Пользователь уже логинился через этого провайдера
	if identity != nil {
		user, err := s.users.GetUserByID(ctx, identity.UserID)
		if err != nil {
			return nil, sl.Err(op, err)
		}

		if user.Status != domain.UserActive {
			log.Warn("oauth login for non-active user",
				slog.String("user_id", user.ID.String()),
				slog.String("status", string(user.Status)),
			)
			return nil, domain.ErrUserBlocked
		}

		userWithRoles, err := s.users.GetUserWithRoles(ctx, user.ID, app.ID)
		if err != nil {
			return nil, sl.Err(op, err)
		}

		log.Info("existing user logged in via oauth",
			slog.String("user_id", user.ID.String()),
		)

		return s.createSessionAndTokens(ctx, user.ID, app.ID, extractRoleCodes(userWithRoles), LoginInput{
			AppCode:   in.AppCode,
			UserAgent: in.UserAgent,
			IPAddress: in.IPAddress,
		}, now)
	}

	// Первый вход — создаём нового пользователя
	defaultRole, err := s.roles.GetRoleByCode(ctx, app.ID, string(domain.DefaultUserRole))
	if err != nil {
		if errors.Is(err, domain.ErrRoleNotFound) {
			log.Error("default USER role not configured",
				slog.String("app_id", app.ID.String()),
			)
			return nil, fmt.Errorf("%w: role USER not configured for app %s", domain.ErrInternal, in.AppCode)
		}
		return nil, sl.Err(op, err)
	}

	var finalTokens *TokenPair

	err = s.txManager.Transactional(ctx, func(ctx context.Context) error {
		newUser := &domain.User{Status: domain.UserActive}
		if err := s.users.CreateUser(ctx, newUser); err != nil {
			return sl.Err(op, err)
		}

		// Привязываем провайдера — userInfo.ID это ProviderUserID от Google/GitHub
		newIdentity := &domain.UserIdentity{
			UserID:         newUser.ID,
			Provider:       provider,
			ProviderUserID: userInfo.ID,
		}
		if err := s.users.CreateIdentity(ctx, newIdentity); err != nil {
			if errors.Is(err, domain.ErrIdentityAlreadyExists) {
				return domain.ErrIdentityAlreadyExists
			}
			return sl.Err(op, err)
		}

		// Если провайдер отдал email — сохраняем без пароля
		if userInfo.Email != "" {
			cred := &domain.Credential{
				UserID:        newUser.ID,
				Email:         userInfo.Email,
				PasswordHash:  "",   // нет пароля у OAuth пользователя
				EmailVerified: true, // провайдер подтвердил email
			}
			if err := s.users.CreateCredentials(ctx, cred); err != nil {
				if errors.Is(err, domain.ErrEmailAlreadyExists) {
					log.Warn("email from oauth already taken",
						slog.String("email", userInfo.Email),
					)
					return domain.ErrEmailAlreadyExists
				}
				return sl.Err(op, err)
			}
		}

		if err := s.users.AssignRole(ctx, newUser.ID, defaultRole.ID); err != nil {
			return sl.Err(op, fmt.Errorf("assign default role: %w", err))
		}

		tokens, err := s.createSessionAndTokens(ctx, newUser.ID, app.ID,
			[]string{string(domain.DefaultUserRole)},
			LoginInput{
				AppCode:   in.AppCode,
				UserAgent: in.UserAgent,
				IPAddress: in.IPAddress,
			}, now)
		if err != nil {
			return sl.Err(op, err)
		}

		finalTokens = tokens
		return nil
	})

	if err != nil {
		return nil, err
	}

	log.Info("new user created via oauth",
		slog.String("provider", in.Provider),
		slog.String("provider_user_id", userInfo.ID),
	)

	return finalTokens, nil
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

	if oldToken.IsRevoked() || oldToken.IsExpired(now) {
		return nil, domain.ErrInvalidCredentials
	}

	//Проверяем статус пользователя — мог быть заблокирован после выдачи токена
	user, err := s.users.GetUserByID(ctx, oldToken.UserID)
	if err != nil {
		return nil, sl.Err(op, err)
	}
	if user.Status != domain.UserActive {
		return nil, domain.ErrUserBlocked
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

	err = s.txManager.Transactional(ctx, func(ctx context.Context) error {
		if err := s.tokens.RevokeToken(ctx, oldToken.ID); err != nil {
			return err
		}
		return s.tokens.CreateToken(ctx, &newToken)
	})
	if err != nil {
		return nil, sl.Err(op, err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
	}, nil
}

func (s *Service) Logout(ctx context.Context, sessionID uuid.UUID) error {
	const op = "auth.Service.Logout"

	err := s.txManager.Transactional(ctx, func(ctx context.Context) error {
		if err := s.sessions.RevokeSession(ctx, sessionID); err != nil {
			if errors.Is(err, domain.ErrSessionRevoked) {
				return nil
			}
			return err
		}
		return s.tokens.RevokeAllBySession(ctx, sessionID)
	})
	if err != nil {
		return sl.Err(op, err)
	}

	s.log.Info("user logged out", slog.String("op", op), slog.String("session_id", sessionID.String()))
	return nil
}

func (s *Service) LogoutAll(ctx context.Context, userID, appID uuid.UUID) error {
	const op = "auth.Service.LogoutAll"

	err := s.txManager.Transactional(ctx, func(ctx context.Context) error {
		if err := s.sessions.RevokeAllByUser(ctx, userID, appID); err != nil {
			return err
		}
		return s.tokens.RevokeAllByUser(ctx, userID, appID)
	})

	if err != nil {
		return sl.Err(op, err)
	}

	s.log.Info("all sessions revoked",
		slog.String("op", op),
		slog.String("user_id", userID.String()),
		slog.String("app_id", appID.String()),
	)
	return nil
}

func (s *Service) ValidateAccessToken(ctx context.Context, token string) (*jwt.AccessClaims, error) {
	const op = "auth.Service.ValidateAccessToken"

	claims, err := s.jwt.ParseAccessToken(token)

	if err != nil {
		if errors.Is(err, jwt.ErrExpiredToken) {
			return nil, domain.ErrRefreshTokenExpired
		}
		return nil, domain.ErrInvalidCredentials
	}

	return claims, nil
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

// validateRegisterInput verifies the correctness of the registration input data
func validateRegisterInput(in RegisterInput) error {
	if in.Email == "" {
		return domain.ErrEmailRequired
	}
	if !isValidEmail(in.Email) {
		return domain.ErrInvalidEmailFormat
	}

	if in.Password == "" {
		return domain.ErrPasswordRequired
	}
	if len(in.Password) < MinPasswordLength {
		return domain.ErrPasswordTooWeak
	}

	if in.AppCode == "" {
		return domain.ErrAppCodeRequired
	}

	if net.ParseIP(in.IPAddress) == nil {
		return domain.ErrInvalidIPAddress
	}

	return nil
}

// isValidEmail performs simple validation of the email format
func isValidEmail(email string) bool {
	atIndex := strings.Index(email, "@")
	if atIndex == -1 {
		return false
	}

	dotIndex := strings.LastIndex(email, ".")
	if dotIndex == -1 || dotIndex < atIndex+1 || dotIndex == len(email)-1 {
		return false
	}

	return atIndex > 0 && dotIndex > atIndex+1
}
