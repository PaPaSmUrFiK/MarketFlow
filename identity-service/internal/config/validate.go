package config

import (
	"fmt"
)

func validate(cfg *Config) error {
	if cfg.Env == "" {
		return fmt.Errorf("env is required")
	}

	if !cfg.Env.IsValid() {
		return fmt.Errorf("env must be one of: local, dev, prod")
	}

	// --- Database ---
	if cfg.Database.Host == "" {
		return fmt.Errorf("database.host is required")
	}

	if cfg.Database.Port < 1 || cfg.Database.Port > 65535 {
		return fmt.Errorf("database.port must be between 1 and 65535")
	}

	if cfg.Database.Name == "" {
		return fmt.Errorf("database.name is required")
	}

	if cfg.Database.MaxOpenConn <= 0 {
		return fmt.Errorf("database.max_open_conn must be > 0")
	}

	if cfg.Database.MaxIdleConn <= 0 {
		return fmt.Errorf("database.max_idle_conn must be > 0")
	}

	if cfg.Database.MaxIdleConn > cfg.Database.MaxOpenConn {
		return fmt.Errorf("database.max_idle_conn cannot exceed max_open_conn")
	}

	if cfg.Database.ConnMaxLifetime <= 0 {
		return fmt.Errorf("database.conn_max_lifetime must be > 0")
	}

	// --- GRPC ---
	if cfg.GRPC.Port < 1 || cfg.GRPC.Port > 65535 {
		return fmt.Errorf("grpc.port must be between 1 and 65535")
	}

	if cfg.GRPC.Timeout <= 0 {
		return fmt.Errorf("grpc.timeout must be > 0")
	}

	// --- JWT ---
	if cfg.JWT.AccessTTL <= 0 {
		return fmt.Errorf("jwt.access_ttl must be > 0")
	}

	if cfg.JWT.RefreshTTL <= 0 {
		return fmt.Errorf("jwt.refresh_ttl must be > 0")
	}

	if cfg.JWT.AccessTTL >= cfg.JWT.RefreshTTL {
		return fmt.Errorf("jwt.access_ttl must be less than jwt.refresh_ttl")
	}

	if cfg.JWT.Issuer == "" {
		return fmt.Errorf("jwt.issuer is required")
	}

	// --- Secrets ---
	if cfg.Secrets.Database.User == "" {
		return fmt.Errorf("DB_USER environment variable is required")
	}

	if cfg.Secrets.Database.Password == "" {
		return fmt.Errorf("DB_PASSWORD environment variable is required")
	}

	if len(cfg.Secrets.JWT.Secret) == 0 {
		return fmt.Errorf("JWT_SECRET environment variable is required")
	}

	if len(cfg.Secrets.JWT.Secret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 bytes")
	}

	return nil
}
