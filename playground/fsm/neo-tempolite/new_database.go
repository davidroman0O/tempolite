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
	Outputs              []interface{}
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
)

// EntityStatus defines the status of the entity
type EntityStatus string

const (
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
	WorkflowID *int
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
	QueueID   *int
	QueueName *string
	Version   *Version
	ChildID   *int
	ChildType *EntityType
	RunID     *int
	// Add more options as needed
}

type ActivityEntityGetterOptions struct {
	IncludeData bool
	// Add more flags as needed
}

type ActivityEntitySetterOptions struct {
	ParentWorkflowID *int
	ParentRunID      *int
	// Add more options as needed
}

type SagaEntityGetterOptions struct {
	IncludeData bool
	// Add more flags as needed
}

type SagaEntitySetterOptions struct {
	ParentWorkflowID *int
	// Add more options as needed
}

type SideEffectEntityGetterOptions struct {
	IncludeData bool
	// Add more flags as needed
}

type SideEffectEntitySetterOptions struct {
	ParentWorkflowID *int
	// Add more options as needed
}

type QueueGetterOptions struct {
	IncludeWorkflows bool
	// Add more flags as needed
}

type QueueSetterOptions struct {
	WorkflowIDs []int
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

// Core workflow options and types
type WorkflowOptions struct {
	Queue            string
	RetryPolicy      *RetryPolicy
	VersionOverrides map[string]int
}

type ContinueAsNewError struct {
	Options *WorkflowOptions
	Args    []interface{}
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
	ID        int
	EntityID  int
	ChangeID  string
	Version   int
	Data      map[string]interface{}
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Run struct {
	ID          int
	Status      RunStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Entities    []*WorkflowEntity
	Hierarchies []*Hierarchy
}

type Queue struct {
	ID        int
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	Entities  []*WorkflowEntity
}

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

// Base entity type
type BaseEntity struct {
	ID          int
	QueueID     int
	HandlerName string
	Type        EntityType
	Status      EntityStatus
	StepID      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	RunID       int
	RetryPolicy retryPolicyInternal
	RetryState  RetryState
}

// Entity Data structures
type WorkflowData struct {
	ID            int      `json:"id,omitempty"`
	EntityID      int      `json:"entity_id,omitempty"`
	Duration      string   `json:"duration,omitempty"`
	Paused        bool     `json:"paused"`
	Resumable     bool     `json:"resumable"`
	Inputs        [][]byte `json:"inputs,omitempty"`
	IsRoot        bool     `json:"is_root"`
	ContinuedFrom *int     `json:"continued_from,omitempty"`
}

type ActivityData struct {
	ID       int      `json:"id,omitempty"`
	EntityID int      `json:"entity_id,omitempty"`
	Inputs   [][]byte `json:"inputs,omitempty"`
	Output   [][]byte `json:"output,omitempty"`
}

type SagaData struct {
	ID               int      `json:"id,omitempty"`
	EntityID         int      `json:"entity_id,omitempty"`
	Compensating     bool     `json:"compensating"`
	CompensationData [][]byte `json:"compensation_data,omitempty"`
}

type SideEffectData struct {
	ID       int `json:"id,omitempty"`
	EntityID int `json:"entity_id,omitempty"`
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
	WorkflowData *WorkflowData
	Edges        *WorkflowEntityEdges
}

type ActivityEntity struct {
	BaseEntity
	ActivityData *ActivityData
}

type SagaEntity struct {
	BaseEntity
	SagaData *SagaData
}

type SideEffectEntity struct {
	BaseEntity
	SideEffectData *SideEffectData
}

// Base execution type
type BaseExecution struct {
	ID          int
	EntityID    int
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
	ID            int        `json:"id,omitempty"`
	ExecutionID   int        `json:"execution_id,omitempty"`
	LastHeartbeat *time.Time `json:"last_heartbeat,omitempty"`
	Outputs       [][]byte   `json:"outputs,omitempty"`
}

type ActivityExecutionData struct {
	ID            int        `json:"id,omitempty"`
	ExecutionID   int        `json:"execution_id,omitempty"`
	LastHeartbeat *time.Time `json:"last_heartbeat,omitempty"`
	Outputs       [][]byte   `json:"outputs,omitempty"`
}

type SagaExecutionData struct {
	ID            int        `json:"id,omitempty"`
	ExecutionID   int        `json:"execution_id,omitempty"`
	LastHeartbeat *time.Time `json:"last_heartbeat,omitempty"`
	Output        [][]byte   `json:"output,omitempty"`
	HasOutput     bool       `json:"hasOutput"`
}

type SideEffectExecutionData struct {
	ID          int      `json:"id,omitempty"`
	ExecutionID int      `json:"execution_id,omitempty"`
	Outputs     [][]byte `json:"outputs,omitempty"`
}

// Execution types
type WorkflowExecution struct {
	BaseExecution
	WorkflowExecutionData *WorkflowExecutionData
}

type ActivityExecution struct {
	BaseExecution
	ActivityExecutionData *ActivityExecutionData
}

type SagaExecution struct {
	BaseExecution
	SagaExecutionData *SagaExecutionData
}

type SideEffectExecution struct {
	BaseExecution
	SideEffectExecutionData *SideEffectExecutionData
}

// Database interface
type Database interface {

	// WORKFLOW-RELATED OPERATIONS
	// Workflow Entity
	AddWorkflowEntity(entity *WorkflowEntity) (int, error)
	GetWorkflowEntity(id int, opts ...WorkflowEntityGetOption) (*WorkflowEntity, error)
	HasWorkflowEntity(id int) (bool, error)
	UpdateWorkflowEntity(entity *WorkflowEntity) error
	GetWorkflowEntityProperties(id int, getters ...WorkflowEntityPropertyGetter) error
	SetWorkflowEntityProperties(id int, setters ...WorkflowEntityPropertySetter) error

	// Workflow Data
	AddWorkflowData(entityID int, data *WorkflowData) (int, error)
	GetWorkflowData(id int) (*WorkflowData, error)
	HasWorkflowData(id int) (bool, error)
	GetWorkflowDataByEntityID(entityID int) (*WorkflowData, error)
	HasWorkflowDataByEntityID(entityID int) (bool, error)
	GetWorkflowDataProperties(entityID int, getters ...WorkflowDataPropertyGetter) error
	SetWorkflowDataProperties(entityID int, setters ...WorkflowDataPropertySetter) error

	// Workflow Execution
	AddWorkflowExecution(exec *WorkflowExecution) (int, error)
	GetWorkflowExecution(id int, opts ...WorkflowExecutionGetOption) (*WorkflowExecution, error)
	GetWorkflowExecutions(entityID int) ([]*WorkflowExecution, error)
	HasWorkflowExecution(id int) (bool, error)
	GetWorkflowExecutionLatestByEntityID(entityID int) (*WorkflowExecution, error)
	GetWorkflowExecutionProperties(id int, getters ...WorkflowExecutionPropertyGetter) error
	SetWorkflowExecutionProperties(id int, setters ...WorkflowExecutionPropertySetter) error

	// Workflow Execution Data
	AddWorkflowExecutionData(executionID int, data *WorkflowExecutionData) (int, error)
	GetWorkflowExecutionData(id int) (*WorkflowExecutionData, error)
	HasWorkflowExecutionData(id int) (bool, error)
	GetWorkflowExecutionDataByExecutionID(executionID int) (*WorkflowExecutionData, error)
	HasWorkflowExecutionDataByExecutionID(executionID int) (bool, error)
	GetWorkflowExecutionDataProperties(entityID int, getters ...WorkflowExecutionDataPropertyGetter) error
	SetWorkflowExecutionDataProperties(entityID int, setters ...WorkflowExecutionDataPropertySetter) error

	// ACTIVITY-RELATED OPERATIONS
	// Activity Entity
	AddActivityEntity(entity *ActivityEntity) (int, error)
	GetActivityEntity(id int, opts ...ActivityEntityGetOption) (*ActivityEntity, error)
	HasActivityEntity(id int) (bool, error)
	UpdateActivityEntity(entity *ActivityEntity) error
	GetActivityEntityProperties(id int, getters ...ActivityEntityPropertyGetter) error
	SetActivityEntityProperties(id int, setters ...ActivityEntityPropertySetter) error

	// Activity Data
	AddActivityData(entityID int, data *ActivityData) (int, error)
	GetActivityData(id int) (*ActivityData, error)
	HasActivityData(id int) (bool, error)
	GetActivityDataByEntityID(entityID int) (*ActivityData, error)
	HasActivityDataByEntityID(entityID int) (bool, error)
	GetActivityDataProperties(entityID int, getters ...ActivityDataPropertyGetter) error
	SetActivityDataProperties(entityID int, setters ...ActivityDataPropertySetter) error

	// Activity Execution
	AddActivityExecution(exec *ActivityExecution) (int, error)
	GetActivityExecution(id int, opts ...ActivityExecutionGetOption) (*ActivityExecution, error)
	HasActivityExecution(id int) (bool, error)
	GetActivityExecutions(entityID int) ([]*ActivityExecution, error)
	GetActivityExecutionLatestByEntityID(entityID int) (*ActivityExecution, error)
	GetActivityExecutionProperties(id int, getters ...ActivityExecutionPropertyGetter) error
	SetActivityExecutionProperties(id int, setters ...ActivityExecutionPropertySetter) error

	// Activity Execution Data
	AddActivityExecutionData(executionID int, data *ActivityExecutionData) (int, error)
	GetActivityExecutionData(id int) (*ActivityExecutionData, error)
	HasActivityExecutionData(id int) (bool, error)
	GetActivityExecutionDataByExecutionID(executionID int) (*ActivityExecutionData, error)
	HasActivityExecutionDataByExecutionID(executionID int) (bool, error)
	GetActivityExecutionDataProperties(entityID int, getters ...ActivityExecutionDataPropertyGetter) error
	SetActivityExecutionDataProperties(entityID int, setters ...ActivityExecutionDataPropertySetter) error

	// SAGA-RELATED OPERATIONS
	// Saga Entity
	AddSagaEntity(entity *SagaEntity) (int, error)
	GetSagaEntity(id int, opts ...SagaEntityGetOption) (*SagaEntity, error)
	HasSagaEntity(id int) (bool, error)
	UpdateSagaEntity(entity *SagaEntity) error
	GetSagaEntityProperties(id int, getters ...SagaEntityPropertyGetter) error
	SetSagaEntityProperties(id int, setters ...SagaEntityPropertySetter) error

	// Saga Data
	AddSagaData(entityID int, data *SagaData) (int, error)
	GetSagaData(id int) (*SagaData, error)
	HasSagaData(id int) (bool, error)
	GetSagaDataByEntityID(entityID int) (*SagaData, error)
	HasSagaDataByEntityID(entityID int) (bool, error)
	GetSagaDataProperties(entityID int, getters ...SagaDataPropertyGetter) error
	SetSagaDataProperties(entityID int, setters ...SagaDataPropertySetter) error

	// Saga Execution
	AddSagaExecution(exec *SagaExecution) (int, error)
	GetSagaExecution(id int, opts ...SagaExecutionGetOption) (*SagaExecution, error)
	HasSagaExecution(id int) (bool, error)
	GetSagaExecutionProperties(id int, getters ...SagaExecutionPropertyGetter) error
	SetSagaExecutionProperties(id int, setters ...SagaExecutionPropertySetter) error

	// Saga Execution Data
	AddSagaExecutionData(executionID int, data *SagaExecutionData) (int, error)
	GetSagaExecutionData(id int) (*SagaExecutionData, error)
	HasSagaExecutionData(id int) (bool, error)
	GetSagaExecutionDataByExecutionID(executionID int) (*SagaExecutionData, error)
	HasSagaExecutionDataByExecutionID(executionID int) (bool, error)
	GetSagaExecutionDataProperties(entityID int, getters ...SagaExecutionDataPropertyGetter) error
	SetSagaExecutionDataProperties(entityID int, setters ...SagaExecutionDataPropertySetter) error

	// SIDE-EFFECT-RELATED OPERATIONS
	// SideEffect Entity
	AddSideEffectEntity(entity *SideEffectEntity) (int, error)
	GetSideEffectEntity(id int, opts ...SideEffectEntityGetOption) (*SideEffectEntity, error)
	HasSideEffectEntity(id int) (bool, error)
	UpdateSideEffectEntity(entity *SideEffectEntity) error
	GetSideEffectEntityProperties(id int, getters ...SideEffectEntityPropertyGetter) error
	SetSideEffectEntityProperties(id int, setters ...SideEffectEntityPropertySetter) error

	// SideEffect Data
	AddSideEffectData(entityID int, data *SideEffectData) (int, error)
	GetSideEffectData(id int) (*SideEffectData, error)
	HasSideEffectData(id int) (bool, error)
	GetSideEffectDataByEntityID(entityID int) (*SideEffectData, error)
	HasSideEffectDataByEntityID(entityID int) (bool, error)
	GetSideEffectDataProperties(entityID int, getters ...SideEffectDataPropertyGetter) error
	SetSideEffectDataProperties(entityID int, setters ...SideEffectDataPropertySetter) error

	// SideEffect Execution
	AddSideEffectExecution(exec *SideEffectExecution) (int, error)
	GetSideEffectExecution(id int, opts ...SideEffectExecutionGetOption) (*SideEffectExecution, error)
	HasSideEffectExecution(id int) (bool, error)
	GetSideEffectExecutionProperties(id int, getters ...SideEffectExecutionPropertyGetter) error
	SetSideEffectExecutionProperties(id int, setters ...SideEffectExecutionPropertySetter) error

	// SideEffect Execution Data
	AddSideEffectExecutionData(executionID int, data *SideEffectExecutionData) (int, error)
	GetSideEffectExecutionData(id int) (*SideEffectExecutionData, error)
	HasSideEffectExecutionData(id int) (bool, error)
	GetSideEffectExecutionDataByExecutionID(executionID int) (*SideEffectExecutionData, error)
	HasSideEffectExecutionDataByExecutionID(executionID int) (bool, error)
	GetSideEffectExecutionDataProperties(entityID int, getters ...SideEffectExecutionDataPropertyGetter) error
	SetSideEffectExecutionDataProperties(entityID int, setters ...SideEffectExecutionDataPropertySetter) error

	// RUN-RELATED OPERATIONS
	AddRun(run *Run) (int, error)
	GetRun(id int, opts ...RunGetOption) (*Run, error)
	HasRun(id int) (bool, error)
	UpdateRun(run *Run) error
	GetRunProperties(id int, getters ...RunPropertyGetter) error
	SetRunProperties(id int, setters ...RunPropertySetter) error

	// VERSION-RELATED OPERATIONS
	AddVersion(version *Version) (int, error)
	GetVersion(id int, opts ...VersionGetOption) (*Version, error)
	HasVersion(id int) (bool, error)
	UpdateVersion(version *Version) error
	GetVersionProperties(id int, getters ...VersionPropertyGetter) error
	SetVersionProperties(id int, setters ...VersionPropertySetter) error

	// HIERARCHY-RELATED OPERATIONS
	AddHierarchy(hierarchy *Hierarchy) (int, error)
	GetHierarchy(id int, opts ...HierarchyGetOption) (*Hierarchy, error)
	GetHierarchyByParentEntity(parentEntityID int, childStepID string, specificType EntityType) (*Hierarchy, error)
	HasHierarchy(id int) (bool, error)
	UpdateHierarchy(hierarchy *Hierarchy) error
	GetHierarchyProperties(id int, getters ...HierarchyPropertyGetter) error
	SetHierarchyProperties(id int, setters ...HierarchyPropertySetter) error
	// specific one, might refactor that
	GetHierarchiesByChildEntity(childEntityID int) ([]*Hierarchy, error)
	GetHierarchiesByParentEntity(parentEntityID int) ([]*Hierarchy, error)

	// QUEUE-RELATED OPERATIONS
	AddQueue(queue *Queue) (int, error)
	GetQueue(id int, opts ...QueueGetOption) (*Queue, error)
	GetQueueByName(name string, opts ...QueueGetOption) (*Queue, error)
	HasQueue(id int) (bool, error)
	HasQueueName(name string) (bool, error)
	UpdateQueue(queue *Queue) error
	GetQueueProperties(id int, getters ...QueuePropertyGetter) error
	SetQueueProperties(id int, setters ...QueuePropertySetter) error
}

// RetryPolicy helper functions
func DefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxAttempts: 1, // 1 means no retries
		MaxInterval: 100 * time.Millisecond,
	}
}

func DefaultRetryPolicyInternal() *retryPolicyInternal {
	return &retryPolicyInternal{
		MaxAttempts: 1, // 1 means no retries
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

	// Fill default values if zero
	if rp.MaxAttempts == 0 {
		rp.MaxAttempts = 1
	}

	if rp.MaxInterval == 0 {
		rp.MaxInterval = 5 * time.Minute
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

func convertSingleOutputFromSerialization(outputType reflect.Type, executionOutput []byte) (interface{}, error) {
	buf := bytes.NewBuffer(executionOutput)

	decodedObj := reflect.New(outputType).Interface()

	if err := rtl.Decode(buf, decodedObj); err != nil {
		return nil, err
	}

	return reflect.ValueOf(decodedObj).Elem().Interface(), nil
}

// Copy functions
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

func copyBaseEntity(base *BaseEntity) *BaseEntity {
	if base == nil {
		return nil
	}

	return &BaseEntity{
		ID:          base.ID,
		HandlerName: base.HandlerName,
		Type:        base.Type,
		Status:      base.Status,
		QueueID:     base.QueueID,
		StepID:      base.StepID,
		CreatedAt:   base.CreatedAt,
		UpdatedAt:   base.UpdatedAt,
		RunID:       base.RunID,
		RetryPolicy: base.RetryPolicy,
		RetryState:  base.RetryState,
	}
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

// Entity Data copy functions
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

func copyWorkflowEntityEdges(e *WorkflowEntityEdges) *WorkflowEntityEdges {
	if e == nil {
		return nil
	}
	copy := *e
	if e.Queue != nil {
		copy.Queue = copyQueue(e.Queue)
	}
	if e.Versions != nil {
		copy.Versions = make([]*Version, len(e.Versions))
		for i, v := range e.Versions {
			copy.Versions[i] = copyVersion(v)
		}
	}
	if e.ActivityChildren != nil {
		copy.ActivityChildren = make([]*ActivityEntity, len(e.ActivityChildren))
		for i, a := range e.ActivityChildren {
			copy.ActivityChildren[i] = copyActivityEntity(a)
		}
	}
	if e.SagaChildren != nil {
		copy.SagaChildren = make([]*SagaEntity, len(e.SagaChildren))
		for i, s := range e.SagaChildren {
			copy.SagaChildren[i] = copySagaEntity(s)
		}
	}
	if e.SideEffectChildren != nil {
		copy.SideEffectChildren = make([]*SideEffectEntity, len(e.SideEffectChildren))
		for i, se := range e.SideEffectChildren {
			copy.SideEffectChildren[i] = copySideEffectEntity(se)
		}
	}
	return &copy
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

// Execution Data copy functions
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
		Edges:        copyWorkflowEntityEdges(entity.Edges),
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
