package config

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

func MustLoad() *Config {

	path := fetchConfigPath()
	if path == "" {
		panic("config path is empty")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config path does not exist: " + path)
	}

	cfgFile, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	var cfg Config
	if err := yaml.Unmarshal(cfgFile, &cfg); err != nil {
		panic(err)
	}

	applyDefaults(&cfg)

	cfg.Secrets = loadSecrets(&cfg)

	if err := validate(&cfg); err != nil {
		panic(err)
	}

	return &cfg
}

func loadSecrets(cfg *Config) Secrets {
	s := Secrets{
		Database: DatabaseSecrets{
			User:     mustGetEnv("DB_USER"),
			Password: mustGetEnv("DB_PASSWORD"),
		},
		JWT: JWTSecrets{
			Secret: []byte(mustGetEnv("JWT_SECRET")),
		},
	}

	// OAuth секреты — опциональны, загружаем только если провайдер включён
	if cfg.OAuth.Google.Enabled {
		s.OAuth.GoogleClientID = mustGetEnv("GOOGLE_CLIENT_ID")
		s.OAuth.GoogleClientSecret = mustGetEnv("GOOGLE_CLIENT_SECRET")
	}
	if cfg.OAuth.GitHub.Enabled {
		s.OAuth.GitHubClientID = mustGetEnv("GITHUB_CLIENT_ID")
		s.OAuth.GitHubClientSecret = mustGetEnv("GITHUB_CLIENT_SECRET")
	}

	return s
}

func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("environment variable %s is required", key))
	}
	return v
}

// fetchConfigPath fetches config path from command line flag or environment variable.
// Priority: flag > env > default.
// Default value is empty string.
func fetchConfigPath() string {
	var res string

	//--config="path/to/config.yaml"
	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}

func applyDefaults(cfg *Config) {
	if cfg.Database.MaxOpenConn == 0 {
		cfg.Database.MaxOpenConn = DefaultMaxOpenConns
	}
	if cfg.Database.MaxIdleConn == 0 {
		cfg.Database.MaxIdleConn = DefaultMaxIdleConns
	}
	if cfg.Database.ConnMaxLifetime == 0 {
		cfg.Database.ConnMaxLifetime = DefaultConnLifetime
	}
	if cfg.GRPC.Timeout == 0 {
		cfg.GRPC.Timeout = DefaultGRPCTimeout
	}
}
