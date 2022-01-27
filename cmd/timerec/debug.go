package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/thomasbuchinger/timerec/api"
	"github.com/thomasbuchinger/timerec/internal/server"
	"gopkg.in/yaml.v2"
)

var debugCmd = &cobra.Command{
	Use:   "debug user|jobs|templates",
	Short: "Shows API objects",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		embeddedServer := server.NewServer()
		var response interface{}
		var err error

		switch args[0] {
		case "user":
			response, err = embeddedServer.StateProvider.GetUser(api.User{Name: "me"})
		case "jobs":
			response, err = embeddedServer.StateProvider.ListJobs()
		case "templates":
			response, err = embeddedServer.TemplateProvider.GetTemplates()
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
