package api

import "time"

type Record struct {
	Id          string
	Name        string
	CustomerRef string
	Description string

	Start time.Time
	End   time.Time
}
