package api

import "time"

type Profile struct {
	ActivityName    string    `yaml:"activity_name"`
	ActivityComment string    `yaml:"activity_comment,omitempty"`
	ActivityStart   time.Time `yaml:"activity_start,omitempty"`
	ActivityTimer   time.Time `yaml:"activity_timer,omitempty"`
}
