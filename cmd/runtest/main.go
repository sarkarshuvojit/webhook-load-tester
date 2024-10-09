package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/sarkarshuvojit/webhook-load-tester/internal/utils"
	"github.com/sarkarshuvojit/webhook-load-tester/pkg/types"
	"github.com/sarkarshuvojit/webhook-load-tester/pkg/webhook_tester"
	"gopkg.in/yaml.v3"
)

var isVerbose bool
var configPath string

var DEFAULT_WAITING_TIMEOUT = time.Duration(30) * time.Second

func setupFlags() {
	flag.BoolVar(&isVerbose, "v", false, "Enable Verbosity")
	flag.StringVar(&configPath, "f", "wlt-run-defatult.yaml", "Path to your test config file")
	flag.Parse()
}

func setupLogger(isVerbose bool) {
	if *&isVerbose {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})))
	} else {
		slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})))
	}
}

func initialize() {
	setupFlags()
	setupLogger(isVerbose)
}

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

func main() {
	initialize()

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
