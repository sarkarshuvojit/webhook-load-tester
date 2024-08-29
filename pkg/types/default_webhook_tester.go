package types

type DefaultWebhookTester struct{}

func NewDefaultWebhookTester() *DefaultWebhookTester {
	return &DefaultWebhookTester{}
}

// InitReceiver implements WebhookTester.
func (*DefaultWebhookTester) InitReceiver() error {
	panic("unimplemented")
}

// InitRequests implements WebhookTester.
func (*DefaultWebhookTester) InitRequests() error {
	panic("unimplemented")
}

// InitTestSetup implements WebhookTester.
func (*DefaultWebhookTester) InitTestSetup() error {
	panic("unimplemented")
}

// LoadConfig implements WebhookTester.
func (*DefaultWebhookTester) LoadConfig() error {
	panic("unimplemented")
}

// PostProcess implements WebhookTester.
func (*DefaultWebhookTester) PostProcess() error {
	panic("unimplemented")
}

// WaitForResults implements WebhookTester.
func (*DefaultWebhookTester) WaitForResults() error {
	panic("unimplemented")
}

var _ WebhookTester = (*DefaultWebhookTester)(nil)
