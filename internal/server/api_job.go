package server

import (
	"context"
	"fmt"
	"time"

	"github.com/thomasbuchinger/timerec/api"
)

type SearchJobParams struct {
	Name          string        `json:"name"`
	StartedAfter  time.Duration `json:"start_after,omitempty" default:"-24h"`
	StartedBefore time.Duration `json:"start_before,omitempty" default:"0s"`
}

type UpdateJobParams struct {
	Name        string `json:"name,omitempty"`
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
	item, err := mgr.StateProvider.GetJob(api.Job{
		Name: params.Name,
	})
	if err == nil {
		return JobResponse{Success: true, Created: false, Job: item}, nil
	}
	if err != nil && err.Error() == "NOT_FOUND" {
		return JobResponse{Success: false, Created: false, Job: api.Job{}}, nil
	}

	mgr.Logger.Error(err)
	return JobResponse{}, ResponseError{
		Type:    ProviderError,
		Message: fmt.Sprintf("Unable to find Job '%s'", item.Name),
		Cause:   err,
	}
}

func (mgr *TimerecServer) CreateJobIfMissing(ctx context.Context, params SearchJobParams) (JobResponse, error) {
	response, err := mgr.GetJob(ctx, params)
	if err != nil {
		mgr.Logger.Error(err)
		return JobResponse{}, err
	}
	if response.Success {
		return response, nil
	}

	new, err := mgr.StateProvider.CreateJob(api.NewJob(params.Name))
	if err != nil {
		mgr.Logger.Error(err)
		return JobResponse{}, ResponseError{
			Type:    ProviderError,
			Message: fmt.Sprintf("Unable to create Job '%s'", params.Name),
			Cause:   err,
		}
	}
	mgr.Logger.Infof("Created Job: %s", new.Name)
	return JobResponse{Success: true, Created: true, Job: new}, nil
}

func (mgr *TimerecServer) UpdateJob(ctx context.Context, params UpdateJobParams) (JobResponse, error) {
	// Check if Job exists
	response, err := mgr.GetJob(
		ctx,
		SearchJobParams{Name: params.Name, StartedAfter: -24 * time.Hour, StartedBefore: time.Duration(0)},
	)
	if err != nil {
		return JobResponse{}, err
	}
	if !response.Success {
		mgr.Logger.Warnf("Job with name '%s' does not exist", params.Name)
		return JobResponse{}, ResponseError{
			Type:    BadRequest,
			Message: "Job does not exist",
			Cause:   err,
		}
	}

	// Update job according to Template
	job := response.Job
	if params.Template != "" {
		templateExists, _ := mgr.TemplateProvider.HasTemplate(params.Template)
		if templateExists {
			tmpl, _ := mgr.TemplateProvider.GetTemplate(params.Template)
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

	saved, err := mgr.StateProvider.UpdateJob(job)
	if err != nil {
		mgr.Logger.Error(err)

		return JobResponse{}, ResponseError{
			Type:    ProviderError,
			Message: "Unable to save Job",
			Cause:   err,
		}
	}
	mgr.Logger.Infof("Updated Job: %s", saved.Name)
	return JobResponse{Success: true, Created: false, Job: saved}, nil
}

func (mgr *TimerecServer) CompleteJob(ctx context.Context, params CompleteJobParams) (JobResponse, error) {
	response, err := mgr.GetJob(
		ctx,
		params.SearchJobParams,
	)
	if err != nil {
		return JobResponse{}, err
	}
	if !response.Success {
		mgr.Logger.Error("Unable to find Job")
		return JobResponse{}, ResponseError{
			Type:    BadRequest,
			Message: "Job does not exist",
			Cause:   err,
		}
	}
	Job := response.Job
	err = Job.Validate()
	if err != nil {
		mgr.Logger.Errorf("Job not valid: %s", err.Error())
		return JobResponse{}, ResponseError{
			Type:    ValidationError,
			Message: err.Error(),
			Cause:   err,
		}
	}

	for _, rec := range Job.ConvertToRecords() {
		_, err = mgr.TimeProvider.SaveRecord(rec)
		if err != nil {
			mgr.Logger.Errorw("unable to save Record", "error", err, "record", rec, "title", rec.Title)
			return JobResponse{}, ResponseError{
				Type:    ProviderError,
				Message: fmt.Sprintf("Unable to save Record '%s'", rec.Title),
				Cause:   err,
			}
		}
	}

	deleted, err := mgr.StateProvider.DeleteJob(Job)
	if err != nil {
		mgr.Logger.Error(err)
		return JobResponse{}, ResponseError{
			Type:    ProviderError,
			Message: "Unable to delete Job",
			Cause:   err,
		}
	}

	mgr.Logger.Infof("Completed Job: %s", Job.Name)
	return JobResponse{Success: true, Created: false, Job: deleted}, nil
}
