package room

import (
	"context"
	"testing"

	"hexagon/errs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type roomRepoStub struct {
	createRoomCalled     bool
	createAmenityCalled  bool
	createInventoryCalled bool
	createRoom           func(ctx context.Context, r Room) (Room, error)
	createAmenity        func(ctx context.Context, a RoomAmenity) (RoomAmenity, error)
	createInventory      func(ctx context.Context, inv RoomInventory) (RoomInventory, error)
}

func (r *roomRepoStub) CreateRoom(ctx context.Context, room Room) (Room, error) {
	r.createRoomCalled = true
	return r.createRoom(ctx, room)
}

func (r *roomRepoStub) CreateAmenity(ctx context.Context, amenity RoomAmenity) (RoomAmenity, error) {
	r.createAmenityCalled = true
	return r.createAmenity(ctx, amenity)
}

func (r *roomRepoStub) CreateInventory(ctx context.Context, inv RoomInventory) (RoomInventory, error) {
	r.createInventoryCalled = true
	return r.createInventory(ctx, inv)
}

func TestUsecase_AddRoom_DefaultStatus(t *testing.T) {
	captured := Room{}
	repo := &roomRepoStub{
		createRoom: func(ctx context.Context, r Room) (Room, error) {
			captured = r
			return r, nil
		},
	}
	uc := NewUsecase(repo)

	room := Room{
		HotelID:      "h-1",
		Name:         "Room",
		BasePrice:    100,
		MaxAdult:     2,
		MaxChild:     1,
		MaxOccupancy: 3,
		Images:       []RoomImage{{URL: "https://image"}},
	}

	created, err := uc.AddRoom(context.Background(), room)
	require.NoError(t, err)
	assert.True(t, repo.createRoomCalled)
	assert.Equal(t, RoomStatusActive, captured.Status)
	assert.Equal(t, RoomStatusActive, created.Status)
}

func TestUsecase_AddRoom_Invalid(t *testing.T) {
	repo := &roomRepoStub{createRoom: func(ctx context.Context, r Room) (Room, error) { return r, nil }}
	uc := NewUsecase(repo)

	_, err := uc.AddRoom(context.Background(), Room{})
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(err))
	assert.False(t, repo.createRoomCalled)
}

func TestUsecase_AddAmenity_Validation(t *testing.T) {
	repo := &roomRepoStub{createAmenity: func(ctx context.Context, a RoomAmenity) (RoomAmenity, error) { return a, nil }}
	uc := NewUsecase(repo)

	_, err := uc.AddAmenity(context.Background(), RoomAmenity{})
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(err))
	assert.False(t, repo.createAmenityCalled)
}

func TestUsecase_AddInventory_Validation(t *testing.T) {
	repo := &roomRepoStub{createInventory: func(ctx context.Context, inv RoomInventory) (RoomInventory, error) { return inv, nil }}
	uc := NewUsecase(repo)

	_, err := uc.AddInventory(context.Background(), RoomInventory{})
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(err))
	assert.False(t, repo.createInventoryCalled)
}
