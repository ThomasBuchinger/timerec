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
	err := params.MakeValid()
	if err != nil {
		return ActivityResponse{}, ResponseError{
			Type:    ValidationError,
			Message: err.Error(),
			Cause:   err,
		}
	}

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
	err := params.MakeValid()
	if err != nil {
		return ActivityResponse{}, ResponseError{
			Type:    ValidationError,
			Message: err.Error(),
			Cause:   err,
		}
	}

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
		time.Now().Add(params.EstimateDuration).Round(user.Settings.RoundTo),
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

	mgr.Logger.Infof("Extend Activity %s by: %s", user.Activity.ActivityName, params.EstimateDuration)
	return ActivityResponse{Success: true, Activity: saved.Activity}, nil
}

func (mgr *TimerecServer) FinishActivity(ctx context.Context, params FinishActivityParams) (JobResponse, error) {
	err := params.MakeValid()
	if err != nil {
		return JobResponse{}, ResponseError{
			Type:    ValidationError,
			Message: err.Error(),
			Cause:   err,
		}
	}

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
