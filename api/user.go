package api

import (
	"fmt"
	"time"
)

type User struct {
	Name     string
	Inactive bool
	Activity Activity
	Settings Settings
}
type Settings struct {
	HelloTimer      time.Duration `json:"hello_timer,omitempty"`
	DefaultEstimate time.Duration `json:"default_estimate,omitempty"`
	RoundTo         time.Duration `json:"round_to,omitempty"`
	MissedWorkAlarm time.Duration `json:"alarm,omitempty"`
	Weekdays        []string      `json:"weekdays,omitempty"`
}

type Activity struct {
	ActivityName    string    `yaml:"activity_name" json:"activity_name"`
	ActivityComment string    `yaml:"activity_comment,omitempty" json:"activity_comment,omitempty"`
	ActivityStart   time.Time `yaml:"activity_start,omitempty" json:"activity_start,omitempty"`
	ActivityTimer   time.Time `yaml:"activity_timer,omitempty" json:"activity_timer,omitempty"`
}

func (a *Activity) CheckActivityActive() error {
	if a.ActivityName == "" {
		return fmt.Errorf("no activity")
	}
	return nil
}

func (a *Activity) CheckNoActivityActive() error {
	if err := a.CheckActivityActive(); err == nil {
		return fmt.Errorf("activity '%s' active", a.ActivityName)
	}
	return nil
}

func NewDefaultUser(name string) User {
	roundTo, _ := time.ParseDuration("15m")
	missedWorkAlarm, _ := time.ParseDuration("12h")
	defaultEstimate, _ := time.ParseDuration("1h")
	helloTimer, _ := time.ParseDuration("1h")

	new := User{
		Name:     name,
		Inactive: false,
		Activity: Activity{},
		Settings: Settings{
			RoundTo:         roundTo,
			HelloTimer:      helloTimer,
			DefaultEstimate: defaultEstimate,
			MissedWorkAlarm: missedWorkAlarm,
			Weekdays:        []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"},
		},
	}
	return new
}

func (a *Activity) AddComment(comment string) {
	if comment == "" {
		return
	}
	if a.ActivityComment == "" {
		a.ActivityComment = comment
		return
	}
	a.ActivityComment = a.ActivityComment + "\n" + comment
}

func (p *User) SetActivity(name string, comment string, start time.Time, timer time.Time) {
	p.Activity.ActivityName = name
	p.Activity.ActivityComment = comment
	p.Activity.ActivityStart = start
	p.Activity.ActivityTimer = timer
}

func (p *User) ClearActivity() {
	p.Activity.ActivityName = ""
	p.Activity.ActivityComment = ""
	p.Activity.ActivityStart = time.Time{}
	p.Activity.ActivityTimer = time.Time{}
}
