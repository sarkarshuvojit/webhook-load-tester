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

type WebhookTesterv2 interface {
	LoadConfig() error
	StartReceiver() (context.CancelFunc, error)
	FireRequests() error
	WaitForResults() error
	PostProcess() error
}
