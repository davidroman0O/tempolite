package tempolite

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/stephenfire/go-rtl"
)

// RetryPolicy defines retry behavior configuration
type RetryPolicy struct {
	MaxAttempts        int
	InitialInterval    time.Duration
	BackoffCoefficient float64
	MaxInterval        time.Duration
}

// RetryState tracks the state of retries
type RetryState struct {
	Attempts int `json:"attempts"`
}

// retryPolicyInternal matches the database schema, using int64 for intervals
type retryPolicyInternal struct {
	MaxAttempts        int     `json:"max_attempts"`
	InitialInterval    int64   `json:"initial_interval"`
	BackoffCoefficient float64 `json:"backoff_coefficient"`
	MaxInterval        int64   `json:"max_interval"`
}

var DefaultVersion int = 0

// Version tracks entity versions
type Version struct {
	ID       int
	EntityID int
	ChangeID string
	Version  int
	Data     map[string]interface{} // Additional data if needed
}

// Hierarchy tracks parent-child relationships between entities
type Hierarchy struct {
	ID                int
	RunID             int
	ParentEntityID    int
	ChildEntityID     int
	ParentExecutionID int
	ChildExecutionID  int
	ParentStepID      string
	ChildStepID       string
	ParentType        string
	ChildType         string
}

// Run represents a workflow execution group
type Run struct {
	ID          int
	Status      string // "Pending", "Running", "Completed", etc.
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Entities    []*Entity
	Hierarchies []*Hierarchy
}

// EntityType defines the type of the entity
type EntityType string

const (
	EntityTypeWorkflow   EntityType = "Workflow"
	EntityTypeActivity   EntityType = "Activity"
	EntityTypeSaga       EntityType = "Saga"
	EntityTypeSideEffect EntityType = "SideEffect"
)

// EntityStatus defines the status of the entity
type EntityStatus string

const (
	StatusPending   EntityStatus = "Pending"
	StatusQueued    EntityStatus = "Queued"
	StatusRunning   EntityStatus = "Running"
	StatusPaused    EntityStatus = "Paused"
	StatusCancelled EntityStatus = "Cancelled"
	StatusCompleted EntityStatus = "Completed"
	StatusFailed    EntityStatus = "Failed"
)

// Entity represents an executable component
type Entity struct {
	ID             int
	HandlerName    string
	Type           EntityType
	Status         EntityStatus
	StepID         string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	RunID          int
	Run            *Run
	Executions     []*Execution
	QueueID        int    // optional
	Queue          *Queue // optional
	Versions       []*Version
	WorkflowData   *WorkflowData   // optional
	ActivityData   *ActivityData   // optional
	SagaData       *SagaData       // optional
	SideEffectData *SideEffectData // optional
	HandlerInfo    *HandlerInfo    // handler metadata
	RetryState     *RetryState
	RetryPolicy    *retryPolicyInternal
	Paused         bool
	Resumable      bool
}

// ExecutionStatus defines the status of an execution
type ExecutionStatus string

const (
	ExecutionStatusPending   ExecutionStatus = "Pending"
	ExecutionStatusQueued    ExecutionStatus = "Queued"
	ExecutionStatusRunning   ExecutionStatus = "Running"
	ExecutionStatusRetried   ExecutionStatus = "Retried"
	ExecutionStatusPaused    ExecutionStatus = "Paused"
	ExecutionStatusCancelled ExecutionStatus = "Cancelled"
	ExecutionStatusCompleted ExecutionStatus = "Completed"
	ExecutionStatusFailed    ExecutionStatus = "Failed"
)

// Execution represents a single execution attempt
type Execution struct {
	ID                      int
	EntityID                int
	Entity                  *Entity
	StartedAt               time.Time
	CompletedAt             *time.Time
	Status                  ExecutionStatus
	Error                   string
	CreatedAt               time.Time
	UpdatedAt               time.Time
	WorkflowExecutionData   *WorkflowExecutionData   // optional
	ActivityExecutionData   *ActivityExecutionData   // optional
	SagaExecutionData       *SagaExecutionData       // optional
	SideEffectExecutionData *SideEffectExecutionData // optional
	Attempt                 int
}

// Queue represents a work queue
type Queue struct {
	ID        int
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	Entities  []*Entity
}

type ActivityData struct {
	Timeout      int64      `json:"timeout,omitempty"`
	MaxAttempts  int        `json:"max_attempts"`
	ScheduledFor *time.Time `json:"scheduled_for,omitempty"`
	Input        [][]byte   `json:"input,omitempty"`
	Output       [][]byte   `json:"output,omitempty"`
	Attempt      int        `json:"attempt"`
}

type SagaData struct {
	Compensating     bool     `json:"compensating"`
	CompensationData [][]byte `json:"compensation_data,omitempty"`
}

type SideEffectData struct {
	// No fields as per ent schema
}

type ActivityExecutionData struct {
	LastHeartbeat *time.Time `json:"last_heartbeat,omitempty"`
	Outputs       [][]byte   `json:"outputs,omitempty"`
}

type SagaExecutionData struct {
	LastHeartbeat *time.Time `json:"last_heartbeat,omitempty"`
	Output        [][]byte   `json:"output,omitempty"`
	HasOutput     bool       `json:"hasOutput"`
}

type SideEffectExecutionData struct {
	Outputs [][]byte `json:"outputs,omitempty"`
}

type RunInfo struct {
	RunID  int
	Status string
}

type WorkflowInfo struct {
	EntityID int
	Status   EntityStatus
	Run      *RunInfo
}

// Serialization functions

func convertInputsForSerialization(executionInputs []interface{}) ([][]byte, error) {
	inputs := [][]byte{}

	for _, input := range executionInputs {
		buf := new(bytes.Buffer)

		// Get the real value
		if reflect.TypeOf(input).Kind() == reflect.Ptr {
			input = reflect.ValueOf(input).Elem().Interface()
		}

		if err := rtl.Encode(input, buf); err != nil {
			return nil, err
		}
		inputs = append(inputs, buf.Bytes())
	}

	return inputs, nil
}

func convertOutputsForSerialization(executionOutputs []interface{}) ([][]byte, error) {
	outputs := [][]byte{}

	for _, output := range executionOutputs {
		buf := new(bytes.Buffer)

		// Get the real value
		if reflect.TypeOf(output).Kind() == reflect.Ptr {
			output = reflect.ValueOf(output).Elem().Interface()
		}

		if err := rtl.Encode(output, buf); err != nil {
			return nil, err
		}
		outputs = append(outputs, buf.Bytes())
	}

	return outputs, nil
}

func convertInputsFromSerialization(handlerInfo HandlerInfo, executionInputs [][]byte) ([]interface{}, error) {
	inputs := []interface{}{}

	for idx, inputType := range handlerInfo.ParamTypes {
		buf := bytes.NewBuffer(executionInputs[idx])

		// Get the pointer of the type of the parameter that we target
		decodedObj := reflect.New(inputType).Interface()

		if err := rtl.Decode(buf, decodedObj); err != nil {
			return nil, err
		}

		inputs = append(inputs, reflect.ValueOf(decodedObj).Elem().Interface())
	}

	return inputs, nil
}

func convertOutputsFromSerialization(handlerInfo HandlerInfo, executionOutputs [][]byte) ([]interface{}, error) {
	output := []interface{}{}

	for idx, outputType := range handlerInfo.ReturnTypes {
		buf := bytes.NewBuffer(executionOutputs[idx])

		// Get the pointer of the type of the parameter that we target
		decodedObj := reflect.New(outputType).Interface()

		if err := rtl.Decode(buf, decodedObj); err != nil {
			return nil, err
		}

		output = append(output, reflect.ValueOf(decodedObj).Elem().Interface())
	}

	return output, nil
}

func convertSingleOutputFromSerialization(outputType reflect.Type, executionOutput []byte) (interface{}, error) {
	buf := bytes.NewBuffer(executionOutput)

	decodedObj := reflect.New(outputType).Interface()

	if err := rtl.Decode(buf, decodedObj); err != nil {
		return nil, err
	}

	return reflect.ValueOf(decodedObj).Elem().Interface(), nil
}

// Future represents an asynchronous result.
type Future struct {
	results    []interface{}
	err        error
	done       chan struct{}
	workflowID int
}

func (f *Future) WorkflowID() int {
	return f.workflowID
}

func NewFuture(workflowID int) *Future {
	return &Future{
		done:       make(chan struct{}),
		workflowID: workflowID,
	}
}

func (f *Future) Get(out ...interface{}) error {
	<-f.done
	if f.err != nil {
		return f.err
	}

	if len(out) == 0 {
		return nil
	}

	// Handle the case where we have a single result
	if len(f.results) == 1 && len(out) == 1 {
		val := reflect.ValueOf(out[0])
		if val.Kind() != reflect.Ptr {
			return fmt.Errorf("output parameter must be a pointer")
		}
		val = val.Elem()

		result := reflect.ValueOf(f.results[0])
		if !result.Type().AssignableTo(val.Type()) {
			return fmt.Errorf("cannot assign type %v to %v", result.Type(), val.Type())
		}

		val.Set(result)
		return nil
	}

	if len(out) > len(f.results) {
		return fmt.Errorf("number of outputs (%d) exceeds number of results (%d)", len(out), len(f.results))
	}

	for i := 0; i < len(out); i++ {
		val := reflect.ValueOf(out[i])
		if val.Kind() != reflect.Ptr {
			return fmt.Errorf("output parameter %d must be a pointer", i)
		}
		val = val.Elem()

		result := reflect.ValueOf(f.results[i])
		if !result.Type().AssignableTo(val.Type()) {
			return fmt.Errorf("cannot assign type %v to %v for parameter %d", result.Type(), val.Type(), i)
		}

		val.Set(result)
	}

	return nil
}

func (f *Future) setResult(results []interface{}) {
	log.Printf("Future.setResult called with results: %v", results)
	f.results = results
	close(f.done)
}

func (f *Future) setError(err error) {
	log.Printf("Future.setError called with error: %v", err)
	f.err = err
	close(f.done)
}

// ActivityOptions provides options for activities, including retry policies.
type ActivityOptions struct {
	RetryPolicy *RetryPolicy
}

var ErrPaused = errors.New("execution paused")

// ContinueAsNewError indicates that the workflow should restart with new inputs.
type ContinueAsNewError struct {
	Options *WorkflowOptions
	Args    []interface{}
}

func (e *ContinueAsNewError) Error() string {
	return "workflow is continuing as new"
}
