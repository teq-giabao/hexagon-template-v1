package httpserver_test

import (
	"context"
	"hexagon/contact"
	"hexagon/httpserver"
	"hexagon/postgres"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	pgcontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/gorm"
)

func MustCreateServer(t testing.TB, db *gorm.DB) *httpserver.Server {
	t.Helper()

	contactService := contact.NewUsecase(postgres.NewContactRepository(db))

	server := httpserver.Default(testConfig())
	server.ContactService = contactService

	return server
}

// setupTestDatabase creates a new testcontainer PostgreSQL database and returns a GORM DB connection
func MustCreateTestDatabase(t testing.TB) *gorm.DB {
	t.Helper()
	ctx := context.Background()
	dbName, dbUser, dbPass := "test_contact", "test", "testpass"
	postgre, err := pgcontainer.RunContainer(ctx,
		testcontainers.WithImage("docker.io/postgres:15.2-alpine"),
		pgcontainer.WithDatabase(dbName),
		pgcontainer.WithUsername(dbUser),
		pgcontainer.WithPassword(dbPass),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(3*time.Second)),
	)
	assert.NoError(t, err, "failed to start postgres container")
	t.Cleanup(func() {
		err := postgre.Terminate(ctx)
		assert.NoError(t, err, "failed to terminate postgres container")
	})

	host, port := extractHostAndPort(t, ctx, postgre)
	db, err := postgres.NewConnection(postgres.Options{
		DBName:   dbName,
		DBUser:   dbUser,
		Password: dbPass,
		Host:     host,
		Port:     port.Port(),
	})
	assert.NoError(t, err, "failed to connect to postgres database")

	return db
}

func extractHostAndPort(t testing.TB, ctx context.Context, postgre *pgcontainer.PostgresContainer) (string, nat.Port) {
	t.Helper()
	host, err := postgre.Host(ctx)
	assert.NoError(t, err, "failed to get container host")

	port, err := postgre.MappedPort(ctx, "5432")
	assert.NoError(t, err, "failed to get mapped port")
	return host, port
}

// migrateTestDatabase runs all migration files against the test database
func MigrateTestDatabase(t testing.TB, db *gorm.DB, migrationPath string) {
	t.Helper()
	migrations := &migrate.FileMigrationSource{
		Dir: migrationPath,
	}

	sqlDB, err := db.DB()
	assert.NoError(t, err, "failed to get sql.DB from gorm.DB")

	_, err = migrate.Exec(sqlDB, "postgres", migrations, migrate.Up)
	assert.NoError(t, err, "failed to run database migrations")
}
