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

var end_duration time.Duration
var finTaskCmd = &cobra.Command{
	Use:   "fin NAME",
	Short: "Finish a task and record the time spent on it",
	Long:  `Finishing a Task will actually log the time in clockodo.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cli.FinishTask(args[0], end_duration)
	},
}

func init() {
	rootCmd.AddCommand(finTaskCmd)

	finTaskCmd.Flags().DurationVar(&end_duration, "end", time.Duration(0), "When did you finish?")
}
