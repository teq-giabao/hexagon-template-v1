// nolint: funlen
package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"hexagon/pkg/config"
)

func TestLoadConfig(t *testing.T) {
	t.Run("loads config from environment variables", func(t *testing.T) {
		// Setup environment variables
		envVars := map[string]string{
			"APP_ENV":                  "test",
			"PORT":                     "8080",
			"SENTRY_DSN":               "https://test@sentry.io/123",
			"ALLOW_ORIGINS":            "*",
			"DB_DRIVER":                "postgres",
			"DB_NAME":                  "testdb",
			"DB_HOST":                  "localhost",
			"DB_PORT":                  "5432",
			"DB_USER":                  "testuser",
			"DB_PASS":                  "testpass",
			"ENABLE_SSL":               "true",
			"DDB_REGION":               "ap-southeast-1",
			"DDB_ENDPOINT":             "http://localhost:8000",
			"DDB_CONTACTS_TABLE":       "contacts",
			"DDB_USERS_TABLE":          "users",
			"DDB_LOGIN_ATTEMPTS_TABLE": "login_attempts",
		}

		// Set environment variables
		for key, value := range envVars {
			t.Setenv(key, value)
		}

		// Load config
		cfg, err := config.LoadConfig()

		// Assertions
		require.NoError(t, err)
		require.NotNil(t, cfg)

		assert.Equal(t, "test", cfg.AppEnv)
		assert.Equal(t, 8080, cfg.Port)
		assert.Equal(t, "https://test@sentry.io/123", cfg.SentryDSN)
		assert.Equal(t, "*", cfg.AllowOrigins)
		assert.Equal(t, "testdb", cfg.DB.Name)
		assert.Equal(t, "localhost", cfg.DB.Host)
		assert.Equal(t, 5432, cfg.DB.Port)
		assert.Equal(t, "testuser", cfg.DB.User)
		assert.Equal(t, "testpass", cfg.DB.Pass)
		assert.True(t, cfg.DB.EnableSSL)
		assert.Equal(t, "postgres", cfg.DB.Driver)
		assert.Equal(t, "ap-southeast-1", cfg.DynamoDB.Region)
		assert.Equal(t, "http://localhost:8000", cfg.DynamoDB.Endpoint)
		assert.Equal(t, "contacts", cfg.DynamoDB.ContactsTable)
		assert.Equal(t, "users", cfg.DynamoDB.UsersTable)
		assert.Equal(t, "login_attempts", cfg.DynamoDB.LoginAttemptsTable)
	})

	t.Run("handles invalid port number", func(t *testing.T) {
		t.Setenv("PORT", "invalid")

		cfg, err := config.LoadConfig()

		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "load config error")
	})

	t.Run("handles invalid boolean value", func(t *testing.T) {
		t.Setenv("ENABLE_SSL", "not-a-boolean")

		cfg, err := config.LoadConfig()

		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "load config error")
	})

	t.Run("handles invalid DB port", func(t *testing.T) {
		t.Setenv("DB_PORT", "not-a-number")

		cfg, err := config.LoadConfig()

		assert.Error(t, err)
		assert.Nil(t, cfg)
		assert.Contains(t, err.Error(), "load config error")
	})
}
