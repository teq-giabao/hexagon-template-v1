package search

import "context"

type Service interface {
	SearchHotels(ctx context.Context, criteria Criteria) ([]HotelSearchResult, error)
	SearchHotelRooms(ctx context.Context, hotelID string, criteria Criteria) (HotelRoomSearchResult, error)
	SearchHotelRoomCombinations(ctx context.Context, hotelID string, criteria Criteria, maxCombinations int) (HotelRoomCombinationsResult, error)
}

type Repository interface {
	SearchHotels(ctx context.Context, criteria Criteria) ([]HotelSearchResult, error)
	SearchHotelRooms(ctx context.Context, hotelID string, criteria Criteria) (HotelRoomSearchResult, error)
	SearchHotelRoomCombinations(ctx context.Context, hotelID string, criteria Criteria, maxCombinations int) (HotelRoomCombinationsResult, error)
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

func (u *Usecase) SearchHotelRooms(ctx context.Context, hotelID string, criteria Criteria) (HotelRoomSearchResult, error) {
	if err := criteria.Validate(); err != nil {
		return HotelRoomSearchResult{}, err
	}

	return u.repo.SearchHotelRooms(ctx, hotelID, criteria)
}

func (u *Usecase) SearchHotelRoomCombinations(ctx context.Context, hotelID string, criteria Criteria, maxCombinations int) (HotelRoomCombinationsResult, error) {
	if err := criteria.Validate(); err != nil {
		return HotelRoomCombinationsResult{}, err
	}

	return u.repo.SearchHotelRoomCombinations(ctx, hotelID, criteria, maxCombinations)
}
