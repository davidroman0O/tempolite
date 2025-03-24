package client

import (
	"time"
)

// Option is a functional option for configuring operations
type Option func(interface{})

// WorkflowOptions contains options for workflow execution
type WorkflowOptions struct {
	// ID is the optional externally provided ID
	ID string

	// Queue is the name of the queue to execute the workflow on
	Queue string

	// RetryPolicy defines how the workflow should be retried on failure
	RetryPolicy *ClientRetryPolicy

	// Timeout is the maximum time the workflow can run
	Timeout time.Duration

	// DeferExecution indicates if the workflow should be created but not started
	DeferExecution bool

	// Memo is unstructured metadata for the workflow
	Memo map[string]interface{}

	// ParentWorkflowID is the ID of the parent workflow if this is a child workflow
	ParentWorkflowID string

	// CronSchedule is a cron expression for scheduled workflows
	CronSchedule string

	// WaitForCancellation indicates whether to wait for cancellation to complete
	WaitForCancellation bool
}

// ClientRetryPolicy defines how operations should be retried on failure
type ClientRetryPolicy struct {
	// MaxAttempts is the maximum number of retry attempts
	MaxAttempts int

	// InitialInterval is the initial retry interval
	InitialInterval time.Duration

	// MaxInterval is the maximum retry interval
	MaxInterval time.Duration

	// BackoffCoefficient is the coefficient for backoff
	BackoffCoefficient float64

	// NonRetryableErrors are errors that should not be retried
	NonRetryableErrors []string
}

// ActivityOptions contains options for activity execution
type ActivityOptions struct {
	// TaskQueue is the name of the queue to execute the activity on
	TaskQueue string

	// RetryPolicy defines how the activity should be retried on failure
	RetryPolicy *ClientRetryPolicy

	// Timeout is the maximum time the activity can run
	Timeout time.Duration

	// HeartbeatTimeout defines when an activity is considered failed due to timeout
	HeartbeatTimeout time.Duration

	// HeartbeatInterval defines how often the activity should heartbeat
	HeartbeatInterval time.Duration

	// WaitForCancellation indicates whether to wait for cancellation to complete
	WaitForCancellation bool
}

// SignalOptions contains options for signal operations
type SignalOptions struct {
	// CorrelationID is an optional ID to track the signal
	CorrelationID string

	// Timeout is the maximum time to wait for delivery
	Timeout time.Duration

	// RetryPolicy defines how the signal should be retried on failure
	RetryPolicy *ClientRetryPolicy
}

// SagaOptions contains options for saga execution
type SagaOptions struct {
	// Parallel indicates whether steps can be executed in parallel
	Parallel bool

	// StrictMode indicates whether to fail on any compensation failure
	StrictMode bool

	// CompensationRetryPolicy defines how the compensation should be retried on failure
	CompensationRetryPolicy *ClientRetryPolicy
}

// SignalWaitOptions contains options for waiting on signals
type SignalWaitOptions struct {
	// Timeout is the maximum time to wait for the signal
	Timeout time.Duration

	// ContinueOnError indicates whether to continue execution after timeout
	ContinueOnError bool
}

// WithID sets the workflow ID
func WithID(id string) Option {
	return func(o interface{}) {
		if opts, ok := o.(*WorkflowOptions); ok {
			opts.ID = id
		}
	}
}

// WithQueue sets the queue name
func WithQueue(queue string) Option {
	return func(o interface{}) {
		if opts, ok := o.(*WorkflowOptions); ok {
			opts.Queue = queue
		} else if opts, ok := o.(*ActivityOptions); ok {
			opts.TaskQueue = queue
		}
	}
}

// WithRetryPolicy sets the retry policy
func WithRetryPolicy(policy *ClientRetryPolicy) Option {
	return func(o interface{}) {
		if opts, ok := o.(*WorkflowOptions); ok {
			opts.RetryPolicy = policy
		} else if opts, ok := o.(*ActivityOptions); ok {
			opts.RetryPolicy = policy
		} else if opts, ok := o.(*SignalOptions); ok {
			opts.RetryPolicy = policy
		} else if opts, ok := o.(*SagaOptions); ok {
			opts.CompensationRetryPolicy = policy
		}
	}
}

// WithTimeout sets the timeout
func WithTimeout(timeout time.Duration) Option {
	return func(o interface{}) {
		if opts, ok := o.(*WorkflowOptions); ok {
			opts.Timeout = timeout
		} else if opts, ok := o.(*ActivityOptions); ok {
			opts.Timeout = timeout
		} else if opts, ok := o.(*SignalOptions); ok {
			opts.Timeout = timeout
		} else if opts, ok := o.(*SignalWaitOptions); ok {
			opts.Timeout = timeout
		}
	}
}

// WithDeferExecution defers workflow execution
func WithDeferExecution() Option {
	return func(o interface{}) {
		if opts, ok := o.(*WorkflowOptions); ok {
			opts.DeferExecution = true
		}
	}
}

// WithMemo adds metadata to a workflow
func WithMemo(key string, value interface{}) Option {
	return func(o interface{}) {
		if opts, ok := o.(*WorkflowOptions); ok {
			if opts.Memo == nil {
				opts.Memo = make(map[string]interface{})
			}
			opts.Memo[key] = value
		}
	}
}

// WithHeartbeatTimeout sets the heartbeat timeout
func WithHeartbeatTimeout(timeout time.Duration) Option {
	return func(o interface{}) {
		if opts, ok := o.(*ActivityOptions); ok {
			opts.HeartbeatTimeout = timeout
		}
	}
}

// WithHeartbeatInterval sets the heartbeat interval
func WithHeartbeatInterval(interval time.Duration) Option {
	return func(o interface{}) {
		if opts, ok := o.(*ActivityOptions); ok {
			opts.HeartbeatInterval = interval
		}
	}
}

// WithParallelExecution enables parallel saga execution
func WithParallelExecution() Option {
	return func(o interface{}) {
		if opts, ok := o.(*SagaOptions); ok {
			opts.Parallel = true
		}
	}
}

// WithStrictMode enables strict mode for saga execution
func WithStrictMode() Option {
	return func(o interface{}) {
		if opts, ok := o.(*SagaOptions); ok {
			opts.StrictMode = true
		}
	}
}

// WithCorrelationID sets a correlation ID for signals
func WithCorrelationID(id string) Option {
	return func(o interface{}) {
		if opts, ok := o.(*SignalOptions); ok {
			opts.CorrelationID = id
		}
	}
}

// WithContinueOnError continues execution after a signal timeout
func WithContinueOnError() Option {
	return func(o interface{}) {
		if opts, ok := o.(*SignalWaitOptions); ok {
			opts.ContinueOnError = true
		}
	}
}

// WithCronSchedule sets a cron schedule for a workflow
func WithCronSchedule(schedule string) Option {
	return func(o interface{}) {
		if opts, ok := o.(*WorkflowOptions); ok {
			opts.CronSchedule = schedule
		}
	}
}

// WithWaitForCancellation waits for cancellation to complete
func WithWaitForCancellation() Option {
	return func(o interface{}) {
		if opts, ok := o.(*WorkflowOptions); ok {
			opts.WaitForCancellation = true
		} else if opts, ok := o.(*ActivityOptions); ok {
			opts.WaitForCancellation = true
		}
	}
}

// DefaultRetryPolicy returns a default retry policy
func DefaultRetryPolicy() *ClientRetryPolicy {
	return &ClientRetryPolicy{
		MaxAttempts:        3,
		InitialInterval:    time.Second,
		MaxInterval:        time.Minute,
		BackoffCoefficient: 2.0,
	}
}

// DefaultWorkflowOptions returns default workflow options
func DefaultWorkflowOptions() WorkflowOptions {
	return WorkflowOptions{
		Queue:       "default",
		RetryPolicy: DefaultRetryPolicy(),
		Timeout:     time.Hour,
	}
}

// DefaultActivityOptions returns default activity options
func DefaultActivityOptions() ActivityOptions {
	return ActivityOptions{
		TaskQueue:         "default",
		RetryPolicy:       DefaultRetryPolicy(),
		Timeout:           time.Minute * 5,
		HeartbeatTimeout:  time.Minute,
		HeartbeatInterval: time.Second * 30,
	}
}

// DefaultSignalOptions returns default signal options
func DefaultSignalOptions() SignalOptions {
	return SignalOptions{
		Timeout:     time.Minute,
		RetryPolicy: DefaultRetryPolicy(),
	}
}

// DefaultSagaOptions returns default saga options
func DefaultSagaOptions() SagaOptions {
	return SagaOptions{
		Parallel:                false,
		StrictMode:              false,
		CompensationRetryPolicy: DefaultRetryPolicy(),
	}
}

// DefaultSignalWaitOptions returns default signal wait options
func DefaultSignalWaitOptions() SignalWaitOptions {
	return SignalWaitOptions{
		Timeout:         time.Hour * 24,
		ContinueOnError: false,
	}
}
