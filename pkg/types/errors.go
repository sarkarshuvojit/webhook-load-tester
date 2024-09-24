package types

import "errors"

var (
	TimedOutWaitingForResultsErr = errors.New("timed out waiting for results")
	UnsupportedOutputErr         = errors.New("Unsupported output format")
)
