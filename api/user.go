package api

import (
	"fmt"
	"time"
)

type User struct {
	Activity Activity
	Settings Settings
}
type Settings struct {
	RoundTo time.Duration
}

type Activity struct {
	ActivityName    string    `yaml:"activity_name" json:"activity_name"`
	ActivityComment string    `yaml:"activity_comment,omitempty" json:"activity_comment"`
	ActivityStart   time.Time `yaml:"activity_start,omitempty" json:"activity_start"`
	ActivityTimer   time.Time `yaml:"activity_timer,omitempty" json:"activity_timer"`
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

func (p *User) GetRoundTo() time.Duration {
	if p.Settings.RoundTo == time.Duration(0) {
		ret, _ := time.ParseDuration("5m")
		return ret
	}
	return p.Settings.RoundTo
}

func (p *User) ClearActivity() {
	p.Activity.ActivityName = ""
	p.Activity.ActivityComment = ""
	p.Activity.ActivityStart = time.Time{}
	p.Activity.ActivityTimer = time.Time{}
}
