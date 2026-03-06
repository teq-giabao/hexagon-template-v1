package httpserver

import "time"

const isoDateLayout = "2006-01-02"

func parseISODate(value string) (time.Time, error) {
	now := time.Now()
	return time.ParseInLocation(isoDateLayout, value, now.Location())
}

func parseISODateRange(from, to string) (time.Time, time.Time, error) {
	fromDate, err := parseISODate(from)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	toDate, err := parseISODate(to)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return fromDate, toDate, nil
}
