package httpserver_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"hexagon/httpserver"
	"hexagon/room"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockRoomService struct {
	mock.Mock
}

func (m *MockRoomService) AddRoom(ctx context.Context, r room.Room) (room.Room, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(room.Room), args.Error(1)
}

func (m *MockRoomService) AddAmenity(ctx context.Context, amenity room.RoomAmenity) (room.RoomAmenity, error) {
	args := m.Called(ctx, amenity)
	return args.Get(0).(room.RoomAmenity), args.Error(1)
}

func (m *MockRoomService) AddInventory(ctx context.Context, inv room.RoomInventory) (room.RoomInventory, error) {
	args := m.Called(ctx, inv)
	return args.Get(0).(room.RoomInventory), args.Error(1)
}

func TestRoomRoutes_AddRoom(t *testing.T) {
	svc := new(MockRoomService)
	server := httpserver.Default(testConfig())
	server.RoomService = svc

	payload := map[string]any{
		"hotelId":      "h-1",
		"name":         "Deluxe",
		"basePrice":    120,
		"maxAdult":     2,
		"maxChild":     1,
		"maxOccupancy": 3,
		"images":       []map[string]any{{"url": "https://cdn.example.com/room.jpg", "isCover": true}},
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	created := room.Room{ID: "r-1", Name: "Deluxe"}
	svc.On("AddRoom", mock.Anything, mock.Anything).Return(created, nil).Once()

	req := httptest.NewRequest(http.MethodPost, "/api/rooms", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	server.Router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), "\"id\":\"r-1\"")
	svc.AssertExpectations(t)
}

func TestRoomRoutes_AddAmenity(t *testing.T) {
	svc := new(MockRoomService)
	server := httpserver.Default(testConfig())
	server.RoomService = svc

	payload := map[string]any{
		"code": "wifi",
		"name": "WiFi",
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	created := room.RoomAmenity{ID: "a-1", Code: "wifi", Name: "WiFi"}
	svc.On("AddAmenity", mock.Anything, mock.Anything).Return(created, nil).Once()

	req := httptest.NewRequest(http.MethodPost, "/api/room-amenities", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	server.Router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), "\"id\":\"a-1\"")
	svc.AssertExpectations(t)
}

func TestRoomRoutes_AddInventory_MissingRoomID(t *testing.T) {
	svc := new(MockRoomService)
	server := httpserver.Default(testConfig())
	server.RoomService = svc

	payload := map[string]any{
		"date":           "2026-04-01",
		"totalInventory": 10,
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/rooms//inventories", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	server.Router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRoomRoutes_AddInventory(t *testing.T) {
	svc := new(MockRoomService)
	server := httpserver.Default(testConfig())
	server.RoomService = svc

	payload := map[string]any{
		"date":            "2026-04-01",
		"totalInventory":  10,
		"heldInventory":   1,
		"bookedInventory": 2,
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	created := room.RoomInventory{ID: "inv-1", RoomID: "r-1"}
	svc.On("AddInventory", mock.Anything, mock.Anything).Return(created, nil).Once()

	req := httptest.NewRequest(http.MethodPost, "/api/rooms/r-1/inventories", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	server.Router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), "\"id\":\"inv-1\"")
	svc.AssertExpectations(t)
}
