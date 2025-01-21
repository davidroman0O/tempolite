package tests

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/davidroman0O/tempolite"
)

func TestWorkflowContinueAsNew(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	defer database.SaveAsJSON("./jsons/workflows_continueasnew_execute.json")
	ctx := context.Background()
	var orchestrator *tempolite.Orchestrator

	orchestrator = tempolite.NewOrchestrator(
		ctx,
		database,
		registry,
		tempolite.WithContinueAsNew(func(queueName string, queueID tempolite.QueueID, workflowID tempolite.WorkflowEntityID, runID tempolite.RunID, workflowFunc interface{}, options *tempolite.WorkflowOptions, args ...interface{}) tempolite.Future {
			return orchestrator.Execute(workflowFunc, nil, args...)
		}),
	)

	var flagTriggered atomic.Bool
	var workflowFunc func(ctx tempolite.WorkflowContext) error
	workflowFunc = func(ctx tempolite.WorkflowContext) error {
		if !flagTriggered.Load() {
			flagTriggered.Store(true)
			if err := ctx.ContinueAsNew(nil); err != nil {
				return err
			}
		}
		return nil
	}

	future := orchestrator.Execute(workflowFunc, nil)

	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

	if future.WorkflowID() == 0 {
		t.Fatal("workflow ID should not be 0")
	}

	if future.WorkflowExecutionID() == 0 {
		t.Fatal("workflow execution ID should not be 0")
	}

	if err := orchestrator.WaitFor(future.WorkflowID(), tempolite.StatusCompleted); err != nil {
		t.Fatal(err)
	}

	<-time.After(1 * time.Second)
}
