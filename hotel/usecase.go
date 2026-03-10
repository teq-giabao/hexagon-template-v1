package hotel

import "context"

type Service interface {
	ListHotels(ctx context.Context) ([]Hotel, error)
	GetHotelByID(ctx context.Context, id string) (Hotel, error)
	AddHotel(ctx context.Context, h Hotel) (Hotel, error)
}

type Repository interface {
	List(ctx context.Context) ([]Hotel, error)
	GetByID(ctx context.Context, id string) (Hotel, error)
	Create(ctx context.Context, h Hotel) (Hotel, error)
}

type Usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) *Usecase {
	return &Usecase{repo: repo}
}

func (uc *Usecase) ListHotels(ctx context.Context) ([]Hotel, error) {
	return uc.repo.List(ctx)
}

func (uc *Usecase) GetHotelByID(ctx context.Context, id string) (Hotel, error) {
	if err := ValidateID(id); err != nil {
		return Hotel{}, err
	}

	return uc.repo.GetByID(ctx, id)
}

func (uc *Usecase) AddHotel(ctx context.Context, h Hotel) (Hotel, error) {
	h.Rating = 0
	if h.DefaultChildMaxAge == 0 {
		h.DefaultChildMaxAge = 11
	}

	if err := h.ValidateForCreate(); err != nil {
		return Hotel{}, err
	}

	return uc.repo.Create(ctx, h)
}
