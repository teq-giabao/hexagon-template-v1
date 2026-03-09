package postgres

import (
	"context"
	"errors"
	"time"

	"hexagon/hotel"

	"gorm.io/gorm"
)

type HotelModel struct {
	ID                 string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name               string `gorm:"not null"`
	Description        string
	Address            string `gorm:"not null"`
	City               string `gorm:"not null"`
	Rating             float64
	CheckInTime        string    `gorm:"type:time;not null"`
	CheckOutTime       string    `gorm:"type:time;not null"`
	DefaultChildMaxAge int       `gorm:"not null;default:11"`
	CreatedAt          time.Time `gorm:"not null;autoCreateTime"`
	UpdatedAt          time.Time `gorm:"not null;autoUpdateTime"`

	Images         []HotelImageModel         `gorm:"foreignKey:HotelID"`
	PaymentOptions []HotelPaymentOptionModel `gorm:"foreignKey:HotelID"`
}

func (HotelModel) TableName() string {
	return "hotels"
}

type HotelImageModel struct {
	ID      string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	HotelID string `gorm:"type:uuid;not null;index"`
	URL     string `gorm:"not null"`
	IsCover bool   `gorm:"not null;default:false"`
}

func (HotelImageModel) TableName() string {
	return "hotel_images"
}

type HotelPaymentOptionModel struct {
	ID            string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	HotelID       string `gorm:"type:uuid;not null;index"`
	PaymentOption string `gorm:"not null"`
	Enabled       bool   `gorm:"not null;default:true"`
}

func (HotelPaymentOptionModel) TableName() string {
	return "hotel_payment_options"
}

type HotelRepository struct {
	db *gorm.DB
}

func NewHotelRepository(db *gorm.DB) *HotelRepository {
	return &HotelRepository{db: db}
}

func (r *HotelRepository) List(ctx context.Context) ([]hotel.Hotel, error) {
	var models []HotelModel
	if err := r.db.WithContext(ctx).
		Preload("Images").
		Preload("PaymentOptions").
		Order("created_at DESC").
		Find(&models).Error; err != nil {
		return nil, err
	}

	hotels := make([]hotel.Hotel, len(models))
	for i := range models {
		hotels[i] = toDomainHotel(models[i])
	}

	return hotels, nil
}

func (r *HotelRepository) GetByID(ctx context.Context, id string) (hotel.Hotel, error) {
	var model HotelModel
	if err := r.db.WithContext(ctx).
		Preload("Images").
		Preload("PaymentOptions").
		Where("id = ?", id).
		First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return hotel.Hotel{}, hotel.ErrHotelNotFound
		}

		return hotel.Hotel{}, err
	}

	return toDomainHotel(model), nil
}

func (r *HotelRepository) Create(ctx context.Context, h hotel.Hotel) (hotel.Hotel, error) {
	var created HotelModel

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		model := HotelModel{
			Name:               h.Name,
			Description:        h.Description,
			Address:            h.Address,
			City:               h.City,
			Rating:             h.Rating,
			CheckInTime:        formatClock(h.CheckInTime),
			CheckOutTime:       formatClock(h.CheckOutTime),
			DefaultChildMaxAge: h.DefaultChildMaxAge,
		}
		if err := tx.Create(&model).Error; err != nil {
			return err
		}

		if len(h.Images) > 0 {
			images := make([]HotelImageModel, len(h.Images))
			for i := range h.Images {
				images[i] = HotelImageModel{
					ID:      h.Images[i].ID,
					HotelID: model.ID,
					URL:     h.Images[i].URL,
					IsCover: h.Images[i].IsCover,
				}
			}

			if err := tx.Create(&images).Error; err != nil {
				return err
			}
		}

		if len(h.PaymentOptions) > 0 {
			options := make([]HotelPaymentOptionModel, len(h.PaymentOptions))
			for i := range h.PaymentOptions {
				options[i] = HotelPaymentOptionModel{
					ID:            h.PaymentOptions[i].ID,
					HotelID:       model.ID,
					PaymentOption: string(h.PaymentOptions[i].PaymentOption),
					Enabled:       h.PaymentOptions[i].Enabled,
				}
			}

			if err := tx.Create(&options).Error; err != nil {
				return err
			}
		}

		created = model

		return nil
	})
	if err != nil {
		return hotel.Hotel{}, err
	}

	return r.GetByID(ctx, created.ID)
}

func toDomainHotel(model HotelModel) hotel.Hotel {
	checkInTime, _ := parseClock(model.CheckInTime)
	checkOutTime, _ := parseClock(model.CheckOutTime)

	images := make([]hotel.HotelImage, len(model.Images))
	for i := range model.Images {
		images[i] = hotel.HotelImage{
			ID:      model.Images[i].ID,
			HotelID: model.Images[i].HotelID,
			URL:     model.Images[i].URL,
			IsCover: model.Images[i].IsCover,
		}
	}

	paymentOptions := make([]hotel.HotelPaymentOption, len(model.PaymentOptions))
	for i := range model.PaymentOptions {
		paymentOptions[i] = hotel.HotelPaymentOption{
			ID:            model.PaymentOptions[i].ID,
			HotelID:       model.PaymentOptions[i].HotelID,
			PaymentOption: hotel.PaymentOption(model.PaymentOptions[i].PaymentOption),
			Enabled:       model.PaymentOptions[i].Enabled,
		}
	}

	return hotel.Hotel{
		ID:                 model.ID,
		Name:               model.Name,
		Description:        model.Description,
		Address:            model.Address,
		City:               model.City,
		Rating:             model.Rating,
		CheckInTime:        checkInTime,
		CheckOutTime:       checkOutTime,
		DefaultChildMaxAge: model.DefaultChildMaxAge,
		Images:             images,
		PaymentOptions:     paymentOptions,
		CreatedAt:          model.CreatedAt,
		UpdatedAt:          model.UpdatedAt,
	}
}

func formatClock(t time.Time) string {
	return t.Format("15:04:05")
}

func parseClock(value string) (time.Time, error) {
	if t, err := time.Parse("15:04:05", value); err == nil {
		return t, nil
	}

	return time.Parse("15:04", value)
}
