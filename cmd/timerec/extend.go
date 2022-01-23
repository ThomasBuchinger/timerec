package main

import (
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var extendActivityCmd = &cobra.Command{
	Use:   "extend NAME --est ESTIMATE [COMMENT]",
	Short: "Set a new estimate for your current Job.",
	Long:  `Set a new estimate for your currently Job Task and restart the timer.`,
	Example: `  # Not finished yet, but probably in about an hour
  timerec extend TICKET-13 --est 1h
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
