// Package context provides the execution context for workflows and activities.
package context

// SideEffectBuilder provides a fluent API for configuring and executing side effects.
type SideEffectBuilder interface {
	// Execution
	Run(fn interface{}, args ...interface{}) SideEffectFuture
	RunIf(condition bool, fn interface{}, args ...interface{}) SideEffectFuture
}

// SideEffectFuture represents the result of an asynchronous side effect execution.
type SideEffectFuture interface {
	// Result handling
	Get(result interface{}) error
	IsDone() bool

	// Status information
	Error() error
}

// SideEffectContext is the context provided to side effect functions.
type SideEffectContext interface {
	// Context information
	WorkflowID() string
	SideEffectID() string
}

// sideEffectContextImpl implements the SideEffectContext interface.
type sideEffectContextImpl struct {
	workflowID   string
	sideEffectID string
}

// NewSideEffectContext creates a new side effect context.
func NewSideEffectContext(workflowID, sideEffectID string) SideEffectContext {
	return &sideEffectContextImpl{
		workflowID:   workflowID,
		sideEffectID: sideEffectID,
	}
}

// WorkflowID returns the ID of the workflow that initiated the side effect.
func (sc *sideEffectContextImpl) WorkflowID() string {
	return sc.workflowID
}

// SideEffectID returns the ID of the side effect.
func (sc *sideEffectContextImpl) SideEffectID() string {
	return sc.sideEffectID
}

// sideEffectBuilderImpl implements the SideEffectBuilder interface.
type sideEffectBuilderImpl struct {
	workflowContext *workflowContextImpl
	sideEffectName  string
}

// Run executes the side effect with the given function and arguments.
func (sb *sideEffectBuilderImpl) Run(fn interface{}, args ...interface{}) SideEffectFuture {
	return &sideEffectFutureImpl{
		sideEffectBuilder: sb,
		sideEffectFn:      fn,
		args:              args,
	}
}

// RunIf executes the side effect if the condition is true.
func (sb *sideEffectBuilderImpl) RunIf(condition bool, fn interface{}, args ...interface{}) SideEffectFuture {
	if !condition {
		return &sideEffectFutureImpl{
			sideEffectBuilder: sb,
			completed:         true,
		}
	}
	return sb.Run(fn, args...)
}

// sideEffectFutureImpl implements the SideEffectFuture interface.
type sideEffectFutureImpl struct {
	sideEffectBuilder *sideEffectBuilderImpl
	sideEffectFn      interface{}
	args              []interface{}
	result            interface{}
	err               error
	completed         bool
}

// Get retrieves the result of the side effect.
func (sf *sideEffectFutureImpl) Get(result interface{}) error {
	if sf.completed {
		// TODO: Use serialization to populate the result with sf.result
		return sf.err
	}

	// Execute the side effect (for first execution)
	// or retrieve the recorded result (for replay)
	// This is synchronous because side effects are deterministic wrappers
	// TODO: Implement actual execution/retrieval

	return nil
}

// IsDone returns whether the side effect has completed.
func (sf *sideEffectFutureImpl) IsDone() bool {
	return sf.completed
}

// Error returns any error that occurred during the side effect.
func (sf *sideEffectFutureImpl) Error() error {
	return sf.err
}

// complete marks the future as complete with the given result and error.
func (sf *sideEffectFutureImpl) complete(result interface{}, err error) {
	sf.result = result
	sf.err = err
	sf.completed = true
}
