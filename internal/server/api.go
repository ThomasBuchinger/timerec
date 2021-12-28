package server

import (
	"time"

	"github.com/thomasbuchinger/timerec/api"
)

type StartActivityParams struct {
	ActivityName     string
	Comment          string
	StartDuration    time.Duration
	EstimateDuration time.Duration
}
type ExtendActivityParams struct {
	Estimate     time.Duration
	Comment      string
	ResetComment bool
}
type FinishActivityParams struct {
	WorkItemName string
	ActivityName string
	Comment      string
	EndDuration  time.Duration
}

type ActivityResponse struct {
	Success  bool
	Err      error
	Activity api.Activity
}

type GetWorkItemParams struct {
	Name          string
	StartedAfter  time.Duration
	StartedBefore time.Duration
}

type UpdateWorkItemParams struct {
	Name        string
	Template    string
	Title       string
	Description string
	Project     string
	Task        string
}

type CompleteWorkItemParams struct {
	GetWorkItemParams
	Status WorkItemStatus
}

type WorkItemStatus string

const (
	WorkItemStatusCancel WorkItemStatus = "canceled"
	WorkItemStatusFinish WorkItemStatus = "finished"
)

type WorkItemResponse struct {
	Success bool
	Created bool
	Err     error
	Item    api.WorkItem
}

func (mgr *TimerecServer) GetActivity(userName string) ActivityResponse {
	user, err := mgr.StateProvider.GetUser()
	if err != nil {
		mgr.loggerv2.Errorw("cannot read User", "name", userName, "error", err)
		return ActivityResponse{Success: false, Err: err}
	}
	return ActivityResponse{Success: true, Activity: user.Activity}
}

func (mgr *TimerecServer) StartActivity(userName string, params StartActivityParams) ActivityResponse {
	user, err := mgr.StateProvider.GetUser()
	if err != nil {
		mgr.loggerv2.Errorw("cannot read User", "name", userName, "error", err)
		return ActivityResponse{Success: false, Err: err}
	}
	err2 := user.Activity.CheckNoActivityActive()
	if err2 != nil {
		mgr.loggerv2.Errorw("finish any active WorkItems, before starting a new one", "error", err2)
		return ActivityResponse{Success: false, Err: err2}
	}

	mgr.loggerv2.Debugf("Setting active WorkItem to '%s'...", params.ActivityName)
	user.SetActivity(
		params.ActivityName,
		params.Comment,
		time.Now().Add(params.StartDuration).Round(user.GetRoundTo()),
		time.Now().Add(params.EstimateDuration).Round(user.GetRoundTo()),
	)
	saved, err3 := mgr.StateProvider.UpdateUser(user)
	if err3 != nil {
		mgr.loggerv2.Errorw("Unable to save User", "name", userName, "error", err3)
	}
	mgr.loggerv2.Infof("Start Activity on: %s", params.ActivityName)
	return ActivityResponse{Success: true, Activity: saved.Activity}
}

func (mgr *TimerecServer) ExtendActivity(userName string, params ExtendActivityParams) ActivityResponse {
	user, err := mgr.StateProvider.GetUser()
	if err != nil {
		mgr.loggerv2.Errorw("cannot read User", "name", userName, "error", err)
		return ActivityResponse{Success: false, Err: err}
	}
	err2 := user.Activity.CheckActivityActive()
	if err2 != nil {
		mgr.loggerv2.Errorw("no active WorkItem", "error", err2)
		return ActivityResponse{Success: false, Err: err2}
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
	saved, err3 := mgr.StateProvider.UpdateUser(user)
	if err3 != nil {
		mgr.loggerv2.Errorw("Unable to update WorkItem", "error", err3)
		return ActivityResponse{Success: false, Err: err3}
	}

	mgr.loggerv2.Infof("Extend Activity %s by: %s", user.Activity.ActivityName, params.Estimate)
	return ActivityResponse{Success: true, Activity: saved.Activity}
}

func (mgr *TimerecServer) FinishActivity(userName string, params FinishActivityParams) WorkItemResponse {
	itemResp := mgr.GetWorkItem(GetWorkItemParams{
		Name:          params.WorkItemName,
		StartedAfter:  -24 * time.Hour,
		StartedBefore: 0,
	})
	if !itemResp.Success {
		return itemResp
	}
	workItem := itemResp.Item
	user, err2 := mgr.StateProvider.GetUser()
	if err2 != nil {
		mgr.loggerv2.Errorw("unable to read User", "error", err2)
		return WorkItemResponse{Success: false, Err: err2}
	}
	err3 := user.Activity.CheckActivityActive()
	if err3 != nil {
		mgr.loggerv2.Warn("Called FinishActivity, but no active actifiy found. Nothing to do \n")
		return WorkItemResponse{Success: true, Item: itemResp.Item}
	}

	user.Activity.AddComment(params.Comment)
	workItem.Update(api.WorkItem{
		Name: workItem.Name,
		Activities: []api.TimeEntry{
			{
				Start:   user.Activity.ActivityStart,
				End:     time.Now().Add(params.EndDuration).Round(user.GetRoundTo()),
				Comment: user.Activity.ActivityComment,
			},
		},
	})
	saved, err4 := mgr.StateProvider.UpdateWorkItem(workItem)
	if err4 != nil {
		mgr.loggerv2.Errorw("unable to update WorkItem", "name", workItem.Name, "error", err4)
		return WorkItemResponse{Success: false, Err: err4}
	}
	user.ClearActivity()
	_, err5 := mgr.StateProvider.UpdateUser(user)
	if err5 != nil {
		mgr.loggerv2.Errorw("unable to update User", "error", err5)
		return WorkItemResponse{Success: false, Err: err5}
	}

	mgr.loggerv2.Infof("Finished Activity on WorkItem: %s", saved.Name)
	return WorkItemResponse{Success: true, Item: saved}
}

func (mgr *TimerecServer) GetWorkItem(params GetWorkItemParams) WorkItemResponse {
	item, err := mgr.StateProvider.GetWorkItem(api.WorkItem{
		Name: params.Name,
	})
	if err != nil {
		mgr.loggerv2.Errorw("Unable to find WorkItems", "name", item.Name, "error", err)
		return WorkItemResponse{Success: false, Err: err}
	}
	return WorkItemResponse{Success: true, Item: item}
}

func (mgr *TimerecServer) CreateWorkItemIfMissing(params GetWorkItemParams) WorkItemResponse {
	getResp := mgr.GetWorkItem(params)
	if getResp.Success {
		return getResp
	}
	new, err := mgr.StateProvider.CreateWorkItem(api.NewWorkItem(params.Name))
	if err != nil {
		mgr.loggerv2.Errorw("unable to create WorkItem", "error", err)
	}
	mgr.loggerv2.Infof("created WorkItem: %s", new.Name)
	return WorkItemResponse{Success: true, Created: true, Item: new}
}

func (mgr *TimerecServer) UpdateWorkItem(params UpdateWorkItemParams) WorkItemResponse {
	// Check if WorkItem exists
	itemResp := mgr.GetWorkItem(GetWorkItemParams{
		Name:          params.Name,
		StartedAfter:  -24 * time.Hour,
		StartedBefore: time.Duration(0),
	})
	if !itemResp.Success {
		mgr.loggerv2.Warnf("WorkItem with name '%s' does not exist", params.Name)
		return itemResp
	}

	// Update WorkItem according to Template
	workItem := itemResp.Item
	if params.Template != "" {
		templateExists, _ := mgr.TemplateProvider.HasTemplate(params.Template)
		if templateExists {
			tmpl, _ := mgr.TemplateProvider.GetTemplate(params.Template)
			workItem.Update(api.WorkItem{
				RecordTemplate: tmpl,
			})
		} else {
			mgr.loggerv2.Warnf("template '%s' not found", params.Template)
		}
	}

	// Update WorkItem with values
	workItem.Update(api.WorkItem{
		RecordTemplate: api.RecordTemplate{
			Title:       params.Title,
			Description: params.Description,
			Project:     params.Project,
			Task:        params.Task,
		},
	})

	saved, err := mgr.StateProvider.UpdateWorkItem(workItem)
	if err != nil {
		mgr.loggerv2.Errorw("unable to save WorkItem", "error", err)
		return WorkItemResponse{Success: false, Err: err}
	}
	mgr.loggerv2.Infof("Updated WorkItem: %s", saved.Name)
	return WorkItemResponse{Success: true, Item: saved}
}

func (mgr *TimerecServer) CompleteWorkItem(params CompleteWorkItemParams) WorkItemResponse {
	itemResp := mgr.GetWorkItem(params.GetWorkItemParams)
	if !itemResp.Success {
		mgr.loggerv2.Errorw("unable to find WorkItem", "error", itemResp.Err)
		return itemResp
	}

	workItem := itemResp.Item
	err := workItem.Validate()
	if err != nil {
		mgr.loggerv2.Errorf("WorkItem not valid: %s", err.Error())
		return WorkItemResponse{Success: false, Err: err}
	}

	for _, rec := range workItem.ConvertToRecords() {
		_, err = mgr.TimeProvider.SaveRecord(rec)
		if err != nil {
			mgr.loggerv2.Errorw("unable to save Record", "error", err, "record", rec, "title", rec.Title)
			return WorkItemResponse{Success: false, Err: err}
		}
	}

	mgr.StateProvider.DeleteWorkItem(workItem)
	mgr.loggerv2.Infof("Completed WorkItem: %s", workItem.Name)
	return WorkItemResponse{Success: true, Item: workItem}
}
