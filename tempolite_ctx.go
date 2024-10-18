package tempolite

import "context"

type TempoliteContext interface {
	context.Context
	RunID() string
	EntityID() string
	ExecutionID() string
	EntityType() string
	StepID() string
}
