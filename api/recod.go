package api

import "time"

type Record struct {
	Id          string
	UserName    string
	Title       string
	Description string
	Project     string
	Task        string

	Start time.Time
	End   time.Time
}
