package client

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/viper"
	"github.com/thomasbuchinger/timerec/api"
)

type ClientObject struct {
	logger     *log.Logger
	restclient RestClient
}

func NewClient() ClientObject {
	client := ClientObject{}
	client.logger = log.Default()
	client.restclient = RestClient{}
	return client
}

func (c *ClientObject) StartActivity(activityName string, comment string, start_duration time.Duration, estimate_duration time.Duration) {
	profile, err := c.restclient.GetActivity()
	if err != nil {
		c.Panic(11, "unable to fetch profile", err)
	}
	err2 := CheckNoActivityActive(profile)
	if err2 != nil {
		c.Panic(12, "finish any active tasks, before starting a new one", fmt.Errorf("active task: %s", err2.Error()))
	}

	c.logger.Printf("Setting active task to '%s'...\n", activityName)
	profile, err = c.restclient.SetActivity(activityName, comment, time.Now().Add(start_duration), time.Now().Add(estimate_duration))
	if err != nil {
		c.Panic(13, "unable to set active task", err)
	}

	PrintActivity(profile)
}

func (c *ClientObject) ExtendActivity(estimate_duration time.Duration, comment string, reset bool) {
	profile, err := c.restclient.GetActivity()
	if err != nil {
		c.Panic(11, "unable to fetch profile", err)
	}
	err2 := CheckActivityActive(profile)
	if err2 != nil {
		c.Panic(14, "No active task", fmt.Errorf("no active Task"))
	}
	if comment != "" {
		profile.ActivityComment = profile.ActivityComment + "\n" + comment
	}
	profile, err = c.restclient.SetActivity(profile.ActivityName, profile.ActivityComment, profile.ActivityStart, time.Now().Add(estimate_duration))
	if err != nil {
		c.Panic(15, "unable to update active task", err)
	}

	PrintActivity(profile)
}

func (c *ClientObject) FinishActivity(taskName string, _activityName string, comment string, endDuration time.Duration) {
	_, ok, err := c.restclient.FindTaskByName(taskName)
	if err != nil {
		c.logger.Fatal(err)
		return
	} else if !ok {
		c.logger.Fatalf("Unable to find Task '%s' name", taskName)
		return
	}

	profile, err := c.restclient.GetActivity()
	if err != nil {
		c.logger.Fatal(err)
		return
	}
	err2 := CheckActivityActive(profile)
	if err2 != nil {
		c.logger.Printf("Called FinishActivity, but no active activity found. Nothing to do\n")
		return
	}

	if comment != "" {
		profile.ActivityComment = profile.ActivityComment + "\n" + comment
	}
	_, err = c.restclient.AddActivityToTask(taskName, profile.ActivityComment, profile.ActivityStart, toTimestamp(endDuration))
	if err != nil {
		c.logger.Fatalf("Error adding current activity to task '%s': %s", taskName, err.Error())
	}
	err = c.restclient.ClearActivity()
	if err != nil {
		fmt.Printf("Cannot reset Activity: %s", err.Error())
		return
	}
}

func (c *ClientObject) ActivityInfo() {
	profile, err := c.restclient.GetActivity()
	if err != nil {
		c.logger.Fatal(err)
		return
	}

	PrintActivity(profile)
}

func (c *ClientObject) EnsureTaskExists(name string) {
	_, ok, err := c.restclient.FindTaskByName(name)
	if err != nil {
		c.Panic(10, "fatal error creating querying tasks", err)
	}
	if ok {
		return
	}

	c.logger.Printf("Creating Task %s\n...", name)
	c.restclient.NewTask(name)
}

func (c *ClientObject) UpdateTask(name, template, title, description, project, task string) {
	existing, ok, err := c.restclient.FindTaskByName(name)
	if err != nil {
		c.Panic(17, "unable to fetch task", err)
	}
	if !ok {
		c.Panic(18, "task does not exist", nil)
	}

	var templates []api.RecordTemplate
	err = viper.UnmarshalKey("templates", &templates)
	// templates, err := c.restclient.ListTemplates()
	if err != nil {
		c.Panic(19, "unable to fetch templates", err)
	}

	_, err = c.restclient.UpdateTask(CombineTask(existing, templates, template, title, description, project, task))
	if err != nil {
		c.Panic(20, "error updating task", err)
	}
}

func (c *ClientObject) CompleteTask(name string) {
	task, ok, err := c.restclient.FindTaskByName(name)
	if err != nil {
		c.logger.Fatal(err)
		return
	} else if !ok {
		c.logger.Fatalf("Unable to find Task '%s' name", name)
		return
	}

	err = task.Validate()
	if err != nil {
		c.logger.Fatalf("task '%s' is invalid: %s", name, err.Error())
		return
	}

	records := task.ConvertToRecords()
	for _, record := range records {
		_, err = c.restclient.SaveRecord(record)
		if err != nil {
			c.logger.Fatalf("Error saving record to API", err)
			return
		}
	}

	c.restclient.Deleteask(task)
	c.logger.Println("Completed Task " + name)
}

func (c *ClientObject) Wait() {
	profile, err := c.restclient.GetActivity()
	if err != nil {
		c.logger.Fatal(err)
		return
	}

	if err := CheckActivityActive(profile); err != nil {
		return
	}
	time.Sleep(time.Until(profile.ActivityTimer))
}
