// Package date_util contains utility methods for Time instances.
package date_util

import (
	"time"
)

const (
	// Format as yyyyMMdd
	YMDFormat = "20060102"
)

// Clock is the interface that wraps the Now method.
type Clock interface {
	Now() time.Time
}

// SystemClock provides the current time.
type SystemClock struct {
}

func (s SystemClock) Now() time.Time {
	return time.Now()
}

// TimeToDate returns t with the time of day zeroed out and the time zone GMT.
func TimeToDate(t time.Time) time.Time {
	y, m, d := t.Date()
	return YMD(y, int(m), d)
}

// YMD creates a new time.Time object in UTC time zone from year, month, day.
func YMD(year int, month int, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}
