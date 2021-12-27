package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/thomasbuchinger/timerec/internal/client"
	"gopkg.in/yaml.v2"
)

var debugCmd = &cobra.Command{
	Use:   "debug profile|items|templates",
	Short: "Shows API objects",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		rest := client.RestClient{}
		var response interface{}
		var err error

		switch args[0] {
		case "profile":
			response, err = rest.GetActivity()
		case "items":
			response, err = rest.ListWorkItems()
		case "templates":
			response, err = rest.ListTemplates()
		}
		if err != nil {
			fmt.Println(err)
			return
		}
		data, err := yaml.Marshal(response)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(string(data))
	},
}

func init() {
	rootCmd.AddCommand(debugCmd)

}
