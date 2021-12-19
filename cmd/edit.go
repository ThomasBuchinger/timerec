package cmd

import (
	"github.com/spf13/cobra"
)

var editTaskCmd = &cobra.Command{
	Use:   "edit NAME",
	Short: "Edit a task",
	Long: `Tasks can be worked on with only a name. BUT more details are required to save them permanently

To update an existing task
`,
	Example: `  #
  timerec edit TICKET-13 --template my-project -n --description ""
	`,
	Args: cobra.ExactArgs(1),
	Run:  EditTaskRun,
}

func init() {
	rootCmd.AddCommand(editTaskCmd)

	AddEditTaskFlags(editTaskCmd)
}

func AddEditTaskFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("template", "t", "", "Use an existing template")
	cmd.Flags().BoolP("use-name-as-title", "n", false, "Use the name as title")
	cmd.Flags().BoolP("interactive", "i", false, "Start interactive editor")

	cmd.Flags().String("project", "", "Which project are you working on?")
	cmd.Flags().String("task", "", "Which task in the project?")
	cmd.Flags().String("title", "", "In a few words. What are you working on?")
	cmd.Flags().String("desc", "", "Any additional Details?")
}

func EditTaskRun(cmd *cobra.Command, args []string) {
	template, err1 := cmd.Flags().GetString("template")
	useNameAsTitle, err2 := cmd.Flags().GetBool("use-name-as-title")
	title, err3 := cmd.Flags().GetString("title")
	description, err4 := cmd.Flags().GetString("desc")
	project, err5 := cmd.Flags().GetString("project")
	taskInProject, err6 := cmd.Flags().GetString("task")

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil || err6 != nil {
		cli.Panic(10, "CLI parse error", nil)
	}

	if useNameAsTitle {
		title = args[0]
	}

	cli.EnsureTaskExists(args[0])
	cli.UpdateTask(args[0], template, title, description, project, taskInProject)

}
