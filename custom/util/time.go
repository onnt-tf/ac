package util

import "time"

// UTCNow returns the current time in UTC.
func UTCNow() time.Time {
	return time.Now().UTC()
}
