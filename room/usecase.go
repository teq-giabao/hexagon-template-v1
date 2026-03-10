package room

import "context"

type Service interface {
	AddRoom(ctx context.Context, r Room) (Room, error)
	AddAmenity(ctx context.Context, amenity RoomAmenity) (RoomAmenity, error)
	AddInventory(ctx context.Context, inv RoomInventory) (RoomInventory, error)
}

type Repository interface {
	CreateRoom(ctx context.Context, r Room) (Room, error)
	CreateAmenity(ctx context.Context, amenity RoomAmenity) (RoomAmenity, error)
	CreateInventory(ctx context.Context, inv RoomInventory) (RoomInventory, error)
}

type Usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) *Usecase {
	return &Usecase{repo: repo}
}

func (uc *Usecase) AddRoom(ctx context.Context, r Room) (Room, error) {
	if r.Status == "" {
		r.Status = RoomStatusActive
	}

	if err := r.ValidateForCreate(); err != nil {
		return Room{}, err
	}

	return uc.repo.CreateRoom(ctx, r)
}

func (uc *Usecase) AddAmenity(ctx context.Context, amenity RoomAmenity) (RoomAmenity, error) {
	if err := amenity.ValidateForCreate(); err != nil {
		return RoomAmenity{}, err
	}

	return uc.repo.CreateAmenity(ctx, amenity)
}

func (uc *Usecase) AddInventory(ctx context.Context, inv RoomInventory) (RoomInventory, error) {
	if err := inv.ValidateForCreate(); err != nil {
		return RoomInventory{}, err
	}

	return uc.repo.CreateInventory(ctx, inv)
}
