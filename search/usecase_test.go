package search

import (
	"context"
	"testing"

	"hexagon/errs"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type searchRepoStub struct {
	called bool
	searchHotels func(ctx context.Context, criteria Criteria) ([]HotelSearchResult, error)
	searchRooms func(ctx context.Context, hotelID string, criteria Criteria) (HotelRoomSearchResult, error)
	searchCombos func(ctx context.Context, hotelID string, criteria Criteria, maxCombos int) (HotelRoomCombinationsResult, error)
}

func (s *searchRepoStub) SearchHotels(ctx context.Context, criteria Criteria) ([]HotelSearchResult, error) {
	s.called = true
	return s.searchHotels(ctx, criteria)
}

func (s *searchRepoStub) SearchHotelRooms(ctx context.Context, hotelID string, criteria Criteria) (HotelRoomSearchResult, error) {
	s.called = true
	return s.searchRooms(ctx, hotelID, criteria)
}

func (s *searchRepoStub) SearchHotelRoomCombinations(ctx context.Context, hotelID string, criteria Criteria, maxCombos int) (HotelRoomCombinationsResult, error) {
	s.called = true
	return s.searchCombos(ctx, hotelID, criteria, maxCombos)
}

func TestUsecase_SearchHotels_InvalidCriteria(t *testing.T) {
	repo := &searchRepoStub{searchHotels: func(ctx context.Context, criteria Criteria) ([]HotelSearchResult, error) {
		return nil, nil
	}}
	uc := NewUsecase(repo)

	_, err := uc.SearchHotels(context.Background(), Criteria{})
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(err))
	assert.False(t, repo.called)
}

func TestUsecase_SearchHotels_Success(t *testing.T) {
	repo := &searchRepoStub{searchHotels: func(ctx context.Context, criteria Criteria) ([]HotelSearchResult, error) {
		return []HotelSearchResult{{HotelID: "h-1"}}, nil
	}}
	uc := NewUsecase(repo)

	criteria := validCriteria()
	res, err := uc.SearchHotels(context.Background(), criteria)
	require.NoError(t, err)
	assert.Len(t, res, 1)
}
