package tempolite

import "time"

type tempoliteConfig struct {
	path        *string
	destructive bool
	logger      Logger

	initialWorkflowsWorkers    int
	initialActivityWorkers     int
	initialSideEffectWorkers   int
	initialTransctionWorkers   int
	initialCompensationWorkers int
}

type tempoliteOption func(*tempoliteConfig)

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
}

type tempoliteWorkflowOptions []tempoliteWorkflowOption

func WorkflowConfig(opts ...tempoliteWorkflowOption) tempoliteWorkflowOptions {
	return opts
}

type tempoliteWorkflowOption func(*tempoliteWorkflowConfig)

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
