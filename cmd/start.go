package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var startTaskCmd = &cobra.Command{
	Use:   "start NAME --start START [--est ESTIMATE] [additinal args are interpreted as comment]",
	Short: "Start to work on a task",
	Long: `Record when work on this task was started. This will automatically set the NAME as your active Task

If no task with this name exists, a basic task-object will be created.
Use '--start' and '--est' to record the begin and estimated finish , relative to right now.
'--est' has no effect, except reminding you to finish the task or update your estimate


Example:
    # Started to work on TICKET-13 15 minutes ago, and be reminded in 1h
    timerec start TICKET-13 --start -15m --est 1h
`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		est, err1 := cmd.Flags().GetDuration("est")
		defaultEst, err2 := cmd.Flags().GetDuration("default-estimate")
		if err1 != nil && est != -1 && err2 == nil {
			est = defaultEst
			fmt.Printf("No Esimate given, using default %s\n", defaultEst.String())
		}
		start, err1 := cmd.Flags().GetDuration("start")
		if err1 != nil {
			cli.Panic(1, "CLI parse error ", err1)
		}

		cli.EnsureTaskExists(args[0])
		EditTaskRun(cmd, args)
		cli.StartActivity(args[0], strings.Join(args[1:], " "), start, est)
	},
}

func init() {
	rootCmd.AddCommand(startTaskCmd)

	startTaskCmd.Flags().Duration("start", time.Duration(0), "When did you start?")
	startTaskCmd.Flags().Duration("est", time.Duration(-1), "How long is it going to take?")
	startTaskCmd.MarkFlagRequired("start")

	AddEditTaskFlags(startTaskCmd)
}
