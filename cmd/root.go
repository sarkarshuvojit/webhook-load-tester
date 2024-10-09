/*
Copyright © 2024 Shuvojit Sarkar <s15sarkar@yahoo.com>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "webhook-load-tester",
	Short: "A tool for testing and load testing webhook-based APIs",
	Long: `Webhook Load Tester is a powerful tool designed to help developers test and load test asynchronous APIs that use webhooks.

Key features:
- Mock server to capture webhook responses
- YAML-based configuration for easy test setup
- End-to-end timing measurements
- Load testing capabilities for asynchronous APIs
- Ngrok support for local testing
- Visual representation of test results

This tool is particularly useful for:
- Developers working with webhook-based APIs
- QA engineers testing asynchronous systems
- DevOps professionals managing API performance
`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose logs")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
}
