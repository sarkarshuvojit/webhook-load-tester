package main

import (
	"log/slog"
	"os"

	"com.github/sarkarshuvojit/webhook-load-tester/pkg/types"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))
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
