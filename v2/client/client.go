// Package client provides the external API for interacting with the Tempolite engine.
// It handles workflow execution, signals, and workflow control operations.
package client

// Client is the main interface for interacting with the Tempolite engine.
// It provides methods for starting workflows, sending signals, and controlling workflows.
type Client struct {
	// Internal dependencies
	engine      interface{} // Will be properly typed later
	persistence interface{} // Will be properly typed later
}

// NewClient creates a new client with the given dependencies.
func NewClient(engine interface{}, persistence interface{}) *Client {
	return &Client{
		engine:      engine,
		persistence: persistence,
	}
}

// NewWorkflow creates a workflow builder for the given workflow function.
func (c *Client) NewWorkflow(workflowFn interface{}) *WorkflowBuilder {
	return &WorkflowBuilder{
		client:     c,
		workflowFn: workflowFn,
	}
}

// GetWorkflowFuture returns a future for a running workflow.
func (c *Client) GetWorkflowFuture(workflowID string) *Future {
	return &Future{
		client:     c,
		workflowID: workflowID,
	}
}

// PauseWorkflow pauses a running workflow.
func (c *Client) PauseWorkflow(workflowID string) *Future {
	return &Future{
		client:     c,
		workflowID: workflowID,
		operation:  "pause",
	}
}

// ResumeWorkflow resumes a paused workflow.
func (c *Client) ResumeWorkflow(workflowID string) *Future {
	return &Future{
		client:     c,
		workflowID: workflowID,
		operation:  "resume",
	}
}

// CancelWorkflow cancels a workflow.
func (c *Client) CancelWorkflow(workflowID string) *Future {
	return &Future{
		client:     c,
		workflowID: workflowID,
		operation:  "cancel",
	}
}

// GetWorkflowStatus returns the status of a workflow.
func (c *Client) GetWorkflowStatus(workflowID string) (string, error) {
	// Get the workflow status from persistence
	return "Running", nil // Placeholder
}

// Signal creates a signal builder for sending signals to workflows.
func (c *Client) Signal(name string) *SignalBuilder {
	return &SignalBuilder{
		client:     c,
		signalName: name,
	}
}

// SignalTyped sends a typed signal to a workflow.
func (c *Client) SignalTyped(workflowID string, signalName string, payload interface{}) *Future {
	return c.Signal(signalName).
		WithPayload(payload).
		ToWorkflow(workflowID).
		Send()
}

// WorkflowBuilder provides a fluent API for configuring and executing workflows.
type WorkflowBuilder struct {
	client     *Client
	workflowFn interface{}
	retries    *RetryPolicy
	timeout    int
	queue      string
}

// WithRetries configures the retry policy for the workflow.
func (wb *WorkflowBuilder) WithRetries(policy RetryPolicy) *WorkflowBuilder {
	wb.retries = &policy
	return wb
}

// WithTimeout configures the timeout for the workflow.
func (wb *WorkflowBuilder) WithTimeout(timeout int) *WorkflowBuilder {
	wb.timeout = timeout
	return wb
}

// WithQueue configures the queue for the workflow.
func (wb *WorkflowBuilder) WithQueue(queue string) *WorkflowBuilder {
	wb.queue = queue
	return wb
}

// Execute executes the workflow with the given arguments.
func (wb *WorkflowBuilder) Execute(args ...interface{}) string {
	// Submit the workflow to the engine for execution
	// Return the workflow ID
	return "workflow-id" // Placeholder
}

// RetryPolicy defines how retries are handled.
type RetryPolicy struct {
	MaxAttempts int
	MaxInterval int
}

// SignalBuilder provides a fluent API for configuring and sending signals.
type SignalBuilder struct {
	client        *Client
	signalName    string
	payload       interface{}
	correlationID string
	workflowID    string
	workflowIDs   []string
}

// WithPayload configures the payload for the signal.
func (sb *SignalBuilder) WithPayload(payload interface{}) *SignalBuilder {
	sb.payload = payload
	return sb
}

// WithCorrelationID configures the correlation ID for the signal.
func (sb *SignalBuilder) WithCorrelationID(correlationID string) *SignalBuilder {
	sb.correlationID = correlationID
	return sb
}

// ToWorkflow configures the target workflow for the signal.
func (sb *SignalBuilder) ToWorkflow(workflowID string) *SignalBuilder {
	sb.workflowID = workflowID
	return sb
}

// ToWorkflows configures multiple target workflows for the signal.
func (sb *SignalBuilder) ToWorkflows(workflowIDs []string) *SignalBuilder {
	sb.workflowIDs = workflowIDs
	return sb
}

// Send sends the signal to the configured workflow(s).
func (sb *SignalBuilder) Send() *Future {
	// Send the signal to the workflow(s)
	return &Future{
		client:     sb.client,
		workflowID: sb.workflowID,
		operation:  "signal",
	}
}
