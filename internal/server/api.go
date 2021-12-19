package server

import (
	"time"

	"github.com/thomasbuchinger/timerec/api"
)

func GetProfile() (int, api.Profile) {
	mgr := NewServer()
	p, _ := mgr.stateProvider.GetProfile()
	return 200, p
}
func SetActivity(new api.Profile) (int, api.Profile) {
	mgr := NewServer()
	old, _ := mgr.stateProvider.GetProfile()

	if new.ActivityComment != "" {
		old.ActivityComment = new.ActivityComment
	}
	if new.ActivityName != "" {
		old.ActivityName = new.ActivityName
	}
	if !new.ActivityStart.IsZero() {
		old.ActivityStart = new.ActivityStart
	}
	if !new.ActivityTimer.IsZero() {
		old.ActivityTimer = new.ActivityTimer
	}

	p, _ := mgr.stateProvider.UpdateProfile(old)
	return 200, p
}

func ClearActivity() (int, api.Profile) {
	mgr := NewServer()
	profile, err := mgr.stateProvider.GetProfile()
	if err != nil {
		return 404, api.Profile{}
	}
	profile.ActivityName = ""
	profile.ActivityComment = ""
	profile.ActivityStart = time.Time{}
	profile.ActivityTimer = time.Time{}
	updated, err := mgr.stateProvider.UpdateProfile(profile)
	if err != nil {
		return 500, api.Profile{}
	}
	return 200, updated
}

func GetTemplates() (int, []api.RecordTemplate) {
	mgr := NewServer()

	ret, err := mgr.stateProvider.GetTemplates()
	if err != nil {
		return 500, ret
	}
	return 200, ret
}

func NewTask(new api.Task) (int, api.Task) {
	mgr := NewServer()
	allTasks, err := mgr.stateProvider.ListTasks()
	if err != nil {
		return 500, api.Task{}
	}

	for _, t := range allTasks {
		if new.Name == t.Name {
			// CONFLICT
			return 500, api.Task{}
		}
	}

	new.CreatedAt = time.Now()
	ret, err := mgr.stateProvider.CreateTask(new)
	if err != nil {
		return 500, ret
	}
	return 200, ret
}

func ListTasks() (int, []api.Task) {
	mgr := NewServer()

	ret, err := mgr.stateProvider.ListTasks()
	if err != nil {
		return 500, ret
	}
	return 200, ret
}

func UpdateTask(t api.Task) (int, api.Task) {
	mgr := NewServer()

	old, err := mgr.stateProvider.GetTask(t)
	if err != nil {
		return 404, api.Task{}
	}

	old.Update(t)
	updated, err := mgr.stateProvider.UpdateTask(old)
	if err != nil {
		return 500, api.Task{}
	}
	return 200, updated
}

func DeleteTask(t api.Task) (int, api.Task) {
	mgr := NewServer()

	deleted, err := mgr.stateProvider.DeleteTask(t)
	if err != nil {
		return 500, api.Task{}
	}
	return 200, deleted
}

func Reconcile() int {
	mgr := NewServer()
	mgr.Reconcile()
	return 200
}

func SaveRecord(rec api.Record) (int, api.Record) {
	mgr := NewServer()

	ret, err := mgr.backend.SaveRecord(rec)
	if err != nil {
		return 500, ret
	}
	return 200, ret

}
