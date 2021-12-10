package api

import "time"

type TimeEntry struct {
	Comment string
	Start   time.Time
	End     time.Time
}
