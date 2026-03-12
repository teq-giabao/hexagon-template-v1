package search

import (
	"math"
	"sort"
)

type RoomCandidate struct {
	RoomID         string
	Name           string
	Description    string
	BasePrice      float64
	MaxAdult       int
	MaxChild       int
	MaxOccupancy   int
	AvailableCount int
}

type guestRequest struct {
	adults       int
	childrenAges []int
}

// MinimumRequiredRooms uses backtracking (DFS) to find the minimum number
// of rooms required to accommodate a specific group of guests.
// It tries to split guests into rooms starting from the requested room count
// up to the total number of available rooms.
func MinimumRequiredRooms(
	candidates []RoomCandidate,
	startRoomCount int,
	adults int,
	childrenAges []int,
	childMaxAge int,
) int {
	totalAvailable := 0
	for i := range candidates {
		totalAvailable += candidates[i].AvailableCount
	}

	for roomCount := startRoomCount; roomCount <= totalAvailable; roomCount++ {
		requests := splitGuestsAcrossRooms(roomCount, adults, childrenAges)
		if canAllocateRequestedRooms(candidates, requests, childMaxAge) {
			return roomCount
		}
	}

	return 0
}

// splitGuestsAcrossRooms uniformly distributes adults and children across the requested
// number of rooms. It ensures that rooms receive an evenly distributed load.
func splitGuestsAcrossRooms(roomCount, adults int, childrenAges []int) []guestRequest {
	if roomCount <= 0 {
		return nil
	}

	rooms := make([]guestRequest, roomCount)
	for i := 0; i < roomCount; i++ {
		rooms[i] = guestRequest{adults: 0, childrenAges: make([]int, 0)}
	}

	for i := 0; i < adults; i++ {
		rooms[i%roomCount].adults++
	}

	for i := range childrenAges {
		rooms[i%roomCount].childrenAges = append(rooms[i%roomCount].childrenAges, childrenAges[i])
	}

	return rooms
}

// canAllocateRequestedRooms acts as the core backtracking algorithm.
// It verifies if a given distribution of guests (requests) can fit perfectly into
// the available candidate rooms without leaving unaccompanied children.
func canAllocateRequestedRooms(candidates []RoomCandidate, requests []guestRequest, childMaxAge int) bool {
	if len(requests) == 0 {
		return true
	}

	if hasAnyUnaccompaniedChildren(requests, childMaxAge) {
		return false
	}

	order, compat, ok := buildCompatibilityOrder(candidates, requests, childMaxAge)
	if !ok {
		return false
	}

	remaining := make([]int, len(candidates))
	for i := range candidates {
		remaining[i] = candidates[i].AvailableCount
	}

	var dfs func(pos int) bool
	dfs = func(pos int) bool {
		if pos == len(order) {
			return true
		}

		reqIdx := order[pos]
		for _, candidateIdx := range compat[reqIdx] {
			if remaining[candidateIdx] <= 0 {
				continue
			}

			remaining[candidateIdx]--

			if dfs(pos + 1) {
				return true
			}

			remaining[candidateIdx]++
		}

		return false
	}

	return dfs(0)
}

// buildCompatibilityOrder determines which candidate rooms can fulfill each guest request.
// It sorts the requests by the number of compatible rooms (ascending) to optimize
// the DFS branch pruning (most constrained requests are assigned first).
func buildCompatibilityOrder(candidates []RoomCandidate, requests []guestRequest, childMaxAge int) ([]int, [][]int, bool) {
	order := make([]int, len(requests))
	for i := range requests {
		order[i] = i
	}

	compat := make([][]int, len(requests))

	for i := range requests {
		for j := range candidates {
			if candidateCanFit(candidates[j], requests[i], childMaxAge) && candidates[j].AvailableCount > 0 {
				compat[i] = append(compat[i], j)
			}
		}

		if len(compat[i]) == 0 {
			return nil, nil, false
		}
	}

	sort.Slice(order, func(i, j int) bool { return len(compat[order[i]]) < len(compat[order[j]]) })

	return order, compat, true
}

// candidateCanFit evaluates whether a single RoomCandidate has enough
// MaxAdult, MaxChild, and MaxOccupancy to accommodate the given guest request.
func candidateCanFit(c RoomCandidate, req guestRequest, childMaxAge int) bool {
	adults, children := normalizeGuest(req, childMaxAge)
	total := adults + children

	return adults <= c.MaxAdult && children <= c.MaxChild && total <= c.MaxOccupancy
}

func normalizeGuest(req guestRequest, childMaxAge int) (int, int) {
	adults := req.adults
	children := 0

	for i := range req.childrenAges {
		if req.childrenAges[i] > childMaxAge {
			adults++
		} else {
			children++
		}
	}

	return adults, children
}

func hasUnaccompaniedChildren(req guestRequest, childMaxAge int) bool {
	adults, children := normalizeGuest(req, childMaxAge)

	return children > 0 && adults == 0
}

func hasAnyUnaccompaniedChildren(requests []guestRequest, childMaxAge int) bool {
	for i := range requests {
		if hasUnaccompaniedChildren(requests[i], childMaxAge) {
			return true
		}
	}

	return false
}

// RequiredSingleRoomTypeQuantity calculates the number of rooms needed if the guests
// were to stay exclusively in a single specific room type.
// This is critical for quickly assessing if a room type can handle the full capacity
// without exceeding its maximum capability limits.
func RequiredSingleRoomTypeQuantity(candidate RoomCandidate, requestedRoomCount, adults int, childrenAges []int, childMaxAge int) (int, bool) {
	if candidate.AvailableCount <= 0 {
		return 0, false
	}

	normalizedAdults, normalizedChildren := normalizeGuestCounts(adults, childrenAges, childMaxAge)
	if !canHostNormalizedGuests(candidate, normalizedAdults, normalizedChildren) {
		return 0, false
	}

	quantity := maxInt(
		requestedRoomCount,
		ceilDiv(normalizedAdults, candidate.MaxAdult),
		ceilDiv(normalizedChildren, candidate.MaxChild),
		ceilDiv(normalizedAdults+normalizedChildren, candidate.MaxOccupancy),
	)

	if quantity <= 0 || quantity > candidate.AvailableCount {
		return 0, false
	}

	return quantity, true
}

func normalizeGuestCounts(adults int, childrenAges []int, childMaxAge int) (int, int) {
	normalizedAdults := adults
	normalizedChildren := 0

	for i := range childrenAges {
		if childrenAges[i] > childMaxAge {
			normalizedAdults++
		} else {
			normalizedChildren++
		}
	}

	return normalizedAdults, normalizedChildren
}

func canHostNormalizedGuests(candidate RoomCandidate, normalizedAdults, normalizedChildren int) bool {
	if normalizedChildren > 0 && normalizedAdults == 0 {
		return false
	}

	if candidate.MaxAdult <= 0 && normalizedAdults > 0 {
		return false
	}

	if candidate.MaxChild <= 0 && normalizedChildren > 0 {
		return false
	}

	if candidate.MaxOccupancy <= 0 && normalizedAdults+normalizedChildren > 0 {
		return false
	}

	return true
}

func maxInt(values ...int) int {
	max := 0
	for i := range values {
		if values[i] > max {
			max = values[i]
		}
	}

	return max
}

func ceilDiv(numerator, denominator int) int {
	if numerator <= 0 {
		return 0
	}

	if denominator <= 0 {
		return math.MaxInt
	}

	return (numerator + denominator - 1) / denominator
}
