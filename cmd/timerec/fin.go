package main

import (
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var finTaskCmd = &cobra.Command{
	Use:   "fin NAME --end END [COMMENT]",
	Short: "Finish a Job and record the time spent on it",
	Long:  `Finishing a Job will actually log the time in clockodo.`,
	Example: `
# Going to finish working on TICKET-13 in 10 minutes, just have to write a nice commit message
./timerec fin TICKET-13 --end 10m
	`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		endDuration, err1 := cmd.Flags().GetDuration("end")
		if err1 != nil {
			cli.Panic(1, "CLI parse error", nil)
		}

		EditTaskRun(cmd, args)
		cli.FinishActivity(args[0], args[0], strings.Join(args[1:], " "), endDuration)
		cli.CompleteJob(args[0])
	},
}

func init() {
	rootCmd.AddCommand(finTaskCmd)

	finTaskCmd.Flags().Duration("end", time.Duration(0), "When did you finish?")
	finTaskCmd.MarkFlagRequired("end")
	AddEditTaskFlags(finTaskCmd)

}
