package client

import (
	"context"
	"errors"
	"fmt"
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

func (c *ClientObject) exitIfError(err error, success bool, message string) {
	if err == nil {
		return
	}

	respErr := &server.ResponseError{}
	if errors.As(err, respErr) {
		c.Panic(10, respErr.Message, respErr.Cause)
	}
	if err != nil {
		c.Panic(11, "GenericError: ", err)
	}
	if !success {
		c.Panic(13, message, nil)
	}

}

func (c *ClientObject) EnsureUserExists(name string) {
	resp, err := c.embeddedServer.CreateUserIfMissing(
		context.TODO(),
		server.SearchUserParams{
			Name:     name,
			Inactive: false,
		},
	)
	c.exitIfError(err, resp.Success, "Unable to create User")

	if resp.Created {
		c.logger.Printf("Creating User '%s'...\n", name)
	}
}

func (c *ClientObject) StartActivity(activityName string, comment string, start_duration time.Duration, estimate_duration time.Duration) {
	c.EnsureUserExists("me")
	resp, err := c.embeddedServer.StartActivity(
		context.TODO(),
		server.StartActivityParams{
			UserName:         "me",
			ActivityName:     activityName,
			Comment:          comment,
			StartDuration:    start_duration,
			EstimateDuration: estimate_duration,
		},
	)
	c.exitIfError(err, resp.Success, "Unable to StartActivity")
	fmt.Println(FormatActivity(resp.Activity))
}

func (c *ClientObject) ExtendActivity(estimate_duration time.Duration, comment string, reset bool) {
	c.EnsureUserExists("me")
	resp, err := c.embeddedServer.ExtendActivity(
		context.TODO(),
		server.ExtendActivityParams{
			UserName:         "me",
			EstimateDuration: estimate_duration,
			Comment:          comment,
			ResetComment:     reset,
		},
	)
	c.exitIfError(err, resp.Success, "Unable to ExtendActivity")
	fmt.Println(FormatActivity(resp.Activity))
}

func (c *ClientObject) FinishActivity(taskName string, _activityName string, comment string, endDuration time.Duration) {
	c.EnsureUserExists("me")
	resp, err := c.embeddedServer.FinishActivity(
		context.TODO(),
		server.FinishActivityParams{
			UserName:     "me",
			JobName:      taskName,
			ActivityName: "",
			Comment:      comment,
			EndDuration:  endDuration,
		},
	)
	c.exitIfError(err, resp.Success, "Unable to FinishActivity")
}

func (c *ClientObject) ActivityInfo() {
	resp := server.ActivityResponse{}
	c.EnsureUserExists("me")
	resp, err := c.embeddedServer.GetActivity(
		context.TODO(),
		server.GetUserParams{
			UserName: "me",
		},
	)
	c.exitIfError(err, resp.Success, "Unable to GetActivity")
	fmt.Println(FormatActivity(resp.Activity))
}

func (c *ClientObject) EnsureJobkExists(name string) {
	resp, err := c.embeddedServer.CreateJobIfMissing(
		context.TODO(),
		server.SearchJobParams{
			Name:          name,
			StartedAfter:  -24 * time.Hour,
			StartedBefore: time.Duration(0),
		},
	)
	c.exitIfError(err, resp.Success, "Unable to create Job")

	if resp.Created {
		c.logger.Printf("Creating Job '%s'...\n", name)
	}
}

func (c *ClientObject) UpdateJob(name, template, title, description, project, task string) {
	c.EnsureJobkExists(name)

	resp, err := c.embeddedServer.UpdateJob(
		context.TODO(),
		server.UpdateJobParams{
			Name:        name,
			Template:    template,
			Title:       title,
			Description: description,
			Project:     project,
			Task:        task,
		},
	)
	c.exitIfError(err, resp.Success, "Unable to UpdateJob")
}

func (c *ClientObject) CompleteJob(name string) {
	resp, err := c.embeddedServer.CompleteJob(
		context.TODO(),
		server.CompleteJobParams{
			Status: server.JobStatusFinish,
			SearchJobParams: server.SearchJobParams{
				Name:          name,
				StartedAfter:  -24 * time.Hour,
				StartedBefore: time.Duration(0),
			},
		},
	)
	c.exitIfError(err, resp.Success, "Unable to CompleteJob")
}

func (c *ClientObject) Wait() {
	resp, err := c.embeddedServer.GetActivity(
		context.TODO(),
		server.GetUserParams{
			UserName: "me",
		},
	)
	c.exitIfError(err, resp.Success, "Unable to GetActivity")

	if err := resp.Activity.CheckNoActivityActive(); err == nil {
		return
	}
	time.Sleep(time.Until(resp.Activity.ActivityTimer))
}

func (c *ClientObject) ReconcileServer() {
	result := c.embeddedServer.ReconcileOnce(context.TODO())
	if result.Requeue {
		time.AfterFunc(result.RetryAfter, func() {
			c.embeddedServer.ReconcileOnce(context.TODO())
		})
	}
}
