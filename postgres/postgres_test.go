package postgres_test

import (
	"context"
	"hexagon/postgres"
	"testing"
	"time"

	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	pgcontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/gorm"
)

type Info struct {
	CurrentUser string `db:"current_user"`
}

func TestConnection(t *testing.T) {
	dbName, dbUser, dbPass := "test1", "test1", "123456"
	db := CreateConnection(t, dbName, dbUser, dbPass)
	MigrateTestDatabase(t, db, "../migrations")

	var info Info
	err := db.Raw("SELECT current_user").Scan(&info).Error
	assert.NoError(t, err)
	assert.Equal(t, dbUser, info.CurrentUser)
}

func TestNewConnection_Error(t *testing.T) {
	// Use invalid options to force a connection failure
	opts := postgres.Options{
		DBName:   "nonexistent",
		DBUser:   "invaliduser",
		Password: "wrongpass",
		Host:     "invalidhost", // Non-existent host to ensure failure
		Port:     "5432",
		SSLMode:  true,
	}

	_, err := postgres.NewConnection(opts)
	assert.Error(t, err) // Assert that an error is returned
}

func MigrateTestDatabase(t testing.TB, db *gorm.DB, migrationPath string) {
	t.Helper()

	migrations := &migrate.FileMigrationSource{
		Dir: migrationPath,
	}

	sqlDB, err := db.DB()
	assert.NoError(t, err)

	_, err = migrate.Exec(sqlDB, "postgres", migrations, migrate.Up)
	assert.NoError(t, err)
}

func CreateConnection(t testing.TB, dbName string, dbUser string, dbPass string) *gorm.DB {
	cont := SetupPostgresContainer(t, dbName, dbUser, dbPass)
	host, _ := cont.Host(context.Background())
	port, _ := cont.MappedPort(context.Background(), "5432")

	db, err := postgres.NewConnection(postgres.Options{
		DBName:   dbName,
		DBUser:   dbUser,
		Password: dbPass,
		Host:     host,
		Port:     port.Port(),
	})
	assert.NoError(t, err)

	return db
}

func SetupPostgresContainer(t testing.TB, dbname, user, password string) testcontainers.Container {
	ctx := context.Background()
	postgre, err := pgcontainer.RunContainer(ctx,
		testcontainers.WithImage("docker.io/postgres:15.2-alpine"),
		pgcontainer.WithDatabase(dbname),
		pgcontainer.WithUsername(user),
		pgcontainer.WithPassword(password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(3*time.Second)),
	)
	assert.NoError(t, err)

	t.Cleanup(func() {
		assert.NoError(t, postgre.Terminate(ctx))
	})

	return postgre
}
