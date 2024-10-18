package tempolite

type TransactionContext[T Identifier] struct {
	TempoliteContext
	tp *Tempolite[T]
}

func (w TransactionContext[T]) GetActivity(id string) (*ActivityInfo[T], error) {
	// todo: implement
	return nil, nil
}

func (w TransactionContext[T]) EntityType() string {
	return "transaction"
}

type CompensationContext[T Identifier] struct {
	TempoliteContext
	tp *Tempolite[T]
}

func (w CompensationContext[T]) EntityType() string {
	return "compensation"
}
