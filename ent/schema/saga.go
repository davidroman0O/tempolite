package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Saga holds the schema definition for the Saga entity.
type Saga struct {
	ent.Schema
}

// Fields of the Saga.
func (Saga) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			Unique(),
		field.String("run_id"),
		field.String("step_id").NotEmpty(),
		field.Enum("status").
			Values("Pending", "Running", "Completed", "Failed", "Compensating", "Compensated").
			Default("Pending"),
		field.String("queue_name").
			Default("default").
			NotEmpty(),
		field.JSON("saga_definition", SagaDefinitionData{}).
			Comment("Stores the full saga definition including transactions and compensations"),
		field.String("error").
			Optional(),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Saga.
func (Saga) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("steps", SagaExecution.Type),
	}
}

type SagaStepPair struct {
	TransactionHandlerName  string `json:"transaction_handler_name"`
	CompensationHandlerName string `json:"compensation_handler_name"`
}

type SagaDefinitionData struct {
	Steps []SagaStepPair `json:"steps"`
}
