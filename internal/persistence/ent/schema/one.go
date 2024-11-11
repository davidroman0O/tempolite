// Package schema defines the ent schemas for the Tempolite workflow engine
package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// BaseMixin provides common fields for all entities
type BaseMixin struct {
	mixin.Schema
}

func (BaseMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// RetryPolicy defines retry behavior configuration
type RetryPolicy struct {
	MaxAttempts        int     `json:"max_attempts"`
	InitialInterval    int64   `json:"initial_interval"`
	BackoffCoefficient float64 `json:"backoff_coefficient"`
	MaxInterval        int64   `json:"max_interval"`
}

// RetryState
type RetryState struct {
	Attempts int `json:"attempts"`
}

// Run represents a workflow execution instance
type Run struct {
	ent.Schema
}

func (Run) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

func (Run) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("status").
			Values("Pending", "Running", "Completed", "Failed", "Cancelled").
			Default("Pending"),
	}
}

func (Run) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("entities", Entity.Type),
		edge.To("hierarchies", Hierarchy.Type),
	}
}

// Entity is the base schema for all executable components
type Entity struct {
	ent.Schema
}

func (Entity) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

func (Entity) Fields() []ent.Field {
	return []ent.Field{
		field.String("handler_name").
			NotEmpty(),
		field.Enum("type").
			Values("Workflow", "Activity", "Saga", "SideEffect").
			Immutable(),
		field.Enum("status").
			Values("Pending", "Running", "Completed", "Failed", "Retried", "Cancelled", "Paused").
			Default("Pending"),
		field.String("step_id").
			Unique(),
	}
}

func (Entity) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("run", Run.Type).
			Ref("entities").
			Unique().
			Required(),
		edge.To("executions", Execution.Type),
		edge.From("queue", Queue.Type).
			Ref("entities").
			Unique(),
		edge.To("versions", Version.Type),
		edge.To("workflow_data", WorkflowData.Type).
			Unique(),
		edge.To("activity_data", ActivityData.Type).
			Unique(),
		edge.To("saga_data", SagaData.Type).
			Unique(),
		edge.To("side_effect_data", SideEffectData.Type).
			Unique(),
	}
}

// Execution represents a single execution attempt
type Execution struct {
	ent.Schema
}

func (Execution) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

func (Execution) Fields() []ent.Field {
	return []ent.Field{
		field.Time("started_at").
			Default(time.Now),
		field.Time("completed_at").
			Optional().
			Nillable(),
		field.Enum("status").
			Values("Pending", "Running", "Completed", "Failed", "Retried", "Cancelled", "Paused").
			Default("Pending"),
	}
}

func (Execution) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("entity", Entity.Type).
			Ref("executions").
			Unique().
			Required(),
		edge.To("workflow_execution", WorkflowExecution.Type).
			Unique(),
		edge.To("activity_execution", ActivityExecution.Type).
			Unique(),
		edge.To("saga_execution", SagaExecution.Type).
			Unique(),
		edge.To("side_effect_execution", SideEffectExecution.Type).
			Unique(),
	}
}

// Queue represents a work queue
type Queue struct {
	ent.Schema
}

func (Queue) Mixin() []ent.Mixin {
	return []ent.Mixin{
		BaseMixin{},
	}
}

func (Queue) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			Unique(),
	}
}

func (Queue) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("entities", Entity.Type),
	}
}

// Version tracks entity versions
type Version struct {
	ent.Schema
}

func (Version) Fields() []ent.Field {
	return []ent.Field{
		field.Int("version").
			Positive(),
		field.JSON("data", map[string]interface{}{}),
	}
}

func (Version) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("entity", Entity.Type).
			Ref("versions").
			Unique().
			Required(),
	}
}

// Hierarchy tracks parent-child relationships
type Hierarchy struct {
	ent.Schema
}

func (Hierarchy) Fields() []ent.Field {
	return []ent.Field{
		field.Int("run_id"),
		field.Int("parent_entity_id"),
		field.Int("child_entity_id"),
		field.Int("parent_execution_id"),
		field.Int("child_execution_id"),
		field.String("parent_step_id"),
		field.String("child_step_id"),
		field.Enum("childType").
			Values("Workflow", "Activity", "Saga", "SideEffect").
			Immutable(),
		field.Enum("parentType").
			Values("Workflow", "Activity", "Saga", "SideEffect").
			Immutable(),
	}
}

func (Hierarchy) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("run", Run.Type).
			Ref("hierarchies").
			Field("run_id").
			Unique().
			Required(),
		edge.To("parent_entity", Entity.Type).
			Unique().
			Required().
			Field("parent_entity_id"),
		edge.To("child_entity", Entity.Type).
			Unique().
			Required().
			Field("child_entity_id"),
	}
}

// Component-specific data schemas

type WorkflowData struct {
	ent.Schema
}

func (WorkflowData) Fields() []ent.Field {
	return []ent.Field{
		field.String("duration").
			Optional(),
		field.Bool("paused").
			Default(false),
		field.Bool("resumable").
			Default(false),
		field.JSON("retry_state", &RetryState{}).
			Default(&RetryState{Attempts: 0}),
		field.JSON("retry_policy", &RetryPolicy{}).
			Default(&RetryPolicy{
				MaxAttempts:        1,
				InitialInterval:    1000000000, // 1 second in nanoseconds
				BackoffCoefficient: 2.0,
				MaxInterval:        300000000000, // 5 minutes in nanoseconds
			}),
		field.JSON("input", [][]byte{}).
			Optional(),
	}
}

func (WorkflowData) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("entity", Entity.Type).
			Ref("workflow_data").
			Unique().
			Required(),
	}
}

type ActivityData struct {
	ent.Schema
}

func (ActivityData) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("timeout").
			Optional(),
		field.Int("max_attempts").
			Default(1),
		field.Time("scheduled_for").
			Optional(),
		field.JSON("retry_policy", &RetryPolicy{}).
			Default(&RetryPolicy{
				MaxAttempts:        1,
				InitialInterval:    1000000000, // 1 second in nanoseconds
				BackoffCoefficient: 2.0,
				MaxInterval:        300000000000, // 5 minutes in nanoseconds
			}),
		field.JSON("input", [][]byte{}).
			Optional(),
		field.JSON("output", [][]byte{}).
			Optional(),
	}
}

func (ActivityData) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("entity", Entity.Type).
			Ref("activity_data").
			Unique().
			Required(),
	}
}

type SagaData struct {
	ent.Schema
}

func (SagaData) Fields() []ent.Field {
	return []ent.Field{
		field.Bool("compensating").
			Default(false),
		field.JSON("compensation_data", [][]byte{}).
			Optional(),
	}
}

func (SagaData) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("entity", Entity.Type).
			Ref("saga_data").
			Unique().
			Required(),
	}
}

type SideEffectData struct {
	ent.Schema
}

func (SideEffectData) Fields() []ent.Field {
	return []ent.Field{
		field.JSON("input", [][]byte{}).
			Optional(),
		field.JSON("output", [][]byte{}).
			Optional(),
	}
}

func (SideEffectData) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("entity", Entity.Type).
			Ref("side_effect_data").
			Unique().
			Required(),
	}
}

// Execution-specific schemas

type WorkflowExecution struct {
	ent.Schema
}

func (WorkflowExecution) Fields() []ent.Field {
	return []ent.Field{}
}

func (WorkflowExecution) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("execution", Execution.Type).
			Ref("workflow_execution").
			Unique().
			Required(),
		edge.To("execution_data", WorkflowExecutionData.Type).
			Unique(),
	}
}

type ActivityExecution struct {
	ent.Schema
}

func (ActivityExecution) Fields() []ent.Field {
	return []ent.Field{
		field.Int("attempt").
			Default(1),
		field.JSON("input", [][]byte{}).
			Optional(),
	}
}

func (ActivityExecution) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("execution", Execution.Type).
			Ref("activity_execution").
			Unique().
			Required(),
		edge.To("execution_data", ActivityExecutionData.Type).
			Unique(),
	}
}

type SagaExecution struct {
	ent.Schema
}

func (SagaExecution) Fields() []ent.Field {
	return []ent.Field{
		field.Enum("step_type").
			Values("transaction", "compensation"),
		field.JSON("compensation_data", []byte{}).
			Optional(),
	}
}

func (SagaExecution) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("execution", Execution.Type).
			Ref("saga_execution").
			Unique().
			Required(),
		edge.To("execution_data", SagaExecutionData.Type).
			Unique(),
	}
}

type SideEffectExecution struct {
	ent.Schema
}

func (SideEffectExecution) Fields() []ent.Field {
	return []ent.Field{
		field.JSON("result", []byte{}).
			Optional(),
	}
}

func (SideEffectExecution) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("execution", Execution.Type).
			Ref("side_effect_execution").
			Unique().
			Required(),
		edge.To("execution_data", SideEffectExecutionData.Type).
			Unique(),
	}
}

// Execution Data schemas

type WorkflowExecutionData struct {
	ent.Schema
}

func (WorkflowExecutionData) Fields() []ent.Field {
	return []ent.Field{
		field.String("error").
			Optional(),
		field.JSON("output", [][]byte{}).
			Optional(),
	}
}

func (WorkflowExecutionData) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("workflow_execution", WorkflowExecution.Type).
			Ref("execution_data").
			Unique().
			Required(),
	}
}

type ActivityExecutionData struct {
	ent.Schema
}

func (ActivityExecutionData) Fields() []ent.Field {
	return []ent.Field{
		field.JSON("heartbeats", [][]byte{}).
			Optional(),
		field.Time("last_heartbeat").
			Optional().
			Nillable(),
		field.JSON("progress", []byte{}).
			Optional(),
		field.JSON("execution_details", []byte{}).
			Optional(),
	}
}

func (ActivityExecutionData) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("activity_execution", ActivityExecution.Type).
			Ref("execution_data").
			Unique().
			Required(),
	}
}

type SagaExecutionData struct {
	ent.Schema
}

func (SagaExecutionData) Fields() []ent.Field {
	return []ent.Field{
		field.JSON("transaction_history", [][]byte{}).
			Optional(),
		field.JSON("compensation_history", [][]byte{}).
			Optional(),
		field.Time("last_transaction").
			Optional().
			Nillable(),
	}
}

func (SagaExecutionData) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("saga_execution", SagaExecution.Type).
			Ref("execution_data").
			Unique().
			Required(),
	}
}

type SideEffectExecutionData struct {
	ent.Schema
}

func (SideEffectExecutionData) Fields() []ent.Field {
	return []ent.Field{
		field.Time("effect_time").
			Optional().
			Nillable(),
		field.JSON("effect_metadata", []byte{}).
			Optional(),
		field.JSON("execution_context", []byte{}).
			Optional(),
	}
}

func (SideEffectExecutionData) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("side_effect_execution", SideEffectExecution.Type).
			Ref("execution_data").
			Unique().
			Required(),
	}
}

// Add SQL table annotations for all schemas
func (Run) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "runs"},
	}
}

func (Entity) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "entities"},
	}
}

func (Execution) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "executions"},
	}
}

func (Queue) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "queues"},
	}
}

func (Version) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "versions"},
	}
}

func (Hierarchy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "hierarchies"},
	}
}

func (WorkflowData) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "workflow_data"},
	}
}

func (ActivityData) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "activity_data"},
	}
}

func (SagaData) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "saga_data"},
	}
}

func (SideEffectData) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "side_effect_data"},
	}
}

func (WorkflowExecution) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "workflow_executions"},
	}
}

func (ActivityExecution) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "activity_executions"},
	}
}

func (SagaExecution) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "saga_executions"},
	}
}

func (SideEffectExecution) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "side_effect_executions"},
	}
}

func (WorkflowExecutionData) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "workflow_execution_data"},
	}
}

func (ActivityExecutionData) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "activity_execution_data"},
	}
}

func (SagaExecutionData) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "saga_execution_data"},
	}
}

func (SideEffectExecutionData) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "side_effect_execution_data"},
	}
}
