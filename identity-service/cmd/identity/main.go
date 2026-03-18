package main

import (
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/app"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/config"
	"github.com/joho/godotenv"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	//инициализировать объект конфига
	// Load .env file if exists
	if _, err := os.Stat(".env"); err == nil {
		_ = godotenv.Load()
	}
	cfg := config.MustLoad()

	//инициализировать логгер
	log := setupLogger(cfg.Env)
	slog.SetDefault(log)

	log.Info("starting application",
		slog.String("env", string(cfg.Env)),
	)

	//инициализировать приложение(app)
	//позволяет запускать приложение не только с точки входа(main.go), но и допустим с тестов

	application := app.New(log, cfg)

	//запустить gRPC-сервер приложения
	go application.GRPCSrv.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	stopSignal := <-stop

	log.Info("stopping application", slog.String("signal", stopSignal.String()))

	application.GRPCSrv.Stop()

	log.Info("application stopped")
}

func setupLogger(env config.Environment) *slog.Logger {
	var handler slog.Handler

	switch env {
	case config.EnvLocal:
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})

	case config.EnvDev:
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})

	case config.EnvProd:
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})

	default:
		panic("unknown environment: " + string(env))
	}

	return slog.New(handler)
}
