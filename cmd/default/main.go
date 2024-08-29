package main

import (
	"com.github/sarkarshuvojit/webhook-load-tester/pkg/types"
)

func main() {
	wt := types.NewDefaultWebhookTester()
	wt.InitTestSetup()
}
