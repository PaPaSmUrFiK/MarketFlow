package postgres

import (
	"context"
	"fmt"
	"github.com/PaPaSmUrFiK/MarketFlow/identity-service/internal/lib/sl"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

func New(ctx context.Context,
	dsn string,
	maxOpen, maxIdle int,
	lifetime time.Duration,
) (*pgxpool.Pool, error) {
	const op = "postgres.New"
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, sl.Err(op, err)
	}

	// MaxConns — максимальное количество соединений в пуле.
	// При высокой нагрузке запросы ждут свободного соединения.
	cfg.MaxConns = int32(maxOpen)

	// MinConns — минимум "прогретых" соединений.
	// pgxpool держит их открытыми чтобы не тратить время на handshake
	// при первых запросах после простоя.
	cfg.MinConns = int32(maxIdle)

	// MaxConnLifetime — соединение закрывается и пересоздаётся через это время.
	// Защита от "протухших" соединений и балансировки нагрузки на стороне Postgres.
	cfg.MaxConnLifetime = lifetime

	// NewWithConfig создаёт пул с нашей конфигурацией.
	// Само подключение происходит лениво — при первом запросе.
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("pgxpool.NewWithConfig: %w", err)
	}

	// Ping проверяет что соединение реально работает прямо сейчас.
	// Без Ping приложение запустится но упадёт при первом запросе к БД.
	if err := pool.Ping(ctx); err != nil {
		pool.Close() // закрываем пул если Ping не прошёл
		return nil, fmt.Errorf("pool.Ping: %w", err)
	}

	return pool, nil
}
