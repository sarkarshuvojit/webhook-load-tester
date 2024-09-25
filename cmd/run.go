/*
Copyright © 2024 Shuvojit Sarkar <s15sarkar@yahoo.com>
*/
package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"com.github/sarkarshuvojit/webhook-load-tester/internal/utils"
	"com.github/sarkarshuvojit/webhook-load-tester/pkg/types"
	"com.github/sarkarshuvojit/webhook-load-tester/pkg/webhook_tester"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func setupLogger(isVerbose bool) {
	if *&isVerbose {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})))
	} else {
		slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})))
	}
}

var DEFAULT_WAITING_TIMEOUT = time.Duration(30) * time.Second

func setDefaults(config *types.InputConfig) {
	if config.Test.Timeout == 0 {
		config.Test.Timeout = int(DEFAULT_WAITING_TIMEOUT.Seconds())
	}
}

func loadConfig(filepath string) (*types.InputConfig, error) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		return nil, errors.New("Could not find config file: " + filepath)
	}

	var config types.InputConfig
	if strings.HasSuffix(filepath, ".yaml") || strings.HasSuffix(filepath, ".yml") {
		err = yaml.Unmarshal(content, &config)
	} else if strings.HasSuffix(filepath, ".json") {
		err = json.Unmarshal(content, &config)
	} else {
		return nil, errors.New("Invalid filetype")
	}
	if err != nil {
		return nil, err
	}

	setDefaults(&config)

	return &config, nil
}

func runTest(configPath string) {
	config, err := loadConfig(configPath)
	if err != nil {
		utils.PPrinter.Error("Failed due to: ", err.Error())
		os.Exit(1)
	}
	utils.PPrinter.Info("Config loaded successfully...")
	slog.Debug("Starting with config", "config", config)

	wt := webhook_tester.NewDefaultWebhookTester(config)
	if err = wt.LoadConfig(); err != nil {
		utils.PPrinter.Error("Failed to load config due to: %v", err.Error())
		os.Exit(1)
	}

	utils.PPrinter.Info("Started receiver...")
	wt.StartReceiver()
	utils.PPrinter.Info("Firing requests...")
	wt.FireRequests()
	utils.PPrinter.Info(fmt.Sprintf("Waiting for responses for %ds...", config.Test.Timeout))
	if err := wt.WaitForResults(); err != nil {
		utils.PPrinter.Warning(fmt.Sprintf("Timed out waiting for %ds", config.Test.Timeout))
	} else {
		utils.PPrinter.Success("Received webhook responses within timeout.")
	}
	utils.PPrinter.Info("Starting post processing...")
	if err := wt.PostProcess(); err != nil {
		utils.PPrinter.Error(fmt.Sprintf("Failed to post process: %v", err))
	} else {
		utils.PPrinter.Success("Post processing complete.")
	}
}

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Execute a specific webhook test configuration",
	Long: `The run command executes a specific webhook test configuration defined in a YAML file.

This command allows you to:
- Run a single test scenario
- Measure the complete API flow from initial call to webhook response
- Capture and display webhook responses
- Generate detailed timing and performance metrics

Usage:
  webhook-load-tester run --config <path-to-config-file.yaml>

Example:
  webhook-load-tester run --config ./tests/payment-api-test.yaml`,
	PreRun: func(cmd *cobra.Command, args []string) {
		isVerbose, _ := cmd.Flags().GetBool("verbose")
		setupLogger(isVerbose)
	},
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("run called")
		configPath, _ := cmd.Flags().GetString("config")
		runTest(configPath)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringP("config", "c", "wlt.yaml", "Path to create the new test config")
	runCmd.MarkFlagRequired("config")
}
