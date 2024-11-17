package tempolite

import (
	"context"
	"log"
)

// SagaContext provides context for saga execution.
type SagaContext struct {
	ctx          context.Context
	orchestrator *Orchestrator
	workflowID   int
	stepID       string
}

func (ctx *WorkflowContext) Saga(stepID string, saga *SagaDefinition) *SagaInfo {
	if err := ctx.checkPause(); err != nil {
		log.Printf("WorkflowContext.Saga paused at stepID: %s", stepID)
		sagaInfo := &SagaInfo{
			err:  err,
			done: make(chan struct{}),
		}
		close(sagaInfo.done)
		return sagaInfo
	}

	// Execute the saga
	return ctx.orchestrator.executeSaga(ctx, stepID, saga)
}
