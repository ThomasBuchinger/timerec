package api

import "time"

type Profile struct {
	OpenTasks []string

	ActivityName  string
	ActivityStart time.Time
	ActivityTimer time.Time
}
