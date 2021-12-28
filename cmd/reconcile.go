package cmd

import (
	"github.com/spf13/cobra"
)

// reconcileCmd represents the wait command
var reconcileCmd = &cobra.Command{
	Use:   "reconcile",
	Short: "Trigger server-side reconsiliation",
	Long:  `Run any-open server-side code. Needs to be run manually when using an embedded server.`,
	Run: func(cmd *cobra.Command, args []string) {
		cli.ReconcileServer()
	},
}

func init() {
	rootCmd.AddCommand(reconcileCmd)
}
