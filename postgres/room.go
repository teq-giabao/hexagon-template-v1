package postgres

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"hexagon/room"

	"gorm.io/gorm"
)

type RoomModel struct {
	ID           string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	HotelID      string `gorm:"type:uuid;not null;index"`
	Name         string `gorm:"not null"`
	Description  string
	BasePrice    float64   `gorm:"type:numeric(12,2);not null"`
	MaxAdult     int       `gorm:"not null"`
	MaxChild     int       `gorm:"not null"`
	MaxOccupancy int       `gorm:"not null"`
	BedOptions   []byte    `gorm:"type:jsonb;not null;default:'[]'"`
	SizeSqm      int       `gorm:"not null;default:0"`
	Status       string    `gorm:"type:varchar(32);not null"`
	CreatedAt    time.Time `gorm:"not null;autoCreateTime"`
	UpdatedAt    time.Time `gorm:"not null;autoUpdateTime"`

	Images []RoomImageModel `gorm:"foreignKey:RoomID"`
}

func (RoomModel) TableName() string { return "rooms" }

type RoomImageModel struct {
	ID      string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	RoomID  string `gorm:"type:uuid;not null;index"`
	URL     string `gorm:"not null"`
	IsCover bool   `gorm:"not null;default:false"`
}

func (RoomImageModel) TableName() string { return "room_images" }

type RoomAmenityModel struct {
	ID          string `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Code        string `gorm:"not null;uniqueIndex"`
	Name        string `gorm:"not null"`
	Description string
	Icon        string
	CreatedAt   time.Time `gorm:"not null;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"not null;autoUpdateTime"`
}

func (RoomAmenityModel) TableName() string { return "room_amenities" }

type RoomAmenityMapModel struct {
	ID        string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	RoomID    string    `gorm:"type:uuid;not null;index"`
	AmenityID string    `gorm:"type:uuid;not null;index"`
	CreatedAt time.Time `gorm:"not null;autoCreateTime"`
}

func (RoomAmenityMapModel) TableName() string { return "room_amenity_maps" }

type RoomInventoryModel struct {
	ID              string    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	RoomID          string    `gorm:"type:uuid;not null;index"`
	Date            time.Time `gorm:"type:date;not null"`
	TotalInventory  int       `gorm:"not null"`
	HeldInventory   int       `gorm:"not null;default:0"`
	BookedInventory int       `gorm:"not null;default:0"`
}

func (RoomInventoryModel) TableName() string { return "room_inventories" }

type RoomRepository struct {
	db *gorm.DB
}

func NewRoomRepository(db *gorm.DB) *RoomRepository { return &RoomRepository{db: db} }

func (r *RoomRepository) CreateRoom(ctx context.Context, rm room.Room) (room.Room, error) {
	model := RoomModel{
		HotelID:      rm.HotelID,
		Name:         rm.Name,
		Description:  rm.Description,
		BasePrice:    rm.BasePrice,
		MaxAdult:     rm.MaxAdult,
		MaxChild:     rm.MaxChild,
		MaxOccupancy: rm.MaxOccupancy,
		BedOptions:   emptyJSONIfNil(rm.BedOptions),
		SizeSqm:      rm.SizeSqm,
		Status:       string(rm.Status),
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&model).Error; err != nil {
			return err
		}

		if len(rm.Images) > 0 {
			images := make([]RoomImageModel, len(rm.Images))
			for i := range rm.Images {
				images[i] = RoomImageModel{RoomID: model.ID, URL: rm.Images[i].URL, IsCover: rm.Images[i].IsCover}
			}

			if err := tx.Create(&images).Error; err != nil {
				return err
			}
		}

		return createRoomAmenityMaps(tx, model.ID, rm.AmenityIDs)
	})
	if err != nil {
		return room.Room{}, err
	}

	var created RoomModel
	if err := r.db.WithContext(ctx).Preload("Images").Where("id = ?", model.ID).First(&created).Error; err != nil {
		return room.Room{}, err
	}

	amenities, err := r.listRoomAmenities(ctx, created.ID)
	if err != nil {
		return room.Room{}, err
	}

	result := toDomainRoom(created)
	result.Amenities = amenities

	return result, nil
}

func (r *RoomRepository) CreateAmenity(ctx context.Context, amenity room.RoomAmenity) (room.RoomAmenity, error) {
	model := RoomAmenityModel{
		Code:        amenity.Code,
		Name:        amenity.Name,
		Description: amenity.Description,
		Icon:        amenity.Icon,
	}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return room.RoomAmenity{}, err
	}

	return room.RoomAmenity{
		ID:          model.ID,
		Code:        model.Code,
		Name:        model.Name,
		Description: model.Description,
		Icon:        model.Icon,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}, nil
}

func (r *RoomRepository) CreateInventory(ctx context.Context, inv room.RoomInventory) (room.RoomInventory, error) {
	model := RoomInventoryModel{
		RoomID:          inv.RoomID,
		Date:            inv.Date,
		TotalInventory:  inv.TotalInventory,
		HeldInventory:   inv.HeldInventory,
		BookedInventory: inv.BookedInventory,
	}
	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return room.RoomInventory{}, err
	}

	return room.RoomInventory{
		ID:              model.ID,
		RoomID:          model.RoomID,
		Date:            model.Date,
		TotalInventory:  model.TotalInventory,
		HeldInventory:   model.HeldInventory,
		BookedInventory: model.BookedInventory,
	}, nil
}

func (r *RoomRepository) listRoomAmenities(ctx context.Context, roomID string) ([]room.RoomAmenity, error) {
	var models []RoomAmenityModel

	err := r.db.WithContext(ctx).
		Table("room_amenities AS ra").
		Select("ra.*").
		Joins("JOIN room_amenity_maps AS ram ON ram.amenity_id = ra.id").
		Where("ram.room_id = ?", roomID).
		Scan(&models).Error
	if err != nil {
		return nil, err
	}

	result := make([]room.RoomAmenity, len(models))
	for i := range models {
		result[i] = room.RoomAmenity{
			ID:          models[i].ID,
			Code:        models[i].Code,
			Name:        models[i].Name,
			Description: models[i].Description,
			Icon:        models[i].Icon,
			CreatedAt:   models[i].CreatedAt,
			UpdatedAt:   models[i].UpdatedAt,
		}
	}

	return result, nil
}

func toDomainRoom(model RoomModel) room.Room {
	images := make([]room.RoomImage, len(model.Images))
	for i := range model.Images {
		images[i] = room.RoomImage{ID: model.Images[i].ID, RoomID: model.Images[i].RoomID, URL: model.Images[i].URL, IsCover: model.Images[i].IsCover}
	}

	bed := json.RawMessage(model.BedOptions)
	if len(bed) == 0 {
		bed = json.RawMessage("[]")
	}

	return room.Room{
		ID:           model.ID,
		HotelID:      model.HotelID,
		Name:         model.Name,
		Description:  model.Description,
		BasePrice:    model.BasePrice,
		MaxAdult:     model.MaxAdult,
		MaxChild:     model.MaxChild,
		MaxOccupancy: model.MaxOccupancy,
		BedOptions:   bed,
		SizeSqm:      model.SizeSqm,
		Status:       room.RoomStatus(model.Status),
		Images:       images,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}
}

func emptyJSONIfNil(value []byte) []byte {
	if len(value) == 0 {
		return []byte("[]")
	}

	return value
}

func createRoomAmenityMaps(tx *gorm.DB, roomID string, amenityIDs []string) error {
	unique := uniqueTrimmedStrings(amenityIDs)
	if len(unique) == 0 {
		return nil
	}

	maps := make([]RoomAmenityMapModel, len(unique))
	for i := range unique {
		maps[i] = RoomAmenityMapModel{
			RoomID:    roomID,
			AmenityID: unique[i],
		}
	}

	return tx.Create(&maps).Error
}

func uniqueTrimmedStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))

	for i := range values {
		v := strings.TrimSpace(values[i])
		if v == "" {
			continue
		}

		if _, ok := seen[v]; ok {
			continue
		}

		seen[v] = struct{}{}

		result = append(result, v)
	}

	return result
}
