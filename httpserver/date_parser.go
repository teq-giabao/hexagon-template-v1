package httpserver

import "time"

const isoDateLayout = "2006-01-02"

func isoDate(value string) (time.Time, error) {
	now := time.Now()
	return time.ParseInLocation(isoDateLayout, value, now.Location())
}
