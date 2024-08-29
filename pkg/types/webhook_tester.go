package types

type WebhookTester interface {
	LoadConfig() error
	InitTestSetup() error
	InitReceiver() error
	InitRequests() error
	WaitForResults() error
	PostProcess() error
}
