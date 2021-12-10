/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"time"

	"github.com/spf13/cobra"
)

var start_duration, estimate_duration time.Duration

// startCmd represents the start command
var startTaskCmd = &cobra.Command{
	Use:   "start NAME --start START --est ESTIMATE",
	Short: "Start to work on a task",
	Long: `Record when work on this task was started. This will automatically set the NAME as your active Task

	If no task with this name exists, a basic task-object will be created.
Use '--start' and '--est' to record the begin and estimated finish , relative to right now.
'--est' has no effect, except reminding you to finish the task or update your estimate


Example:
    # Started to work on TICKET-13 15 minutes ago, and be reminded in 1h
    timerec start TICKET-13 --start -15m --est 1h
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cli.StartTask(args[0], start_duration, estimate_duration)
	},
}

func init() {
	rootCmd.AddCommand(startTaskCmd)

	startTaskCmd.Flags().DurationVar(&start_duration, "start", time.Duration(0), "When did you start?")
	startTaskCmd.Flags().DurationVar(&estimate_duration, "est", time.Duration(0), "How long is it going to take?")

}
