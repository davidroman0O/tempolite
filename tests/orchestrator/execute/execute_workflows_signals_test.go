package tests

import (
	"context"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	tempolite "github.com/davidroman0O/tempolite"
)

func TestWorkflowExecuteSignals(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	defer database.SaveAsJSON("./jsons/workflows_basic_execute_signals.json")
	ctx := context.Background()

	snmap := sync.Map{}

	orchestrator := tempolite.NewOrchestrator(
		ctx,
		database,
		registry,
		tempolite.WithSignalNew(
			func(
				workflowID tempolite.WorkflowEntityID,
				workflowExecutionID tempolite.WorkflowExecutionID,
				signalEntityID tempolite.SignalEntityID,
				signalExecutionID tempolite.SignalExecutionID,
				signal string,
				future tempolite.Future) error {

				snmap.Store(signal, future)

				return nil
			}),
		tempolite.WithSignalRemove(
			func(
				workflowID tempolite.WorkflowEntityID,
				workflowExecutionID tempolite.WorkflowExecutionID,
				signalEntityID tempolite.SignalEntityID,
				signalExecutionID tempolite.SignalExecutionID,
				signal string) error {

				snmap.Delete(signal)

				return nil
			}),
	)

	var flagTriggered atomic.Bool
	var value atomic.Int32

	workflowFunc := func(ctx tempolite.WorkflowContext) error {
		flagTriggered.Store(true)
		var life int
		if err := ctx.Signal("life", &life); err != nil {
			return err
		}
		value.Store(int32(life))
		return nil
	}

	future := orchestrator.Execute(workflowFunc, nil)

	go func() {
		<-time.After(time.Second)
		signalFuture, ok := snmap.Load("life")
		if !ok {
			t.Fatal("signal not found")
		}
		signalFuture.(tempolite.Future).SetResult([]interface{}{42})
	}()

	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

	if future.WorkflowID() == 0 {
		t.Fatal("workflow ID should not be 0")
	}

	if future.WorkflowExecutionID() == 0 {
		t.Fatal("workflow execution ID should not be 0")
	}

	if !flagTriggered.Load() {
		t.Fatal("workflowFunc was not triggered")
	}
	if value.Load() != 42 {
		t.Fatalf("expected value %d, got %d", 42, value.Load())
	}

	we, err := database.GetWorkflowEntity(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if we.Status != tempolite.StatusCompleted {
		t.Fatalf("expected status %s, got %s", tempolite.StatusCompleted, we.Status)
	}

	if we.StepID != "root" {
		t.Fatalf("expected stepID %s, got %s", "root", we.StepID)
	}

	if we.RunID != 1 {
		t.Fatalf("expected runID %d, got %d", 1, we.RunID)
	}

	execs, err := database.GetWorkflowExecutions(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if len(execs) != 1 {
		t.Fatalf("expected 1 execution, got %d", len(execs))
	}

	sort.Slice(execs, func(i, j int) bool {
		return execs[i].ID < execs[j].ID
	})

	if execs[0].Status != tempolite.ExecutionStatusCompleted {
		t.Fatalf("expected status %s, got %s", tempolite.ExecutionStatusCompleted, execs[0].Status)
	}

	if execs[0].Error != "" {
		t.Fatal("expected no error")
	}
}

// func TestWorkflowExecuteSignalsFailure(t *testing.T) {
// 	registry := tempolite.NewRegistry()
// 	database := tempolite.NewMemoryDatabase()
// 	defer database.SaveAsJSON("./jsons/workflows_basic_execute_failure.json")
// 	ctx := context.Background()

// 	orchestrator := tempolite.NewOrchestrator(ctx, database, registry)

// 	workflowFunc := func(ctx tempolite.WorkflowContext) error {
// 		var life string
// 		if err := ctx.Signal("life", &life); err != nil {
// 			return err
// 		}
// 		value.Store(int32(life))
// 		return nil
// 	}

// 	future := orchestrator.Execute(workflowFunc, nil)

// 	if err := future.Get(); err == nil {
// 		t.Fatal("expected error")
// 	}

// 	we, err := database.GetWorkflowEntity(future.WorkflowID())
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if we.Status != tempolite.StatusFailed {
// 		t.Fatalf("expected status %s, got %s", tempolite.StatusFailed, we.Status)
// 	}

// 	if we.StepID != "root" {
// 		t.Fatalf("expected stepID %s, got %s", "root", we.StepID)
// 	}

// 	if we.RunID != 1 {
// 		t.Fatalf("expected runID %d, got %d", 1, we.RunID)
// 	}

// 	execs, err := database.GetWorkflowExecutions(future.WorkflowID())
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if len(execs) != 1 {
// 		t.Fatalf("expected 1 execution, got %d", len(execs))
// 	}

// 	sort.Slice(execs, func(i, j int) bool {
// 		return execs[i].ID < execs[j].ID
// 	})

// 	if execs[0].Status != tempolite.ExecutionStatusFailed {
// 		t.Fatalf("expected status %s, got %s", tempolite.ExecutionStatusFailed, execs[0].Status)
// 	}

// 	if execs[0].Error == "" {
// 		t.Fatal("expected error")
// 	}
// }

// func TestWorkflowExecuteSignalsPanic(t *testing.T) {
// 	registry := tempolite.NewRegistry()
// 	database := tempolite.NewMemoryDatabase()
// 	defer database.SaveAsJSON("./jsons/workflows_basic_execute_panic.json")
// 	ctx := context.Background()

// 	orchestrator := tempolite.NewOrchestrator(ctx, database, registry)

// 	workflowFunc := func(ctx tempolite.WorkflowContext) error {
// 		panic("on purpose")
// 	}

// 	future := orchestrator.Execute(workflowFunc, nil)

// 	if err := future.Get(); err == nil {
// 		t.Fatal("expected error")
// 	}

// 	we, err := database.GetWorkflowEntity(future.WorkflowID())
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if we.Status != tempolite.StatusFailed {
// 		t.Fatalf("expected status %s, got %s", tempolite.StatusFailed, we.Status)
// 	}

// 	if we.StepID != "root" {
// 		t.Fatalf("expected stepID %s, got %s", "root", we.StepID)
// 	}

// 	if we.RunID != 1 {
// 		t.Fatalf("expected runID %d, got %d", 1, we.RunID)
// 	}

// 	execs, err := database.GetWorkflowExecutions(future.WorkflowID())
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if len(execs) != 1 {
// 		t.Fatalf("expected 1 execution, got %d", len(execs))
// 	}

// 	sort.Slice(execs, func(i, j int) bool {
// 		return execs[i].ID < execs[j].ID
// 	})

// 	if execs[0].Status != tempolite.ExecutionStatusFailed {
// 		t.Fatalf("expected status %s, got %s", tempolite.ExecutionStatusFailed, execs[0].Status)
// 	}

// 	if execs[0].Error == "" {
// 		t.Fatal("expected error")
// 	}
// 	if execs[0].StackTrace == nil {
// 		t.Fatal("expected error")
// 	}
// }
