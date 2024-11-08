package types

import "errors"

var (
	ErrWorkflowPaused = errors.New("workflow paused") // It is not an error but we use it as an error
)
