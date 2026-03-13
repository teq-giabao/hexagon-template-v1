package postgres

import (
	"context"
	"errors"
	"time"

	"hexagon/booking"
	"hexagon/hotel"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type BookingModel struct {
	ID              string     `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	HotelID         string     `gorm:"type:uuid;not null;index"`
	RoomID          string     `gorm:"type:uuid;not null;index"`
	CheckInDate     time.Time  `gorm:"type:date;not null"`
	CheckOutDate    time.Time  `gorm:"type:date;not null"`
	Nights          int        `gorm:"not null"`
	RoomCount       int        `gorm:"not null"`
	GuestCount      int        `gorm:"not null"`
	NightlyPrice    float64    `gorm:"type:numeric(12,2);not null"`
	TotalPrice      float64    `gorm:"type:numeric(12,2);not null"`
	Status          string     `gorm:"type:varchar(32);not null"`
	PaymentOption   string     `gorm:"type:varchar(64)"`
	PaymentStatus   string     `gorm:"type:varchar(32);not null;default:'unpaid'"`
	HoldExpiresAt   *time.Time `gorm:"type:timestamptz"`
	PaymentDeadline *time.Time `gorm:"type:timestamptz"`
	CancelledAt     *time.Time `gorm:"type:timestamptz"`
	CancellationFee float64    `gorm:"type:numeric(12,2);not null;default:0"`
	RefundAmount    float64    `gorm:"type:numeric(12,2);not null;default:0"`
	CreatedAt       time.Time  `gorm:"not null;autoCreateTime"`
	UpdatedAt       time.Time  `gorm:"not null;autoUpdateTime"`
}

func (BookingModel) TableName() string { return "bookings" }

type BookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

func (r *BookingRepository) CreatePending(ctx context.Context, req booking.CreateRequest, holdDuration time.Duration) (booking.Checkout, error) {
	var checkout booking.Checkout
	var room RoomModel

	nights := nightsBetween(req.CheckInDate, req.CheckOutDate)
	if nights <= 0 {
		return checkout, booking.ErrCheckOutBeforeCheckIn
	}

	holdUntil := time.Now().Add(holdDuration)

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ?", req.RoomID).First(&room).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return booking.ErrRoomNotFound
			}
			return err
		}

		if room.Status != "active" {
			return booking.ErrRoomUnavailable
		}

		inventories, err := lockRoomInventories(tx, req.RoomID, req.CheckInDate, req.CheckOutDate)
		if err != nil {
			return err
		}

		if len(inventories) != nights {
			return booking.ErrRoomUnavailable
		}

		for i := range inventories {
			available := inventories[i].TotalInventory - inventories[i].HeldInventory - inventories[i].BookedInventory
			if available < req.RoomCount {
				return booking.ErrRoomUnavailable
			}
		}

		if err := updateHeldInventory(tx, req.RoomID, req.CheckInDate, req.CheckOutDate, req.RoomCount); err != nil {
			return err
		}

		nightly := room.BasePrice
		total := nightly * float64(nights) * float64(req.RoomCount)

		model := BookingModel{
			HotelID:       room.HotelID,
			RoomID:        room.ID,
			CheckInDate:   req.CheckInDate,
			CheckOutDate:  req.CheckOutDate,
			Nights:        nights,
			RoomCount:     req.RoomCount,
			GuestCount:    req.GuestCount,
			NightlyPrice:  nightly,
			TotalPrice:    total,
			Status:        string(booking.BookingStatusPending),
			PaymentStatus: string(booking.PaymentStatusUnpaid),
			HoldExpiresAt: &holdUntil,
		}

		if err := tx.Create(&model).Error; err != nil {
			return err
		}

		checkout.Booking = toDomainBooking(model)
		options, err := enabledPaymentOptionsWithDB(tx, model.HotelID)
		if err != nil {
			return err
		}
		checkout.PaymentOptions = options

		return nil
	})
	if err != nil {
		return booking.Checkout{}, err
	}

	return checkout, nil
}

func (r *BookingRepository) GetByID(ctx context.Context, id string) (booking.Booking, error) {
	var model BookingModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return booking.Booking{}, booking.ErrBookingNotFound
		}
		return booking.Booking{}, err
	}

	return toDomainBooking(model), nil
}

func (r *BookingRepository) SetPaymentOption(ctx context.Context, id string, option hotel.PaymentOption, paymentDeadline *time.Time) (booking.Booking, error) {
	moment := time.Now()

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var model BookingModel
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&model).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return booking.ErrBookingNotFound
			}
			return err
		}

		switch booking.BookingStatus(model.Status) {
		case booking.BookingStatusExpired:
			return booking.ErrBookingExpired
		case booking.BookingStatusCancelled:
			return booking.ErrBookingCancelled
		case booking.BookingStatusConfirmed:
			return booking.ErrBookingNotPending
		case booking.BookingStatusPending:
			// ok
		default:
			return booking.ErrBookingNotPending
		}

		if model.HoldExpiresAt != nil && !model.HoldExpiresAt.After(moment) {
			if err := expirePendingBooking(tx, model, moment); err != nil {
				return err
			}
			return booking.ErrBookingExpired
		}

		updates := map[string]interface{}{
			"payment_option": string(option),
			"updated_at":     moment,
		}

		switch option {
		case hotel.PaymentOptionImmediate:
			updates["payment_deadline"] = model.HoldExpiresAt
			if err := tx.Model(&BookingModel{}).Where("id = ?", model.ID).Updates(updates).Error; err != nil {
				return err
			}
		case hotel.PaymentOptionPayAtHotel, hotel.PaymentOptionDeferred:
			if err := moveHeldToBooked(tx, model.RoomID, model.CheckInDate, model.CheckOutDate, model.RoomCount); err != nil {
				return err
			}

			updates["status"] = string(booking.BookingStatusConfirmed)
			updates["hold_expires_at"] = nil
			if option == hotel.PaymentOptionDeferred {
				updates["payment_deadline"] = paymentDeadline
			} else {
				updates["payment_deadline"] = nil
			}

			if err := tx.Model(&BookingModel{}).Where("id = ?", model.ID).Updates(updates).Error; err != nil {
				return err
			}
		default:
			return booking.ErrPaymentOptionInvalid
		}

		return nil
	})
	if err != nil {
		return booking.Booking{}, err
	}

	return r.GetByID(ctx, id)
}

func (r *BookingRepository) MarkPaid(ctx context.Context, id string) (booking.Booking, error) {
	moment := time.Now()

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var model BookingModel
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&model).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return booking.ErrBookingNotFound
			}
			return err
		}

		switch booking.BookingStatus(model.Status) {
		case booking.BookingStatusExpired:
			return booking.ErrBookingExpired
		case booking.BookingStatusCancelled:
			return booking.ErrBookingCancelled
		}

		if booking.PaymentStatus(model.PaymentStatus) == booking.PaymentStatusPaid {
			return nil
		}

		if booking.BookingStatus(model.Status) == booking.BookingStatusPending {
			if model.HoldExpiresAt != nil && !model.HoldExpiresAt.After(moment) {
				if err := expirePendingBooking(tx, model, moment); err != nil {
					return err
				}
				return booking.ErrBookingExpired
			}

			if err := moveHeldToBooked(tx, model.RoomID, model.CheckInDate, model.CheckOutDate, model.RoomCount); err != nil {
				return err
			}

			model.Status = string(booking.BookingStatusConfirmed)
			model.HoldExpiresAt = nil
		}

		model.PaymentStatus = string(booking.PaymentStatusPaid)
		model.PaymentDeadline = nil
		model.UpdatedAt = moment

		return tx.Save(&model).Error
	})
	if err != nil {
		return booking.Booking{}, err
	}

	return r.GetByID(ctx, id)
}

func (r *BookingRepository) Cancel(ctx context.Context, id string, cancellationFee float64) (booking.Booking, error) {
	moment := time.Now()

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var model BookingModel
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&model).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return booking.ErrBookingNotFound
			}
			return err
		}

		switch booking.BookingStatus(model.Status) {
		case booking.BookingStatusExpired:
			return booking.ErrBookingExpired
		case booking.BookingStatusCancelled:
			return booking.ErrBookingCancelled
		}

		if booking.BookingStatus(model.Status) == booking.BookingStatusPending {
			if model.HoldExpiresAt != nil && !model.HoldExpiresAt.After(moment) {
				if err := expirePendingBooking(tx, model, moment); err != nil {
					return err
				}
				return booking.ErrBookingExpired
			}

			if err := updateHeldInventory(tx, model.RoomID, model.CheckInDate, model.CheckOutDate, -model.RoomCount); err != nil {
				return err
			}
		} else if booking.BookingStatus(model.Status) == booking.BookingStatusConfirmed {
			if err := updateBookedInventory(tx, model.RoomID, model.CheckInDate, model.CheckOutDate, -model.RoomCount); err != nil {
				return err
			}
		}

		refund := 0.0
		if booking.PaymentStatus(model.PaymentStatus) == booking.PaymentStatusPaid {
			refund = model.TotalPrice - cancellationFee
			if refund < 0 {
				refund = 0
			}
			model.PaymentStatus = string(booking.PaymentStatusRefunded)
		}

		model.Status = string(booking.BookingStatusCancelled)
		model.CancellationFee = cancellationFee
		model.RefundAmount = refund
		model.CancelledAt = &moment
		model.UpdatedAt = moment

		return tx.Save(&model).Error
	})
	if err != nil {
		return booking.Booking{}, err
	}

	return r.GetByID(ctx, id)
}

func (r *BookingRepository) ExpireStale(ctx context.Context, now time.Time) (int, error) {
	expiredCount := 0

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var pending []BookingModel
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("status = ? AND hold_expires_at IS NOT NULL AND hold_expires_at <= ?", string(booking.BookingStatusPending), now).
			Find(&pending).Error; err != nil {
			return err
		}

		for i := range pending {
			if err := updateHeldInventory(tx, pending[i].RoomID, pending[i].CheckInDate, pending[i].CheckOutDate, -pending[i].RoomCount); err != nil {
				return err
			}
			pending[i].Status = string(booking.BookingStatusExpired)
			pending[i].UpdatedAt = now
			if err := tx.Save(&pending[i]).Error; err != nil {
				return err
			}
			expiredCount++
		}

		var overdue []BookingModel
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("status = ? AND payment_deadline IS NOT NULL AND payment_deadline <= ? AND payment_status != ?", string(booking.BookingStatusConfirmed), now, string(booking.PaymentStatusPaid)).
			Find(&overdue).Error; err != nil {
			return err
		}

		for i := range overdue {
			if err := updateBookedInventory(tx, overdue[i].RoomID, overdue[i].CheckInDate, overdue[i].CheckOutDate, -overdue[i].RoomCount); err != nil {
				return err
			}
			overdue[i].Status = string(booking.BookingStatusExpired)
			overdue[i].UpdatedAt = now
			if err := tx.Save(&overdue[i]).Error; err != nil {
				return err
			}
			expiredCount++
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return expiredCount, nil
}

func enabledPaymentOptionsWithDB(db *gorm.DB, hotelID string) ([]hotel.PaymentOption, error) {
	if hotelID == "" {
		return nil, nil
	}

	type row struct {
		PaymentOption string
	}

	var rows []row
	if err := db.Table("hotel_payment_options").
		Select("payment_option").
		Where("hotel_id = ? AND enabled = TRUE", hotelID).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	options := make([]hotel.PaymentOption, 0, len(rows))
	for i := range rows {
		options = append(options, hotel.PaymentOption(rows[i].PaymentOption))
	}

	return options, nil
}

func toDomainBooking(model BookingModel) booking.Booking {
	return booking.Booking{
		ID:              model.ID,
		HotelID:         model.HotelID,
		RoomID:          model.RoomID,
		CheckInDate:     model.CheckInDate,
		CheckOutDate:    model.CheckOutDate,
		Nights:          model.Nights,
		RoomCount:       model.RoomCount,
		GuestCount:      model.GuestCount,
		NightlyPrice:    model.NightlyPrice,
		TotalPrice:      model.TotalPrice,
		Status:          booking.BookingStatus(model.Status),
		PaymentOption:   hotel.PaymentOption(model.PaymentOption),
		PaymentStatus:   booking.PaymentStatus(model.PaymentStatus),
		HoldExpiresAt:   model.HoldExpiresAt,
		PaymentDeadline: model.PaymentDeadline,
		CancelledAt:     model.CancelledAt,
		CancellationFee: model.CancellationFee,
		RefundAmount:    model.RefundAmount,
		CreatedAt:       model.CreatedAt,
		UpdatedAt:       model.UpdatedAt,
	}
}

func nightsBetween(checkIn, checkOut time.Time) int {
	nights := int(checkOut.Sub(checkIn).Hours() / 24)
	if nights <= 0 {
		return 0
	}
	return nights
}

func lockRoomInventories(tx *gorm.DB, roomID string, checkIn, checkOut time.Time) ([]RoomInventoryModel, error) {
	var inventories []RoomInventoryModel

	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("room_id = ? AND date >= ? AND date < ?", roomID, checkIn, checkOut).
		Order("date asc").
		Find(&inventories).Error
	if err != nil {
		return nil, err
	}

	return inventories, nil
}

func updateHeldInventory(tx *gorm.DB, roomID string, checkIn, checkOut time.Time, delta int) error {
	return updateInventoryField(tx, roomID, checkIn, checkOut, "held_inventory", delta)
}

func updateBookedInventory(tx *gorm.DB, roomID string, checkIn, checkOut time.Time, delta int) error {
	return updateInventoryField(tx, roomID, checkIn, checkOut, "booked_inventory", delta)
}

func updateInventoryField(tx *gorm.DB, roomID string, checkIn, checkOut time.Time, field string, delta int) error {
	result := tx.Model(&RoomInventoryModel{}).
		Where("room_id = ? AND date >= ? AND date < ?", roomID, checkIn, checkOut).
		Update(field, gorm.Expr(field+" + ?", delta))
	if result.Error != nil {
		return result.Error
	}

	nights := nightsBetween(checkIn, checkOut)
	if result.RowsAffected != int64(nights) {
		return booking.ErrRoomUnavailable
	}

	return nil
}

func moveHeldToBooked(tx *gorm.DB, roomID string, checkIn, checkOut time.Time, roomCount int) error {
	result := tx.Model(&RoomInventoryModel{}).
		Where("room_id = ? AND date >= ? AND date < ?", roomID, checkIn, checkOut).
		Updates(map[string]interface{}{
			"held_inventory":   gorm.Expr("held_inventory - ?", roomCount),
			"booked_inventory": gorm.Expr("booked_inventory + ?", roomCount),
		})
	if result.Error != nil {
		return result.Error
	}

	nights := nightsBetween(checkIn, checkOut)
	if result.RowsAffected != int64(nights) {
		return booking.ErrRoomUnavailable
	}

	return nil
}

func expirePendingBooking(tx *gorm.DB, model BookingModel, now time.Time) error {
	if err := updateHeldInventory(tx, model.RoomID, model.CheckInDate, model.CheckOutDate, -model.RoomCount); err != nil {
		return err
	}

	model.Status = string(booking.BookingStatusExpired)
	model.UpdatedAt = now

	return tx.Save(&model).Error
}
