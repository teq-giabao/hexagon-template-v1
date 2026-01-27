package postgres_test

import (
	"context"
	"hexagon/contact"
	"hexagon/postgres"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestContactRepository_CreateContact(t *testing.T) {
	// Arrange - Setup shared database container and connection
	dbName, dbUser, dbPass := "contact_test", "testuser", "testpass"
	db := CreateConnection(t, dbName, dbUser, dbPass)
	MigrateTestDatabase(t, db, "../migrations")

	t.Run("successfully creates a contact", func(t *testing.T) {
		// Arrange
		cleanupContactDatabase(t, db)
		repo := postgres.NewContactRepository(db)
		testContact := contact.Contact{Name: "John Doe", Phone: "1234567890"}

		// Act
		err := repo.CreateContact(context.Background(), testContact)

		// Assert
		require.NoError(t, err)
		assertContactExists(t, db, testContact)
	})

	t.Run("creates multiple contacts", func(t *testing.T) {
		// Arrange
		cleanupContactDatabase(t, db)
		repo := postgres.NewContactRepository(db)
		contacts := []contact.Contact{
			{Name: "Alice Smith", Phone: "1111111111"},
			{Name: "Bob Johnson", Phone: "2222222222"},
			{Name: "Charlie Brown", Phone: "3333333333"},
		}

		// Act
		for _, c := range contacts {
			err := repo.CreateContact(context.Background(), c)
			require.NoError(t, err)
		}

		// Assert
		assertContacts(t, contacts, db)
	})
}

func TestContactRepository_AllContacts(t *testing.T) {
	// Arrange - Setup shared database container and connection
	dbName, dbUser, dbPass := "contact_all_test", "testuser", "testpass"
	db := CreateConnection(t, dbName, dbUser, dbPass)
	MigrateTestDatabase(t, db, "../migrations")

	t.Run("returns all contacts", func(t *testing.T) {
		// Arrange
		cleanupContactDatabase(t, db)
		repo := postgres.NewContactRepository(db)
		expectedContacts := []contact.Contact{
			{Name: "Alice Smith", Phone: "1111111111"},
			{Name: "Bob Johnson", Phone: "2222222222"},
			{Name: "Charlie Brown", Phone: "3333333333"},
		}
		mustCreateContacts(t, db, expectedContacts)

		// Act
		contacts, err := repo.AllContacts(context.Background())

		// Assert
		require.NoError(t, err)
		assert.Len(t, contacts, len(expectedContacts))
		assert.Equal(t, expectedContacts, contacts)
	})

	t.Run("returns empty list when no contacts exist", func(t *testing.T) {
		// Arrange
		cleanupContactDatabase(t, db)
		repo := postgres.NewContactRepository(db)

		// Act
		contacts, err := repo.AllContacts(context.Background())

		// Assert
		require.NoError(t, err)
		assertNoContacts(t, contacts)
	})

	t.Run("fails with closed database connection", func(t *testing.T) {
		// Arrange
		cleanupContactDatabase(t, db)
		repo := postgres.NewContactRepository(db)
		mustCloseDBConnection(db)

		// Act
		_, err := repo.AllContacts(context.Background())

		// Assert
		assert.Error(t, err)
	})
}

func mustCloseDBConnection(db *gorm.DB) {
	sqlDB, _ := db.DB()
	sqlDB.Close()
}

func mustCreateContacts(t *testing.T, db *gorm.DB, expectedContacts []contact.Contact) {
	for _, c := range expectedContacts {
		err := db.Create(&postgres.ContactModel{
			Name:  c.Name,
			Phone: c.Phone,
		}).Error
		require.NoError(t, err)
	}
}

func assertNoContacts(t *testing.T, contacts []contact.Contact) {
	t.Helper()
	assert.Empty(t, contacts)
}

func assertContacts(t *testing.T, contacts []contact.Contact, db *gorm.DB) {
	for _, expected := range contacts {
		assertContactExists(t, db, expected)
	}
}

// assertContactExists verifies that a contact exists in the database with correct values
func assertContactExists(t testing.TB, db *gorm.DB, expected contact.Contact) {
	t.Helper()
	var model postgres.ContactModel
	result := db.Where("name = ? AND phone = ?", expected.Name, expected.Phone).First(&model)
	require.NoError(t, result.Error, "contact should exist in database")
	assert.Equal(t, expected.Name, model.Name)
	assert.Equal(t, expected.Phone, model.Phone)
	assert.NotZero(t, model.ID)
}

// cleanupContactDatabase truncates all tables to ensure test isolation
func cleanupContactDatabase(t testing.TB, db *gorm.DB) {
	t.Helper()
	err := db.Exec("TRUNCATE TABLE contacts RESTART IDENTITY CASCADE").Error
	require.NoError(t, err)
}
