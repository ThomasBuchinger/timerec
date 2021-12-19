package client

import (
	"fmt"
	"os"
	"time"

	"github.com/thomasbuchinger/timerec/api"
)

func (c *ClientObject) Panic(code int, message string, err error) {
	c.logger.Fatalln(message)
	c.logger.Fatalln(err)
	os.Exit(code)
}

// func strToPtr(str string) *string {
// 	if str == "" {
// 		return nil
// 	} else {
// 		return &str
// 	}
// }

func toTimestamp(duration time.Duration) time.Time {
	return time.Now().Add(duration)
}

func CheckNoActivityActive(profile api.Profile) error {
	err := CheckActivityActive(profile)
	if err == nil {
		return fmt.Errorf(profile.ActivityName)
	}
	profile.ActivityComment = ""
	profile.ActivityStart = time.Time{}
	profile.ActivityTimer = time.Time{}
	return nil
}

func CheckActivityActive(profile api.Profile) error {
	if profile.ActivityName == "" {
		return fmt.Errorf("no activity")
	}
	return nil
}

func PrintActivity(profile api.Profile) {
	err := CheckActivityActive(profile)
	if err != nil {
		fmt.Println("No Task active")
		return
	}

	roundToSecond, _ := time.ParseDuration("1m")
	start_h, start_m, _ := profile.ActivityStart.Clock()
	start_dur := time.Since(profile.ActivityStart).Round(roundToSecond).String()
	fin_h, fin_m, _ := profile.ActivityTimer.Clock()
	fin_dur := time.Until(profile.ActivityTimer).Round(roundToSecond).String
	dur := profile.ActivityTimer.Sub(profile.ActivityStart).Round(roundToSecond).String()
	fmt.Printf("Working on:     %s\n", profile.ActivityName)
	fmt.Printf("Started:        %d:%d (%s ago)\n", start_h, start_m, start_dur)
	fmt.Printf("Est. to finish: %d:%d (%s)\n", fin_h, fin_m, fin_dur())
	fmt.Printf("Duration:       %s\n", dur)

}

func CombineTask(existing api.Task, template_list []api.RecordTemplate, template string, title string, description string, project string, task string) api.Task {
	newTask := api.Task{Name: existing.Name}
	// Set Values according to Template
	if template != "" {
		for _, v := range template_list {
			if v.TemplateName == template {
				newTask.RecordTemplate.Project = v.Project
				newTask.RecordTemplate.Task = v.Task
				newTask.RecordTemplate.Title = v.Title
				newTask.RecordTemplate.Description = v.Description
				break
			}
		}
	}

	// Update newTask with existing values
	newTask.Update(existing)

	// Overwrite anything set explicitly
	if title != "" {
		newTask.RecordTemplate.Title = title
	}
	if description != "" {
		newTask.RecordTemplate.Description = description
	}
	if project != "" {
		newTask.RecordTemplate.Project = project
	}
	if task != "" {
		newTask.RecordTemplate.Task = task
	}

	return newTask
}
