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

func NewWorkItem(new api.WorkItem) (int, api.WorkItem) {
	mgr := NewServer()
	allTasks, err := mgr.stateProvider.ListWorkItems()
	if err != nil {
		return 500, api.WorkItem{}
	}

	for _, t := range allTasks {
		if new.Name == t.Name {
			// CONFLICT
			return 500, api.WorkItem{}
		}
	}

	new.CreatedAt = time.Now()
	ret, err := mgr.stateProvider.CreateWorkItem(new)
	if err != nil {
		return 500, ret
	}
	return 200, ret
}

func ListWorkItems() (int, []api.WorkItem) {
	mgr := NewServer()

	ret, err := mgr.stateProvider.ListWorkItems()
	if err != nil {
		return 500, ret
	}
	return 200, ret
}

func UpdateWorkItems(t api.WorkItem) (int, api.WorkItem) {
	mgr := NewServer()

	old, err := mgr.stateProvider.GetWorkItem(t)
	if err != nil {
		return 404, api.WorkItem{}
	}

	old.Update(t)
	updated, err := mgr.stateProvider.UpdateWorkItem(old)
	if err != nil {
		return 500, api.WorkItem{}
	}
	return 200, updated
}

func DeleteWorkItems(t api.WorkItem) (int, api.WorkItem) {
	mgr := NewServer()

	deleted, err := mgr.stateProvider.DeleteWorkItem(t)
	if err != nil {
		return 500, api.WorkItem{}
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
