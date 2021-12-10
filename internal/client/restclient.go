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
	} else if code == 500 {
		return fmt.Errorf("500 - Server Error")
	} else {
		return nil
	}

}

func (r *RestClient) NewTask(name string) (api.Task, error) {
	new := &api.Task{
		Name:        name,
		CustomerRef: "usual",
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

func (r *RestClient) SetActivity(name string, start_ts time.Time, end_ts time.Time) (api.Profile, error) {
	roundTo5Minutes, _ := time.ParseDuration("5m")
	act := api.Profile{
		ActivityName:  name,
		ActivityStart: start_ts.Round(roundTo5Minutes),
		ActivityTimer: end_ts.Round(roundTo5Minutes),
	}
	code, profile := server.SetActivity(act)
	return profile, r.httpCodeToError(code)
}

func (r *RestClient) GetActivity() (api.Profile, error) {
	code, profile := server.GetProfile()
	return profile, r.httpCodeToError(code)
}

func (r *RestClient) SaveRecords(records []api.Record) error {
	roundTo5Minutes, _ := time.ParseDuration("5m")
	for _, rec := range records {
		rec.Start = rec.Start.Round(roundTo5Minutes)
		rec.End = rec.End.Round(roundTo5Minutes)

		code, _ := server.SaveRecord(rec)
		err := r.httpCodeToError(code)
		if err != nil {
			return fmt.Errorf("Cannot Save Record: %s", rec.Name)
		}
	}
	return nil
}
