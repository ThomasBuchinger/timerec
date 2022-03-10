package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/thomasbuchinger/timerec/internal/server"
	"github.com/thomasbuchinger/timerec/internal/server/providers"
	"gopkg.in/yaml.v2"
)

var debugCmd = &cobra.Command{
	Use:   "debug user|jobs|templates",
	Short: "Shows API objects",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		embeddedServer := server.NewServer()
		var response interface{}
		state, err := embeddedServer.StateProvider.Refresh("me")
		if err != nil {
			fmt.Println(err)
			return
		}

		switch args[0] {
		case "user":
			response, err = providers.ListUsers(&state)
		case "jobs":
			response, err = providers.ListJobs(&state)
		case "templates":
			response, err = providers.ListTemplates(&state)
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
