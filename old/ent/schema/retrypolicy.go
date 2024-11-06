package schema

import (
	"time"
)

// RetryPolicy represents retry configurations.
type RetryPolicy struct {
	MaximumAttempts    int           `json:"maximum_attempts"`
	InitialInterval    time.Duration `json:"initial_interval"`
	BackoffCoefficient float64       `json:"backoff_coefficient"`
	MaximumInterval    time.Duration `json:"maximum_interval"`
}
