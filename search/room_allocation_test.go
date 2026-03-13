package search

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitGuestsAcrossRooms(t *testing.T) {
	rooms := splitGuestsAcrossRooms(2, 3, []int{5, 6, 7})
	assert.Len(t, rooms, 2)
	assert.Equal(t, 2, rooms[0].adults)
	assert.Equal(t, 1, rooms[1].adults)
	assert.Len(t, rooms[0].childrenAges, 2)
	assert.Len(t, rooms[1].childrenAges, 1)
}

func TestMinimumRequiredRooms(t *testing.T) {
	candidates := []RoomCandidate{
		{RoomID: "r1", MaxAdult: 2, MaxChild: 1, MaxOccupancy: 3, AvailableCount: 1},
		{RoomID: "r2", MaxAdult: 2, MaxChild: 1, MaxOccupancy: 3, AvailableCount: 1},
	}

	count := MinimumRequiredRooms(candidates, 1, 3, []int{}, 12)
	assert.Equal(t, 2, count)

	count = MinimumRequiredRooms(candidates, 1, 2, []int{5}, 12)
	assert.Equal(t, 1, count)
}

func TestMinimumRequiredRooms_UnaccompaniedChildren(t *testing.T) {
	candidates := []RoomCandidate{{RoomID: "r1", MaxAdult: 1, MaxChild: 1, MaxOccupancy: 2, AvailableCount: 1}}
	count := MinimumRequiredRooms(candidates, 1, 0, []int{5}, 12)
	assert.Equal(t, 0, count)
}

func TestRequiredSingleRoomTypeQuantity(t *testing.T) {
	candidate := RoomCandidate{RoomID: "r1", MaxAdult: 2, MaxChild: 1, MaxOccupancy: 3, AvailableCount: 3}

	qty, ok := RequiredSingleRoomTypeQuantity(candidate, 1, 3, []int{}, 12)
	assert.True(t, ok)
	assert.Equal(t, 2, qty)

	candidate.AvailableCount = 1
	qty, ok = RequiredSingleRoomTypeQuantity(candidate, 1, 3, []int{}, 12)
	assert.False(t, ok)
	assert.Equal(t, 0, qty)
}

func TestNormalizeGuestAndUnaccompanied(t *testing.T) {
	adults, children := normalizeGuest(guestRequest{adults: 1, childrenAges: []int{5, 18}}, 12)
	assert.Equal(t, 2, adults)
	assert.Equal(t, 1, children)

	assert.True(t, hasUnaccompaniedChildren(guestRequest{adults: 0, childrenAges: []int{5}}, 12))
	assert.False(t, hasUnaccompaniedChildren(guestRequest{adults: 1, childrenAges: []int{5}}, 12))
}

func TestCeilDiv(t *testing.T) {
	assert.Equal(t, 0, ceilDiv(0, 5))
	assert.Equal(t, math.MaxInt, ceilDiv(5, 0))
	assert.Equal(t, 3, ceilDiv(7, 3))
}
