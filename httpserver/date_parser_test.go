package httpserver

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestISODate(t *testing.T) {
	value := time.Now().Format("2006-01-02")
	parsed, err := isoDate(value)
	assert.NoError(t, err)
	assert.Equal(t, value, parsed.Format("2006-01-02"))

	_, err = isoDate("invalid")
	assert.Error(t, err)
}

func TestParseClockTime(t *testing.T) {
	parsed, err := parseClockTime("15:04")
	assert.NoError(t, err)
	assert.Equal(t, 15, parsed.Hour())
	assert.Equal(t, 4, parsed.Minute())

	parsed, err = parseClockTime("15:04:05")
	assert.NoError(t, err)
	assert.Equal(t, 5, parsed.Second())

	_, err = parseClockTime("99:99")
	assert.Error(t, err)
}

func TestParseHotelTimes(t *testing.T) {
	_, _, err := parseHotelTimes("15:00", "12:00")
	assert.NoError(t, err)

	_, _, err = parseHotelTimes("bad", "12:00")
	assert.Error(t, err)
}
