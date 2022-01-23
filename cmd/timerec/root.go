package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/thomasbuchinger/timerec/internal/client"

	"github.com/spf13/viper"
)

func main() {
	Execute()
}

var cfgFile string
var cli client.ClientObject

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "timerec",
	Short: "simple time recording app, that works",
	// Run: func(cmd *cobra.Command, args []string) {
	// 	cli.FinishActivity("hello", "hello", "comment", defaultEstimate)
	// },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.timerec.yaml)")
	rootCmd.PersistentFlags().Duration("default-estimate", time.Duration(0), "Set default value for estimates (useful to configure in config file)")
	viper.BindPFlag("settings.default_estimate", rootCmd.Flags().Lookup("default-estimate"))

	cli = client.NewClient()
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".timerec" (without extension).
		currentDir, _ := os.Getwd()
		viper.AddConfigPath(home)
		viper.AddConfigPath(currentDir)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".timerec")
		viper.SetConfigName("timerec-config.yaml")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
