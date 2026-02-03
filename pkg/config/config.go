package config

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

var Empty = new(Config)

type Config struct {
	AppEnv       string `envconfig:"APP_ENV"`
	Port         int    `envconfig:"PORT"`
	SentryDSN    string `envconfig:"SENTRY_DSN"`
	AllowOrigins string `envconfig:"ALLOW_ORIGINS"`

	DB struct {
		Name      string `envconfig:"DB_NAME"`
		Host      string `envconfig:"DB_HOST"`
		Port      int    `envconfig:"DB_PORT"`
		User      string `envconfig:"DB_USER"`
		Pass      string `envconfig:"DB_PASS"`
		EnableSSL bool   `envconfig:"ENABLE_SSL"`
	}
	Auth struct {
		JWTSecret  string `envconfig:"AUTH_JWT_SECRET"`
		TokenTTL   int    `envconfig:"AUTH_TOKEN_TTL"`
		RefreshTTL int    `envconfig:"AUTH_REFRESH_TTL"`
	}
}

func LoadConfig() (*Config, error) {
	// load default .env file, ignore the error
	_ = godotenv.Load()

	cfg := new(Config)
	err := envconfig.Process("", cfg)
	if err != nil {
		return nil, fmt.Errorf("load config error: %v", err)
	}

	return cfg, nil
}
