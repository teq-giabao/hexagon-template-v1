package httpserver_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"hexagon/httpserver"
	"hexagon/search"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockSearchService struct {
	mock.Mock
}

func (m *MockSearchService) SearchHotels(ctx context.Context, criteria search.Criteria) ([]search.HotelSearchResult, error) {
	args := m.Called(ctx, criteria)
	return args.Get(0).([]search.HotelSearchResult), args.Error(1)
}

func (m *MockSearchService) SearchHotelRooms(ctx context.Context, hotelID string, criteria search.Criteria) (search.HotelRoomSearchResult, error) {
	args := m.Called(ctx, hotelID, criteria)
	return args.Get(0).(search.HotelRoomSearchResult), args.Error(1)
}

func (m *MockSearchService) SearchHotelRoomCombinations(ctx context.Context, hotelID string, criteria search.Criteria, maxCombinations int) (search.HotelRoomCombinationsResult, error) {
	args := m.Called(ctx, hotelID, criteria, maxCombinations)
	return args.Get(0).(search.HotelRoomCombinationsResult), args.Error(1)
}

func TestSearchRoutes_SearchHotels_Pagination(t *testing.T) {
	svc := new(MockSearchService)
	server := httpserver.Default(testConfig())
	server.SearchService = svc

	checkIn := time.Now().Add(24 * time.Hour).Format("2006-01-02")
	checkOut := time.Now().Add(48 * time.Hour).Format("2006-01-02")
	payload := map[string]any{
		"query":      "ha noi",
		"checkInAt":  checkIn,
		"checkOutAt": checkOut,
		"roomCount":  1,
		"adultCount": 2,
		"page":       2,
		"pageSize":   1,
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	results := []search.HotelSearchResult{
		{HotelID: "h-1", Name: "A"},
		{HotelID: "h-2", Name: "B"},
	}
	svc.On("SearchHotels", mock.Anything, mock.Anything).Return(results, nil).Once()

	req := httptest.NewRequest(http.MethodPost, "/api/search/hotels", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.Router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "\"hotelId\":\"h-2\"")
	svc.AssertExpectations(t)
}

func TestSearchRoutes_SearchHotelRooms_MissingHotelID(t *testing.T) {
	svc := new(MockSearchService)
	server := httpserver.Default(testConfig())
	server.SearchService = svc

	payload := map[string]any{
		"checkInAt":  "2026-04-01",
		"checkOutAt": "2026-04-03",
		"roomCount":  1,
		"adultCount": 2,
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/search/hotels//rooms", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.Router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestSearchRoutes_SearchHotelRoomCombinations(t *testing.T) {
	svc := new(MockSearchService)
	server := httpserver.Default(testConfig())
	server.SearchService = svc

	payload := map[string]any{
		"checkInAt":     "2026-04-01",
		"checkOutAt":    "2026-04-03",
		"roomCount":     1,
		"adultCount":    2,
		"maxCombinations": 2,
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	result := search.HotelRoomCombinationsResult{HotelID: "h-1", RequestedRoomCount: 1}
	svc.On("SearchHotelRoomCombinations", mock.Anything, "h-1", mock.Anything, 2).Return(result, nil).Once()

	req := httptest.NewRequest(http.MethodPost, "/api/search/hotels/h-1/room-combinations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.Router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "\"hotelId\":\"h-1\"")
	svc.AssertExpectations(t)
}
