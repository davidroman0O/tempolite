package tempolite

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/stephenfire/go-rtl"
)

// Error definitions
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
	ErrSagaValueNotFound           = errors.New("saga value not found")
)

// State and Trigger definitions
type state string

const (
	StateIdle          state = "Idle"
	StateExecuting     state = "Executing"
	StateCompleted     state = "Completed"
	StateFailed        state = "Failed"
	StateRetried       state = "Retried"
	StatePaused        state = "Paused"
	StateTransactions  state = "Transactions"
	StateCompensations state = "Compensations"
	StateFinalize      state = "Finalize"
)

type trigger string

const (
	TriggerStart      trigger = "Start"
	TriggerComplete   trigger = "Complete"
	TriggerFail       trigger = "Fail"
	TriggerPause      trigger = "Pause"
	TriggerResume     trigger = "Resume"
	TriggerCompensate trigger = "Compensate"
	TriggerFinalize   trigger = "Finalize"
)

// Core types and constants remain the same
var DefaultVersion int = 0

var DefaultQueue string = "default"

type WorkflowOutput struct {
	Outputs []interface{}

	ContinueAsNewOptions *WorkflowOptions
	ContinueAsNewArgs    []interface{}
	Paused               bool
	StrackTrace          *string
}

type ActivityOutput struct {
	Outputs     []interface{}
	StrackTrace *string
	Paused      bool
}

// Status and Type enums remain the same as before
type RunStatus string

const (
	RunStatusPending   RunStatus = "Pending"
	RunStatusRunning   RunStatus = "Running"
	RunStatusPaused    RunStatus = "Paused"
	RunStatusCancelled RunStatus = "Cancelled"
	RunStatusCompleted RunStatus = "Completed"
	RunStatusFailed    RunStatus = "Failed"
)

type EntityType string

var (
	EntityWorkflow   EntityType = "workflow"
	EntityActivity   EntityType = "activity"
	EntitySaga       EntityType = "saga"
	EntitySideEffect EntityType = "side_effect"
	EntitySignal     EntityType = "signal"
)

// EntityStatus defines the status of the entity
type EntityStatus string

const (
	StatusNone        EntityStatus = "None"
	StatusPending     EntityStatus = "Pending"
	StatusQueued      EntityStatus = "Queued"
	StatusRunning     EntityStatus = "Running"
	StatusPaused      EntityStatus = "Paused"
	StatusCancelled   EntityStatus = "Cancelled"
	StatusCompleted   EntityStatus = "Completed"
	StatusFailed      EntityStatus = "Failed"
	StatusCompensated EntityStatus = "Compensated"
)

// ExecutionStatus defines the status of an execution
type ExecutionStatus string

const (
	ExecutionStatusNone        ExecutionStatus = "None"
	ExecutionStatusPending     ExecutionStatus = "Pending"
	ExecutionStatusQueued      ExecutionStatus = "Queued"
	ExecutionStatusRunning     ExecutionStatus = "Running"
	ExecutionStatusRetried     ExecutionStatus = "Retried"
	ExecutionStatusPaused      ExecutionStatus = "Paused"
	ExecutionStatusCancelled   ExecutionStatus = "Cancelled"
	ExecutionStatusCompleted   ExecutionStatus = "Completed"
	ExecutionStatusFailed      ExecutionStatus = "Failed"
	ExecutionStatusCompensated ExecutionStatus = "Compensated"
)

// ID type definitions
type WorkflowEntityID int
type WorkflowExecutionID int
type WorkflowDataID int
type WorkflowExecutionDataID int
type ActivityEntityID int
type ActivityExecutionID int
type ActivityDataID int
type ActivityExecutionDataID int
type SagaEntityID int
type SagaExecutionID int
type SagaDataID int
type SagaExecutionDataID int
type SideEffectEntityID int
type SideEffectExecutionID int
type SideEffectDataID int
type SideEffectExecutionDataID int
type SignalEntityID int
type SignalExecutionID int
type SignalDataID int
type SignalExecutionDataID int
type SagaValueID int
type RunID int
type VersionID int
type HierarchyID int
type QueueID int

// Add to existing ExecutionType
type ExecutionType string

const (
	ExecutionTypeTransaction  ExecutionType = "transaction"
	ExecutionTypeCompensation ExecutionType = "compensation"
)

// Base entity options
type BaseEntityGetterOptions struct {
	// Add flags if needed
}

type BaseEntitySetterOptions struct {
	// Add options if needed
}

// Base execution options
type BaseExecutionGetterOptions struct {
	IncludeError     bool
	IncludeStarted   bool
	IncludeCompleted bool
	// Add more flags if needed
}

type BaseExecutionSetterOptions struct {
	// Add options if needed
}

// Options structures for all property operations
type RunGetterOptions struct {
	IncludeWorkflows   bool
	IncludeHierarchies bool
	// Add more flags as needed
}

type RunSetterOptions struct {
	WorkflowID *WorkflowEntityID
	// Add more options as needed
}

type WorkflowEntityGetterOptions struct {
	IncludeVersion  bool
	IncludeQueue    bool
	IncludeChildren bool
	IncludeRun      bool
	IncludeData     bool
	// Add more flags as needed
}

type WorkflowEntitySetterOptions struct {
	QueueID   *QueueID
	QueueName *string
	Version   *Version
	ChildID   *int
	ChildType *EntityType
	RunID     *RunID
	// Add more options as needed
}

type ActivityEntityGetterOptions struct {
	IncludeData bool
	// Add more flags as needed
}

type ActivityEntitySetterOptions struct {
	ParentWorkflowID *WorkflowEntityID
	ParentRunID      *RunID
	// Add more options as needed
}

type SagaEntityGetterOptions struct {
	IncludeData bool
	// Add more flags as needed
}

type SagaEntitySetterOptions struct {
	ParentWorkflowID *WorkflowEntityID
	// Add more options as needed
}

type SideEffectEntityGetterOptions struct {
	IncludeData bool
	// Add more flags as needed
}

type SideEffectEntitySetterOptions struct {
	ParentWorkflowID *WorkflowEntityID
	// Add more options as needed
}

type QueueGetterOptions struct {
	IncludeWorkflows bool
	// Add more flags as needed
}

type QueueSetterOptions struct {
	WorkflowIDs []WorkflowEntityID
	// Add more options as needed
}

type VersionGetterOptions struct {
	IncludeData bool
	// Add more flags as needed
}

type VersionSetterOptions struct {
	// Add more options as needed
}

type hierarchyTripleIdentity struct {
	ParentEntityID int        // if a parent
	ChildStepID    string     // got this child step id
	SpecificType   EntityType // of this type
	// then we should have a unique hierarchy, therefore might find a result
}

type HierarchyGetterOptions struct {
	ChildrenEntityOf *hierarchyTripleIdentity
	// Add flags as needed
}

type HierarchySetterOptions struct {
	// Add options as needed
}

// Entity Data options
type WorkflowDataGetterOptions struct {
	IncludeInputs bool
	// Add more flags as needed
}

type WorkflowDataSetterOptions struct {
	// Add options as needed
}

type ActivityDataGetterOptions struct {
	IncludeInputs  bool
	IncludeOutputs bool
	// Add more flags as needed
}

type ActivityDataSetterOptions struct {
	// Add options as needed
}

type SagaDataGetterOptions struct {
	IncludeCompensationData bool
	// Add more flags as needed
}

type SagaDataSetterOptions struct {
	// Add options as needed
}

type SideEffectDataGetterOptions struct {
	// Add flags as needed
}

type SideEffectDataSetterOptions struct {
	// Add options as needed
}

// Execution Data options
type WorkflowExecutionDataGetterOptions struct {
	IncludeOutputs bool
	// Add more flags as needed
}

type WorkflowExecutionDataSetterOptions struct {
	// Add options as needed
}

type ActivityExecutionDataGetterOptions struct {
	IncludeOutputs bool
	// Add more flags as needed
}

type ActivityExecutionDataSetterOptions struct {
	// Add options as needed
}

type SagaExecutionDataGetterOptions struct {
	IncludeOutput bool
	// Add more flags as needed
}

type SagaExecutionDataSetterOptions struct {
	// Add options as needed
}

type SideEffectExecutionDataGetterOptions struct {
	IncludeOutputs bool
	// Add more flags as needed
}

type SideEffectExecutionDataSetterOptions struct {
	// Add options as needed
}

// New option structs for Execution properties
type WorkflowExecutionGetterOptions struct {
	IncludeData bool
	// Add more options as needed
}

type WorkflowExecutionSetterOptions struct {
	// Add options as needed
}

type ActivityExecutionGetterOptions struct {
	IncludeData bool
	// Add more options as needed
}

type ActivityExecutionSetterOptions struct {
	// Add options as needed
}

type SagaExecutionGetterOptions struct {
	IncludeData bool
	// Add more options as needed
}

type SagaExecutionSetterOptions struct {
	// Add options as needed
}

type SideEffectExecutionGetterOptions struct {
	IncludeData bool
	// Add more options as needed
}

type SideEffectExecutionSetterOptions struct {
	// Add options as needed
}

// Signal entity options
type SignalEntityGetterOptions struct {
	IncludeData bool
}

type SignalEntitySetterOptions struct {
}

type SignalEntityGetOption func(*SignalEntityGetterOptions) error

// Signal execution options
type SignalExecutionGetterOptions struct {
	IncludeData bool
}

type SignalExecutionSetterOptions struct {
}

type SignalExecutionGetOption func(*SignalExecutionGetterOptions) error

// Signal data options
type SignalDataGetterOptions struct {
	IncludeName bool
}

type SignalDataSetterOptions struct {
}

// Signal execution data options
type SignalExecutionDataGetterOptions struct {
	IncludeValue bool
}

type SignalExecutionDataSetterOptions struct {
}

// Option get functions
type WorkflowEntityGetOption func(*WorkflowEntityGetterOptions) error
type ActivityEntityGetOption func(*ActivityEntityGetterOptions) error
type SagaEntityGetOption func(*SagaEntityGetterOptions) error
type SideEffectEntityGetOption func(*SideEffectEntityGetterOptions) error

type WorkflowExecutionGetOption func(*WorkflowExecutionGetterOptions) error
type ActivityExecutionGetOption func(*ActivityExecutionGetterOptions) error
type SagaExecutionGetOption func(*SagaExecutionGetterOptions) error
type SideEffectExecutionGetOption func(*SideEffectExecutionGetterOptions) error

type RunGetOption func(*RunGetterOptions) error
type VersionGetOption func(*VersionGetterOptions) error
type QueueGetOption func(*QueueGetterOptions) error
type HierarchyGetOption func(*HierarchyGetterOptions) error

// Option function types
type BaseEntityPropertyGetterOption func(*BaseEntityGetterOptions) error
type BaseEntityPropertySetterOption func(*BaseEntitySetterOptions) error

type BaseExecutionPropertyGetterOption func(*BaseExecutionGetterOptions) error
type BaseExecutionPropertySetterOption func(*BaseExecutionSetterOptions) error

type RunPropertyGetterOption func(*RunGetterOptions) error
type RunPropertySetterOption func(*RunSetterOptions) error

type WorkflowEntityPropertyGetterOption func(*WorkflowEntityGetterOptions) error
type WorkflowEntityPropertySetterOption func(*WorkflowEntitySetterOptions) error

type ActivityEntityPropertyGetterOption func(*ActivityEntityGetterOptions) error
type ActivityEntityPropertySetterOption func(*ActivityEntitySetterOptions) error

type SagaEntityPropertyGetterOption func(*SagaEntityGetterOptions) error
type SagaEntityPropertySetterOption func(*SagaEntitySetterOptions) error

type SideEffectEntityPropertyGetterOption func(*SideEffectEntityGetterOptions) error
type SideEffectEntityPropertySetterOption func(*SideEffectEntitySetterOptions) error

type QueuePropertyGetterOption func(*QueueGetterOptions) error
type QueuePropertySetterOption func(*QueueSetterOptions) error

type VersionPropertyGetterOption func(*VersionGetterOptions) error
type VersionPropertySetterOption func(*VersionSetterOptions) error

type HierarchyPropertyGetterOption func(*HierarchyGetterOptions) error
type HierarchyPropertySetterOption func(*HierarchySetterOptions) error

// Execution property option types - Adding these new types
type WorkflowExecutionPropertyGetterOption func(*WorkflowExecutionGetterOptions) error
type WorkflowExecutionPropertySetterOption func(*WorkflowExecutionSetterOptions) error

type ActivityExecutionPropertyGetterOption func(*ActivityExecutionGetterOptions) error
type ActivityExecutionPropertySetterOption func(*ActivityExecutionSetterOptions) error

type SagaExecutionPropertyGetterOption func(*SagaExecutionGetterOptions) error
type SagaExecutionPropertySetterOption func(*SagaExecutionSetterOptions) error

type SideEffectExecutionPropertyGetterOption func(*SideEffectExecutionGetterOptions) error
type SideEffectExecutionPropertySetterOption func(*SideEffectExecutionSetterOptions) error

// Data property option types
type WorkflowDataPropertyGetterOption func(*WorkflowDataGetterOptions) error
type WorkflowDataPropertySetterOption func(*WorkflowDataSetterOptions) error

type ActivityDataPropertyGetterOption func(*ActivityDataGetterOptions) error
type ActivityDataPropertySetterOption func(*ActivityDataSetterOptions) error

type SagaDataPropertyGetterOption func(*SagaDataGetterOptions) error
type SagaDataPropertySetterOption func(*SagaDataSetterOptions) error

type SideEffectDataPropertyGetterOption func(*SideEffectDataGetterOptions) error
type SideEffectDataPropertySetterOption func(*SideEffectDataSetterOptions) error

// Execution Data property option types
type WorkflowExecutionDataPropertyGetterOption func(*WorkflowExecutionDataGetterOptions) error
type WorkflowExecutionDataPropertySetterOption func(*WorkflowExecutionDataSetterOptions) error

type ActivityExecutionDataPropertyGetterOption func(*ActivityExecutionDataGetterOptions) error
type ActivityExecutionDataPropertySetterOption func(*ActivityExecutionDataSetterOptions) error

type SagaExecutionDataPropertyGetterOption func(*SagaExecutionDataGetterOptions) error
type SagaExecutionDataPropertySetterOption func(*SagaExecutionDataSetterOptions) error

type SideEffectExecutionDataPropertyGetterOption func(*SideEffectExecutionDataGetterOptions) error
type SideEffectExecutionDataPropertySetterOption func(*SideEffectExecutionDataSetterOptions) error

// Property getter/setter function types
type BaseEntityPropertyGetter func(*BaseEntity) (BaseEntityPropertyGetterOption, error)
type BaseEntityPropertySetter func(*BaseEntity) (BaseEntityPropertySetterOption, error)

type BaseExecutionPropertyGetter func(*BaseExecution) (BaseExecutionPropertyGetterOption, error)
type BaseExecutionPropertySetter func(*BaseExecution) (BaseExecutionPropertySetterOption, error)

type RunPropertyGetter func(*Run) (RunPropertyGetterOption, error)
type RunPropertySetter func(*Run) (RunPropertySetterOption, error)

type WorkflowEntityPropertyGetter func(*WorkflowEntity) (WorkflowEntityPropertyGetterOption, error)
type WorkflowEntityPropertySetter func(*WorkflowEntity) (WorkflowEntityPropertySetterOption, error)

type ActivityEntityPropertyGetter func(*ActivityEntity) (ActivityEntityPropertyGetterOption, error)
type ActivityEntityPropertySetter func(*ActivityEntity) (ActivityEntityPropertySetterOption, error)

type SagaEntityPropertyGetter func(*SagaEntity) (SagaEntityPropertyGetterOption, error)
type SagaEntityPropertySetter func(*SagaEntity) (SagaEntityPropertySetterOption, error)

type SideEffectEntityPropertyGetter func(*SideEffectEntity) (SideEffectEntityPropertyGetterOption, error)
type SideEffectEntityPropertySetter func(*SideEffectEntity) (SideEffectEntityPropertySetterOption, error)

type QueuePropertyGetter func(*Queue) (QueuePropertyGetterOption, error)
type QueuePropertySetter func(*Queue) (QueuePropertySetterOption, error)

type VersionPropertyGetter func(*Version) (VersionPropertyGetterOption, error)
type VersionPropertySetter func(*Version) (VersionPropertySetterOption, error)

type HierarchyPropertyGetter func(*Hierarchy) (HierarchyPropertyGetterOption, error)
type HierarchyPropertySetter func(*Hierarchy) (HierarchyPropertySetterOption, error)

// Execution property types - Fixed these
type WorkflowExecutionPropertyGetter func(*WorkflowExecution) (WorkflowExecutionPropertyGetterOption, error)
type WorkflowExecutionPropertySetter func(*WorkflowExecution) (WorkflowExecutionPropertySetterOption, error)

type ActivityExecutionPropertyGetter func(*ActivityExecution) (ActivityExecutionPropertyGetterOption, error)
type ActivityExecutionPropertySetter func(*ActivityExecution) (ActivityExecutionPropertySetterOption, error)

type SagaExecutionPropertyGetter func(*SagaExecution) (SagaExecutionPropertyGetterOption, error)
type SagaExecutionPropertySetter func(*SagaExecution) (SagaExecutionPropertySetterOption, error)

type SideEffectExecutionPropertyGetter func(*SideEffectExecution) (SideEffectExecutionPropertyGetterOption, error)
type SideEffectExecutionPropertySetter func(*SideEffectExecution) (SideEffectExecutionPropertySetterOption, error)

// Data property types
type WorkflowDataPropertyGetter func(*WorkflowData) (WorkflowDataPropertyGetterOption, error)
type WorkflowDataPropertySetter func(*WorkflowData) (WorkflowDataPropertySetterOption, error)

type ActivityDataPropertyGetter func(*ActivityData) (ActivityDataPropertyGetterOption, error)
type ActivityDataPropertySetter func(*ActivityData) (ActivityDataPropertySetterOption, error)

type SagaDataPropertyGetter func(*SagaData) (SagaDataPropertyGetterOption, error)
type SagaDataPropertySetter func(*SagaData) (SagaDataPropertySetterOption, error)

type SideEffectDataPropertyGetter func(*SideEffectData) (SideEffectDataPropertyGetterOption, error)
type SideEffectDataPropertySetter func(*SideEffectData) (SideEffectDataPropertySetterOption, error)

// Execution Data property types
type WorkflowExecutionDataPropertyGetter func(*WorkflowExecutionData) (WorkflowExecutionDataPropertyGetterOption, error)
type WorkflowExecutionDataPropertySetter func(*WorkflowExecutionData) (WorkflowExecutionDataPropertySetterOption, error)

type ActivityExecutionDataPropertyGetter func(*ActivityExecutionData) (ActivityExecutionDataPropertyGetterOption, error)
type ActivityExecutionDataPropertySetter func(*ActivityExecutionData) (ActivityExecutionDataPropertySetterOption, error)

type SagaExecutionDataPropertyGetter func(*SagaExecutionData) (SagaExecutionDataPropertyGetterOption, error)
type SagaExecutionDataPropertySetter func(*SagaExecutionData) (SagaExecutionDataPropertySetterOption, error)

type SideEffectExecutionDataPropertyGetter func(*SideEffectExecutionData) (SideEffectExecutionDataPropertyGetterOption, error)
type SideEffectExecutionDataPropertySetter func(*SideEffectExecutionData) (SideEffectExecutionDataPropertySetterOption, error)

// Signals
type SignalEntityPropertyGetter func(*SignalEntity) (SignalEntityPropertyGetterOption, error)
type SignalEntityPropertySetter func(*SignalEntity) (SignalEntityPropertySetterOption, error)
type SignalEntityPropertyGetterOption func(*SignalEntityGetterOptions) error
type SignalEntityPropertySetterOption func(*SignalEntitySetterOptions) error

type SignalDataPropertyGetter func(*SignalData) (SignalDataPropertyGetterOption, error)
type SignalDataPropertySetter func(*SignalData) (SignalDataPropertySetterOption, error)
type SignalDataPropertyGetterOption func(*SignalDataGetterOptions) error
type SignalDataPropertySetterOption func(*SignalDataSetterOptions) error

type SignalExecutionPropertyGetter func(*SignalExecution) (SignalExecutionPropertyGetterOption, error)
type SignalExecutionPropertySetter func(*SignalExecution) (SignalExecutionPropertySetterOption, error)
type SignalExecutionPropertyGetterOption func(*SignalExecutionGetterOptions) error
type SignalExecutionPropertySetterOption func(*SignalExecutionSetterOptions) error

type SignalExecutionDataPropertyGetter func(*SignalExecutionData) (SignalExecutionDataPropertyGetterOption, error)
type SignalExecutionDataPropertySetter func(*SignalExecutionData) (SignalExecutionDataPropertySetterOption, error)
type SignalExecutionDataPropertyGetterOption func(*SignalExecutionDataGetterOptions) error
type SignalExecutionDataPropertySetterOption func(*SignalExecutionDataSetterOptions) error

// Core workflow options and types
type WorkflowOptions struct {
	Queue            string
	RetryPolicy      *RetryPolicy
	VersionOverrides map[string]int
	DeferExecution   bool // If true, only creates entity without starting execution
}

type ContinueAsNewError struct {
	Options *WorkflowOptions
	Args    []interface{}
	// Internal tracking of parent workflow for version inheritance
	parentWorkflowID int
}

func (e ContinueAsNewError) Error() string {
	return "workflow is continuing as new"
}

var ErrPaused = errors.New("execution paused")

// RetryPolicy and state
type RetryPolicy struct {
	MaxAttempts uint64
	MaxInterval time.Duration
}

type RetryState struct {
	Attempts uint64 `json:"attempts"`
	Timeout  int64  `json:"timeout,omitempty"`
}

type retryPolicyInternal struct {
	MaxAttempts uint64 `json:"max_attempts"`
	MaxInterval int64  `json:"max_interval"`
}

// Handler types
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

// Core entity structures
type Version struct {
	ID        VersionID
	EntityID  WorkflowEntityID       // Workflow ID this version belongs to
	ChangeID  string                 // Identifier for the change (e.g., "format-change")
	Version   int                    // The actual version number
	Data      map[string]interface{} // Additional version metadata if needed
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Run struct {
	ID          RunID
	Status      RunStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Entities    []*WorkflowEntity
	Hierarchies []*Hierarchy
}

type Queue struct {
	ID        QueueID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	Entities  []*WorkflowEntity
}

type Hierarchy struct {
	ID                HierarchyID
	RunID             RunID
	ParentEntityID    int
	ChildEntityID     int
	ParentExecutionID int
	ChildExecutionID  int
	ParentStepID      string
	ChildStepID       string
	ParentType        EntityType
	ChildType         EntityType
}

// Base entity type
type BaseEntity struct {
	QueueID     QueueID
	HandlerName string
	Type        EntityType
	Status      EntityStatus
	StepID      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	RunID       RunID
	RetryPolicy retryPolicyInternal
	RetryState  RetryState
}

// Entity Data structures
type WorkflowData struct {
	ID        WorkflowDataID   `json:"id,omitempty"`
	EntityID  WorkflowEntityID `json:"entity_id,omitempty"`
	Duration  string           `json:"duration,omitempty"`
	Paused    bool             `json:"paused"`
	Resumable bool             `json:"resumable"`
	Inputs    [][]byte         `json:"inputs,omitempty"`
	IsRoot    bool             `json:"is_root"`

	// ContinueAsNew reference to the parent workflow we come from
	ContinuedFrom *WorkflowEntityID `json:"continued_from,omitempty"`

	// ContinueAsNew reference to the parent workflow execution we come from
	ContinuedExecutionFrom *WorkflowExecutionID `json:"continued_execution_from,omitempty"`

	// Cross-queue reference to the step id we used
	WorkflowStepID *string `json:"workflow_step_id,omitempty"`

	// Cross-queue reference to the workflow we used
	WorkflowFrom *WorkflowEntityID `json:"workflow_from,omitempty"`

	// Cross-queue reference to the workflow execution we used
	WorkflowExecutionFrom *WorkflowExecutionID `json:"workflow_execution_from,omitempty"`

	Versions map[string]int // Tracks versions used in this workflow
}

type ActivityData struct {
	ID       ActivityDataID   `json:"id,omitempty"`
	EntityID ActivityEntityID `json:"entity_id,omitempty"`
	Inputs   [][]byte         `json:"inputs,omitempty"`
	Output   [][]byte         `json:"output,omitempty"`
}

type SagaData struct {
	ID       SagaDataID   `json:"id,omitempty"`
	EntityID SagaEntityID `json:"entity_id,omitempty"`
}

type SideEffectData struct {
	ID       SideEffectDataID   `json:"id,omitempty"`
	EntityID SideEffectEntityID `json:"entity_id,omitempty"`
	// No fields as per schema
}

// Entity types
type WorkflowEntityEdges struct {
	WorkflowChildren   []*WorkflowEntity
	ActivityChildren   []*ActivityEntity
	SagaChildren       []*SagaEntity
	SideEffectChildren []*SideEffectEntity
	Versions           []*Version
	Queue              *Queue
}

type WorkflowEntity struct {
	BaseEntity
	ID           WorkflowEntityID
	WorkflowData *WorkflowData
	Edges        *WorkflowEntityEdges
}

type ActivityEntity struct {
	BaseEntity
	ID           ActivityEntityID
	ActivityData *ActivityData
}

type SagaEntity struct {
	BaseEntity
	ID       SagaEntityID
	SagaData *SagaData
}

type SideEffectEntity struct {
	BaseEntity
	ID             SideEffectEntityID
	SideEffectData *SideEffectData
}

type SignalEntity struct {
	BaseEntity
	ID         SignalEntityID
	SignalData *SignalData
}

type SignalExecution struct {
	BaseExecution
	ID                  SignalExecutionID
	EntityID            SignalEntityID
	SignalExecutionData *SignalExecutionData
}

// Base execution type
type BaseExecution struct {
	StartedAt   time.Time
	CompletedAt *time.Time
	Status      ExecutionStatus
	Error       string
	StackTrace  *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Execution Data structures
type WorkflowExecutionData struct {
	ID            WorkflowExecutionDataID `json:"id,omitempty"`
	ExecutionID   WorkflowExecutionID     `json:"execution_id,omitempty"`
	LastHeartbeat *time.Time              `json:"last_heartbeat,omitempty"`
	Outputs       [][]byte                `json:"outputs,omitempty"`
}

type ActivityExecutionData struct {
	ID            ActivityExecutionDataID `json:"id,omitempty"`
	ExecutionID   ActivityExecutionID     `json:"execution_id,omitempty"`
	LastHeartbeat *time.Time              `json:"last_heartbeat,omitempty"`
	Outputs       [][]byte                `json:"outputs,omitempty"`
}

type SagaExecutionData struct {
	ID            SagaExecutionDataID `json:"id,omitempty"`
	ExecutionID   SagaExecutionID     `json:"execution_id,omitempty"`
	LastHeartbeat *time.Time          `json:"last_heartbeat,omitempty"`
	StepIndex     int                 `json:"step_index"` // Which step this execution is for
}

type SignalData struct {
	ID       SignalDataID   `json:"id,omitempty"`
	EntityID SignalEntityID `json:"entity_id,omitempty"`
	Name     string         `json:"name"`
}

type SignalExecutionData struct {
	ID          SignalExecutionDataID `json:"id,omitempty"`
	ExecutionID SignalExecutionID     `json:"execution_id,omitempty"`
	Value       []byte                `json:"value,omitempty"`
	Kind        uint                  `json:"kind"`
}

type SagaValue struct {
	ID                  SagaValueID         `json:"id,omitempty"`
	SagaExecutionID     SagaExecutionID     `json:"saga_executionID"`
	WorkflowExecutionID WorkflowExecutionID `json:"workflow_executionID"`
	SagaEntityID        SagaEntityID        `json:"saga_entityID"`
	Key                 string              `json:"key"`

	Value []byte `json:"value"` // TODO: this is WRONG
}

type SideEffectExecutionData struct {
	ID          SideEffectExecutionDataID `json:"id,omitempty"`
	ExecutionID SideEffectExecutionID     `json:"execution_id,omitempty"`
	Outputs     [][]byte                  `json:"outputs,omitempty"`
}

// Execution types
type WorkflowExecution struct {
	BaseExecution
	ID                    WorkflowExecutionID
	WorkflowEntityID      WorkflowEntityID
	WorkflowExecutionData *WorkflowExecutionData
}

type ActivityExecution struct {
	BaseExecution
	ID                    ActivityExecutionID
	ActivityEntityID      ActivityEntityID
	ActivityExecutionData *ActivityExecutionData
}

type SagaExecution struct {
	BaseExecution
	ID                SagaExecutionID
	SagaEntityID      SagaEntityID
	ExecutionType     ExecutionType `json:"execution_type"` // transaction or compensation
	SagaExecutionData *SagaExecutionData
}
type SideEffectExecution struct {
	BaseExecution
	ID                      SideEffectExecutionID
	SideEffectEntityID      SideEffectEntityID
	SideEffectExecutionData *SideEffectExecutionData
}

type BaseEntityFilter struct {
	RunID           *RunID
	QueueID         *QueueID
	HandlerName     *string
	Type            *EntityType
	Status          *EntityStatus
	StepID          *string
	AfterCreatedAt  *time.Time
	BeforeCreatedAt *time.Time
	AfterUpdatedAt  *time.Time
	BeforeUpdatedAt *time.Time
}

type BaseExecutionFilter struct {
	Status          *ExecutionStatus
	Error           *bool
	AfterStartedAt  *time.Time
	BeforeStartedAt *time.Time
	AfterCreatedAt  *time.Time
	BeforeCreatedAt *time.Time
	AfterUpdatedAt  *time.Time
	BeforeUpdatedAt *time.Time
}

type RunFilter struct {
	Status          *RunStatus
	AfterCreatedAt  *time.Time
	BeforeCreatedAt *time.Time
	AfterUpdatedAt  *time.Time
	BeforeUpdatedAt *time.Time
}

type VersionFilter struct {
	EntityID        *WorkflowEntityID
	ChangeID        *string
	Version         *int
	AfterCreatedAt  *time.Time
	BeforeCreatedAt *time.Time
}

type HierarchyFilter struct {
	RunID          *RunID
	ParentEntityID *int
	ChildEntityID  *int
	ParentType     *EntityType
	ChildType      *EntityType
	ParentStepID   *string
	ChildStepID    *string
}

type QueueFilter struct {
	Name            *string
	AfterCreatedAt  *time.Time
	BeforeCreatedAt *time.Time
	AfterUpdatedAt  *time.Time
	BeforeUpdatedAt *time.Time
}

type Pagination struct {
	Page     int
	PageSize int
}

type Paginated[T any] struct {
	Data       []*T // List of T for current page
	Total      int  // Total number of T matching filter
	TotalPages int  // Total number of pages
	Page       int  // Current page number
	PageSize   int  // Number of items per page
}

// WorkflowComponent represents a component (child) of a workflow with its metadata
type WorkflowComponent struct {
	Type       EntityType           // The type of the component (workflow, activity, saga, etc)
	EntityID   int                  // The ID of the entity
	StepID     string               // The step ID where this component was created
	Status     EntityStatus         // The current status of the component
	Components []*WorkflowComponent // Nested components (for sub-workflows)
}

// Database interface
type Database interface {
	GetWorkflowSubWorkflows(workflowID WorkflowEntityID) ([]*WorkflowEntity, error)
	GetWorkflowComponents(workflowID WorkflowEntityID) (*WorkflowComponent, error)

	// Listing operations
	ListRuns(page *Pagination, filter *RunFilter) (*Paginated[Run], error)
	ListVersions(page *Pagination, filter *VersionFilter) (*Paginated[Version], error)
	ListHierarchies(page *Pagination, filter *HierarchyFilter) (*Paginated[Hierarchy], error)
	ListQueues(page *Pagination, filter *QueueFilter) (*Paginated[Queue], error)

	// Entity listings
	ListWorkflowEntities(page *Pagination, filter *BaseEntityFilter) (*Paginated[WorkflowEntity], error)
	ListActivityEntities(page *Pagination, filter *BaseEntityFilter) (*Paginated[ActivityEntity], error)
	ListSideEffectEntities(page *Pagination, filter *BaseEntityFilter) (*Paginated[SideEffectEntity], error)
	ListSagaEntities(page *Pagination, filter *BaseEntityFilter) (*Paginated[SagaEntity], error)
	ListSignalEntities(page *Pagination, filter *BaseEntityFilter) (*Paginated[SignalEntity], error)

	// Execution listings
	ListWorkflowExecutions(page *Pagination, filter *BaseExecutionFilter) (*Paginated[WorkflowExecution], error)
	ListActivityExecutions(page *Pagination, filter *BaseExecutionFilter) (*Paginated[ActivityExecution], error)
	ListSideEffectExecutions(page *Pagination, filter *BaseExecutionFilter) (*Paginated[SideEffectExecution], error)
	ListSagaExecutions(page *Pagination, filter *BaseExecutionFilter) (*Paginated[SagaExecution], error)
	ListSignalExecutions(page *Pagination, filter *BaseExecutionFilter) (*Paginated[SignalExecution], error)

	// WORKFLOW-RELATED OPERATIONS
	// Workflow Entity
	AddWorkflowEntity(entity *WorkflowEntity) (WorkflowEntityID, error)
	GetWorkflowEntity(id WorkflowEntityID, opts ...WorkflowEntityGetOption) (*WorkflowEntity, error)
	HasWorkflowEntity(id WorkflowEntityID) (bool, error)
	UpdateWorkflowEntity(entity *WorkflowEntity) error
	GetWorkflowEntityProperties(id WorkflowEntityID, getters ...WorkflowEntityPropertyGetter) error
	SetWorkflowEntityProperties(id WorkflowEntityID, setters ...WorkflowEntityPropertySetter) error
	// specific
	CountWorkflowEntityByQueue(queueID QueueID) (int, error)
	CountWorkflowEntityByQueueByStatus(queueID QueueID, status EntityStatus) (int, error)

	// Workflow Data
	AddWorkflowData(entityID WorkflowEntityID, data *WorkflowData) (WorkflowDataID, error)
	GetWorkflowDataIDByEntityID(entityID WorkflowEntityID) (WorkflowDataID, error)
	GetWorkflowData(id WorkflowDataID) (*WorkflowData, error)
	HasWorkflowData(id WorkflowDataID) (bool, error)
	GetWorkflowDataProperties(id WorkflowDataID, getters ...WorkflowDataPropertyGetter) error
	SetWorkflowDataProperties(id WorkflowDataID, setters ...WorkflowDataPropertySetter) error
	GetWorkflowDataByEntityID(entityID WorkflowEntityID) (*WorkflowData, error)
	HasWorkflowDataByEntityID(entityID WorkflowEntityID) (bool, error)
	GetWorkflowDataPropertiesByEntityID(entityID WorkflowEntityID, getters ...WorkflowDataPropertyGetter) error
	SetWorkflowDataPropertiesByEntityID(entityID WorkflowEntityID, setters ...WorkflowDataPropertySetter) error

	// Workflow Execution
	AddWorkflowExecution(entity WorkflowEntityID, exec *WorkflowExecution) (WorkflowExecutionID, error)
	GetWorkflowExecution(id WorkflowExecutionID, opts ...WorkflowExecutionGetOption) (*WorkflowExecution, error)
	GetWorkflowExecutions(entityID WorkflowEntityID) ([]*WorkflowExecution, error)
	HasWorkflowExecution(id WorkflowExecutionID) (bool, error)
	GetWorkflowExecutionLatestByEntityID(entityID WorkflowEntityID) (*WorkflowExecution, error)
	GetWorkflowExecutionProperties(id WorkflowExecutionID, getters ...WorkflowExecutionPropertyGetter) error
	SetWorkflowExecutionProperties(id WorkflowExecutionID, setters ...WorkflowExecutionPropertySetter) error
	// specific
	CountWorkflowExecutionByQueue(queueID QueueID) (int, error)
	CountWorkflowExecutionByQueueByStatus(queueID QueueID, status ExecutionStatus) (int, error)

	// Workflow Execution Data
	AddWorkflowExecutionData(executionID WorkflowExecutionID, data *WorkflowExecutionData) (WorkflowExecutionDataID, error)
	GetWorkflowExecutionDataIDByExecutionID(executionID WorkflowExecutionID) (WorkflowExecutionDataID, error)
	GetWorkflowExecutionData(id WorkflowExecutionDataID) (*WorkflowExecutionData, error)
	HasWorkflowExecutionData(id WorkflowExecutionDataID) (bool, error)
	GetWorkflowExecutionDataProperties(id WorkflowExecutionDataID, getters ...WorkflowExecutionDataPropertyGetter) error
	SetWorkflowExecutionDataProperties(id WorkflowExecutionDataID, setters ...WorkflowExecutionDataPropertySetter) error
	GetWorkflowExecutionDataByExecutionID(executionID WorkflowExecutionID) (*WorkflowExecutionData, error)
	HasWorkflowExecutionDataByExecutionID(executionID WorkflowExecutionID) (bool, error)
	GetWorkflowExecutionDataPropertiesByExecutionID(executionID WorkflowExecutionID, getters ...WorkflowExecutionDataPropertyGetter) error
	SetWorkflowExecutionDataPropertiesByExecutionID(executionID WorkflowExecutionID, setters ...WorkflowExecutionDataPropertySetter) error

	// ACTIVITY-RELATED OPERATIONS
	// Activity Entity
	AddActivityEntity(entity *ActivityEntity, parentWorkflowID WorkflowEntityID) (ActivityEntityID, error)
	GetActivityEntity(id ActivityEntityID, opts ...ActivityEntityGetOption) (*ActivityEntity, error)
	GetActivityEntities(workflowID WorkflowEntityID, opts ...ActivityEntityGetOption) ([]*ActivityEntity, error)
	HasActivityEntity(id ActivityEntityID) (bool, error)
	UpdateActivityEntity(entity *ActivityEntity) error
	GetActivityEntityProperties(id ActivityEntityID, getters ...ActivityEntityPropertyGetter) error
	SetActivityEntityProperties(id ActivityEntityID, setters ...ActivityEntityPropertySetter) error

	// Activity Data
	AddActivityData(entityID ActivityEntityID, data *ActivityData) (ActivityDataID, error)
	GetActivityData(id ActivityDataID) (*ActivityData, error)
	HasActivityData(id ActivityDataID) (bool, error)
	GetActivityDataIDByEntityID(entityID ActivityEntityID) (ActivityDataID, error)
	GetActivityDataProperties(id ActivityDataID, getters ...ActivityDataPropertyGetter) error
	SetActivityDataProperties(id ActivityDataID, setters ...ActivityDataPropertySetter) error
	HasActivityDataByEntityID(entityID ActivityEntityID) (bool, error)
	GetActivityDataByEntityID(entityID ActivityEntityID) (*ActivityData, error)
	GetActivityDataPropertiesByEntityID(entityID ActivityEntityID, getters ...ActivityDataPropertyGetter) error
	SetActivityDataPropertiesByEntityID(entityID ActivityEntityID, setters ...ActivityDataPropertySetter) error

	// Activity Execution
	AddActivityExecution(exec *ActivityExecution) (ActivityExecutionID, error)
	GetActivityExecution(id ActivityExecutionID, opts ...ActivityExecutionGetOption) (*ActivityExecution, error)
	HasActivityExecution(id ActivityExecutionID) (bool, error)
	GetActivityExecutions(entityID ActivityEntityID) ([]*ActivityExecution, error)
	GetActivityExecutionLatestByEntityID(entityID ActivityEntityID) (*ActivityExecution, error)
	GetActivityExecutionProperties(id ActivityExecutionID, getters ...ActivityExecutionPropertyGetter) error
	SetActivityExecutionProperties(id ActivityExecutionID, setters ...ActivityExecutionPropertySetter) error

	// Activity Execution Data
	AddActivityExecutionData(executionID ActivityExecutionID, data *ActivityExecutionData) (ActivityExecutionDataID, error)
	GetActivityExecutionData(id ActivityExecutionDataID) (*ActivityExecutionData, error)
	HasActivityExecutionData(id ActivityExecutionDataID) (bool, error)
	GetActivityExecutionDataIDByExecutionID(executionID ActivityExecutionID) (ActivityExecutionDataID, error)
	GetActivityExecutionDataProperties(id ActivityExecutionDataID, getters ...ActivityExecutionDataPropertyGetter) error
	SetActivityExecutionDataProperties(id ActivityExecutionDataID, setters ...ActivityExecutionDataPropertySetter) error
	GetActivityExecutionDataByExecutionID(executionID ActivityExecutionID) (*ActivityExecutionData, error)
	HasActivityExecutionDataByExecutionID(executionID ActivityExecutionID) (bool, error)
	GetActivityExecutionDataPropertiesByExecutionID(executionID ActivityExecutionID, getters ...ActivityExecutionDataPropertyGetter) error
	SetActivityExecutionDataPropertiesByExecutionID(executionID ActivityExecutionID, setters ...ActivityExecutionDataPropertySetter) error

	// SAGA-RELATED OPERATIONS
	// Saga Entity
	AddSagaEntity(entity *SagaEntity, parentWorkflowID WorkflowEntityID) (SagaEntityID, error)
	GetSagaEntity(id SagaEntityID, opts ...SagaEntityGetOption) (*SagaEntity, error)
	GetSagaEntities(workflowID WorkflowEntityID, opts ...SagaEntityGetOption) ([]*SagaEntity, error)
	HasSagaEntity(id SagaEntityID) (bool, error)
	UpdateSagaEntity(entity *SagaEntity) error
	GetSagaEntityProperties(id SagaEntityID, getters ...SagaEntityPropertyGetter) error
	SetSagaEntityProperties(id SagaEntityID, setters ...SagaEntityPropertySetter) error

	// Saga Data
	AddSagaData(entityID SagaEntityID, data *SagaData) (SagaDataID, error)
	GetSagaData(id SagaDataID) (*SagaData, error)
	HasSagaData(id SagaDataID) (bool, error)
	GetSagaDataIDByEntityID(entityID SagaEntityID) (SagaDataID, error)
	GetSagaDataProperties(id SagaDataID, getters ...SagaDataPropertyGetter) error
	SetSagaDataProperties(id SagaDataID, setters ...SagaDataPropertySetter) error
	GetSagaDataByEntityID(entityID SagaEntityID) (*SagaData, error)
	HasSagaDataByEntityID(entityID SagaEntityID) (bool, error)
	GetSagaDataPropertiesByEntityID(entityID SagaEntityID, getters ...SagaDataPropertyGetter) error
	SetSagaDataPropertiesByEntityID(entityID SagaEntityID, setters ...SagaDataPropertySetter) error

	// Saga Execution
	AddSagaExecution(exec *SagaExecution) (SagaExecutionID, error)
	GetSagaExecution(id SagaExecutionID, opts ...SagaExecutionGetOption) (*SagaExecution, error)
	GetSagaExecutions(entityID SagaEntityID) ([]*SagaExecution, error)
	HasSagaExecution(id SagaExecutionID) (bool, error)
	GetSagaExecutionProperties(id SagaExecutionID, getters ...SagaExecutionPropertyGetter) error
	SetSagaExecutionProperties(id SagaExecutionID, setters ...SagaExecutionPropertySetter) error

	// Saga Execution Data
	AddSagaExecutionData(executionID SagaExecutionID, data *SagaExecutionData) (SagaExecutionDataID, error)
	GetSagaExecutionData(id SagaExecutionDataID) (*SagaExecutionData, error)
	HasSagaExecutionData(id SagaExecutionDataID) (bool, error)
	GetSagaExecutionDataProperties(id SagaExecutionDataID, getters ...SagaExecutionDataPropertyGetter) error
	SetSagaExecutionDataProperties(id SagaExecutionDataID, setters ...SagaExecutionDataPropertySetter) error
	GetSagaExecutionDataIDByExecutionID(executionID SagaExecutionID) (SagaExecutionDataID, error)
	GetSagaExecutionDataByExecutionID(executionID SagaExecutionID) (*SagaExecutionData, error)
	HasSagaExecutionDataByExecutionID(executionID SagaExecutionID) (bool, error)
	GetSagaExecutionDataPropertiesByEntities(entityID SagaExecutionID, getters ...SagaExecutionDataPropertyGetter) error
	SetSagaExecutionDataPropertiesByEntities(entityID SagaExecutionID, setters ...SagaExecutionDataPropertySetter) error

	// Saga Context operations
	SetSagaValue(sagaEntityID SagaEntityID, sagaExecID SagaExecutionID, key string, value []byte) (SagaValueID, error)
	GetSagaValue(id SagaValueID) (*SagaValue, error)
	GetSagaValueByKey(sagaEntityID SagaEntityID, key string) ([]byte, error)
	GetSagaValuesByEntity(sagaEntityID SagaEntityID) ([]*SagaValue, error)
	GetSagaValuesByExecution(sagaExecID SagaExecutionID) ([]*SagaValue, error)

	// SIDE-EFFECT-RELATED OPERATIONS
	// SideEffect Entity
	AddSideEffectEntity(entity *SideEffectEntity, parentWorkflowID WorkflowEntityID) (SideEffectEntityID, error)
	GetSideEffectEntity(id SideEffectEntityID, opts ...SideEffectEntityGetOption) (*SideEffectEntity, error)
	GetSideEffectEntities(workflowID WorkflowEntityID, opts ...SideEffectEntityGetOption) ([]*SideEffectEntity, error)
	HasSideEffectEntity(id SideEffectEntityID) (bool, error)
	UpdateSideEffectEntity(entity *SideEffectEntity) error
	GetSideEffectEntityProperties(id SideEffectEntityID, getters ...SideEffectEntityPropertyGetter) error
	SetSideEffectEntityProperties(id SideEffectEntityID, setters ...SideEffectEntityPropertySetter) error

	// SideEffect Data
	AddSideEffectData(entityID SideEffectEntityID, data *SideEffectData) (SideEffectDataID, error)
	GetSideEffectData(id SideEffectDataID) (*SideEffectData, error)
	HasSideEffectData(id SideEffectDataID) (bool, error)
	GetSideEffectDataProperties(id SideEffectDataID, getters ...SideEffectDataPropertyGetter) error
	SetSideEffectDataProperties(id SideEffectDataID, setters ...SideEffectDataPropertySetter) error
	GetSideEffectDataIDByEntityID(entityID SideEffectEntityID) (SideEffectDataID, error)
	GetSideEffectDataByEntityID(entityID SideEffectEntityID) (*SideEffectData, error)
	HasSideEffectDataByEntityID(entityID SideEffectEntityID) (bool, error)
	GetSideEffectDataPropertiesByEntityID(entityID SideEffectEntityID, getters ...SideEffectDataPropertyGetter) error
	SetSideEffectDataPropertiesByEntityID(entityID SideEffectEntityID, setters ...SideEffectDataPropertySetter) error

	// SideEffect Execution
	AddSideEffectExecution(exec *SideEffectExecution) (SideEffectExecutionID, error)
	GetSideEffectExecution(id SideEffectExecutionID, opts ...SideEffectExecutionGetOption) (*SideEffectExecution, error)
	HasSideEffectExecution(id SideEffectExecutionID) (bool, error)
	GetSideEffectExecutions(entityID SideEffectEntityID) ([]*SideEffectExecution, error)
	GetSideEffectExecutionProperties(id SideEffectExecutionID, getters ...SideEffectExecutionPropertyGetter) error
	SetSideEffectExecutionProperties(id SideEffectExecutionID, setters ...SideEffectExecutionPropertySetter) error

	// SideEffect Execution Data
	AddSideEffectExecutionData(executionID SideEffectExecutionID, data *SideEffectExecutionData) (SideEffectExecutionDataID, error)
	GetSideEffectExecutionData(id SideEffectExecutionDataID) (*SideEffectExecutionData, error)
	HasSideEffectExecutionData(id SideEffectExecutionDataID) (bool, error)
	GetSideEffectExecutionDataIDByExecutionID(executionID SideEffectEntityID) (SideEffectDataID, error)
	GetSideEffectExecutionDataByExecutionID(executionID SideEffectExecutionID) (*SideEffectExecutionData, error)
	HasSideEffectExecutionDataByExecutionID(executionID SideEffectExecutionID) (bool, error)
	GetSideEffectExecutionDataProperties(id SideEffectExecutionDataID, getters ...SideEffectExecutionDataPropertyGetter) error
	SetSideEffectExecutionDataProperties(id SideEffectExecutionDataID, setters ...SideEffectExecutionDataPropertySetter) error
	GetSideEffectExecutionDataPropertiesByExecutionID(executionID SideEffectExecutionID, getters ...SideEffectExecutionDataPropertyGetter) error
	SetSideEffectExecutionDataPropertiesByExecutionID(executionID SideEffectExecutionID, setters ...SideEffectExecutionDataPropertySetter) error

	// RUN-RELATED OPERATIONS
	AddRun(run *Run) (RunID, error)
	GetRun(id RunID, opts ...RunGetOption) (*Run, error)
	HasRun(id RunID) (bool, error)
	UpdateRun(run *Run) error
	GetRunProperties(id RunID, getters ...RunPropertyGetter) error
	SetRunProperties(id RunID, setters ...RunPropertySetter) error
	DeleteRunsByStatus(status RunStatus) error
	DeleteRuns(ids ...RunID) error

	// VERSION-RELATED OPERATIONS
	AddVersion(version *Version) (VersionID, error)
	GetVersion(id VersionID, opts ...VersionGetOption) (*Version, error)
	HasVersion(id VersionID) (bool, error)
	UpdateVersion(version *Version) error
	GetVersionByWorkflowAndChangeID(workflowID WorkflowEntityID, changeID string) (*Version, error)
	GetVersionsByWorkflowID(workflowID WorkflowEntityID) ([]*Version, error)
	SetVersion(version *Version) error
	GetVersionProperties(id VersionID, getters ...VersionPropertyGetter) error
	SetVersionProperties(id VersionID, setters ...VersionPropertySetter) error

	// HIERARCHY-RELATED OPERATIONS
	AddHierarchy(hierarchy *Hierarchy) (HierarchyID, error)
	GetHierarchy(id HierarchyID, opts ...HierarchyGetOption) (*Hierarchy, error)
	GetHierarchyByParentEntity(parentEntityID int, childStepID string, specificType EntityType) (*Hierarchy, error)
	GetHierarchiesByParentEntityAndStep(parentEntityID int, childStepID string, specificType EntityType) ([]*Hierarchy, error)
	HasHierarchy(id HierarchyID) (bool, error)
	UpdateHierarchy(hierarchy *Hierarchy) error
	GetHierarchyProperties(id HierarchyID, getters ...HierarchyPropertyGetter) error
	SetHierarchyProperties(id HierarchyID, setters ...HierarchyPropertySetter) error
	// specific one, might refactor that
	GetHierarchiesByChildEntity(childEntityID int) ([]*Hierarchy, error)
	GetHierarchiesByParentEntity(parentEntityID int) ([]*Hierarchy, error)

	// QUEUE-RELATED OPERATIONS
	AddQueue(queue *Queue) (QueueID, error)
	GetQueue(id QueueID, opts ...QueueGetOption) (*Queue, error)
	GetQueueByName(name string, opts ...QueueGetOption) (*Queue, error)
	HasQueue(id QueueID) (bool, error)
	HasQueueName(name string) (bool, error)
	UpdateQueue(queue *Queue) error
	GetQueueProperties(id QueueID, getters ...QueuePropertyGetter) error
	SetQueueProperties(id QueueID, setters ...QueuePropertySetter) error

	// Signal Entity operations
	AddSignalEntity(entity *SignalEntity, parentWorkflowID WorkflowEntityID) (SignalEntityID, error)
	GetSignalEntity(id SignalEntityID, opts ...SignalEntityGetOption) (*SignalEntity, error)
	GetSignalEntities(workflowID WorkflowEntityID, opts ...SignalEntityGetOption) ([]*SignalEntity, error)
	HasSignalEntity(id SignalEntityID) (bool, error)
	UpdateSignalEntity(entity *SignalEntity) error
	GetSignalEntityProperties(id SignalEntityID, getters ...SignalEntityPropertyGetter) error
	SetSignalEntityProperties(id SignalEntityID, setters ...SignalEntityPropertySetter) error

	// Signal Data
	AddSignalData(entityID SignalEntityID, data *SignalData) (SignalDataID, error)
	GetSignalData(id SignalDataID) (*SignalData, error)
	HasSignalData(id SignalDataID) (bool, error)
	GetSignalDataIDByEntityID(entityID SignalEntityID) (SignalDataID, error)
	GetSignalDataByEntityID(entityID SignalEntityID) (*SignalData, error)
	HasSignalDataByEntityID(entityID SignalEntityID) (bool, error)
	GetSignalDataProperties(id SignalDataID, getters ...SignalDataPropertyGetter) error
	SetSignalDataProperties(id SignalDataID, setters ...SignalDataPropertySetter) error
	GetSignalDataPropertiesByEntityID(entityID SignalEntityID, getters ...SignalDataPropertyGetter) error
	SetSignalDataPropertiesByEntityID(entityID SignalEntityID, setters ...SignalDataPropertySetter) error

	// Signal Execution
	AddSignalExecution(exec *SignalExecution) (SignalExecutionID, error)
	GetSignalExecution(id SignalExecutionID, opts ...SignalExecutionGetOption) (*SignalExecution, error)
	GetSignalExecutions(entityID SignalEntityID) ([]*SignalExecution, error)
	HasSignalExecution(id SignalExecutionID) (bool, error)
	GetSignalExecutionLatestByEntityID(entityID SignalEntityID) (*SignalExecution, error)
	GetSignalExecutionProperties(id SignalExecutionID, getters ...SignalExecutionPropertyGetter) error
	SetSignalExecutionProperties(id SignalExecutionID, setters ...SignalExecutionPropertySetter) error

	// Signal Execution Data
	AddSignalExecutionData(executionID SignalExecutionID, data *SignalExecutionData) (SignalExecutionDataID, error)
	GetSignalExecutionData(id SignalExecutionDataID) (*SignalExecutionData, error)
	HasSignalExecutionData(id SignalExecutionDataID) (bool, error)
	GetSignalExecutionDataIDByExecutionID(executionID SignalExecutionID) (SignalExecutionDataID, error)
	GetSignalExecutionDataByExecutionID(executionID SignalExecutionID) (*SignalExecutionData, error)
	HasSignalExecutionDataByExecutionID(executionID SignalExecutionID) (bool, error)
	GetSignalExecutionDataProperties(id SignalExecutionDataID, getters ...SignalExecutionDataPropertyGetter) error
	SetSignalExecutionDataProperties(id SignalExecutionDataID, setters ...SignalExecutionDataPropertySetter) error
	GetSignalExecutionDataPropertiesByExecutionID(executionID SignalExecutionID, getters ...SignalExecutionDataPropertyGetter) error
	SetSignalExecutionDataPropertiesByExecutionID(executionID SignalExecutionID, setters ...SignalExecutionDataPropertySetter) error
}

// Return the output of a workflow if it's completed
func GetWorkflowEntityOutput(id WorkflowEntityID, registry *Registry, db Database) ([]interface{}, error) {
	workflowEntity, err := db.GetWorkflowEntity(id)
	if err != nil {
		return nil, err
	}

	if workflowEntity.Status != StatusCompleted {
		return nil, fmt.Errorf("workflow not completed")
	}

	workflowExecution, err := db.GetWorkflowExecutionLatestByEntityID(id)
	if err != nil {
		return nil, err
	}

	handler, ok := registry.GetWorkflow(workflowEntity.HandlerName)
	if !ok {
		return nil, fmt.Errorf("workflow handler not found")
	}

	data, err := db.GetWorkflowExecutionDataByExecutionID(workflowExecution.ID)
	if err != nil {
		return nil, err
	}

	return convertOutputsFromSerialization(handler, data.Outputs)
}

// RetryPolicy helper functions
func DefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxAttempts: 0, // 0 means no retries
		MaxInterval: 100 * time.Millisecond,
	}
}

func DefaultRetryPolicyInternal() *retryPolicyInternal {
	return &retryPolicyInternal{
		MaxAttempts: 0, // 0 means no retries
		MaxInterval: (100 * time.Millisecond).Nanoseconds(),
	}
}

func getRetryPolicyOrDefault(options *RetryPolicy) (maxAttempts uint64, maxInterval time.Duration) {
	def := DefaultRetryPolicy()
	if options != nil {
		maxInterval = def.MaxInterval
		maxAttempts = def.MaxAttempts
		// fmt.Println("options.MaxAttempts", options.MaxAttempts, "def.MaxAttempts", def.MaxAttempts, options.MaxAttempts != def.MaxAttempts)
		if options.MaxAttempts != def.MaxAttempts {
			maxAttempts = options.MaxAttempts
		} else {
		}
		// fmt.Println("options.MaxInterval", options.MaxInterval, "def.MaxInterval", def.MaxInterval, options.MaxInterval != def.MaxInterval && options.MaxInterval > 0)
		if options.MaxInterval != def.MaxInterval && options.MaxInterval > 0 {
			maxInterval = options.MaxInterval
		}
	} else {
		maxAttempts = def.MaxAttempts
		maxInterval = def.MaxInterval
	}
	return
}

func ToInternalRetryPolicy(rp *RetryPolicy) *retryPolicyInternal {
	if rp == nil {
		return DefaultRetryPolicyInternal()
	}

	return &retryPolicyInternal{
		MaxAttempts: rp.MaxAttempts,
		MaxInterval: rp.MaxInterval.Nanoseconds(),
	}
}

// Serialization helper functions
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

		decodedObj := reflect.New(outputType).Interface()

		if err := rtl.Decode(buf, decodedObj); err != nil {
			return nil, err
		}

		output = append(output, reflect.ValueOf(decodedObj).Elem().Interface())
	}

	return output, nil
}

func convertOutputFromSerialization(outputType reflect.Type, executionOutput []byte) (interface{}, error) {
	buf := bytes.NewBuffer(executionOutput)

	decodedObj := reflect.New(outputType).Interface()

	if err := rtl.Decode(buf, decodedObj); err != nil {
		return nil, err
	}

	return reflect.ValueOf(decodedObj).Elem().Interface(), nil
}

func convertSingleOutputFromSerialization(outputType reflect.Type, executionOutput []byte) (interface{}, error) {
	buf := bytes.NewBuffer(executionOutput)

	decodedObj := reflect.New(outputType).Interface()

	if err := rtl.Decode(buf, decodedObj); err != nil {
		return nil, err
	}

	return reflect.ValueOf(decodedObj).Elem().Interface(), nil
}

func ConvertBack(data []byte, output interface{}) error {
	buf := bytes.NewBuffer(data)
	if reflect.TypeOf(output).Kind() != reflect.Ptr {
		return errors.Join(ErrSerialization, ErrMustPointer)
	}
	if err := rtl.Decode(buf, output); err != nil {
		return errors.Join(ErrSerialization, err)
	}
	return nil
}
