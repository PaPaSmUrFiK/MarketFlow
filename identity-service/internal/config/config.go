package config

import (
	"fmt"
	"time"
)

type Environment string

const (
	EnvLocal Environment = "local"
	EnvDev   Environment = "dev"
	EnvProd  Environment = "prod"
)

func (e Environment) IsValid() bool {
	switch e {
	case EnvLocal, EnvDev, EnvProd:
		return true
	default:
		return false
	}
}

const (
	DefaultMaxOpenConns = 25
	DefaultMaxIdleConns = 10
	DefaultConnLifetime = 30 * time.Minute
	DefaultGRPCTimeout  = 10 * time.Second
)

type Config struct {
	Env      Environment    `yaml:"env"`
	Database DatabaseConfig `yaml:"database"`
	GRPC     GRPCConfig     `yaml:"grpc"`
	JWT      JWTConfig      `yaml:"jwt"`
	Secrets  Secrets        `yaml:"-"`
}

type DatabaseConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	Name            string        `yaml:"name"`
	MaxOpenConn     int           `yaml:"max_open_conn"`
	MaxIdleConn     int           `yaml:"max_idle_conn"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

type JWTConfig struct {
	AccessTTL  time.Duration `yaml:"access_ttl"`
	RefreshTTL time.Duration `yaml:"refresh_ttl"`
	Issuer     string        `yaml:"issuer"`
}

type Secrets struct {
	Database DatabaseSecrets
	JWT      JWTSecrets
}

type DatabaseSecrets struct {
	User     string
	Password string
}

type JWTSecrets struct {
	Secret []byte
}

// GetDSN строит URL-формат который понимает pgx/pgxpool:
// postgres://user:password@host:port/dbname?sslmode=disable
//
// pgxpool.ParseConfig принимает именно URL, а не key=value строку.
// key=value формат (старый GetDSN) работал только с database/sql + std
func (c *Config) GetDSN() string {
	sslMode := "disable"
	if c.Env == EnvProd {
		sslMode = "require"
	}

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Secrets.Database.User,
		c.Secrets.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Name,
		sslMode,
	)
}

// GetJWTSecret returns JWT secret for token operations
func (c *Config) GetJWTSecret() []byte {
	return c.Secrets.JWT.Secret
}
