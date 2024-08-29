package main

import (
	"flag"
	"log/slog"
	"os"

	"com.github/sarkarshuvojit/webhook-load-tester/pkg/types"
)

var isVerbose bool

func setupVerboseFlag() {
	flag.BoolVar(&isVerbose, "verbose", false, "Enable Verbosity")
	flag.Parse()

	if *&isVerbose {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})))
	}
}

func main() {
	setupVerboseFlag()

	wt := types.NewDefaultWebhookTester()
	wt.LoadConfig()
	wt.InitTestSetup()
	cancelReceiver, err := wt.InitReceiver()
	defer cancelReceiver()

	if err != nil {
		slog.Error("Failed to start receiver", "error", err)
	}

	wt.InitRequests()
	wt.WaitForResults()
	wt.PostProcess()
}
