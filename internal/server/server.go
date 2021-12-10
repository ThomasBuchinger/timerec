package server

import (
	"github.com/thomasbuchinger/timerec/api"
)

type Storage interface {
	CreateTask(api.Task) (api.Task, error)
	GetTasks() ([]api.Task, error)
	// UpdateTask(api.Task) (api.Task, error)
	// DeleteTask(api.Task) (api.Task, error)

	AddRecord(api.Record) (api.Record, error)
}

func GetProfile() (int, api.Profile) {
	datasource := &FileStorage{}
	data := datasource.load()

	return 200, data.Profile
}
func SetActivity(activity api.Profile) (int, api.Profile) {
	datasource := &FileStorage{}
	data := datasource.load()
	data.Profile.ActivityName = activity.ActivityName
	data.Profile.ActivityStart = activity.ActivityStart
	data.Profile.ActivityTimer = activity.ActivityTimer
	datasource.store(data)

	return 200, data.Profile
}

func NewTask(t api.Task) (int, api.Task) {
	var datasource Storage
	datasource = &FileStorage{}

	ret, err := datasource.CreateTask(t)
	if err != nil {
		return 500, ret
	}
	return 200, ret
}

func ListTasks() (int, []api.Task) {
	var datasource Storage
	datasource = &FileStorage{}

	ret, err := datasource.GetTasks()
	if err != nil {
		return 500, ret
	}
	return 200, ret
}

func SaveRecord(rec api.Record) (int, api.Record) {
	var datasource Storage
	datasource = &FileStorage{}

	ret, err := datasource.AddRecord(rec)
	if err != nil {
		return 500, ret
	}
	return 200, ret

}
