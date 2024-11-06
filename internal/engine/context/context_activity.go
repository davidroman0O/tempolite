package context

import "context"

type ActivityContext struct {
	context.Context
	activityID  string
	executionID string
	runID       string
	stepID      string
}
