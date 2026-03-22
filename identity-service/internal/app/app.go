package app

import (
	"context"
	appgrpc "github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/app/grpc"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/config"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/domain"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/jwt"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/oauth"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/service/admin"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/service/auth"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/service/user"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/storage/postgres"
	"log/slog"
)

type App struct {
	GRPCSrv *appgrpc.App
}

func New(log *slog.Logger, cfg *config.Config) *App {
	ctx := context.Background()

	pool, err := postgres.New(
		ctx,
		cfg.GetDSN(),
		cfg.Database.MaxOpenConn,
		cfg.Database.MaxIdleConn,
		cfg.Database.ConnMaxLifetime,
	)

	if err != nil {
		panic(err)
	}

	log.Info("connected to postgres database",
		slog.String("host", cfg.Database.Host),
		slog.Int("port", cfg.Database.Port),
		slog.String("db", cfg.Database.Name),
	)

	userRepo := postgres.NewUserRepo(pool)
	sessionRepo := postgres.NewSessionRepo(pool)
	tokenRepo := postgres.NewTokenRepo(pool)
	appRepo := postgres.NewAppRepo(pool)
	roleRepo := postgres.NewRoleRepo(pool)
	txManager := postgres.NewTx(pool)

	jwtManager, err := jwt.NewManager(
		string(cfg.GetJWTSecret()),
		cfg.JWT.Issuer,
		cfg.JWT.AccessTTL,
		cfg.JWT.RefreshTTL,
	)

	if err != nil {
		panic(err)
	}

	oauthProviders := buildOAuthProviders(cfg, log)

	authSvc := auth.New(userRepo, roleRepo, appRepo, sessionRepo, tokenRepo, txManager, oauthProviders, log, jwtManager)
	userSvc := user.New(userRepo, sessionRepo, oauthProviders, log)
	adminSvc := admin.New(userRepo, roleRepo, appRepo, log)

	grpcApp := appgrpc.New(log, authSvc, userSvc, adminSvc, jwtManager, cfg.GRPC.Port)

	return &App{GRPCSrv: grpcApp}
}

func buildOAuthProviders(cfg *config.Config, log *slog.Logger) map[domain.OAuthProvider]oauth.Provider {
	providers := make(map[domain.OAuthProvider]oauth.Provider)

	if cfg.OAuth.Google.Enabled {
		providers[domain.ProviderGoogle] = oauth.NewGoogleProvider(
			cfg.Secrets.OAuth.GoogleClientID,
			cfg.Secrets.OAuth.GoogleClientSecret,
			cfg.OAuth.Google.RedirectURI,
		)
		log.Info("oauth provider enabled", slog.String("provider", "google"))
	}

	if cfg.OAuth.GitHub.Enabled {
		providers[domain.ProviderGitHub] = oauth.NewGitHubProvider(
			cfg.Secrets.OAuth.GitHubClientID,
			cfg.Secrets.OAuth.GitHubClientSecret,
			cfg.OAuth.GitHub.RedirectURI,
		)
		log.Info("oauth provider enabled", slog.String("provider", "github"))
	}

	return providers
}
