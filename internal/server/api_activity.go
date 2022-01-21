package server

import (
	"context"
	"fmt"
	"time"

	"github.com/thomasbuchinger/timerec/api"
)

type GetUserParams struct {
	UserName string
}

type StartActivityParams struct {
	UserName       string `path:"user"`
	ActivityName   string `json:"activity"`
	Comment        string `json:"comment,omitempty"`
	StartString    string `json:"start"`
	ExtimateString string `json:"estimate"`

	StartDuration    time.Duration `json:"start_int"`
	EstimateDuration time.Duration `json:"estimate_int"`
}
type ExtendActivityParams struct {
	UserName     string        `path:"user"`
	Estimate     time.Duration `json:"estimate"`
	Comment      string        `json:"comment,omitempty"`
	ResetComment bool          `json:"reset_comment,omitempty" default:"false"`
}
type FinishActivityParams struct {
	UserName     string        `path:"user"`
	JobName      string        `json:"job"`
	ActivityName string        `json:"activity"`
	Comment      string        `json:"comment,omitempty"`
	EndDuration  time.Duration `json:"end"`
}

type ActivityResponse struct {
	Success  bool         `json:"success"`
	Activity api.Activity `json:"activity,omitempty"`
}

func (mgr *TimerecServer) GetActivity(ctx context.Context, params GetUserParams) (ActivityResponse, error) {
	user, err := mgr.StateProvider.GetUser()
	if err != nil {
		err := ResponseError{
			Type:    ProviderError,
			Message: fmt.Sprintf("Cannot read User '%s'", params.UserName),
			Cause:   err,
		}
		mgr.Logger.Error(err)
		return ActivityResponse{}, err
	}
	return ActivityResponse{Success: true, Activity: user.Activity}, nil
}

func (mgr *TimerecServer) StartActivity(ctx context.Context, params StartActivityParams) (ActivityResponse, error) {
	user, err := mgr.StateProvider.GetUser()
	if err != nil {
		resp := ResponseError{
			Type:    ProviderError,
			Message: fmt.Sprintf("Cannot read User '%s'", params.UserName),
			Cause:   err,
		}
		mgr.Logger.Error(err)
		return ActivityResponse{}, resp
	}
	err = user.Activity.CheckNoActivityActive()
	if err != nil {
		resp := ResponseError{
			Type:    BadRequest,
			Message: "Finish any active Jobs, before starting a new one",
			Cause:   err,
		}
		mgr.Logger.Debugf("Cannot start new Activity: %v", err)
		return ActivityResponse{}, resp
	}

	mgr.Logger.Debugf("Setting active Activity to '%s'...", params.ActivityName)
	user.SetActivity(
		params.ActivityName,
		params.Comment,
		time.Now().Add(params.StartDuration).Round(user.GetRoundTo()),
		time.Now().Add(params.EstimateDuration).Round(user.GetRoundTo()),
	)
	saved, err := mgr.StateProvider.UpdateUser(user)
	if err != nil {
		resp := ResponseError{
			Type:    ProviderError,
			Message: "Error updating User",
			Cause:   err,
		}
		mgr.Logger.Error(err)
		return ActivityResponse{}, resp
	}
	mgr.Logger.Infof("Start working on: %s", params.ActivityName)
	return ActivityResponse{Success: true, Activity: saved.Activity}, nil
}

func (mgr *TimerecServer) ExtendActivity(ctx context.Context, params ExtendActivityParams) (ActivityResponse, error) {
	user, err := mgr.StateProvider.GetUser()
	if err != nil {
		mgr.Logger.Error(err)
		return ActivityResponse{}, ResponseError{
			Type:    ProviderError,
			Message: fmt.Sprintf("Cannot read User '%s'", params.UserName),
			Cause:   err,
		}
	}
	err = user.Activity.CheckActivityActive()
	if err != nil {
		mgr.Logger.Debugf("no active Activity: %v", err)
		return ActivityResponse{}, ResponseError{
			Type:    BadRequest,
			Message: "Cannot extent timer: no active Job",
			Cause:   err,
		}
	}
	if params.ResetComment {
		user.Activity.ActivityComment = params.Comment
	} else {
		user.Activity.AddComment(params.Comment)
	}
	user.SetActivity(
		user.Activity.ActivityName,
		user.Activity.ActivityComment,
		user.Activity.ActivityStart,
		time.Now().Add(params.Estimate).Round(user.Settings.RoundTo),
	)
	saved, err := mgr.StateProvider.UpdateUser(user)
	if err != nil {
		mgr.Logger.Error(err)
		return ActivityResponse{}, ResponseError{
			Type:    ProviderError,
			Message: fmt.Sprintf("Unable to save User '%s'", params.UserName),
			Cause:   err,
		}
	}

	mgr.Logger.Infof("Extend Activity %s by: %s", user.Activity.ActivityName, params.Estimate)
	return ActivityResponse{Success: true, Activity: saved.Activity}, nil
}

func (mgr *TimerecServer) FinishActivity(ctx context.Context, params FinishActivityParams) (JobResponse, error) {
	response, err := mgr.GetJob(
		ctx,
		SearchJobParams{Name: params.JobName, StartedAfter: -24 * time.Hour, StartedBefore: 0},
	)
	if err != nil {
		return JobResponse{}, err
	}
	job, job_is_missing := response.Job, false
	if err == nil && !response.Success {
		// We should exit here, but we might be able to ignore it, if there is no activity to finish anyway
		// A common error, when FinishActivity is called multiple times
		job_is_missing = true
	}

	user, err := mgr.StateProvider.GetUser()
	if err != nil {
		mgr.Logger.Error(err)
		return JobResponse{}, ResponseError{
			Type:    ProviderError,
			Message: fmt.Sprintf("Cannot read User '%s'", params.UserName),
			Cause:   err,
		}
	}
	err = user.Activity.CheckActivityActive()
	if err != nil {
		mgr.Logger.Info("Called FinishActivity, but no active actifiy found. Nothing to do \n")
		return JobResponse{Success: true, Job: job}, nil
	}

	if job_is_missing {
		return JobResponse{}, ResponseError{
			Type:    BadRequest,
			Message: fmt.Sprintf("Job '%s' found", params.JobName),
		}
	}

	user.Activity.AddComment(params.Comment)
	job.Update(api.Job{
		Name: job.Name,
		Activities: []api.TimeEntry{
			{
				Start:   user.Activity.ActivityStart,
				End:     time.Now().Add(params.EndDuration).Round(user.GetRoundTo()),
				Comment: user.Activity.ActivityComment,
			},
		},
	})
	saved, err := mgr.StateProvider.UpdateJob(job)
	if err != nil {
		mgr.Logger.Error(err)
		return JobResponse{}, ResponseError{
			Type:    ProviderError,
			Message: fmt.Sprintf("Unable to update Job '%s'", job.Name),
			Cause:   err,
		}
	}
	user.ClearActivity()
	_, err = mgr.StateProvider.UpdateUser(user)
	if err != nil {
		mgr.Logger.Error(err)
		return JobResponse{}, ResponseError{
			Type:    ProviderError,
			Message: "Unable to update User",
			Cause:   err,
		}
	}

	mgr.Logger.Infof("Finished Activity on Job: %s", saved.Name)
	return JobResponse{Success: true, Job: saved}, nil
}
