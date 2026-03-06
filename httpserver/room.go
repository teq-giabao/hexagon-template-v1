package httpserver

import (
	"hexagon/room"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func (s *Server) RegisterRoomRoutes() {
	s.Router.POST("/api/rooms", s.handleAddRoom)
	s.Router.POST("/api/room-amenities", s.handleAddRoomAmenity)
	s.Router.POST("/api/rooms/:room_id/inventories", s.handleAddRoomInventory)
}

// handleAddRoom godoc
// @Summary Create Room
// @Description Create a room type with images and optional amenityIds mapped in the same transaction.
// @Tags rooms
// @Accept json
// @Produce json
// @Param payload body AddRoomRequest true "Room payload"
// @Success 201 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 404 {object} APIErrorResponse
// @Failure 409 {object} APIErrorResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/rooms [post]
func (s *Server) handleAddRoom(c echo.Context) error {
	var req AddRoomRequest
	if err := c.Bind(&req); err != nil {
		return respondError(c, http.StatusBadRequest, "invalid request body", err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return respondError(c, http.StatusBadRequest, "invalid request body", err.Error())
	}

	created, err := s.RoomService.AddRoom(c.Request().Context(), req.ToRoom())
	if err != nil {
		return err
	}
	return respondCreated(c, toRoomResponse(created))
}

// handleAddRoomAmenity godoc
// @Summary Create Room Amenity
// @Description Create a room amenity in master list.
// @Tags rooms
// @Accept json
// @Produce json
// @Param payload body AddRoomAmenityRequest true "Room amenity payload"
// @Success 201 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 409 {object} APIErrorResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/room-amenities [post]
func (s *Server) handleAddRoomAmenity(c echo.Context) error {
	var req AddRoomAmenityRequest
	if err := c.Bind(&req); err != nil {
		return respondError(c, http.StatusBadRequest, "invalid request body", err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return respondError(c, http.StatusBadRequest, "invalid request body", err.Error())
	}

	created, err := s.RoomService.AddAmenity(c.Request().Context(), req.ToRoomAmenity())
	if err != nil {
		return err
	}
	return respondCreated(c, toRoomAmenityResponse(created))
}

// handleAddRoomInventory godoc
// @Summary Create Room Inventory
// @Description Create inventory for a room on a specific date.
// @Tags rooms
// @Accept json
// @Produce json
// @Param room_id path string true "Room ID"
// @Param payload body AddRoomInventoryRequest true "Room inventory payload"
// @Success 201 {object} APISuccessResponse
// @Failure 400 {object} APIErrorResponse
// @Failure 404 {object} APIErrorResponse
// @Failure 409 {object} APIErrorResponse
// @Failure 500 {object} APIErrorResponse
// @Router /api/rooms/{room_id}/inventories [post]
func (s *Server) handleAddRoomInventory(c echo.Context) error {
	roomID := c.Param("room_id")
	if roomID == "" {
		return respondError(c, http.StatusBadRequest, "invalid room id", "room_id is required")
	}

	var req AddRoomInventoryRequest
	if err := c.Bind(&req); err != nil {
		return respondError(c, http.StatusBadRequest, "invalid request body", err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return respondError(c, http.StatusBadRequest, "invalid request body", err.Error())
	}

	date, err := parseISODate(req.Date)
	if err != nil {
		return respondError(c, http.StatusBadRequest, "invalid request body", "date must be in YYYY-MM-DD format")
	}

	created, err := s.RoomService.AddInventory(c.Request().Context(), req.ToRoomInventory(roomID, date))
	if err != nil {
		return err
	}
	return respondCreated(c, toRoomInventoryResponse(created))
}

func (r AddRoomInventoryRequest) ToRoomInventory(roomID string, date time.Time) room.RoomInventory {
	return room.RoomInventory{
		RoomID:          roomID,
		Date:            date,
		TotalInventory:  r.TotalInventory,
		HeldInventory:   r.HeldInventory,
		BookedInventory: r.BookedInventory,
	}
}
