package booking

import (
	"context"
	"time"

	"hexagon/hotel"
)

const (
	DefaultHoldDuration          = 10 * time.Minute
	DefaultDeferredPaymentWindow = 24 * time.Hour
)

type Checkout struct {
	Booking        Booking
	PaymentOptions []hotel.PaymentOption
}

type Service interface {
	CreateBooking(ctx context.Context, req CreateRequest) (Checkout, error)
	GetBookingByID(ctx context.Context, id string) (Booking, error)
	SelectPaymentOption(ctx context.Context, id string, option hotel.PaymentOption) (Booking, error)
	MarkPaymentPaid(ctx context.Context, id string) (Booking, error)
	CancelBooking(ctx context.Context, id string) (Booking, error)
	ExpireStaleBookings(ctx context.Context) (int, error)
}

type Repository interface {
	CreatePending(ctx context.Context, req CreateRequest, holdDuration time.Duration) (Checkout, error)
	GetByID(ctx context.Context, id string) (Booking, error)
	SetPaymentOption(ctx context.Context, id string, option hotel.PaymentOption, paymentDeadline *time.Time) (Booking, error)
	MarkPaid(ctx context.Context, id string) (Booking, error)
	Cancel(ctx context.Context, id string, cancellationFee float64) (Booking, error)
	ExpireStale(ctx context.Context, now time.Time) (int, error)
}

type Usecase struct {
	repo                  Repository
	holdDuration          time.Duration
	deferredPaymentWindow time.Duration
}

func NewUsecase(repo Repository) *Usecase {
	return &Usecase{
		repo:                  repo,
		holdDuration:          DefaultHoldDuration,
		deferredPaymentWindow: DefaultDeferredPaymentWindow,
	}
}

func NewUsecaseWithConfig(repo Repository, holdDuration, deferredPaymentWindow time.Duration) *Usecase {
	if holdDuration <= 0 {
		holdDuration = DefaultHoldDuration
	}

	if deferredPaymentWindow <= 0 {
		deferredPaymentWindow = DefaultDeferredPaymentWindow
	}

	return &Usecase{
		repo:                  repo,
		holdDuration:          holdDuration,
		deferredPaymentWindow: deferredPaymentWindow,
	}
}

func (uc *Usecase) CreateBooking(ctx context.Context, req CreateRequest) (Checkout, error) {
	if err := req.Validate(); err != nil {
		return Checkout{}, err
	}

	_, _ = uc.repo.ExpireStale(ctx, time.Now())

	return uc.repo.CreatePending(ctx, req, uc.holdDuration)
}

func (uc *Usecase) GetBookingByID(ctx context.Context, id string) (Booking, error) {
	if err := ValidateID(id); err != nil {
		return Booking{}, err
	}

	_, _ = uc.repo.ExpireStale(ctx, time.Now())

	return uc.repo.GetByID(ctx, id)
}

func (uc *Usecase) SelectPaymentOption(ctx context.Context, id string, option hotel.PaymentOption) (Booking, error) {
	if err := ValidateID(id); err != nil {
		return Booking{}, err
	}

	if err := ValidatePaymentOption(option); err != nil {
		return Booking{}, err
	}

	var deadline *time.Time

	if option == hotel.PaymentOptionDeferred {
		value := time.Now().Add(uc.deferredPaymentWindow)
		deadline = &value
	}

	return uc.repo.SetPaymentOption(ctx, id, option, deadline)
}

func (uc *Usecase) MarkPaymentPaid(ctx context.Context, id string) (Booking, error) {
	if err := ValidateID(id); err != nil {
		return Booking{}, err
	}

	return uc.repo.MarkPaid(ctx, id)
}

func (uc *Usecase) CancelBooking(ctx context.Context, id string) (Booking, error) {
	if err := ValidateID(id); err != nil {
		return Booking{}, err
	}

	return uc.repo.Cancel(ctx, id, 0)
}

func (uc *Usecase) ExpireStaleBookings(ctx context.Context) (int, error) {
	return uc.repo.ExpireStale(ctx, time.Now())
}
