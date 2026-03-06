package search

import "context"

type Service interface {
	SearchHotels(ctx context.Context, criteria Criteria) ([]HotelSearchResult, error)
}

type Repository interface {
	SearchHotels(ctx context.Context, criteria Criteria) ([]HotelSearchResult, error)
}

type Usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) *Usecase {
	return &Usecase{repo: repo}
}

func (u *Usecase) SearchHotels(ctx context.Context, criteria Criteria) ([]HotelSearchResult, error) {
	if err := criteria.Validate(); err != nil {
		return nil, err
	}
	return u.repo.SearchHotels(ctx, criteria)
}
