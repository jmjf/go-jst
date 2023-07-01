package common

import "time"

func TruncateTimeToMs(tm time.Time) time.Time {
	return tm.Truncate(time.Millisecond)
}
