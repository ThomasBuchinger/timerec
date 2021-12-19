package client

import (
	"fmt"
	"time"

	"github.com/thomasbuchinger/timerec/api"
	"github.com/thomasbuchinger/timerec/internal/server"
)

type RestClient struct {
}

func (r *RestClient) httpCodeToError(code int) error {
	if code == 400 {
		return fmt.Errorf("400 - Bad Request")
	} else if code == 404 {
		return fmt.Errorf("404 - Not Found")
	} else if code == 500 {
		return fmt.Errorf("500 - Server Error")
	} else {
		return nil
	}

}

func (r *RestClient) SetActivity(name string, comment string, start_ts time.Time, end_ts time.Time) (api.Profile, error) {
	roundTo5Minutes, _ := time.ParseDuration("5m")
	roundedStart := start_ts.Round(roundTo5Minutes)
	roundedEnd := end_ts.Round(roundTo5Minutes)

	act := api.Profile{
		ActivityName:    name,
		ActivityComment: comment,
		ActivityStart:   roundedStart,
		ActivityTimer:   roundedEnd,
	}
	code, profile := server.SetActivity(act)
	return profile, r.httpCodeToError(code)
}

func (r *RestClient) GetActivity() (api.Profile, error) {
	code, profile := server.GetProfile()
	return profile, r.httpCodeToError(code)
}

func (r *RestClient) ClearActivity() error {
	code, _ := server.ClearActivity()
	return r.httpCodeToError(code)
}

func (r *RestClient) ListTemplates() ([]api.RecordTemplate, error) {
	code, templates := server.GetTemplates()
	return templates, r.httpCodeToError(code)
}

func (r *RestClient) NewTask(name string) (api.Task, error) {
	new := &api.Task{
		Name:           name,
		RecordTemplate: api.RecordTemplate{},
		Activities:     []api.TimeEntry{},
	}

	code, t := server.NewTask(*new)
	return t, r.httpCodeToError(code)
}

func (r *RestClient) ListTasks() ([]api.Task, error) {
	code, list := server.ListTasks()
	return list, r.httpCodeToError(code)
}

func (r *RestClient) FindTaskByName(name string) (api.Task, bool, error) {
	list, err := r.ListTasks()
	if err != nil {
		return api.Task{}, false, err
	}

	for _, task := range list {
		if task.Name == name {
			return task, true, nil
		}
	}
	return api.Task{}, false, nil
}

func (r *RestClient) UpdateTask(update api.Task) (api.Task, error) {
	code, updated := server.UpdateTask(update)
	return updated, r.httpCodeToError(code)
}

func (r *RestClient) Deleteask(toDelete api.Task) (api.Task, error) {
	code, deleted := server.DeleteTask(toDelete)
	return deleted, r.httpCodeToError(code)
}

func (r *RestClient) AddActivityToTask(taskName string, comment string, start_ts time.Time, end_ts time.Time) (api.Task, error) {
	roundTo5Minutes, _ := time.ParseDuration("5m")
	roundedStart := start_ts.Round(roundTo5Minutes)
	roundedEnd := end_ts.Round(roundTo5Minutes)
	task := api.Task{
		Name: taskName,
		Activities: []api.TimeEntry{{
			Comment: comment,
			Start:   roundedStart,
			End:     roundedEnd,
		}},
	}

	code, updated := server.UpdateTask(task)
	return updated, r.httpCodeToError(code)
}

func (r *RestClient) SaveRecord(record api.Record) (api.Record, error) {
	code, rec := server.SaveRecord(record)
	return rec, r.httpCodeToError(code)
}
