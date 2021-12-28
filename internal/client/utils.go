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

func EditRecordsPreSendHook(rec []api.Record) []api.Record {
	return []api.Record{}
}

func PrintActivity(activity api.Activity) {
	err := activity.CheckActivityActive()
	if err != nil {
		fmt.Println("No Task active")
		return
	}

	roundToSecond, _ := time.ParseDuration("1m")
	start_h, start_m, _ := activity.ActivityStart.Clock()
	start_dur := time.Since(activity.ActivityStart).Round(roundToSecond).String()
	fin_h, fin_m, _ := activity.ActivityTimer.Clock()
	fin_dur := time.Until(activity.ActivityTimer).Round(roundToSecond).String
	dur := activity.ActivityTimer.Sub(activity.ActivityStart).Round(roundToSecond).String()
	fmt.Printf("Working on:     %s\n", activity.ActivityName)
	fmt.Printf("Started:        %d:%d (%s ago)\n", start_h, start_m, start_dur)
	fmt.Printf("Est. to finish: %d:%d (%s)\n", fin_h, fin_m, fin_dur())
	fmt.Printf("Duration:       %s\n", dur)

}
