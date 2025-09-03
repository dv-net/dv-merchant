package time_conv

import "time"

func FirstDayOfWeek(tm time.Time) time.Time {
	weekday := int64(time.Duration(tm.Weekday()))
	if weekday == 0 {
		weekday = 7
	}

	year, month, day := tm.Date()
	currentZeroDay := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
	hours := -1 * (weekday - 1) * 24

	return currentZeroDay.Add(time.Hour * time.Duration(hours))
}
