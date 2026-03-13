package httpserver_test

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"hexagon/hotel"
	"hexagon/httpserver"
	"hexagon/upload"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockHotelService struct {
	mock.Mock
}

func (m *MockHotelService) ListHotels(ctx context.Context) ([]hotel.Hotel, error) {
	args := m.Called(ctx)
	return args.Get(0).([]hotel.Hotel), args.Error(1)
}

func (m *MockHotelService) GetHotelByID(ctx context.Context, id string) (hotel.Hotel, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(hotel.Hotel), args.Error(1)
}

func (m *MockHotelService) AddHotel(ctx context.Context, h hotel.Hotel) (hotel.Hotel, error) {
	args := m.Called(ctx, h)
	return args.Get(0).(hotel.Hotel), args.Error(1)
}

type MockUploadService struct {
	mock.Mock
}

func (m *MockUploadService) UploadImages(ctx context.Context, folder string, files []upload.File) ([]upload.UploadedFile, error) {
	args := m.Called(ctx, folder, mock.Anything)
	if f, ok := args.Get(0).([]upload.UploadedFile); ok {
		return f, args.Error(1)
	}
	return nil, args.Error(1)
}

func TestHotelRoutes_ListHotels(t *testing.T) {
	svc := new(MockHotelService)
	server := httpserver.Default(testConfig())
	server.HotelService = svc

	hotels := []hotel.Hotel{{ID: "h-1", Name: "Hotel A"}}
	svc.On("ListHotels", mock.Anything).Return(hotels, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/hotels", nil)
	rec := httptest.NewRecorder()
	server.Router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "\"id\":\"h-1\"")
	svc.AssertExpectations(t)
}

func TestHotelRoutes_GetByID(t *testing.T) {
	svc := new(MockHotelService)
	server := httpserver.Default(testConfig())
	server.HotelService = svc

	h := hotel.Hotel{ID: "h-1", Name: "Hotel A"}
	svc.On("GetHotelByID", mock.Anything, "h-1").Return(h, nil).Once()

	req := httptest.NewRequest(http.MethodGet, "/api/hotels/h-1", nil)
	rec := httptest.NewRecorder()
	server.Router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "\"id\":\"h-1\"")
	svc.AssertExpectations(t)
}

func TestHotelRoutes_AddHotel(t *testing.T) {
	svc := new(MockHotelService)
	server := httpserver.Default(testConfig())
	server.HotelService = svc

	payload := map[string]any{
		"name":        "Hotel A",
		"description": "desc",
		"address":     "123 Street",
		"city":        "City",
		"checkInTime": "15:00",
		"checkOutTime": "12:00",
		"paymentOptions": []map[string]any{{"paymentOption": "immediate", "enabled": true}},
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	created := hotel.Hotel{ID: "h-1", Name: "Hotel A"}
	svc.On("AddHotel", mock.Anything, mock.Anything).Return(created, nil).Once()

	req := httptest.NewRequest(http.MethodPost, "/api/hotels", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.Router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), "\"id\":\"h-1\"")
	svc.AssertExpectations(t)
}

func TestHotelRoutes_AddHotel_InvalidTime(t *testing.T) {
	svc := new(MockHotelService)
	server := httpserver.Default(testConfig())
	server.HotelService = svc

	payload := map[string]any{
		"name":        "Hotel A",
		"address":     "123 Street",
		"city":        "City",
		"checkInTime": "99:99",
		"checkOutTime": "12:00",
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/hotels", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.Router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "invalid check-in/check-out time")
}

func TestHotelRoutes_UploadImages_NoService(t *testing.T) {
	server := httpserver.Default(testConfig())

	req := httptest.NewRequest(http.MethodPost, "/api/hotels/upload-images", nil)
	rec := httptest.NewRecorder()
	server.Router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotImplemented, rec.Code)
	assert.Contains(t, rec.Body.String(), "upload service is not configured")
}

func TestHotelRoutes_UploadImages_Success(t *testing.T) {
	svc := new(MockUploadService)
	server := httpserver.Default(testConfig())
	server.UploadService = svc

	expected := []upload.UploadedFile{{
		FileName:    "a.jpg",
		URL:         "https://cdn.example.com/a.jpg",
		Size:        3,
		ContentType: "image/jpeg",
	}}
	svc.On("UploadImages", mock.Anything, "hotel-images", mock.Anything).Return(expected, nil).Once()

	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)
	part, err := writer.CreateFormFile("images", "a.jpg")
	require.NoError(t, err)
	_, _ = part.Write([]byte("abc"))
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/api/hotels/upload-images", buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	rec := httptest.NewRecorder()
	server.Router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), "\"fileName\":\"a.jpg\"")
	svc.AssertExpectations(t)
}
