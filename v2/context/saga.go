// Package context provides the execution context for workflows and activities.
package context

// SagaBuilder provides a fluent API for configuring and executing sagas.
type SagaBuilder interface {
	// Configuration
	WithOptions(opts SagaOptions) SagaBuilder
	OnCompensation(handler func(error)) SagaBuilder

	// Step definition
	AddStep(name string, txFn interface{}, compFn interface{}, args ...interface{}) SagaBuilder
	Transaction(fn interface{}) SagaTransactionBuilder
	Compensation(fn interface{}) SagaCompensationBuilder

	// Execution
	Execute() SagaResult
}

// SagaOptions contains configuration options for saga execution.
type SagaOptions struct {
	// Parallel determines whether saga steps can run in parallel
	Parallel bool

	// StrictMode enforces strict error handling
	StrictMode bool
}

// SagaTransactionBuilder provides a fluent API for configuring saga transactions.
type SagaTransactionBuilder interface {
	WithArgs(args ...interface{}) SagaTransactionBuilder
}

// SagaCompensationBuilder provides a fluent API for configuring saga compensations.
type SagaCompensationBuilder interface {
	WithArgs(args ...interface{}) SagaCompensationBuilder
}

// SagaResult represents the result of a saga execution.
type SagaResult interface {
	// Status checks
	IsCompleted() bool
	IsCompensated() bool
	Error() error

	// Result access
	StepResults() map[string]interface{}
}

// SagaContext is the context provided to saga transaction and compensation functions.
type SagaContext interface {
	// Context information
	WorkflowID() string
	SagaID() string
	StepName() string
}

// SagaTransactionContext is the context provided to saga transaction functions.
type SagaTransactionContext interface {
	SagaContext
}

// SagaCompensationContext is the context provided to saga compensation functions.
type SagaCompensationContext interface {
	SagaContext

	// Transaction result access
	GetTransactionResult(result interface{}) error
}

// sagaBuilderImpl implements the SagaBuilder interface.
type sagaBuilderImpl struct {
	workflowContext     *workflowContextImpl
	sagaName            string
	options             SagaOptions
	compensationHandler func(error)
	steps               []sagaStep
}

// sagaStep represents a single step in a saga.
type sagaStep struct {
	name           string
	transactionFn  interface{}
	compensationFn interface{}
	args           []interface{}
}

// WithOptions configures the options for the saga.
func (sb *sagaBuilderImpl) WithOptions(opts SagaOptions) SagaBuilder {
	sb.options = opts
	return sb
}

// OnCompensation configures the compensation handler for the saga.
func (sb *sagaBuilderImpl) OnCompensation(handler func(error)) SagaBuilder {
	sb.compensationHandler = handler
	return sb
}

// AddStep adds a step to the saga with transaction and compensation functions.
func (sb *sagaBuilderImpl) AddStep(name string, txFn interface{}, compFn interface{}, args ...interface{}) SagaBuilder {
	sb.steps = append(sb.steps, sagaStep{
		name:           name,
		transactionFn:  txFn,
		compensationFn: compFn,
		args:           args,
	})
	return sb
}

// Transaction creates a transaction builder for the given function.
func (sb *sagaBuilderImpl) Transaction(fn interface{}) SagaTransactionBuilder {
	return &sagaTransactionBuilderImpl{
		sagaBuilder: sb,
		fn:          fn,
	}
}

// Compensation creates a compensation builder for the given function.
func (sb *sagaBuilderImpl) Compensation(fn interface{}) SagaCompensationBuilder {
	return &sagaCompensationBuilderImpl{
		sagaBuilder: sb,
		fn:          fn,
	}
}

// Execute executes the saga.
func (sb *sagaBuilderImpl) Execute() SagaResult {
	// Create a new saga result
	result := &sagaResultImpl{
		completed:   false,
		compensated: false,
		stepResults: make(map[string]interface{}),
	}

	// Execute the saga steps
	// This is a placeholder implementation
	// In a real implementation, this would execute each step and handle compensations

	return result
}

// sagaTransactionBuilderImpl implements the SagaTransactionBuilder interface.
type sagaTransactionBuilderImpl struct {
	sagaBuilder *sagaBuilderImpl
	fn          interface{}
	args        []interface{}
}

// WithArgs configures the arguments for the transaction function.
func (stb *sagaTransactionBuilderImpl) WithArgs(args ...interface{}) SagaTransactionBuilder {
	stb.args = args
	return stb
}

// sagaCompensationBuilderImpl implements the SagaCompensationBuilder interface.
type sagaCompensationBuilderImpl struct {
	sagaBuilder *sagaBuilderImpl
	fn          interface{}
	args        []interface{}
}

// WithArgs configures the arguments for the compensation function.
func (scb *sagaCompensationBuilderImpl) WithArgs(args ...interface{}) SagaCompensationBuilder {
	scb.args = args
	return scb
}

// sagaContextImpl implements the SagaContext interface.
type sagaContextImpl struct {
	workflowID string
	sagaID     string
	stepName   string
}

// WorkflowID returns the ID of the workflow that initiated the saga.
func (sc *sagaContextImpl) WorkflowID() string {
	return sc.workflowID
}

// SagaID returns the ID of the saga.
func (sc *sagaContextImpl) SagaID() string {
	return sc.sagaID
}

// StepName returns the name of the current step.
func (sc *sagaContextImpl) StepName() string {
	return sc.stepName
}

// sagaTransactionContextImpl implements the SagaTransactionContext interface.
type sagaTransactionContextImpl struct {
	sagaContextImpl
}

// NewSagaTransactionContext creates a new saga transaction context.
func NewSagaTransactionContext(workflowID, sagaID, stepName string) SagaTransactionContext {
	return &sagaTransactionContextImpl{
		sagaContextImpl: sagaContextImpl{
			workflowID: workflowID,
			sagaID:     sagaID,
			stepName:   stepName,
		},
	}
}

// sagaCompensationContextImpl implements the SagaCompensationContext interface.
type sagaCompensationContextImpl struct {
	sagaContextImpl
	transactionResult interface{}
}

// NewSagaCompensationContext creates a new saga compensation context.
func NewSagaCompensationContext(workflowID, sagaID, stepName string, transactionResult interface{}) SagaCompensationContext {
	return &sagaCompensationContextImpl{
		sagaContextImpl: sagaContextImpl{
			workflowID: workflowID,
			sagaID:     sagaID,
			stepName:   stepName,
		},
		transactionResult: transactionResult,
	}
}

// GetTransactionResult retrieves the result of the transaction.
func (sc *sagaCompensationContextImpl) GetTransactionResult(result interface{}) error {
	// TODO: Use serialization to populate the result with sc.transactionResult
	return nil
}

// sagaResultImpl implements the SagaResult interface.
type sagaResultImpl struct {
	completed   bool
	compensated bool
	err         error
	stepResults map[string]interface{}
}

// IsCompleted returns whether the saga completed successfully.
func (sr *sagaResultImpl) IsCompleted() bool {
	return sr.completed
}

// IsCompensated returns whether the saga was compensated.
func (sr *sagaResultImpl) IsCompensated() bool {
	return sr.compensated
}

// Error returns any error that occurred during the saga execution.
func (sr *sagaResultImpl) Error() error {
	return sr.err
}

// StepResults returns the results of the saga steps.
func (sr *sagaResultImpl) StepResults() map[string]interface{} {
	return sr.stepResults
}
