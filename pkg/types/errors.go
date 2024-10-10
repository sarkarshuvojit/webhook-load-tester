package types

import "errors"

var (
	TimedOutWaitingForResultsErr = errors.New("timed out waiting for results")
	UnsupportedOutputErr         = errors.New("Unsupported output format")
	NgrokAuthMissingErr          = errors.New("Ngrok auth token missing from environment. Please set NGROK_AUTHTOKEN to use ngrok")
)
