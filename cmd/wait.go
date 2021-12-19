package cmd

import (
	"github.com/spf13/cobra"
	"github.com/thomasbuchinger/timerec/internal/server"
)

// waitCmd represents the wait command
var waitCmd = &cobra.Command{
	Use:   "wait [--reconcile]",
	Short: "Print current status and wait for the current estimate to timeout.",
	Long: `Print status of the current activity, and wait for activity timer to time out.
Optionally run any server-side reconciliation loops from an embedded server. This will update any configured backend systems.
The Reconciliation loop is rune once at the beginning of this command and once again after the activity timer expired
`,
	Run: func(cmd *cobra.Command, args []string) {
		reconcile, _ := cmd.Flags().GetBool("reconcile")
		if reconcile {
			server.Reconcile()
		}

		cli.ActivityInfo()
		cli.Wait()

		if reconcile {
			server.Reconcile()
		}
	},
}

func init() {
	rootCmd.AddCommand(waitCmd)
	waitCmd.Flags().BoolP("reconcile", "r", false, "Run server-side reconciliation while waiting")
}
