package tempolite

import "context"

type TransactionContext struct {
	context.Context
	tp *Tempolite
}

func (w TransactionContext) EntityType() string {
	return "transaction"
}

type CompensationContext struct {
	context.Context
	tp *Tempolite
}

func (w CompensationContext) EntityType() string {
	return "compensation"
}
