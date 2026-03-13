package main

import (
	"log/slog"
	"os"
	"strconv"
	"time"

	"hexagon/pkg/config"
	"hexagon/postgres"

	"gorm.io/gorm"
)

const (
	demoHotelID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	demoRoomAID = "11111111-1111-1111-1111-111111111111"
	demoRoomBID = "22222222-2222-2222-2222-222222222222"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("cannot load config", "error", err)
		os.Exit(1)
	}

	db, err := postgres.NewConnection(postgres.Options{
		DBName:   cfg.DB.Name,
		DBUser:   cfg.DB.User,
		Password: cfg.DB.Pass,
		Host:     cfg.DB.Host,
		Port:     strconv.Itoa(cfg.DB.Port),
		SSLMode:  cfg.DB.EnableSSL,
	})
	if err != nil {
		logger.Error("cannot connect db", "error", err)
		os.Exit(1)
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := cleanupDemoData(tx); err != nil {
			return err
		}

		if err := insertDemoData(tx); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		logger.Error("cannot seed demo data", "error", err)
		os.Exit(1)
	}

	logger.Info("seed completed", "hotel_id", demoHotelID, "room_ids", []string{demoRoomAID, demoRoomBID})
}

func cleanupDemoData(tx *gorm.DB) error {
	if err := tx.Where("room_id IN ?", []string{demoRoomAID, demoRoomBID}).Delete(&postgres.RoomInventoryModel{}).Error; err != nil {
		return err
	}

	if err := tx.Where("room_id IN ?", []string{demoRoomAID, demoRoomBID}).Delete(&postgres.RoomImageModel{}).Error; err != nil {
		return err
	}

	if err := tx.Where("id IN ?", []string{demoRoomAID, demoRoomBID}).Delete(&postgres.RoomModel{}).Error; err != nil {
		return err
	}

	if err := tx.Where("hotel_id = ?", demoHotelID).Delete(&postgres.HotelImageModel{}).Error; err != nil {
		return err
	}

	if err := tx.Where("hotel_id = ?", demoHotelID).Delete(&postgres.HotelPaymentOptionModel{}).Error; err != nil {
		return err
	}

	if err := tx.Where("id = ?", demoHotelID).Delete(&postgres.HotelModel{}).Error; err != nil {
		return err
	}

	return nil
}

func insertDemoData(tx *gorm.DB) error {
	now := time.Now().UTC()

	if err := insertDemoHotel(tx, now); err != nil {
		return err
	}

	if err := insertDemoRooms(tx, now); err != nil {
		return err
	}

	return insertDemoInventories(tx)
}

func insertDemoHotel(tx *gorm.DB, now time.Time) error {
	hotel := postgres.HotelModel{
		ID:                 demoHotelID,
		Name:               "Hexagon Grand Hanoi",
		Description:        "Demo hotel for search manual QA",
		Address:            "12 Tran Hung Dao",
		City:               "ha noi",
		Rating:             4.6,
		CheckInTime:        "14:00:00",
		CheckOutTime:       "12:00:00",
		DefaultChildMaxAge: 11,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := tx.Create(&hotel).Error; err != nil {
		return err
	}

	paymentOptions := []postgres.HotelPaymentOptionModel{
		{ID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaa01", HotelID: demoHotelID, PaymentOption: "immediate", Enabled: true},
		{ID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaa02", HotelID: demoHotelID, PaymentOption: "pay_at_hotel", Enabled: true},
	}
	if err := tx.Create(&paymentOptions).Error; err != nil {
		return err
	}

	hotelImages := []postgres.HotelImageModel{{
		ID:      "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaa10",
		HotelID: demoHotelID,
		URL:     "https://picsum.photos/seed/hotel-cover/1200/800",
		IsCover: true,
	}}

	return tx.Create(&hotelImages).Error
}

func insertDemoRooms(tx *gorm.DB, now time.Time) error {
	rooms := []postgres.RoomModel{
		{
			ID:           demoRoomAID,
			HotelID:      demoHotelID,
			Name:         "Deluxe Twin",
			Description:  "Demo room type A",
			BasePrice:    1200000,
			MaxAdult:     2,
			MaxChild:     1,
			MaxOccupancy: 3,
			BedOptions:   []byte("[]"),
			SizeSqm:      32,
			Status:       "active",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:           demoRoomBID,
			HotelID:      demoHotelID,
			Name:         "Family Suite",
			Description:  "Demo room type B",
			BasePrice:    1800000,
			MaxAdult:     3,
			MaxChild:     2,
			MaxOccupancy: 5,
			BedOptions:   []byte("[]"),
			SizeSqm:      48,
			Status:       "active",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}
	if err := tx.Create(&rooms).Error; err != nil {
		return err
	}

	roomImages := []postgres.RoomImageModel{
		{ID: "11111111-1111-1111-1111-111111111112", RoomID: demoRoomAID, URL: "https://picsum.photos/seed/room-a/1200/800", IsCover: true},
		{ID: "22222222-2222-2222-2222-222222222223", RoomID: demoRoomBID, URL: "https://picsum.photos/seed/room-b/1200/800", IsCover: true},
	}

	return tx.Create(&roomImages).Error
}

func insertDemoInventories(tx *gorm.DB) error {
	inventories := []postgres.RoomInventoryModel{
		{ID: "31111111-1111-1111-1111-111111111111", RoomID: demoRoomAID, Date: dateOnly(2026, 4, 1), TotalInventory: 5, HeldInventory: 0, BookedInventory: 1},
		{ID: "31111111-1111-1111-1111-111111111112", RoomID: demoRoomAID, Date: dateOnly(2026, 4, 2), TotalInventory: 5, HeldInventory: 0, BookedInventory: 1},
		{ID: "31111111-1111-1111-1111-111111111113", RoomID: demoRoomAID, Date: dateOnly(2026, 4, 3), TotalInventory: 5, HeldInventory: 0, BookedInventory: 1},
		{ID: "32222222-2222-2222-2222-222222222221", RoomID: demoRoomBID, Date: dateOnly(2026, 4, 1), TotalInventory: 3, HeldInventory: 0, BookedInventory: 0},
		{ID: "32222222-2222-2222-2222-222222222222", RoomID: demoRoomBID, Date: dateOnly(2026, 4, 2), TotalInventory: 3, HeldInventory: 0, BookedInventory: 0},
		{ID: "32222222-2222-2222-2222-222222222223", RoomID: demoRoomBID, Date: dateOnly(2026, 4, 3), TotalInventory: 3, HeldInventory: 0, BookedInventory: 0},
	}

	return tx.Create(&inventories).Error
}

func dateOnly(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}
