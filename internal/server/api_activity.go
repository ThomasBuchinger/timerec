package server

import (
	"context"
	"fmt"
	"time"

	"github.com/thomasbuchinger/timerec/api"
	"github.com/thomasbuchinger/timerec/internal/server/providers"
)

type GetUserParams struct {
	UserName string
}

type StartActivityParams struct {
	UserName       string `path:"user"`
	ActivityName   string `json:"activity"`
	Comment        string `json:"comment,omitempty"`
	StartString    string `json:"start"`
	EstimateString string `json:"estimate"`

	StartDuration    time.Duration `json:"start_int"`
	EstimateDuration time.Duration `json:"estimate_int"`
}

func (param *StartActivityParams) MakeValid() error {
	var err error
	if param.StartDuration == time.Duration(0) && param.StartString != "" {
		param.StartDuration, err = time.ParseDuration(param.StartString)
		if err != nil {
			return err
		}
	}
	if param.EstimateDuration == time.Duration(0) && param.EstimateString != "" {
		param.EstimateDuration, err = time.ParseDuration(param.EstimateString)
		if err != nil {
			return err
		}
	}
	return nil
}

type ExtendActivityParams struct {
	UserName       string `path:"user"`
	EstimateString string `json:"estimate"`
	Comment        string `json:"comment,omitempty"`
	ResetComment   bool   `json:"reset_comment,omitempty" default:"false"`

	EstimateDuration time.Duration `json:"estimate_int,omitempty"`
}

func (param *ExtendActivityParams) MakeValid() error {
	var err error
	if param.EstimateDuration == time.Duration(0) && param.EstimateString != "" {
		param.EstimateDuration, err = time.ParseDuration(param.EstimateString)
		if err != nil {
			return err
		}
	}
	return nil
}

type FinishActivityParams struct {
	UserName     string `path:"user"`
	JobName      string `json:"job"`
	ActivityName string `json:"activity"`
	Comment      string `json:"comment,omitempty"`

	EndString   string        `json:"end"`
	EndDuration time.Duration `json:"end_int,omitempty"`
}

func (param *FinishActivityParams) MakeValid() error {
	var err error
	if param.EndDuration == time.Duration(0) && param.EndString != "" {
		param.EndDuration, err = time.ParseDuration(param.EndString)
		if err != nil {
			return err
		}
	}
	return nil
}

type ActivityResponse struct {
	Success  bool         `json:"success"`
	Activity api.Activity `json:"activity,omitempty"`
}

func (mgr *TimerecServer) GetActivity(ctx context.Context, params GetUserParams) (ActivityResponse, error) {
	state, err := mgr.StateProvider.Refresh(params.UserName)
	if err != nil {
		return ActivityResponse{}, mgr.MakeNewResponseError(ProviderError, err, "Unable to query Provider: %s", err.Error())
	}

	user, err := providers.GetUser(&state, api.User{Name: params.UserName})
	if err == providers.ProviderOk {
		return ActivityResponse{Success: true, Activity: user.Activity}, nil
	}

	if err == providers.ProviderNotFound {
		return ActivityResponse{}, mgr.MakeNewResponseError(BadRequest, err, "Cannot read User '%s'", params.UserName)
	}
	return ActivityResponse{}, mgr.MakeNewResponseError(BadRequest, err, "Unexpected Error: %s", params.UserName)
}

func (mgr *TimerecServer) StartActivity(ctx context.Context, params StartActivityParams) (ActivityResponse, error) {
	err := params.MakeValid()
	if err != nil {
		return ActivityResponse{}, mgr.MakeNewResponseError(ValidationError, err, err.Error())
	}
	state, proverr := mgr.StateProvider.Refresh(params.UserName)
	if err != nil {
		return ActivityResponse{}, mgr.MakeNewResponseError(BadRequest, proverr, "Unable to query Provider: %s", proverr.Error())
	}

	user, proverr := providers.GetUser(&state, api.User{Name: params.UserName})
	if proverr != providers.ProviderOk {
		return ActivityResponse{}, mgr.MakeNewResponseError(BadRequest, proverr, "Cannot read User '%s'", params.UserName)
	}

	err = user.Activity.CheckNoActivityActive()
	if err != nil {
		mgr.Logger.Debugf("Cannot start new Activity: %v", err)
		return ActivityResponse{}, mgr.MakeNewResponseError(BadRequest, err, "Finish any active Jobs, before starting a new one")
	}

	mgr.Logger.Debugf("Setting active Activity to '%s'...", params.ActivityName)
	user.SetActivity(
		params.ActivityName,
		params.Comment,
		time.Now().Add(params.StartDuration).Round(user.Settings.RoundTo),
		time.Now().Add(params.EstimateDuration).Round(user.Settings.RoundTo),
	)
	proverr = providers.UpdateUser(&state, user)
	if proverr != providers.ProviderOk {
		return ActivityResponse{}, mgr.MakeNewResponseError(BadRequest, proverr, "Update failed: %v", proverr)
	}
	err = mgr.StateProvider.Save(state.Partition, state)
	if err != nil {
		return ActivityResponse{}, mgr.MakeNewResponseError(ProviderError, err, "Cannot save User: %v", err)
	}
	mgr.Logger.Infof("Start working on: %s", params.ActivityName)
	return ActivityResponse{Success: true, Activity: user.Activity}, nil
}

func (mgr *TimerecServer) ExtendActivity(ctx context.Context, params ExtendActivityParams) (ActivityResponse, error) {
	err := params.MakeValid()
	if err != nil {
		return ActivityResponse{}, mgr.MakeNewResponseError(ValidationError, err, "Invalid Request: %s", err.Error())
	}
	state, err := mgr.StateProvider.Refresh(params.UserName)
	if err != nil {
		return ActivityResponse{}, mgr.MakeNewResponseError(ProviderError, err, "Unable to query Provider: %s", err.Error())
	}

	// Get Activity
	user, proverr := providers.GetUser(&state, api.User{Name: params.UserName})
	if proverr != providers.ProviderOk {
		mgr.Logger.Error(proverr)
		return ActivityResponse{}, mgr.MakeNewResponseError(BadRequest, proverr, "Cannot read User '%s'", params.UserName)
	}
	err = user.Activity.CheckActivityActive()
	if err != nil {
		mgr.Logger.Debugf("no active Activity: %v", err)
		return ActivityResponse{}, mgr.MakeNewResponseError(BadRequest, err, "Cannot extent timer: no active Job")
	}

	// Update Activity
	if params.ResetComment {
		user.Activity.ActivityComment = params.Comment
	} else {
		user.Activity.AddComment(params.Comment)
	}
	user.SetActivity(
		user.Activity.ActivityName,
		user.Activity.ActivityComment,
		user.Activity.ActivityStart,
		time.Now().Add(params.EstimateDuration).Round(user.Settings.RoundTo),
	)
	proverr = providers.UpdateUser(&state, user)
	if proverr != providers.ProviderOk {
		return ActivityResponse{}, mgr.MakeNewResponseError(BadRequest, proverr, "Failed to update User '%s'", params.UserName)
	}

	err = mgr.StateProvider.Save(state.Partition, state)
	if err != nil {
		return ActivityResponse{}, mgr.MakeNewResponseError(ProviderError, err, "Unable to save User '%s'", params.UserName)
	}
	mgr.Logger.Infof("Extend Activity %s by: %s", user.Activity.ActivityName, params.EstimateDuration)
	return ActivityResponse{Success: true, Activity: user.Activity}, nil
}

func (mgr *TimerecServer) FinishActivity(ctx context.Context, params FinishActivityParams) (JobResponse, error) {
	err := params.MakeValid()
	if err != nil {
		return JobResponse{}, mgr.MakeNewResponseError(ValidationError, err, err.Error())
	}
	state, err := mgr.StateProvider.Refresh(params.UserName)
	if err != nil {
		return JobResponse{}, mgr.MakeNewResponseError(ProviderError, err, "Unable to query Provider: %s", err.Error())
	}

	response, err := mgr.GetJob(
		ctx,
		SearchJobParams{Name: params.JobName, Owner: params.UserName, StartedAfter: -24 * time.Hour, StartedBefore: 0},
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

	user, proverr := providers.GetUser(&state, api.User{Name: params.UserName})
	if proverr != providers.ProviderOk {
		return JobResponse{}, mgr.MakeNewResponseError(BadRequest, proverr, "Cannot read User '%s'", params.UserName)
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

	// Update Job & User
	user.Activity.AddComment(params.Comment)
	job.Update(api.Job{
		Name: job.Name,
		Activities: []api.TimeEntry{
			{
				Start:   user.Activity.ActivityStart,
				End:     time.Now().Add(params.EndDuration).Round(user.Settings.RoundTo),
				Comment: user.Activity.ActivityComment,
			},
		},
	})
	proverr = providers.UpdateJob(&state, job)
	if proverr != providers.ProviderOk {
		return JobResponse{}, mgr.MakeNewResponseError(BadRequest, proverr, "Unable to update Job '%s'", job.Name)
	}

	user.ClearActivity()
	proverr = providers.UpdateUser(&state, user)
	if proverr != providers.ProviderOk {
		mgr.Logger.Error(proverr)
		return JobResponse{}, mgr.MakeNewResponseError(BadRequest, proverr, "Unable to update user '%s'", user.Name)
	}
	err = mgr.StateProvider.Save(state.Partition, state)
	if err != nil {
		return JobResponse{}, mgr.MakeNewResponseError(ProviderError, err, "Unable save user '%s'", user.Name)
	}

	mgr.Logger.Infof("Finished Activity on Job: %s", user.Name)
	return JobResponse{Success: true, Job: job}, nil
}
