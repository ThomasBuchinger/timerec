package client

import (
	"log"
	"time"

	"github.com/thomasbuchinger/timerec/internal/server"
)

type ClientObject struct {
	logger *log.Logger
	// restclient     RestClient
	embeddedServer server.TimerecServer
}

func NewClient() ClientObject {
	client := ClientObject{}
	client.logger = log.Default()
	// client.restclient = RestClient{}

	client.embeddedServer = server.NewServer()
	return client
}

func (c *ClientObject) StartActivity(activityName string, comment string, start_duration time.Duration, estimate_duration time.Duration) {
	resp := c.embeddedServer.StartActivity("me", server.StartActivityParams{
		ActivityName:     activityName,
		Comment:          comment,
		StartDuration:    start_duration,
		EstimateDuration: estimate_duration,
	})
	if !resp.Success {
		c.Panic(10, "Unable to start Activity", resp.Err)
	}

	PrintActivity(resp.Activity)
}

func (c *ClientObject) ExtendActivity(estimate_duration time.Duration, comment string, reset bool) {
	resp := c.embeddedServer.ExtendActivity("me", server.ExtendActivityParams{
		Estimate:     estimate_duration,
		Comment:      comment,
		ResetComment: reset,
	})
	if !resp.Success {
		c.Panic(11, "unable to extend Activity", resp.Err)
	}
	PrintActivity(resp.Activity)
}

func (c *ClientObject) FinishActivity(taskName string, _activityName string, comment string, endDuration time.Duration) {
	resp := c.embeddedServer.FinishActivity("me", server.FinishActivityParams{
		WorkItemName: taskName,
		ActivityName: "",
		Comment:      comment,
		EndDuration:  endDuration,
	})

	if !resp.Success {
		c.Panic(12, "unable to finish activity", resp.Err)
	}
}

func (c *ClientObject) ActivityInfo() {
	resp := c.embeddedServer.GetActivity("me")
	if !resp.Success {
		c.Panic(13, "unable to get activity", resp.Err)
	}

	PrintActivity(resp.Activity)
}

func (c *ClientObject) EnsureWorkItemkExists(name string) {
	resp := c.embeddedServer.CreateWorkItemIfMissing(server.GetWorkItemParams{
		Name:          name,
		StartedAfter:  -24 * time.Hour,
		StartedBefore: time.Duration(0),
	})
	if !resp.Success {
		c.Panic(14, "unable to create WorkItem", resp.Err)
	}

	if resp.Created {
		c.logger.Printf("Creating WorkItem '%s'...\n", name)
	}
}

func (c *ClientObject) UpdateWorkItem(name, template, title, description, project, task string) {
	c.embeddedServer.CreateWorkItemIfMissing(server.GetWorkItemParams{
		Name:          name,
		StartedAfter:  -24 * time.Hour,
		StartedBefore: time.Duration(0),
	})

	resp := c.embeddedServer.UpdateWorkItem(server.UpdateWorkItemParams{
		Name:        name,
		Template:    template,
		Title:       title,
		Description: description,
		Project:     project,
		Task:        task,
	})

	if !resp.Success {
		c.Panic(15, "Unable to update WorkItem", resp.Err)
	}
}

func (c *ClientObject) CompleteWorkItem(name string) {
	resp := c.embeddedServer.CompleteWorkItem(server.CompleteWorkItemParams{
		Status: server.WorkItemStatusFinish,
		GetWorkItemParams: server.GetWorkItemParams{
			Name:          name,
			StartedAfter:  -24 * time.Hour,
			StartedBefore: time.Duration(0),
		},
	})

	if !resp.Success {
		c.Panic(16, "unable to complete WorkItem", resp.Err)
	}
}

func (c *ClientObject) Wait() {
	resp := c.embeddedServer.GetActivity("me")
	if !resp.Success {
		c.Panic(17, "unable to get Activity", resp.Err)
	}

	if err := resp.Activity.CheckNoActivityActive(); err == nil {
		return
	}
	time.Sleep(time.Until(resp.Activity.ActivityTimer))
}

func (c *ClientObject) ReconcileServer() {
	c.embeddedServer.Reconcile()
}
