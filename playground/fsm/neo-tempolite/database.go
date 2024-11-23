package tempolite

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/stephenfire/go-rtl"
)

var (
	ErrWorkflowEntityNotFound      = errors.New("workflow entity not found")
	ErrActivityEntityNotFound      = errors.New("activity entity not found")
	ErrSagaEntityNotFound          = errors.New("saga entity not found")
	ErrSideEffectEntityNotFound    = errors.New("side effect entity not found")
	ErrWorkflowExecutionNotFound   = errors.New("workflow execution not found")
	ErrActivityExecutionNotFound   = errors.New("activity execution not found")
	ErrSagaExecutionNotFound       = errors.New("saga execution not found")
	ErrSideEffectExecutionNotFound = errors.New("side effect execution not found")
	ErrRunNotFound                 = errors.New("run not found")
	ErrVersionNotFound             = errors.New("version not found")
	ErrHierarchyNotFound           = errors.New("hierarchy not found")
	ErrQueueNotFound               = errors.New("queue not found")
	ErrQueueExists                 = errors.New("queue already exists")
)

// Options types
type WorkflowOptions struct {
	Queue            string
	RetryPolicy      RetryPolicy
	VersionOverrides map[string]int
}

// ContinueAsNewError indicates that the workflow should restart with new inputs.
type ContinueAsNewError struct {
	Options *WorkflowOptions
	Args    []interface{}
}

func (e *ContinueAsNewError) Error() string {
	return "workflow is continuing as new"
}

var ErrPaused = errors.New("execution paused")

// RunStatus defines the status of a run
type RunStatus string

const (
	RunStatusPending   RunStatus = "Pending"
	RunStatusRunning   RunStatus = "Running"
	RunStatusCompleted RunStatus = "Completed"
	RunStatusFailed    RunStatus = "Failed"
	RunStatusCancelled RunStatus = "Cancelled"
)

// EntityType defines the type of the entity
type EntityType string

const (
	EntityTypeWorkflow   EntityType = "Workflow"
	EntityTypeActivity   EntityType = "Activity"
	EntityTypeSaga       EntityType = "Saga"
	EntityTypeSideEffect EntityType = "SideEffect"
)

// ExecutionStatus defines the status of an execution
type ExecutionStatus string

const (
	ExecutionStatusPending     ExecutionStatus = "Pending"
	ExecutionStatusQueued      ExecutionStatus = "Queued"
	ExecutionStatusRunning     ExecutionStatus = "Running"
	ExecutionStatusRetried     ExecutionStatus = "Retried"
	ExecutionStatusPaused      ExecutionStatus = "Paused"
	ExecutionStatusCancelled   ExecutionStatus = "Cancelled"
	ExecutionStatusCompleted   ExecutionStatus = "Completed"
	ExecutionStatusFailed      ExecutionStatus = "Failed"
	ExecutionStatusCompensated ExecutionStatus = "Compensated" // special use case if contained a saga
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

type HandlerIdentity string

func (h HandlerIdentity) String() string {
	return string(h)
}

type HandlerInfo struct {
	HandlerName     string
	HandlerLongName HandlerIdentity
	Handler         interface{}
	ParamsKinds     []reflect.Kind
	ParamTypes      []reflect.Type
	ReturnTypes     []reflect.Type
	ReturnKinds     []reflect.Kind
	NumIn           int
	NumOut          int
}

// Version tracks entity versions
type Version struct {
	ID       int
	EntityID int
	ChangeID string
	Version  int
	Data     map[string]interface{} // Additional data if needed
}

// Run represents a workflow execution group
type Run struct {
	ID          int
	Status      RunStatus // "Pending", "Running", "Completed", etc.
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Entities    []*WorkflowEntity
	Hierarchies []*Hierarchy
}

// Queue represents a work queue
type Queue struct {
	ID        int
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	Entities  []*WorkflowEntity
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
	ParentType        EntityType
	ChildType         EntityType
}

// Base entity types
type BaseEntity struct {
	ID          int
	HandlerName string
	Type        EntityType
	Status      EntityStatus
	StepID      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	RunID       int
	RetryPolicy retryPolicyInternal
	RetryState  RetryState
	HandlerInfo HandlerInfo
}

// Data structs matching the ent schemas
type WorkflowData struct {
	Duration  string   `json:"duration,omitempty"`
	Paused    bool     `json:"paused"`
	Resumable bool     `json:"resumable"`
	Inputs    [][]byte `json:"inputs,omitempty"`
	Attempt   int      `json:"attempt"`
}

type WorkflowEntity struct {
	BaseEntity
	WorkflowData *WorkflowData
}

// ActivityData contains data specific to activity entities
type ActivityData struct {
	Timeout      int64      `json:"timeout,omitempty"`
	MaxAttempts  int        `json:"max_attempts"`
	ScheduledFor *time.Time `json:"scheduled_for,omitempty"`
	Inputs       [][]byte   `json:"inputs,omitempty"`
	Output       [][]byte   `json:"output,omitempty"`
	Attempt      int        `json:"attempt"`
}

type ActivityEntity struct {
	BaseEntity
	ActivityData *ActivityData
}

// SagaData contains data specific to saga entities
type SagaData struct {
	Compensating     bool     `json:"compensating"`
	CompensationData [][]byte `json:"compensation_data,omitempty"`
}

type SagaEntity struct {
	BaseEntity
	SagaData *SagaData
}

// SideEffectData contains data specific to side effect entities
type SideEffectData struct {
	// No fields as per schema
}

type SideEffectEntity struct {
	BaseEntity
	SideEffectData *SideEffectData
}

// Base execution types
type BaseExecution struct {
	ID          int
	EntityID    int
	StartedAt   time.Time
	CompletedAt *time.Time
	Status      ExecutionStatus
	Error       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Attempt     int
}

// ExecutionData structs matching the ent schemas
type WorkflowExecutionData struct {
	LastHeartbeat *time.Time `json:"last_heartbeat,omitempty"`
	Outputs       [][]byte   `json:"outputs,omitempty"`
}

type WorkflowExecution struct {
	BaseExecution
	WorkflowExecutionData *WorkflowExecutionData
}

// ActivityExecutionData contains execution data for activities
type ActivityExecutionData struct {
	LastHeartbeat *time.Time `json:"last_heartbeat,omitempty"`
	Outputs       [][]byte   `json:"outputs,omitempty"`
}

type ActivityExecution struct {
	BaseExecution
	ActivityExecutionData *ActivityExecutionData
}

// SagaExecutionData contains execution data for sagas
type SagaExecutionData struct {
	LastHeartbeat *time.Time `json:"last_heartbeat,omitempty"`
	Output        [][]byte   `json:"output,omitempty"`
	HasOutput     bool       `json:"hasOutput"`
}

type SagaExecution struct {
	BaseExecution
	SagaExecutionData *SagaExecutionData
}

// SideEffectExecutionData contains execution data for side effects
type SideEffectExecutionData struct {
	Outputs [][]byte `json:"outputs,omitempty"`
}

type SideEffectExecution struct {
	BaseExecution
	SideEffectExecutionData *SideEffectExecutionData
}

// Property getters/setters

type RunPropertyGetter func(*Run) error
type RunPropertySetter func(*Run) error

type VersionPropertyGetter func(*Version) error
type VersionPropertySetter func(*Version) error

type HierarchyPropertyGetter func(*Hierarchy) error
type HierarchyPropertySetter func(*Hierarchy) error

type QueuePropertyGetter func(*Queue) error
type QueuePropertySetter func(*Queue) error

type WorkflowEntityPropertyGetter func(*WorkflowEntity) error
type WorkflowEntityPropertySetter func(*WorkflowEntity) error

type ActivityEntityPropertyGetter func(*ActivityEntity) error
type ActivityEntityPropertySetter func(*ActivityEntity) error

type SagaEntityPropertyGetter func(*SagaEntity) error
type SagaEntityPropertySetter func(*SagaEntity) error

type SideEffectEntityPropertyGetter func(*SideEffectEntity) error
type SideEffectEntityPropertySetter func(*SideEffectEntity) error

type WorkflowExecutionPropertyGetter func(*WorkflowExecution) error
type WorkflowExecutionPropertySetter func(*WorkflowExecution) error

type ActivityExecutionPropertyGetter func(*ActivityExecution) error
type ActivityExecutionPropertySetter func(*ActivityExecution) error

type SagaExecutionPropertyGetter func(*SagaExecution) error
type SagaExecutionPropertySetter func(*SagaExecution) error

type SideEffectExecutionPropertyGetter func(*SideEffectExecution) error
type SideEffectExecutionPropertySetter func(*SideEffectExecution) error

// Data property getters/setters
type WorkflowDataPropertyGetter func(*WorkflowData) error
type WorkflowDataPropertySetter func(*WorkflowData) error

type ActivityDataPropertyGetter func(*ActivityData) error
type ActivityDataPropertySetter func(*ActivityData) error

type SagaDataPropertyGetter func(*SagaData) error
type SagaDataPropertySetter func(*SagaData) error

type SideEffectDataPropertyGetter func(*SideEffectData) error
type SideEffectDataPropertySetter func(*SideEffectData) error

type ActivityExecutionDataPropertyGetter func(*ActivityExecutionData) error
type ActivityExecutionDataPropertySetter func(*ActivityExecutionData) error

type WorkflowExecutionDataPropertyGetter func(*WorkflowExecutionData) error
type WorkflowExecutionDataPropertySetter func(*WorkflowExecutionData) error

type SagaExecutionDataPropertyGetter func(*SagaExecutionData) error
type SagaExecutionDataPropertySetter func(*SagaExecutionData) error

type SideEffectExecutionDataPropertyGetter func(*SideEffectExecutionData) error
type SideEffectExecutionDataPropertySetter func(*SideEffectExecutionData) error

// Base Entity property getters/setters
type BaseEntityPropertyGetter func(*BaseEntity) error
type BaseEntityPropertySetter func(*BaseEntity) error

// Execution property getters/setters
type BaseExecutionPropertyGetter func(*BaseExecution) error
type BaseExecutionPropertySetter func(*BaseExecution) error

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
		if idx >= len(executionInputs) {
			return nil, fmt.Errorf("insufficient execution inputs")
		}
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
		if idx >= len(executionOutputs) {
			return nil, fmt.Errorf("insufficient execution outputs")
		}
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

// Database interface
type Database interface {

	// Run operations
	AddRun(run *Run) (int, error)
	GetRun(id int) (*Run, error)
	UpdateRun(run *Run) error
	GetRunProperties(id int, getters ...RunPropertyGetter) error
	SetRunProperties(id int, setters ...RunPropertySetter) error

	// Version operations
	AddVersion(version *Version) (int, error)
	GetVersion(id int) (*Version, error)
	UpdateVersion(version *Version) error
	GetVersionProperties(id int, getters ...VersionPropertyGetter) error
	SetVersionProperties(id int, setters ...VersionPropertySetter) error

	// Hierarchy operations
	AddHierarchy(hierarchy *Hierarchy) (int, error)
	GetHierarchy(id int) (*Hierarchy, error)
	UpdateHierarchy(hierarchy *Hierarchy) error
	GetHierarchyProperties(id int, getters ...HierarchyPropertyGetter) error
	SetHierarchyProperties(id int, setters ...HierarchyPropertySetter) error

	// Queue operations
	AddQueue(queue *Queue) (int, error)
	GetQueue(id int) (*Queue, error)
	UpdateQueue(queue *Queue) error
	GetQueueProperties(id int, getters ...QueuePropertyGetter) error
	SetQueueProperties(id int, setters ...QueuePropertySetter) error

	// Workflow Entity operations
	AddWorkflowEntity(entity *WorkflowEntity) (int, error)
	GetWorkflowEntity(id int) (*WorkflowEntity, error)
	UpdateWorkflowEntity(entity *WorkflowEntity) error
	GetWorkflowEntityProperties(id int, getters ...WorkflowEntityPropertyGetter) error
	SetWorkflowEntityProperties(id int, setters ...WorkflowEntityPropertySetter) error

	// Activity Entity operations
	AddActivityEntity(entity *ActivityEntity) (int, error)
	GetActivityEntity(id int) (*ActivityEntity, error)
	UpdateActivityEntity(entity *ActivityEntity) error
	GetActivityEntityProperties(id int, getters ...ActivityEntityPropertyGetter) error
	SetActivityEntityProperties(id int, setters ...ActivityEntityPropertySetter) error

	// Saga Entity operations
	AddSagaEntity(entity *SagaEntity) (int, error)
	GetSagaEntity(id int) (*SagaEntity, error)
	UpdateSagaEntity(entity *SagaEntity) error
	GetSagaEntityProperties(id int, getters ...SagaEntityPropertyGetter) error
	SetSagaEntityProperties(id int, setters ...SagaEntityPropertySetter) error

	// SideEffect Entity operations
	AddSideEffectEntity(entity *SideEffectEntity) (int, error)
	GetSideEffectEntity(id int) (*SideEffectEntity, error)
	UpdateSideEffectEntity(entity *SideEffectEntity) error
	GetSideEffectEntityProperties(id int, getters ...SideEffectEntityPropertyGetter) error
	SetSideEffectEntityProperties(id int, setters ...SideEffectEntityPropertySetter) error

	// Workflow Execution operations
	AddWorkflowExecution(exec *WorkflowExecution) (int, error)
	GetWorkflowExecution(id int) (*WorkflowExecution, error)
	GetWorkflowExecutionProperties(id int, getters ...WorkflowExecutionPropertyGetter) error
	SetWorkflowExecutionProperties(id int, setters ...WorkflowExecutionPropertySetter) error

	// Activity Execution operations
	AddActivityExecution(exec *ActivityExecution) (int, error)
	GetActivityExecution(id int) (*ActivityExecution, error)
	GetActivityExecutionProperties(id int, getters ...ActivityExecutionPropertyGetter) error
	SetActivityExecutionProperties(id int, setters ...ActivityExecutionPropertySetter) error

	// Saga Execution operations
	AddSagaExecution(exec *SagaExecution) (int, error)
	GetSagaExecution(id int) (*SagaExecution, error)
	GetSagaExecutionProperties(id int, getters ...SagaExecutionPropertyGetter) error
	SetSagaExecutionProperties(id int, setters ...SagaExecutionPropertySetter) error

	// SideEffect Execution operations
	AddSideEffectExecution(exec *SideEffectExecution) (int, error)
	GetSideEffectExecution(id int) (*SideEffectExecution, error)
	GetSideEffectExecutionProperties(id int, getters ...SideEffectExecutionPropertyGetter) error
	SetSideEffectExecutionProperties(id int, setters ...SideEffectExecutionPropertySetter) error

	// Data properties operations
	GetWorkflowDataProperties(entityID int, getters ...WorkflowDataPropertyGetter) error
	SetWorkflowDataProperties(entityID int, setters ...WorkflowDataPropertySetter) error
	GetActivityDataProperties(entityID int, getters ...ActivityDataPropertyGetter) error
	SetActivityDataProperties(entityID int, setters ...ActivityDataPropertySetter) error
	GetSagaDataProperties(entityID int, getters ...SagaDataPropertyGetter) error
	SetSagaDataProperties(entityID int, setters ...SagaDataPropertySetter) error
	GetSideEffectDataProperties(entityID int, getters ...SideEffectDataPropertyGetter) error
	SetSideEffectDataProperties(entityID int, setters ...SideEffectDataPropertySetter) error

	// Batch operations
	BatchGetQueueProperties(ids []int, getters ...QueuePropertyGetter) error
	BatchSetQueueProperties(ids []int, setters ...QueuePropertySetter) error
	BatchGetHierarchyProperties(ids []int, getters ...HierarchyPropertyGetter) error
	BatchSetHierarchyProperties(ids []int, setters ...HierarchyPropertySetter) error
	BatchGetRunProperties(ids []int, getters ...RunPropertyGetter) error
	BatchSetRunProperties(ids []int, setters ...RunPropertySetter) error
	BatchGetVersionProperties(ids []int, getters ...VersionPropertyGetter) error
	BatchSetVersionProperties(ids []int, setters ...VersionPropertySetter) error

	BatchGetWorkflowEntityProperties(ids []int, getters ...WorkflowEntityPropertyGetter) error
	BatchSetWorkflowEntityProperties(ids []int, setters ...WorkflowEntityPropertySetter) error
	BatchGetActivityEntityProperties(ids []int, getters ...ActivityEntityPropertyGetter) error
	BatchSetActivityEntityProperties(ids []int, setters ...ActivityEntityPropertySetter) error
	BatchGetSagaEntityProperties(ids []int, getters ...SagaEntityPropertyGetter) error
	BatchSetSagaEntityProperties(ids []int, setters ...SagaEntityPropertySetter) error
	BatchGetSideEffectEntityProperties(ids []int, getters ...SideEffectEntityPropertyGetter) error
	BatchSetSideEffectEntityProperties(ids []int, setters ...SideEffectEntityPropertySetter) error

	BatchGetWorkflowExecutionProperties(ids []int, getters ...WorkflowExecutionPropertyGetter) error
	BatchSetWorkflowExecutionProperties(ids []int, setters ...WorkflowExecutionPropertySetter) error
	BatchGetActivityExecutionProperties(ids []int, getters ...ActivityExecutionPropertyGetter) error
	BatchSetActivityExecutionProperties(ids []int, setters ...ActivityExecutionPropertySetter) error
	BatchGetSagaExecutionProperties(ids []int, getters ...SagaExecutionPropertyGetter) error
	BatchSetSagaExecutionProperties(ids []int, setters ...SagaExecutionPropertySetter) error
	BatchGetSideEffectExecutionProperties(ids []int, getters ...SideEffectExecutionPropertyGetter) error
	BatchSetSideEffectExecutionProperties(ids []int, setters ...SideEffectExecutionPropertySetter) error
}

func DefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxAttempts:        1,
		InitialInterval:    time.Second,
		BackoffCoefficient: 2.0,
		MaxInterval:        5 * time.Minute,
	}
}

func DefaultRetryPolicyInternal() *retryPolicyInternal {
	return &retryPolicyInternal{
		MaxAttempts:        1,
		InitialInterval:    time.Second.Nanoseconds(),
		BackoffCoefficient: 2.0,
		MaxInterval:        (5 * time.Minute).Nanoseconds(),
	}
}

func ToInternalRetryPolicy(rp *RetryPolicy) *retryPolicyInternal {
	if rp == nil {
		return DefaultRetryPolicyInternal()
	}

	// Fill default values if zero
	if rp.MaxAttempts == 0 {
		rp.MaxAttempts = 1
	}
	if rp.InitialInterval == 0 {
		rp.InitialInterval = time.Second
	}
	if rp.BackoffCoefficient == 0 {
		rp.BackoffCoefficient = 2.0
	}
	if rp.MaxInterval == 0 {
		rp.MaxInterval = 5 * time.Minute
	}

	return &retryPolicyInternal{
		MaxAttempts:        rp.MaxAttempts,
		InitialInterval:    rp.InitialInterval.Nanoseconds(),
		BackoffCoefficient: rp.BackoffCoefficient,
		MaxInterval:        rp.MaxInterval.Nanoseconds(),
	}
}

func copyRun(run *Run) *Run {
	if run == nil {
		return nil
	}

	copy := &Run{
		ID:        run.ID,
		Status:    run.Status,
		CreatedAt: run.CreatedAt,
		UpdatedAt: run.UpdatedAt,
	}

	if run.Entities != nil {
		copy.Entities = make([]*WorkflowEntity, len(run.Entities))
		for i, entity := range run.Entities {
			copy.Entities[i] = copyWorkflowEntity(entity)
		}
	}

	if run.Hierarchies != nil {
		copy.Hierarchies = make([]*Hierarchy, len(run.Hierarchies))
		for i, hierarchy := range run.Hierarchies {
			copy.Hierarchies[i] = copyHierarchy(hierarchy)
		}
	}

	return copy
}

func copyVersion(version *Version) *Version {
	if version == nil {
		return nil
	}

	copy := &Version{
		ID:       version.ID,
		EntityID: version.EntityID,
		ChangeID: version.ChangeID,
		Version:  version.Version,
	}

	if version.Data != nil {
		copy.Data = make(map[string]interface{}, len(version.Data))
		for k, v := range version.Data {
			copy.Data[k] = v
		}
	}

	return copy
}

func copyHierarchy(hierarchy *Hierarchy) *Hierarchy {
	if hierarchy == nil {
		return nil
	}

	return &Hierarchy{
		ID:                hierarchy.ID,
		RunID:             hierarchy.RunID,
		ParentEntityID:    hierarchy.ParentEntityID,
		ChildEntityID:     hierarchy.ChildEntityID,
		ParentExecutionID: hierarchy.ParentExecutionID,
		ChildExecutionID:  hierarchy.ChildExecutionID,
		ParentStepID:      hierarchy.ParentStepID,
		ChildStepID:       hierarchy.ChildStepID,
		ParentType:        hierarchy.ParentType,
		ChildType:         hierarchy.ChildType,
	}
}

func copyQueue(queue *Queue) *Queue {
	if queue == nil {
		return nil
	}

	copy := &Queue{
		ID:        queue.ID,
		Name:      queue.Name,
		CreatedAt: queue.CreatedAt,
		UpdatedAt: queue.UpdatedAt,
	}

	if queue.Entities != nil {
		copy.Entities = make([]*WorkflowEntity, len(queue.Entities))
		for i, entity := range queue.Entities {
			copy.Entities[i] = copyWorkflowEntity(entity)
		}
	}

	return copy
}

// Copy functions for deep copying
func copyBaseEntity(base *BaseEntity) *BaseEntity {
	if base == nil {
		return nil
	}

	return &BaseEntity{
		ID:          base.ID,
		HandlerName: base.HandlerName,
		Type:        base.Type,
		Status:      base.Status,
		StepID:      base.StepID,
		CreatedAt:   base.CreatedAt,
		UpdatedAt:   base.UpdatedAt,
		RunID:       base.RunID,
		RetryPolicy: base.RetryPolicy, // Value copy
		RetryState:  base.RetryState,  // Value copy
		HandlerInfo: base.HandlerInfo, // Value copy
	}
}

func copyWorkflowData(data *WorkflowData) *WorkflowData {
	if data == nil {
		return nil
	}

	c := *data

	if data.Inputs != nil {
		c.Inputs = make([][]byte, len(data.Inputs))
		for i, input := range data.Inputs {
			inputCopy := make([]byte, len(input))
			copy(inputCopy, input)
			c.Inputs[i] = inputCopy
		}
	}

	return &c
}

func copyActivityData(data *ActivityData) *ActivityData {
	if data == nil {
		return nil
	}

	c := *data

	if data.Inputs != nil {
		c.Inputs = make([][]byte, len(data.Inputs))
		for i, input := range data.Inputs {
			inputCopy := make([]byte, len(input))
			copy(inputCopy, input)
			c.Inputs[i] = inputCopy
		}
	}

	if data.Output != nil {
		c.Output = make([][]byte, len(data.Output))
		for i, output := range data.Output {
			outputCopy := make([]byte, len(output))
			copy(outputCopy, output)
			c.Output[i] = outputCopy
		}
	}

	if data.ScheduledFor != nil {
		scheduledForCopy := *data.ScheduledFor
		c.ScheduledFor = &scheduledForCopy
	}

	return &c
}

func copySagaData(data *SagaData) *SagaData {
	if data == nil {
		return nil
	}

	c := *data

	if data.CompensationData != nil {
		c.CompensationData = make([][]byte, len(data.CompensationData))
		for i, cd := range data.CompensationData {
			cdCopy := make([]byte, len(cd))
			copy(cdCopy, cd)
			c.CompensationData[i] = cdCopy
		}
	}

	return &c
}

func copySideEffectData(data *SideEffectData) *SideEffectData {
	if data == nil {
		return nil
	}
	copy := *data
	return &copy
}

func copyBaseExecution(base *BaseExecution) *BaseExecution {
	if base == nil {
		return nil
	}

	copy := *base

	if base.CompletedAt != nil {
		completedAtCopy := *base.CompletedAt
		copy.CompletedAt = &completedAtCopy
	}

	return &copy
}

func copyWorkflowExecutionData(data *WorkflowExecutionData) *WorkflowExecutionData {
	if data == nil {
		return nil
	}

	c := *data

	if data.LastHeartbeat != nil {
		heartbeatCopy := *data.LastHeartbeat
		c.LastHeartbeat = &heartbeatCopy
	}

	if data.Outputs != nil {
		c.Outputs = make([][]byte, len(data.Outputs))
		for i, output := range data.Outputs {
			outputCopy := make([]byte, len(output))
			copy(outputCopy, output)
			c.Outputs[i] = outputCopy
		}
	}

	return &c
}

func copyActivityExecutionData(data *ActivityExecutionData) *ActivityExecutionData {
	if data == nil {
		return nil
	}

	c := *data

	if data.LastHeartbeat != nil {
		heartbeatCopy := *data.LastHeartbeat
		c.LastHeartbeat = &heartbeatCopy
	}

	if data.Outputs != nil {
		c.Outputs = make([][]byte, len(data.Outputs))
		for i, output := range data.Outputs {
			outputCopy := make([]byte, len(output))
			copy(outputCopy, output)
			c.Outputs[i] = outputCopy
		}
	}

	return &c
}

func copySagaExecutionData(data *SagaExecutionData) *SagaExecutionData {
	if data == nil {
		return nil
	}

	c := *data

	if data.LastHeartbeat != nil {
		heartbeatCopy := *data.LastHeartbeat
		c.LastHeartbeat = &heartbeatCopy
	}

	if data.Output != nil {
		c.Output = make([][]byte, len(data.Output))
		for i, output := range data.Output {
			outputCopy := make([]byte, len(output))
			copy(outputCopy, output)
			c.Output[i] = outputCopy
		}
	}

	return &c
}

func copySideEffectExecutionData(data *SideEffectExecutionData) *SideEffectExecutionData {
	if data == nil {
		return nil
	}

	c := *data

	if data.Outputs != nil {
		c.Outputs = make([][]byte, len(data.Outputs))
		for i, output := range data.Outputs {
			outputCopy := make([]byte, len(output))
			copy(outputCopy, output)
			c.Outputs[i] = outputCopy
		}
	}

	return &c
}

// Entity copy functions
func copyWorkflowEntity(entity *WorkflowEntity) *WorkflowEntity {
	if entity == nil {
		return nil
	}

	copy := WorkflowEntity{
		BaseEntity:   *copyBaseEntity(&entity.BaseEntity),
		WorkflowData: copyWorkflowData(entity.WorkflowData),
	}

	return &copy
}

func copyActivityEntity(entity *ActivityEntity) *ActivityEntity {
	if entity == nil {
		return nil
	}

	copy := ActivityEntity{
		BaseEntity:   *copyBaseEntity(&entity.BaseEntity),
		ActivityData: copyActivityData(entity.ActivityData),
	}

	return &copy
}

func copySagaEntity(entity *SagaEntity) *SagaEntity {
	if entity == nil {
		return nil
	}

	copy := SagaEntity{
		BaseEntity: *copyBaseEntity(&entity.BaseEntity),
		SagaData:   copySagaData(entity.SagaData),
	}

	return &copy
}

func copySideEffectEntity(entity *SideEffectEntity) *SideEffectEntity {
	if entity == nil {
		return nil
	}

	copy := SideEffectEntity{
		BaseEntity:     *copyBaseEntity(&entity.BaseEntity),
		SideEffectData: copySideEffectData(entity.SideEffectData),
	}

	return &copy
}

// Execution copy functions
func copyWorkflowExecution(exec *WorkflowExecution) *WorkflowExecution {
	if exec == nil {
		return nil
	}

	copy := WorkflowExecution{
		BaseExecution:         *copyBaseExecution(&exec.BaseExecution),
		WorkflowExecutionData: copyWorkflowExecutionData(exec.WorkflowExecutionData),
	}

	return &copy
}

func copyActivityExecution(exec *ActivityExecution) *ActivityExecution {
	if exec == nil {
		return nil
	}

	copy := ActivityExecution{
		BaseExecution:         *copyBaseExecution(&exec.BaseExecution),
		ActivityExecutionData: copyActivityExecutionData(exec.ActivityExecutionData),
	}

	return &copy
}

func copySagaExecution(exec *SagaExecution) *SagaExecution {
	if exec == nil {
		return nil
	}

	copy := SagaExecution{
		BaseExecution:     *copyBaseExecution(&exec.BaseExecution),
		SagaExecutionData: copySagaExecutionData(exec.SagaExecutionData),
	}

	return &copy
}

func copySideEffectExecution(exec *SideEffectExecution) *SideEffectExecution {
	if exec == nil {
		return nil
	}

	copy := SideEffectExecution{
		BaseExecution:           *copyBaseExecution(&exec.BaseExecution),
		SideEffectExecutionData: copySideEffectExecutionData(exec.SideEffectExecutionData),
	}

	return &copy
}
