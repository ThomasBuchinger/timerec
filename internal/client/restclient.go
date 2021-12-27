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

func (r *RestClient) NewWorkItem(name string) (api.WorkItem, error) {
	new := &api.WorkItem{
		Name:           name,
		RecordTemplate: api.RecordTemplate{},
		Activities:     []api.TimeEntry{},
	}

	code, t := server.NewWorkItem(*new)
	return t, r.httpCodeToError(code)
}

func (r *RestClient) ListWorkItems() ([]api.WorkItem, error) {
	code, list := server.ListWorkItems()
	return list, r.httpCodeToError(code)
}

func (r *RestClient) FindWorkItemByName(name string) (api.WorkItem, bool, error) {
	list, err := r.ListWorkItems()
	if err != nil {
		return api.WorkItem{}, false, err
	}

	for _, task := range list {
		if task.Name == name {
			return task, true, nil
		}
	}
	return api.WorkItem{}, false, nil
}

func (r *RestClient) UpdateWorkItem(update api.WorkItem) (api.WorkItem, error) {
	code, updated := server.UpdateWorkItems(update)
	return updated, r.httpCodeToError(code)
}

func (r *RestClient) DeleteWorkItem(toDelete api.WorkItem) (api.WorkItem, error) {
	code, deleted := server.DeleteWorkItems(toDelete)
	return deleted, r.httpCodeToError(code)
}

func (r *RestClient) AddActivityToWorkItem(taskName string, comment string, start_ts time.Time, end_ts time.Time) (api.WorkItem, error) {
	roundTo5Minutes, _ := time.ParseDuration("5m")
	roundedStart := start_ts.Round(roundTo5Minutes)
	roundedEnd := end_ts.Round(roundTo5Minutes)
	task := api.WorkItem{
		Name: taskName,
		Activities: []api.TimeEntry{{
			Comment: comment,
			Start:   roundedStart,
			End:     roundedEnd,
		}},
	}

	code, updated := server.UpdateWorkItems(task)
	return updated, r.httpCodeToError(code)
}

func (r *RestClient) SaveRecord(record api.Record) (api.Record, error) {
	code, rec := server.SaveRecord(record)
	return rec, r.httpCodeToError(code)
}
