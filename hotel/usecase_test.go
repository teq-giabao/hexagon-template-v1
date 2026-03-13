package hotel

import (
	"context"
	"testing"

	"hexagon/errs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type hotelRepoStub struct {
	list  func(ctx context.Context) ([]Hotel, error)
	get   func(ctx context.Context, id string) (Hotel, error)
	create func(ctx context.Context, h Hotel) (Hotel, error)
	getCalled   bool
	createCalled bool
}

func (r *hotelRepoStub) List(ctx context.Context) ([]Hotel, error) {
	return r.list(ctx)
}

func (r *hotelRepoStub) GetByID(ctx context.Context, id string) (Hotel, error) {
	r.getCalled = true
	return r.get(ctx, id)
}

func (r *hotelRepoStub) Create(ctx context.Context, h Hotel) (Hotel, error) {
	r.createCalled = true
	return r.create(ctx, h)
}

func TestUsecase_GetHotelByID_Validation(t *testing.T) {
	repo := &hotelRepoStub{
		get: func(ctx context.Context, id string) (Hotel, error) {
			return Hotel{}, nil
		},
	}
	uc := NewUsecase(repo)

	_, err := uc.GetHotelByID(context.Background(), "")
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(err))
	assert.False(t, repo.getCalled)
}

func TestUsecase_AddHotel_DefaultsAndCreate(t *testing.T) {
	captured := Hotel{}
	repo := &hotelRepoStub{
		create: func(ctx context.Context, h Hotel) (Hotel, error) {
			captured = h
			return h, nil
		},
	}
	uc := NewUsecase(repo)

	input := Hotel{
		Name:    "Hotel",
		Address: "Address",
		City:    "City",
		PaymentOptions: []HotelPaymentOption{{PaymentOption: PaymentOptionImmediate}},
	}

	created, err := uc.AddHotel(context.Background(), input)
	require.NoError(t, err)
	assert.True(t, repo.createCalled)
	assert.Equal(t, 0.0, captured.Rating)
	assert.Equal(t, 11, captured.DefaultChildMaxAge)
	assert.Equal(t, input.Name, created.Name)
}
