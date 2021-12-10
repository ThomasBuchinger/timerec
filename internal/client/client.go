package client

import (
	"fmt"
	"log"
	"time"

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

func (c *ClientObject) StartTask(name string, start_duration time.Duration, estimate_duration time.Duration) {
	_, ok, err := c.restclient.FindTaskByName(name)
	if !ok {
		c.logger.Printf("Task does not exist, Creating it now...")
		_, err2 := c.restclient.NewTask(name)
		if err2 != nil {
			c.logger.Fatalf("Error creating Task '%s': %s", name, err2.Error())
			return
		}
	}
	c.logger.Printf("Start working on '%s'...", name)
	profile, err := c.restclient.SetActivity(name, time.Now().Add(start_duration), time.Now().Add(estimate_duration))
	if err != nil {
		c.logger.Fatal(err)
		return
	}
	duration := profile.ActivityTimer.Sub(profile.ActivityStart).String()
	h, m, _ := profile.ActivityTimer.Clock()
	c.logger.Printf("Working on '%s' until %d:%d (duration %s)", profile.ActivityName, h, m, duration)
}

func (c *ClientObject) Wait() {
	roundToSecond, _ := time.ParseDuration("1m")
	profile, err := c.restclient.GetActivity()
	if err != nil {
		c.logger.Fatal(err)
		return
	}

	start_h, start_m, _ := profile.ActivityStart.Clock()
	start_dur := time.Since(profile.ActivityStart).Round(roundToSecond).String()
	fin_h, fin_m, _ := profile.ActivityTimer.Clock()
	fin_dur := time.Until(profile.ActivityTimer).Round(roundToSecond).String
	dur := profile.ActivityTimer.Sub(profile.ActivityStart).Round(roundToSecond).String()
	fmt.Printf("Working on:     %s\n", profile.ActivityName)
	fmt.Printf("Started:        %d:%d (%s ago)\n", start_h, start_m, start_dur)
	fmt.Printf("Est. to finish: %d:%d (%s)\n", fin_h, fin_m, fin_dur())
	fmt.Printf("Duration:       %s\n", dur)

	time.Sleep(time.Until(profile.ActivityTimer))
}

func (c *ClientObject) FinishTask(name string, end_duration time.Duration) {
	task, ok, err := c.restclient.FindTaskByName(name)
	if err != nil {
		c.logger.Fatal(err)
		return
	} else if !ok {
		c.logger.Fatalf("Unable to find Task '%s' name", name)
		return
	}
	profile, err := c.restclient.GetActivity()
	if err != nil {
		c.logger.Fatal(err)
		return
	}

	task.Id = "1234"
	task.CustomerRef = "socradev"
	task.Activities = append(task.Activities, api.TimeEntry{
		Comment: "id did something",
		Start:   profile.ActivityStart,
		End:     time.Now().Add(end_duration),
	})

	err = c.restclient.SaveRecords(task.ConvertToRecords())
	if err != nil {
		c.logger.Fatal(err)
		return
	}
	c.logger.Println("Done")
}
