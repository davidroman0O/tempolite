package schema

import "time"

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	MaximumAttempts    int           `json:"maximum_attempts"`
	InitialInterval    time.Duration `json:"initial_interval"`
	BackoffCoefficient float64       `json:"backoff_coefficient"`
	MaximumInterval    time.Duration `json:"maximum_interval"`
}

// SagaDefinitionData defines the structure of a saga
type SagaStepPair struct {
	TransactionHandlerName  string `json:"transaction_handler_name"`
	CompensationHandlerName string `json:"compensation_handler_name"`
}

type SagaDefinitionData struct {
	Steps []SagaStepPair `json:"steps"`
}
