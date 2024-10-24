package tempolite

type SideEffectContext struct {
	TempoliteContext
	tp           *Tempolite
	sideEffectID string
	executionID  string
	runID        string
	stepID       string
}

func (w SideEffectContext) StepID() string {
	return w.stepID
}

func (w SideEffectContext) RunID() string {
	return w.runID
}

func (w SideEffectContext) EntityID() string {
	return w.sideEffectID
}

func (w SideEffectContext) ExecutionID() string {
	return w.executionID
}

func (w SideEffectContext) EntityType() string {
	return "sideEffect"
}
