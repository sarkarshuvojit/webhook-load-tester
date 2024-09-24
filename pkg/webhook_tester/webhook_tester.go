package webhook_tester

import "context"

type WebhookTester interface {
	LoadConfig() error
	StartReceiver() (context.CancelFunc, error)
	FireRequests() error
	WaitForResults() error
	PostProcess() error
}
