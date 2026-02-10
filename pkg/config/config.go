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
	GRPCPort     int    `envconfig:"GRPC_PORT"`
	SentryDSN    string `envconfig:"SENTRY_DSN"`
	AllowOrigins string `envconfig:"ALLOW_ORIGINS"`

	DB struct {
		Driver    string `envconfig:"DB_DRIVER"`
		Name      string `envconfig:"DB_NAME"`
		Host      string `envconfig:"DB_HOST"`
		Port      int    `envconfig:"DB_PORT"`
		User      string `envconfig:"DB_USER"`
		Pass      string `envconfig:"DB_PASS"`
		EnableSSL bool   `envconfig:"ENABLE_SSL"`
	}
	DynamoDB struct {
		Region             string `envconfig:"DDB_REGION"`
		Endpoint           string `envconfig:"DDB_ENDPOINT"`
		AccessKey          string `envconfig:"DDB_ACCESS_KEY"`
		SecretKey          string `envconfig:"DDB_SECRET_KEY"`
		SessionToken       string `envconfig:"DDB_SESSION_TOKEN"`
		ContactsTable      string `envconfig:"DDB_CONTACTS_TABLE"`
		UsersTable         string `envconfig:"DDB_USERS_TABLE"`
		LoginAttemptsTable string `envconfig:"DDB_LOGIN_ATTEMPTS_TABLE"`
	}
	Auth struct {
		JWTSecret          string `envconfig:"AUTH_JWT_SECRET"`
		TokenTTL           int    `envconfig:"AUTH_TOKEN_TTL"`
		RefreshTTL         int    `envconfig:"AUTH_REFRESH_TTL"`
		GoogleClientID     string `envconfig:"AUTH_GOOGLE_CLIENT_ID"`
		GoogleClientSecret string `envconfig:"AUTH_GOOGLE_CLIENT_SECRET"`
		GoogleRedirectURL  string `envconfig:"AUTH_GOOGLE_REDIRECT_URL"`
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
