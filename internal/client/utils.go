package client

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/thomasbuchinger/timerec/api"
)

func (c *ClientObject) Panic(code int, message string, err error) {
	c.logger.Fatalln(message)
	c.logger.Fatalln(err)
	os.Exit(code)
}

func EditRecordsPreSendHook(rec []api.Record) []api.Record {
	return []api.Record{}
}

func FormatActivity(activity api.Activity) string {
	var builder strings.Builder
	err := activity.CheckActivityActive()
	if err != nil {
		builder.WriteString("No active Activity")
		return builder.String()
	}

	roundToSecond, _ := time.ParseDuration("1m")
	start_h, start_m, _ := activity.ActivityStart.Clock()
	start_dur := time.Since(activity.ActivityStart).Round(roundToSecond).String()
	fin_h, fin_m, _ := activity.ActivityTimer.Clock()
	fin_dur := time.Until(activity.ActivityTimer).Round(roundToSecond).String
	dur := activity.ActivityTimer.Sub(activity.ActivityStart).Round(roundToSecond).String()
	fmt.Fprintf(&builder, "Working on:     %s\n", activity.ActivityName)
	fmt.Fprintf(&builder, "Started:        %d:%d (%s ago)\n", start_h, start_m, start_dur)
	fmt.Fprintf(&builder, "Est. to finish: %d:%d (%s)\n", fin_h, fin_m, fin_dur())
	fmt.Fprintf(&builder, "Duration:       %s\n", dur)
	return builder.String()
}

func FormatUserStatus(user api.User, jobs []api.Job) string {
	var builder strings.Builder
	validationError := user.Activity.CheckNoActivityActive()
	free := validationError == nil
	min, _ := time.ParseDuration("1m")

	builder.WriteString(user.Name)
	builder.WriteString(" is currently")
	if free {
		builder.WriteString(" free.")
	} else {
		builder.WriteString(" working on ")
		builder.WriteString(user.Activity.ActivityName)
		builder.WriteString(". Started on: **")
		builder.WriteString(user.Activity.ActivityStart.Format(time.RFC1123))
		builder.WriteString("**, for the next ")
		builder.WriteString(time.Until(user.Activity.ActivityTimer).Truncate(min).String())
	}

	var jobTitles []string
	for _, j := range jobs {
		jobTitles = append(jobTitles, j.Name)
	}
	builder.WriteString("\nOpen Jobs: ")
	builder.WriteString(strings.Join(jobTitles, ", "))
	return builder.String()
}

func FormatDay(records []api.Record) string {
	// | Title        | Start | End   | Duration | Description      |
	// |--------------|-------|-------|----------|------------------|
	// | Fix Jira-123 | 10:15 | 15:00 |  4h 45m  | Off by one error |
	// | Weekly       | 15:00 | 15:30 |     30m  |                  |
	// | Review PR    | 15:30 | 17:00 |  1h 30m  |                  |
	var builder strings.Builder
	// data := []struct {
	// 	Title string
	// 	Desc  string
	// 	Start time.Time
	// 	End   time.Time
	// }{}

	return builder.String()
}
