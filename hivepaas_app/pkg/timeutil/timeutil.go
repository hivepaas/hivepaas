package timeutil

import "time"

func NowUTC() time.Time {
	return time.Now().UTC()
}

func CurrentDateUTC() Date {
	return NewDate(time.Now().UTC())
}

func CurrentYearUTC() int {
	return NowUTC().Year()
}
