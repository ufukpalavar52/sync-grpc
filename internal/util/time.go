package util

import "time"

func StrToTime(date string) (time.Time, error) {
	return time.Parse(time.RFC3339, date)
}

func StrToTimeWithoutError(date string) time.Time {
	parse, _ := time.Parse(time.RFC3339, date)
	return parse
}
