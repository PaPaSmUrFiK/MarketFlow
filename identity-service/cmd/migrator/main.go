package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/joho/godotenv"
	"os"

	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var migrationsPath string

	flag.StringVar(&migrationsPath, "migrations-path", "./migrations", "path to migrations directory")
	flag.Parse()

	if migrationsPath == "" {
		panic("migrations-path is required")
	}

	if _, err := os.Stat(".env"); err == nil {
		_ = godotenv.Load()
	}

	cfg := config.MustLoad()

	// schema_migrations создаётся в public — identity_user теперь имеет права
	// CREATE TABLE из миграций идут в identity благодаря search_path роли
	dbURL := fmt.Sprintf(
		"pgx5://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Secrets.Database.User,
		cfg.Secrets.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
	)

	m, err := migrate.New("file://"+migrationsPath, dbURL)
	if err != nil {
		panic(fmt.Sprintf("migrate init: %v", err))
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			fmt.Printf("source close error: %v\n", srcErr)
		}
		if dbErr != nil {
			fmt.Printf("db close error: %v\n", dbErr)
		}
	}()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")
			return
		}
		panic(fmt.Sprintf("migrate up: %v", err))
	}

	fmt.Println("migrations applied successfully")

}
