package types

import "context"

type WebhookTester interface {
	LoadConfig() error
	InitTestSetup() error
	InitReceiver() (context.CancelFunc, error)
	InitRequests() error
	WaitForResults() error
	PostProcess() error
}
