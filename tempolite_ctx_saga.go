package tempolite

type TransactionContext struct {
	TempoliteContext
	tp *Tempolite
}

func (w TransactionContext) EntityType() string {
	return "transaction"
}

type CompensationContext struct {
	TempoliteContext
	tp *Tempolite
}

func (w CompensationContext) EntityType() string {
	return "compensation"
}
