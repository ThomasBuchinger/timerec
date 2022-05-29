package server

import (
	"context"
	"time"

	"github.com/thomasbuchinger/timerec/api"
	"github.com/thomasbuchinger/timerec/internal/server/providers"
)

type SearchJobParams struct {
	Name          string        `json:"name"`
	Owner         string        `json:"owner"`
	StartedAfter  time.Duration `json:"start_after,omitempty" default:"-24h"`
	StartedBefore time.Duration `json:"start_before,omitempty" default:"0s"`
}

type UpdateJobParams struct {
	Name        string `json:"name,omitempty"`
	Owner       string `json:"owner"`
	Template    string `json:"template,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Project     string `json:"project,omitempty"`
	Task        string `json:"task,omitempty"`
}

type CompleteJobParams struct {
	SearchJobParams `json:",inline"`
	Status          JobStatus `json:"status"`
}

type JobStatus string

const (
	JobStatusCancel JobStatus = "canceled"
	JobStatusFinish JobStatus = "finished"
)

type JobResponse struct {
	Success bool    `json:"success"`
	Created bool    `json:"created"`
	Job     api.Job `json:"job,omitempty"`
}

func (mgr *TimerecServer) GetJob(ctx context.Context, params SearchJobParams) (JobResponse, error) {
	state, err := mgr.StateProvider.Refresh(params.Owner)
	if err != nil {
		return JobResponse{}, mgr.MakeNewResponseError(ProviderError, err, "Unable to query Provider: %s", err.Error())
	}

	item, proverr := providers.GetJob(&state, api.Job{
		Name:  params.Name,
		Owner: params.Owner,
	})
	if proverr == providers.ProviderOk {
		return JobResponse{Success: true, Created: false, Job: item}, nil
	}
	if proverr == providers.ProviderNotFound {
		return JobResponse{Success: false, Created: false, Job: api.Job{}}, nil
	}

	return JobResponse{}, mgr.MakeNewResponseError(BadRequest, proverr, "Error querying Job '%s'", item.Name)
}

func (mgr *TimerecServer) CreateJobIfMissing(ctx context.Context, params SearchJobParams) (JobResponse, error) {
	state, err := mgr.StateProvider.Refresh(params.Owner)
	if err != nil {
		return JobResponse{}, mgr.MakeNewResponseError(ProviderError, err, "Unable to query Provider: %s", err.Error())
	}

	response, err := mgr.GetJob(ctx, params)
	if err != nil {
		mgr.Logger.Error(err)
		return JobResponse{}, err
	}
	if response.Success {
		return response, nil
	}

	new := api.NewJob(params.Name, params.Owner)
	proverr := providers.CreateJob(&state, new)
	if proverr != providers.ProviderOk {
		return JobResponse{}, mgr.MakeNewResponseError(BadRequest, proverr, "Unable to create Job '%s'", params.Name)
	}
	err = mgr.StateProvider.Save(state.Partition, state)
	if err != nil {
		return JobResponse{}, mgr.MakeNewResponseError(ProviderError, err, "Unable to save Job '%s'", params.Name)
	}

	mgr.Logger.Infof("Created Job: %s", new.Name)
	return JobResponse{Success: true, Created: true, Job: new}, nil
}

func (mgr *TimerecServer) UpdateJob(ctx context.Context, params UpdateJobParams) (JobResponse, error) {
	state, err := mgr.StateProvider.Refresh(params.Owner)
	if err != nil {
		return JobResponse{}, mgr.MakeNewResponseError(ProviderError, err, "Unable to query Provider: %s", err.Error())
	}

	// Check if Job exists
	response, err := mgr.GetJob(
		ctx,
		SearchJobParams{Name: params.Name, Owner: params.Owner, StartedAfter: -24 * time.Hour, StartedBefore: time.Duration(0)},
	)
	if err != nil {
		return JobResponse{}, err
	}
	if !response.Success {
		mgr.Logger.Warnf("Job with name '%s' does not exist", params.Name)
		return JobResponse{}, mgr.MakeNewResponseError(BadRequest, err, "Job deos not exist")
	}

	// Update job according to Template
	job := response.Job
	if params.Template != "" {
		templateExists, _ := providers.HasTemplate(&state, params.Template)
		if templateExists {
			tmpl, _ := providers.GetTemplate(&state, params.Template)
			job.Update(api.Job{
				RecordTemplate: tmpl,
			})
		} else {
			mgr.Logger.Warnf("template '%s' not found", params.Template)
		}
	}

	// Update Job with values
	job.Update(api.Job{
		RecordTemplate: api.RecordTemplate{
			Title:       params.Title,
			Description: params.Description,
			Project:     params.Project,
			Task:        params.Task,
		},
	})

	proverr := providers.UpdateJob(&state, job)
	if proverr != providers.ProviderOk {
		return JobResponse{}, mgr.MakeNewResponseError(BadRequest, proverr, "Unable to update Job '%s'", job.Name)
	}
	err = mgr.StateProvider.Save(state.Partition, state)
	if err != nil {
		return JobResponse{}, mgr.MakeNewResponseError(ProviderError, err, "Unable to save Job '%s'", job.Name)
	}

	mgr.Logger.Infof("Updated Job: %s", job.Name)
	return JobResponse{Success: true, Created: false, Job: job}, nil
}

func (mgr *TimerecServer) CompleteJob(ctx context.Context, params CompleteJobParams) (JobResponse, error) {
	state, err := mgr.StateProvider.Refresh(params.Owner)
	if err != nil {
		return JobResponse{}, mgr.MakeNewResponseError(ProviderError, err, "Unable to query Provider: %s", err.Error())
	}
	response, err := mgr.GetJob(
		ctx,
		params.SearchJobParams,
	)
	if err != nil {
		return JobResponse{}, err
	}
	if !response.Success {
		return JobResponse{}, mgr.MakeNewResponseError(BadRequest, err, "Job does not exist")
	}
	Job := response.Job
	err = Job.Validate()
	if err != nil {
		return JobResponse{}, mgr.MakeNewResponseError(ValidationError, err, "Job not valid: %s", err.Error())
	}

	for _, rec := range Job.ConvertToRecords() {
		_, err = mgr.TimeProvider.SaveRecord(rec)
		if err != nil {
			mgr.Logger.Errorw("unable to save Record", "error", err, "record", rec, "title", rec.Title)
			return JobResponse{}, mgr.MakeNewResponseError(ProviderError, err, "Unable to save Record '%s'", rec.Title)
		}
	}

	state, _ = mgr.StateProvider.Refresh(params.Owner) // Refresh State, because the time provider might have changed the state-file
	deleted, proverr := providers.DeleteJob(&state, Job)
	if proverr != providers.ProviderOk {
		return JobResponse{}, mgr.MakeNewResponseError(BadRequest, proverr, "Unable to delete Job")
	}
	err = mgr.StateProvider.Save(state.Partition, state)
	if err != nil {
		return JobResponse{}, mgr.MakeNewResponseError(ServerError, err, "Unable to save state")
	}

	mgr.Logger.Infof("Completed Job: %s", Job.Name)
	return JobResponse{Success: true, Created: false, Job: deleted}, nil
}
