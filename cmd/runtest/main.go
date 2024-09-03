package main

import (
	"encoding/json"
	"errors"
	"flag"
	"log/slog"
	"os"
	"strings"

	"com.github/sarkarshuvojit/webhook-load-tester/internal/utils"
	"com.github/sarkarshuvojit/webhook-load-tester/pkg/types"
	"gopkg.in/yaml.v3"
)

var isVerbose bool
var configPath string

func setupFlags() {
	flag.BoolVar(&isVerbose, "v", false, "Enable Verbosity")
	flag.StringVar(&configPath, "f", "wlt-run-defatult.yaml", "Path to your test config file")
	flag.Parse()
}

func setupLogger(isVerbose bool) {
	var newLevel slog.Level
	if *&isVerbose {
		newLevel = slog.LevelDebug
	} else {
		newLevel = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: newLevel,
	})))
}

func initialize() {
	setupFlags()
	setupLogger(isVerbose)
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
	return &config, nil
}

func main() {
	initialize()

	config, err := loadConfig(configPath)
	if err != nil {
		utils.PPrinter.Error("Failed due to: ", err.Error())
		os.Exit(1)
	}
	slog.Debug("Starting with config", "config", config)

	wt := types.NewDefaultWebhookTesterv2(config)
	if err = wt.LoadConfig(); err != nil {
		utils.PPrinter.Error("Failed to load config: ", err.Error())
		os.Exit(1)
	}

	wt.StartReceiver()
	wt.FireRequests()
	wt.WaitForResults()
}
