package httpserver_test

import (
	"encoding/json"
	"hexagon/contact"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeepingTrackOfContact(t *testing.T) {
	db := MustCreateTestDatabase(t)
	MigrateTestDatabase(t, db, "../migrations")
	c := contact.Contact{Name: "Charlie", Phone: "3456789012"}
	server := MustCreateServer(t, db)

	server.Router.ServeHTTP(httptest.NewRecorder(), newAddContactRequest(contact.Contact{Name: "Alice", Phone: "1234567890"}))
	server.Router.ServeHTTP(httptest.NewRecorder(), newAddContactRequest(contact.Contact{Name: "Bob", Phone: "2345678901"}))

	t.Run("add new contact", func(t *testing.T) {
		rec := httptest.NewRecorder()
		server.Router.ServeHTTP(rec, newAddContactRequest(c))

		assert.Equal(t, http.StatusCreated, rec.Code, "Expected 201 Created")
	})

	t.Run("list all contacts", func(t *testing.T) {
		rec := httptest.NewRecorder()
		server.Router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/contacts", nil))

		assert.Equal(t, http.StatusOK, rec.Code, "Expected 200 OK")
		assertAddedContact(t, rec, c)
	})
}

func assertAddedContact(t *testing.T, rec *httptest.ResponseRecorder, c contact.Contact) {
	result := []contact.Contact{}
	err := json.NewDecoder(rec.Body).Decode(&result)
	assert.NoError(t, err, "Failed to decode response")
	assert.Len(t, result, 3, "Expected 3 contacts in the list")
	assert.Contains(t, result, c, "Expected contact list to contain the newly added contact")
}
