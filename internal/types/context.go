package types

import "context"

type SharedContext interface {
	context.Context      // base context
	RunID() string       // root ID of execution
	EntityID() string    // parent entity
	ExecutionID() string // current execution
	EntityType() string  // entity type
	StepID() string      // current step
	QueueName() string   // queue name
}
