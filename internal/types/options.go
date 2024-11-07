package types

import "time"

type WorkflowConfig struct {
	RetryMaximumAttempts int
	RetryInitialInterval int64
	// retryBackoffCoefficient float64 // not used YET
	// maximumInterval time.Duration // not used YET
	QueueName string
	Duration  string
}

type WorkflowOptions []workflowOption

func NewWorkflowConfig(opts ...workflowOption) WorkflowOptions {
	return opts
}

type workflowOption func(*WorkflowConfig)

func WithWorkflowRetryInitialInterval(interval time.Duration) workflowOption {
	return func(c *WorkflowConfig) {
		c.RetryInitialInterval = int64(interval)
	}
}

func WithWorkflowContextDuration(duration string) workflowOption {
	return func(c *WorkflowConfig) {
		c.Duration = duration
	}
}

func WithWorkflowQueue(queueName string) workflowOption {
	return func(c *WorkflowConfig) {
		c.QueueName = queueName
	}
}

func WithWorkflowRetryMaximumAttempts(max int) workflowOption {
	return func(c *WorkflowConfig) {
		c.RetryMaximumAttempts = max
	}
}
