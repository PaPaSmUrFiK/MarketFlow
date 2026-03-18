package app

import (
	appgrpc "github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/app/grpc"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/config"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/jwt"
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
	const op = "app.New"

	storage, err := postgres.New(
		cfg.GetDSN(),
		cfg.Database.MaxOpenConn,
		cfg.Database.MaxIdleConn,
		cfg.Database.ConnMaxLifetime,
	)

	if err != nil {
		panic(err)
	}

	userRepo := postgres.NewUserRepo(storage)
	sessionRepo := postgres.NewSessionRepo(storage)
	tokenRepo := postgres.NewTokenRepo(storage)
	appRepo := postgres.NewAppRepo(storage)
	roleRepo := postgres.NewRoleRepo(storage)

	jwtManager, err := jwt.NewManager(
		string(cfg.GetJWTSecret()),
		cfg.JWT.Issuer,
		cfg.JWT.AccessTTL,
		cfg.JWT.RefreshTTL,
	)

	if err != nil {
		panic(err)
	}

	authSvc := auth.New(userRepo, appRepo, sessionRepo, tokenRepo, log, jwtManager)
	userSvc := user.New(userRepo, sessionRepo, log)
	adminSvc := admin.New(userRepo, roleRepo, appRepo, log)

	grpcApp := appgrpc.New(log, authSvc, userSvc, adminSvc, cfg.GRPC.Port)

	return &App{GRPCSrv: grpcApp}
}
