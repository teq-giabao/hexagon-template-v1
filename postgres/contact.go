package postgres

import (
	"context"
	"hexagon/contact"

	"gorm.io/gorm"
)

// ContactModel represents the database model for contacts
type ContactModel struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `gorm:"not null"`
	Phone string `gorm:"not null"`
}

// TableName specifies the table name for GORM
func (ContactModel) TableName() string {
	return "contacts"
}

// ContactRepository implements contact.Repository interface
type ContactRepository struct {
	db *gorm.DB
}

// NewContactRepository creates a new contact repository
func NewContactRepository(db *gorm.DB) *ContactRepository {
	return &ContactRepository{db: db}
}

// CreateContact creates a new contact in the database
func (r *ContactRepository) CreateContact(ctx context.Context, c contact.Contact) error {
	model := ContactModel{
		Name:  c.Name,
		Phone: c.Phone,
	}
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *ContactRepository) AllContacts(ctx context.Context) ([]contact.Contact, error) {
	var models []ContactModel
	if err := r.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, err
	}

	contacts := make([]contact.Contact, len(models))
	for i, model := range models {
		contacts[i] = contact.Contact{
			Name:  model.Name,
			Phone: model.Phone,
		}
	}
	return contacts, nil
}
