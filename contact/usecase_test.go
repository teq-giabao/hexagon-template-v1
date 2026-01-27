package contact_test

import (
	"context"
	"hexagon/contact"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockContactRepository struct {
	mock.Mock
}

func (m *MockContactRepository) CreateContact(ctx context.Context, c contact.Contact) error {
	args := m.Called(ctx, c)
	return args.Error(0)
}

func (m *MockContactRepository) AllContacts(ctx context.Context) ([]contact.Contact, error) {
	args := m.Called(ctx)
	return args.Get(0).([]contact.Contact), args.Error(1)
}

func TestAddContact(t *testing.T) {
	r := new(MockContactRepository)
	uc := contact.NewUsecase(r)

	t.Run("should add new contact to track", func(t *testing.T) {
		c := contact.Contact{Name: "John Doe", Phone: "1234567890"}
		r.On("CreateContact", mock.Anything, c).Return(nil).Times(1)

		err := uc.AddContact(context.Background(), c)

		assert.NoError(t, err, "expected no error when adding contact")
		r.AssertExpectations(t)
	})

	t.Run("should fail on empty name", func(t *testing.T) {
		c := contact.Contact{Name: "", Phone: "1234567890"}

		err := uc.AddContact(context.Background(), c)

		assert.Equal(t, contact.ErrInvalidName, err, "expected error for empty name")
		r.AssertExpectations(t)
	})

	t.Run("should fail on empty phone", func(t *testing.T) {
		c := contact.Contact{Name: "John Doe", Phone: ""}

		err := uc.AddContact(context.Background(), c)

		assert.Equal(t, contact.ErrInvalidPhone, err, "expected error for empty phone")
		r.AssertExpectations(t)
	})
}

func TestListContacts(t *testing.T) {
	r := new(MockContactRepository)
	uc := contact.NewUsecase(r)

	t.Run("should return list of contacts", func(t *testing.T) {
		contacts := []contact.Contact{
			{Name: "John Doe", Phone: "1234567890"},
			{Name: "Jane Smith", Phone: "0987654321"},
		}
		r.On("AllContacts", mock.Anything).Return(contacts, nil).Once()

		result, err := uc.ListContacts(context.Background())

		assert.NoError(t, err, "expected no error when listing contacts")
		assert.Equal(t, contacts, result, "expected returned contacts to match")
		r.AssertExpectations(t)
	})
}
