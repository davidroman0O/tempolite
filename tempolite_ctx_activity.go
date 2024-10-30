package tempolite

import "context"

type ActivityContext struct {
	context.Context
	tp          *Tempolite
	activityID  string
	executionID string
	runID       string
	stepID      string
}

func (w ActivityContext) StepID() string {
	return w.stepID
}

func (w ActivityContext) RunID() string {
	return w.runID
}

func (w ActivityContext) EntityID() string {
	return w.activityID
}

func (w ActivityContext) ExecutionID() string {
	return w.executionID
}

func (w ActivityContext) EntityType() string {
	return "activity"
}

func (w ActivityContext) GetActivity(id ActivityExecutionID) *ActivityExecutionInfo {
	return w.tp.getActivityExecution(w, id, nil)
}
