package types

type SharedContext interface {
	RunID() int         // root ID of execution
	EntityID() int      // parent entity
	ExecutionID() int   // current execution
	EntityType() string // entity type
	StepID() string     // current step
	QueueName() string  // queue name
}
