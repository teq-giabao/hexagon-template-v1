package httpserver_test

import (
	"context"
	"fmt"
	"hexagon/contact"
	"hexagon/httpserver"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockContactService struct {
	mock.Mock
}

func (m *MockContactService) AddContact(ctx context.Context, c contact.Contact) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func (m *MockContactService) ListContacts(ctx context.Context) ([]contact.Contact, error) {
	args := m.Called(ctx)
	return args.Get(0).([]contact.Contact), args.Error(1)
}

func TestAddContact(t *testing.T) {
	server := httpserver.Default(testConfig())
	svc := new(MockContactService)
	server.ContactService = svc
	token, err := signTestToken()
	assert.NoError(t, err)

	t.Run("should returns 201 when added new contact", func(t *testing.T) {
		c := contact.Contact{Name: "Jane Doe", Phone: "0987654321"}
		svc.On("AddContact", mock.Anything, c).Return(nil).Once()
		request := newAddContactRequestWithAuth(c, token)
		recorder := httptest.NewRecorder()

		server.Router.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusCreated, recorder.Code, "Expected 201 Created")
		resp := decodeAPIResponse(t, recorder)
		assert.Equal(t, "201", resp.Code)
		assert.Equal(t, "OK", resp.Message)
		svc.AssertExpectations(t)
	})

	t.Run("should returns 400 when request is invalid", func(t *testing.T) {
		c := contact.Contact{Phone: "0987654321"}
		request := newAddContactRequestWithAuth(c, token)
		recorder := httptest.NewRecorder()

		server.Router.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusBadRequest, recorder.Code, "Expected 400 Bad Request")
		resp := decodeAPIResponse(t, recorder)
		assert.Equal(t, "100010", resp.Code)
		svc.AssertNotCalled(t, "AddContact")
	})

	t.Run("should returns 400 when phone format is invalid", func(t *testing.T) {
		c := contact.Contact{Name: "Jane Doe", Phone: "invalid"}
		request := newAddContactRequestWithAuth(c, token)
		recorder := httptest.NewRecorder()

		server.Router.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusBadRequest, recorder.Code, "Expected 400 Bad Request")
		resp := decodeAPIResponse(t, recorder)
		assert.Equal(t, "100010", resp.Code)
		svc.AssertNotCalled(t, "AddContact")
	})

	t.Run("should returns 400 when JSON is malformed", func(t *testing.T) {
		request := malformedAddContactRequest()
		request.Header.Set("Authorization", "Bearer "+token)
		recorder := httptest.NewRecorder()

		server.Router.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusBadRequest, recorder.Code, "Expected 400 Bad Request for malformed JSON")
		resp := decodeAPIResponse(t, recorder)
		assert.Equal(t, "100010", resp.Code)
		// Service should not be called when binding fails
		svc.AssertNotCalled(t, "AddContact")
	})
}

func TestListContacts(t *testing.T) {
	server := httpserver.Default(testConfig())
	svc := new(MockContactService)
	server.ContactService = svc

	t.Run("should returns 200 with list of contacts", func(t *testing.T) {
		contacts := []contact.Contact{
			{Name: "Alice", Phone: "1234567890"},
			{Name: "Bob", Phone: "2345678901"},
		}
		svc.On("ListContacts", mock.Anything).Return(contacts, nil).Once()
		request := httptest.NewRequest("GET", "/api/contacts", nil)
		recorder := httptest.NewRecorder()

		server.Router.ServeHTTP(recorder, request)

		assertListContacts(t, recorder, contacts)
		svc.AssertExpectations(t)
	})
}

func assertListContacts(t *testing.T, recorder *httptest.ResponseRecorder, contacts []contact.Contact) {
	assert.Equal(t, http.StatusOK, recorder.Code, "Expected 200 OK")
	resp := decodeAPIResponse(t, recorder)
	assert.Equal(t, "200", resp.Code)
	assert.Equal(t, "OK", resp.Message)
	var result struct {
		Data []contact.Contact `json:"data"`
	}
	decodeAPIResult(t, resp.Result, &result)
	assert.Equal(t, contacts, result.Data, "Expected returned contacts to match")
}

func malformedAddContactRequest() *http.Request {
	request := httptest.NewRequest("POST", "/api/contacts", strings.NewReader(`{"name": "John", invalid json`))
	request.Header.Set("Content-Type", "application/json")
	return request
}

func newAddContactRequest(c contact.Contact) *http.Request {
	body := strings.NewReader(fmt.Sprintf(`{"name":"%s","phone":"%s"}`, c.Name, c.Phone))
	request := httptest.NewRequest("POST", "/api/contacts", body)
	request.Header.Set("Content-Type", "application/json")
	return request
}

func newAddContactRequestWithAuth(c contact.Contact, token string) *http.Request {
	request := newAddContactRequest(c)
	request.Header.Set("Authorization", "Bearer "+token)
	return request
}
