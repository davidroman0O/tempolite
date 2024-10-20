package tempolite

type SideEffectContext[T Identifier] struct {
	TempoliteContext
	tp           *Tempolite[T]
	sideEffectID string
	executionID  string
	runID        string
	stepID       string
}

func (w SideEffectContext[T]) StepID() string {
	return w.stepID
}

func (w SideEffectContext[T]) RunID() string {
	return w.runID
}

func (w SideEffectContext[T]) EntityID() string {
	return w.sideEffectID
}

func (w SideEffectContext[T]) ExecutionID() string {
	return w.executionID
}

func (w SideEffectContext[T]) EntityType() string {
	return "sideEffect"
}
