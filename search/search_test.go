package search

import (
	"testing"
	"time"

	"hexagon/errs"

	"github.com/stretchr/testify/assert"
)

func validCriteria() Criteria {
	now := time.Now().UTC()
	return Criteria{
		Query:        "test",
		CheckInDate:  now.Add(24 * time.Hour),
		CheckOutDate: now.Add(48 * time.Hour),
		Adults:       2,
		ChildrenAges: []int{5},
		RoomCount:    1,
		RatingMin:    4,
		AmenityIDs:   []string{"wifi"},
		PaymentOptions: []string{"card"},
	}
}

func TestCriteriaValidate(t *testing.T) {
	c := validCriteria()
	assert.NoError(t, c.Validate())

	bad := c
	bad.CheckInDate = time.Time{}
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(bad.Validate()))

	bad = c
	bad.CheckOutDate = time.Time{}
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(bad.Validate()))

	bad = c
	bad.CheckOutDate = bad.CheckInDate
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(bad.Validate()))

	bad = c
	bad.Adults = 0
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(bad.Validate()))

	bad = c
	bad.ChildrenAges = []int{-1}
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(bad.Validate()))

	bad = c
	bad.RoomCount = 0
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(bad.Validate()))

	bad = c
	bad.AmenityIDs = []string{""}
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(bad.Validate()))

	bad = c
	bad.PaymentOptions = []string{""}
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(bad.Validate()))

	bad = c
	bad.RatingMin = -1
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(bad.Validate()))

	bad = c
	bad.RatingMin = 6
	assert.Equal(t, errs.EINVALID, errs.ErrorCode(bad.Validate()))
}
