/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/sarkarshuvojit/webhook-load-tester/internal/utils"
	"github.com/sarkarshuvojit/webhook-load-tester/pkg/templates"
	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Generate a new webhook test configuration file",
	Long:  `The create command generates a new webhook test configuration file with boilerplate YAML content.`,
	Run: func(cmd *cobra.Command, args []string) {
		path, _ := cmd.Flags().GetString("config")

		if err := templates.CreateTemplate(path); err != nil {
			utils.PPrinter.Error("Error creating file")
		} else {
			utils.PPrinter.Success("Create config file successfully")
		}
	},
}

func init() {
	rootCmd.AddCommand(createCmd)

	createCmd.Flags().StringP("config", "c", "wlt.yaml", "Path to create the new test config")
	createCmd.MarkFlagRequired("config")
}
