package api

import "time"

type Record struct {
	Id          string
	Title       string
	Description string
	Project     string
	Task        string

	Start time.Time
	End   time.Time
}
