// schema/ids.go
package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

type ID int

// Entity IDs
type RunID ID
type QueueID ID
type VersionID ID
type WorkflowEntityID ID
type ActivityEntityID ID
type SagaEntityID ID
type SideEffectEntityID ID
type SignalEntityID ID
type HierarchyID ID

// Execution IDs
type WorkflowExecutionID ID
type ActivityExecutionID ID
type SagaExecutionID ID
type SideEffectExecutionID ID
type SignalExecutionID ID

// Data IDs
type WorkflowDataID ID
type ActivityDataID ID
type SagaDataID ID
type SideEffectDataID ID
type SignalDataID ID
type WorkflowExecutionDataID ID
type ActivityExecutionDataID ID
type SagaExecutionDataID ID
type SideEffectExecutionDataID ID
type SignalExecutionDataID ID
type SagaValueID ID

// Status types
type RunStatus string
type EntityStatus string
type ExecutionStatus string
type EntityType string
type ExecutionType string

// Version types
type VersionChange string
type VersionNumber uint

// Run status constants
const (
	RunStatusPending   RunStatus = "Pending"
	RunStatusRunning   RunStatus = "Running"
	RunStatusPaused    RunStatus = "Paused"
	RunStatusCancelled RunStatus = "Cancelled"
	RunStatusCompleted RunStatus = "Completed"
	RunStatusFailed    RunStatus = "Failed"
)

// Entity status constants
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

// Execution status constants
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

// Entity type constants
const (
	EntityWorkflow   EntityType = "workflow"
	EntityActivity   EntityType = "activity"
	EntitySaga       EntityType = "saga"
	EntitySideEffect EntityType = "side_effect"
	EntitySignal     EntityType = "signal"
)

// Execution type constants
const (
	ExecutionTypeTransaction  ExecutionType = "transaction"
	ExecutionTypeCompensation ExecutionType = "compensation"
)

// RetryPolicy and State
type RetryPolicy struct {
	MaxAttempts uint64 `json:"max_attempts"`
	MaxInterval int64  `json:"max_interval"`
}

type RetryState struct {
	Attempts uint64 `json:"attempts"`
	Timeout  int64  `json:"timeout,omitempty"`
}

// Run schema definition
type Run struct {
	ent.Schema
}

func (Run) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			GoType(RunID(0)),
		field.String("status").
			GoType(RunStatus("")),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (Run) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("workflows", WorkflowEntity.Type),
		edge.To("hierarchies", Hierarchy.Type),
	}
}

// Queue schema definition
type Queue struct {
	ent.Schema
}

func (Queue) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			GoType(QueueID(0)),
		field.String("name").
			Unique(),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (Queue) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("entities", WorkflowEntity.Type),
	}
}

// Version schema definition
type Version struct {
	ent.Schema
}

func (Version) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			GoType(VersionID(0)),
		field.Int("entity_id").
			GoType(WorkflowEntityID(0)),
		field.String("change_id").
			GoType(VersionChange("")),
		field.Int("version").
			GoType(VersionNumber(0)).
			Default(0),
		field.JSON("data", map[string]interface{}{}), // TODO: not used yet, might just be map[string]string
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (Version) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workflow", WorkflowEntity.Type).
			Ref("versions").
			Unique().
			Required(). // Add this
			Field("entity_id"),
	}
}

// WorkflowEntity schema definition
type WorkflowEntity struct {
	ent.Schema
}

func (WorkflowEntity) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			GoType(WorkflowEntityID(0)),
		field.String("handler_name"),
		field.String("type").
			GoType(EntityType("")).
			Default(string(EntityWorkflow)),
		field.String("status").
			GoType(EntityStatus("")).
			Default(string(StatusPending)),
		field.String("step_id"),
		field.Int("run_id").
			GoType(RunID(0)),
		field.JSON("retry_policy", RetryPolicy{}),
		field.JSON("retry_state", RetryState{}),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (WorkflowEntity) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("queue", Queue.Type).
			Ref("entities").
			Unique(),
		edge.From("run", Run.Type).
			Ref("entities").
			Unique().
			Required(). // Add this
			Field("run_id"),
		edge.To("versions", Version.Type),
		edge.To("workflow_data", WorkflowData.Type).
			Unique(),
		edge.To("activity_children", ActivityEntity.Type),
		edge.To("saga_children", SagaEntity.Type),
		edge.To("side_effect_children", SideEffectEntity.Type),
		edge.To("executions", WorkflowExecution.Type),
	}
}

// WorkflowData schema definition
type WorkflowData struct {
	ent.Schema
}

func (WorkflowData) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			GoType(WorkflowDataID(0)),
		field.Int("entity_id").
			GoType(WorkflowEntityID(0)),
		field.String("duration").
			Optional(),
		field.Bool("paused"),
		field.Bool("resumable"),
		field.Bool("is_root"),
		field.Bytes("inputs").
			Optional(),
		field.Int("continued_from").
			Optional().
			Nillable().
			GoType(WorkflowEntityID(0)),
		field.Int("continued_execution_from").
			Optional().
			Nillable().
			GoType(WorkflowExecutionID(0)),
		field.String("workflow_step_id").
			Optional().
			Nillable(),
		field.Int("workflow_from").
			Optional().
			Nillable().
			GoType(WorkflowEntityID(0)),
		field.Int("workflow_execution_from").
			Optional().
			Nillable().
			GoType(WorkflowExecutionID(0)),
		field.JSON("versions", map[string]int{}),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (WorkflowData) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workflow", WorkflowEntity.Type).
			Ref("workflow_data").
			Unique().
			Required(). // Add this
			Field("entity_id"),
	}
}

// WorkflowExecution schema definition
type WorkflowExecution struct {
	ent.Schema
}

func (WorkflowExecution) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			GoType(WorkflowExecutionID(0)),
		field.Int("workflow_entity_id").
			GoType(WorkflowEntityID(0)),
		field.Time("started_at"),
		field.Time("completed_at").
			Optional().
			Nillable(),
		field.String("status").
			GoType(ExecutionStatus("")).
			Default(string(ExecutionStatusPending)),
		field.String("error").
			Optional(),
		field.String("stack_trace").
			Optional(),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (WorkflowExecution) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workflow", WorkflowEntity.Type).
			Ref("executions").
			Unique().
			Required(). // Add this
			Field("workflow_entity_id"),
		edge.To("execution_data", WorkflowExecutionData.Type).
			Unique(),
	}
}

// WorkflowExecutionData schema definition
type WorkflowExecutionData struct {
	ent.Schema
}

func (WorkflowExecutionData) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			GoType(WorkflowExecutionDataID(0)),
		field.Int("execution_id").
			GoType(WorkflowExecutionID(0)),
		field.Time("last_heartbeat").
			Optional().
			Nillable(),
		field.Bytes("outputs").
			Optional(),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (WorkflowExecutionData) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("execution", WorkflowExecution.Type).
			Ref("execution_data").
			Unique().
			Required(). // Add this
			Field("execution_id"),
	}
}

// ActivityEntity schema definition
type ActivityEntity struct {
	ent.Schema
}

func (ActivityEntity) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			GoType(ActivityEntityID(0)),
		field.String("handler_name"),
		field.String("type").
			GoType(EntityType("")).
			Default(string(EntityActivity)),
		field.String("status").
			GoType(EntityStatus("")).
			Default(string(StatusPending)),
		field.String("step_id"),
		field.Int("run_id").
			GoType(RunID(0)),
		field.JSON("retry_policy", RetryPolicy{}),
		field.JSON("retry_state", RetryState{}),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Also need to update ActivityEntity to include executions edge:
func (ActivityEntity) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workflow", WorkflowEntity.Type).
			Ref("activity_children").
			Unique().
			Required(), // Added Required()
		edge.To("activity_data", ActivityData.Type).
			Unique(),
		edge.To("executions", ActivityExecution.Type),
	}
}

// ActivityData schema definition
type ActivityData struct {
	ent.Schema
}

func (ActivityData) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			GoType(ActivityDataID(0)),
		field.Int("entity_id").
			GoType(ActivityEntityID(0)),
		field.Bytes("inputs").
			Optional(),
		field.Bytes("output").
			Optional(),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (ActivityData) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("activity", ActivityEntity.Type).
			Ref("activity_data").
			Unique().
			Required(). // Add this
			Field("entity_id"),
	}
}

// ActivityExecution schema definition
type ActivityExecution struct {
	ent.Schema
}

func (ActivityExecution) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			GoType(ActivityExecutionID(0)),
		field.Int("activity_entity_id").
			GoType(ActivityEntityID(0)),
		field.Time("started_at"),
		field.Time("completed_at").
			Optional().
			Nillable(),
		field.String("status").
			GoType(ExecutionStatus("")).
			Default(string(ExecutionStatusPending)),
		field.String("error").
			Optional(),
		field.String("stack_trace").
			Optional(),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (ActivityExecution) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("activity", ActivityEntity.Type).
			Ref("executions").
			Unique().
			Required().
			Field("activity_entity_id"),
		edge.To("execution_data", ActivityExecutionData.Type).
			Unique(),
	}
}

// ActivityExecutionData schema definition
type ActivityExecutionData struct {
	ent.Schema
}

func (ActivityExecutionData) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			GoType(ActivityExecutionDataID(0)),
		field.Int("execution_id").
			GoType(ActivityExecutionID(0)),
		field.Time("last_heartbeat").
			Optional().
			Nillable(),
		field.Bytes("outputs").
			Optional(),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (ActivityExecutionData) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("execution", ActivityExecution.Type).
			Ref("execution_data").
			Unique().
			Required().
			Field("execution_id"),
	}
}

type SagaEntity struct {
	ent.Schema
}

func (SagaEntity) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			GoType(SagaEntityID(0)),
		field.String("handler_name"),
		field.String("type").
			GoType(EntityType("")).
			Default(string(EntitySaga)),
		field.String("status").
			GoType(EntityStatus("")).
			Default(string(StatusPending)),
		field.String("step_id"),
		field.Int("run_id").
			GoType(RunID(0)),
		field.JSON("retry_policy", RetryPolicy{}),
		field.JSON("retry_state", RetryState{}),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (SagaEntity) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workflow", WorkflowEntity.Type).
			Ref("saga_children").
			Unique().
			Required(),
		edge.To("saga_data", SagaData.Type).
			Unique(),
		edge.To("executions", SagaExecution.Type),
		edge.To("values", SagaValue.Type),
	}
}

type SagaData struct {
	ent.Schema
}

func (SagaData) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			GoType(SagaDataID(0)),
		field.Int("entity_id").
			GoType(SagaEntityID(0)),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (SagaData) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("saga", SagaEntity.Type).
			Ref("saga_data").
			Unique().
			Required().
			Field("entity_id"),
		edge.To("values", SagaValue.Type),
	}
}

type SagaExecution struct {
	ent.Schema
}

func (SagaExecution) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			GoType(SagaExecutionID(0)),
		field.Int("saga_entity_id").
			GoType(SagaEntityID(0)),
		field.String("execution_type").
			GoType(ExecutionType("")).
			Default(string(ExecutionTypeTransaction)),
		field.Time("started_at"),
		field.Time("completed_at").
			Optional().
			Nillable(),
		field.String("status").
			GoType(ExecutionStatus("")).
			Default(string(ExecutionStatusPending)),
		field.String("error").
			Optional(),
		field.String("stack_trace").
			Optional(),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (SagaExecution) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("saga", SagaEntity.Type).
			Ref("executions").
			Unique().
			Required().
			Field("saga_entity_id"),
		edge.To("execution_data", SagaExecutionData.Type).
			Unique(),
		edge.To("values", SagaValue.Type),
	}
}

type SagaExecutionData struct {
	ent.Schema
}

func (SagaExecutionData) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			GoType(SagaExecutionDataID(0)),
		field.Int("execution_id").
			GoType(SagaExecutionID(0)),
		field.Time("last_heartbeat").
			Optional().
			Nillable(),
		field.Int("step_index"),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (SagaExecutionData) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("execution", SagaExecution.Type).
			Ref("execution_data").
			Unique().
			Required().
			Field("execution_id"),
		edge.To("values", SagaValue.Type),
	}
}

type SagaValue struct {
	ent.Schema
}

func (SagaValue) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			GoType(SagaValueID(0)),
		field.String("key"),
		field.Bytes("value"),
		field.Int("saga_entity_id").
			GoType(SagaEntityID(0)),
		field.Int("saga_execution_id").
			GoType(SagaExecutionID(0)),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (SagaValue) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("execution", SagaExecution.Type).
			Ref("values").
			Unique().
			Required().
			Field("saga_execution_id"),
		edge.From("saga_data", SagaData.Type).
			Ref("values").
			Unique(),
		edge.From("execution_data", SagaExecutionData.Type).
			Ref("values").
			Unique(),
	}
}

type SideEffectEntity struct {
	ent.Schema
}

func (SideEffectEntity) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			GoType(SideEffectEntityID(0)),
		field.String("handler_name"),
		field.String("type").
			GoType(EntityType("")).
			Default(string(EntitySideEffect)),
		field.String("status").
			GoType(EntityStatus("")).
			Default(string(StatusPending)),
		field.String("step_id"),
		field.Int("run_id").
			GoType(RunID(0)),
		field.JSON("retry_policy", RetryPolicy{}),
		field.JSON("retry_state", RetryState{}),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (SideEffectEntity) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workflow", WorkflowEntity.Type).
			Ref("side_effect_children").
			Unique().
			Required(),
		edge.To("side_effect_data", SideEffectData.Type).
			Unique(),
		edge.To("executions", SideEffectExecution.Type),
	}
}

type SideEffectData struct {
	ent.Schema
}

func (SideEffectData) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			GoType(SideEffectDataID(0)),
		field.Int("entity_id").
			GoType(SideEffectEntityID(0)),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (SideEffectData) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("side_effect", SideEffectEntity.Type).
			Ref("side_effect_data").
			Unique().
			Required().
			Field("entity_id"),
	}
}

type SideEffectExecution struct {
	ent.Schema
}

func (SideEffectExecution) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			GoType(SideEffectExecutionID(0)),
		field.Int("side_effect_entity_id").
			GoType(SideEffectEntityID(0)),
		field.Time("started_at"),
		field.Time("completed_at").
			Optional().
			Nillable(),
		field.String("status").
			GoType(ExecutionStatus("")).
			Default(string(ExecutionStatusPending)),
		field.String("error").
			Optional(),
		field.String("stack_trace").
			Optional(),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (SideEffectExecution) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("side_effect", SideEffectEntity.Type).
			Ref("executions").
			Unique().
			Required().
			Field("side_effect_entity_id"),
		edge.To("execution_data", SideEffectExecutionData.Type).
			Unique(),
	}
}

type SideEffectExecutionData struct {
	ent.Schema
}

func (SideEffectExecutionData) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			GoType(SideEffectExecutionDataID(0)),
		field.Int("execution_id").
			GoType(SideEffectExecutionID(0)),
		field.Bytes("outputs").
			Optional(),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (SideEffectExecutionData) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("execution", SideEffectExecution.Type).
			Ref("execution_data").
			Unique().
			Required().
			Field("execution_id"),
	}
}

type SignalEntity struct {
	ent.Schema
}

func (SignalEntity) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			GoType(SignalEntityID(0)),
		field.String("handler_name"),
		field.String("type").
			GoType(EntityType("")).
			Default(string(EntitySignal)),
		field.String("status").
			GoType(EntityStatus("")).
			Default(string(StatusPending)),
		field.String("step_id"),
		field.Int("run_id").
			GoType(RunID(0)),
		field.JSON("retry_policy", RetryPolicy{}),
		field.JSON("retry_state", RetryState{}),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (SignalEntity) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("signal_data", SignalData.Type).
			Unique(),
		edge.To("executions", SignalExecution.Type),
	}
}

type SignalData struct {
	ent.Schema
}

func (SignalData) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			GoType(SignalDataID(0)),
		field.Int("entity_id").
			GoType(SignalEntityID(0)),
		field.String("name"),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (SignalData) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("signal", SignalEntity.Type).
			Ref("signal_data").
			Unique().
			Required().
			Field("entity_id"),
	}
}

type SignalExecution struct {
	ent.Schema
}

func (SignalExecution) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			GoType(SignalExecutionID(0)),
		field.Int("entity_id").
			GoType(SignalEntityID(0)),
		field.Time("started_at"),
		field.Time("completed_at").
			Optional().
			Nillable(),
		field.String("status").
			GoType(ExecutionStatus("")).
			Default(string(ExecutionStatusPending)),
		field.String("error").
			Optional(),
		field.String("stack_trace").
			Optional(),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (SignalExecution) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("signal", SignalEntity.Type).
			Ref("executions").
			Unique().
			Required().
			Field("entity_id"),
		edge.To("execution_data", SignalExecutionData.Type).
			Unique(),
	}
}

type SignalExecutionData struct {
	ent.Schema
}

func (SignalExecutionData) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			GoType(SignalExecutionDataID(0)),
		field.Int("execution_id").
			GoType(SignalExecutionID(0)),
		field.Bytes("value").
			Optional(),
		field.Uint("kind"),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (SignalExecutionData) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("execution", SignalExecution.Type).
			Ref("execution_data").
			Unique().
			Required().
			Field("execution_id"),
	}
}

type Hierarchy struct {
	ent.Schema
}

func (Hierarchy) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			GoType(HierarchyID(0)),
		field.Int("run_id").
			GoType(RunID(0)),
		field.Int("parent_entity_id"),
		field.Int("child_entity_id"),
		field.Int("parent_execution_id"),
		field.Int("child_execution_id"),
		field.String("parent_step_id"),
		field.String("child_step_id"),
		field.String("parent_type").
			GoType(EntityType("")),
		field.String("child_type").
			GoType(EntityType("")),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (Hierarchy) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("run", Run.Type).
			Ref("hierarchies").
			Unique().
			Required().
			Field("run_id"),
	}
}
