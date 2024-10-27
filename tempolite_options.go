package tempolite

import "time"

// Queue configuration at startup
type queueConfig struct {
	Name                string
	WorkflowWorkers     int
	ActivityWorkers     int
	SideEffectWorkers   int
	TransactionWorkers  int
	CompensationWorkers int
}

type queueOption func(*queueConfig)

func WithWorkflowWorkers(n int) queueOption {
	return func(c *queueConfig) {
		c.WorkflowWorkers = n
	}
}

func WithActivityWorkers(n int) queueOption {
	return func(c *queueConfig) {
		c.ActivityWorkers = n
	}
}

func WithSideEffectWorkers(n int) queueOption {
	return func(c *queueConfig) {
		c.SideEffectWorkers = n
	}
}

func WithTransactionWorkers(n int) queueOption {
	return func(c *queueConfig) {
		c.TransactionWorkers = n
	}
}

func WithCompensationWorkers(n int) queueOption {
	return func(c *queueConfig) {
		c.CompensationWorkers = n
	}
}

func NewQueue(name string, opts ...queueOption) queueConfig {
	c := queueConfig{
		Name:                name,
		WorkflowWorkers:     1,
		ActivityWorkers:     1,
		SideEffectWorkers:   1,
		TransactionWorkers:  1,
		CompensationWorkers: 1,
	}
	for _, opt := range opts {
		opt(&c)
	}
	return c
}

type tempoliteConfig struct {
	path        *string
	destructive bool
	logger      Logger

	initialWorkflowsWorkers    int
	initialActivityWorkers     int
	initialSideEffectWorkers   int
	initialTransctionWorkers   int
	initialCompensationWorkers int

	// Additional queues
	queues []queueConfig
}

type tempoliteOption func(*tempoliteConfig)

func WithQueueConfig(queue queueConfig) tempoliteOption {
	return func(c *tempoliteConfig) {
		c.queues = append(c.queues, queue)
	}
}

func WithLogger(logger Logger) tempoliteOption {
	return func(c *tempoliteConfig) {
		c.logger = logger
	}
}

func WithPath(path string) tempoliteOption {
	return func(c *tempoliteConfig) {
		c.path = &path
	}
}

func WithMemory() tempoliteOption {
	return func(c *tempoliteConfig) {
		c.path = nil
	}
}

func WithDestructive() tempoliteOption {
	return func(c *tempoliteConfig) {
		c.destructive = true
	}
}

// If you intent to have sub-workflows you should increase this number accordingly
func WithInitialWorkflowsWorkers(n int) tempoliteOption {
	return func(c *tempoliteConfig) {
		c.initialWorkflowsWorkers = n
	}
}

func WithInitialActivityWorkers(n int) tempoliteOption {
	return func(c *tempoliteConfig) {
		c.initialActivityWorkers = n
	}
}

func WithInitialSideEffectWorkers(n int) tempoliteOption {
	return func(c *tempoliteConfig) {
		c.initialSideEffectWorkers = n
	}
}

func WithInitialTransctionWorkers(n int) tempoliteOption {
	return func(c *tempoliteConfig) {
		c.initialTransctionWorkers = n
	}
}

type tempoliteWorkflowConfig struct {
	retryMaximumAttempts    int
	retryInitialInterval    time.Duration
	retryBackoffCoefficient float64
	maximumInterval         time.Duration
	queueName               string
}

type tempoliteWorkflowOptions []tempoliteWorkflowOption

func WorkflowConfig(opts ...tempoliteWorkflowOption) tempoliteWorkflowOptions {
	return opts
}

type tempoliteWorkflowOption func(*tempoliteWorkflowConfig)

func WithQueue(queueName string) tempoliteWorkflowOption {
	return func(c *tempoliteWorkflowConfig) {
		c.queueName = queueName
	}
}

func WithRetryMaximumAttempts(max int) tempoliteWorkflowOption {
	return func(c *tempoliteWorkflowConfig) {
		c.retryMaximumAttempts = max
	}
}

// TODO: to implement
// func WithRetryInitialInterval(interval time.Duration) tempoliteWorkflowOption {
// 	return func(c *tempoliteWorkflowConfig) {
// 		c.retryInitialInterval = interval
// 	}
// }

// TODO: to implement
// func WithRetryBackoffCoefficient(coefficient float64) tempoliteWorkflowOption {
// 	return func(c *tempoliteWorkflowConfig) {
// 		c.retryBackoffCoefficient = coefficient
// 	}
// }

// TODO: to implement
// func WithMaximumInterval(max time.Duration) tempoliteWorkflowOption {
// 	return func(c *tempoliteWorkflowConfig) {
// 		c.maximumInterval = max
// 	}
// }
