package cmd

import (
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var extendActivityCmd = &cobra.Command{
	Use:   "extend --est ESTIMATE [COMMENT]",
	Short: "Set a new estimate for your current task.",
	Long:  `Set a new estimate for your currently active Task and restart the timer.`,
	Example: `  # Almost finished, just need 30m to update the PullRequest
  timerec extend --est 30m
	`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		est, err1 := cmd.Flags().GetDuration("est")
		if err1 != nil {
			cli.Panic(1, "CLI parse error ", err1)
		}
		EditTaskRun(cmd, args)
		cli.ExtendActivity(est, strings.Join(args, " "), false)
	},
}

func init() {
	rootCmd.AddCommand(extendActivityCmd)

	extendActivityCmd.Flags().Duration("est", time.Duration(0), "How long is it going to take?")
	extendActivityCmd.MarkFlagRequired("est")
	AddEditTaskFlags(extendActivityCmd)
}
