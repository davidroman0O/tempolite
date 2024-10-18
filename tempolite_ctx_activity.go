package tempolite

type ActivityContext[T Identifier] struct {
	TempoliteContext
	tp          *Tempolite[T]
	activityID  string
	executionID string
	runID       string
	stepID      string
}

func (w ActivityContext[T]) StepID() string {
	return w.stepID
}

func (w ActivityContext[T]) RunID() string {
	return w.runID
}

func (w ActivityContext[T]) EntityID() string {
	return w.activityID
}

func (w ActivityContext[T]) ExecutionID() string {
	return w.executionID
}

func (w ActivityContext[T]) EntityType() string {
	return "activity"
}

func (w ActivityContext[T]) GetActivity(id ActivityExecutionID) *ActivityExecutionInfo[T] {
	return w.tp.getActivityExecution(w, id, nil)
}
