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

func TestWorkflowExecuteResults(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	defer database.SaveAsJSON("./jsons/workflows_basic_execute_results.json")
	ctx := context.Background()

	orchestrator := tempolite.NewOrchestrator(ctx, database, registry)

	var flagTriggered atomic.Bool

	workflowFunc := func(ctx tempolite.WorkflowContext) (int, error) {
		flagTriggered.Store(true)
		return 42, nil
	}

	future := orchestrator.Execute(workflowFunc, nil)
	var result int
	if err := future.Get(&result); err != nil {
		t.Fatal(err)
	}

	if future.WorkflowID() == 0 {
		t.Fatal("workflow ID should not be 0")
	}

	if future.WorkflowExecutionID() == 0 {
		t.Fatal("workflow execution ID should not be 0")
	}

	if result != 42 {
		t.Fatalf("expected result %d, got %d", 42, result)
	}

	if !flagTriggered.Load() {
		t.Fatal("workflowFunc was not triggered")
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

func TestWorkflowExecuteResultsFailures(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	defer database.SaveAsJSON("./jsons/workflows_basic_execute_results_failure.json")
	ctx := context.Background()

	orchestrator := tempolite.NewOrchestrator(ctx, database, registry)

	var flagTriggered atomic.Bool

	workflowFunc := func(ctx tempolite.WorkflowContext) (int, error) {
		flagTriggered.Store(true)
		return 42, nil
	}

	future := orchestrator.Execute(workflowFunc, nil)
	var result int
	var msg string
	if err := future.Get(&result, &msg); err == nil {
		t.Fatal(err)
	}

	if future.WorkflowID() == 0 {
		t.Fatal("workflow ID should not be 0")
	}

	if future.WorkflowExecutionID() == 0 {
		t.Fatal("workflow execution ID should not be 0")
	}

	if result != 0 {
		t.Fatalf("expected result %d, got %d", 0, result)
	}

	if !flagTriggered.Load() {
		t.Fatal("workflowFunc was not triggered")
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

func TestWorkflowExecuteResultsMany(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	defer database.SaveAsJSON("./jsons/workflows_basic_execute_results_many.json")
	ctx := context.Background()

	orchestrator := tempolite.NewOrchestrator(ctx, database, registry)

	var flagTriggered atomic.Bool

	workflowFunc := func(ctx tempolite.WorkflowContext) (int, string, float32, error) {
		flagTriggered.Store(true)
		return 42, "hello world", 3.14, nil
	}

	future := orchestrator.Execute(workflowFunc, nil)
	var life int
	var msg string
	var pi float32
	if err := future.Get(&life, &msg, &pi); err != nil {
		t.Fatal(err)
	}

	if future.WorkflowID() == 0 {
		t.Fatal("workflow ID should not be 0")
	}

	if future.WorkflowExecutionID() == 0 {
		t.Fatal("workflow execution ID should not be 0")
	}

	if life != 42 {
		t.Fatalf("expected life %d, got %d", 42, life)
	}

	if msg != "hello world" {
		t.Fatalf("expected message %s, got %s", "hello world", msg)
	}

	if pi != 3.14 {
		t.Fatalf("expected pi %f, got %f", 3.14, pi)
	}

	if !flagTriggered.Load() {
		t.Fatal("workflowFunc was not triggered")
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

func TestWorkflowExecuteResultsSideEffect(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	defer database.SaveAsJSON("./jsons/workflows_basic_execute_results_sideeffects.json")
	ctx := context.Background()

	orchestrator := tempolite.NewOrchestrator(ctx, database, registry)

	var flagTriggered atomic.Bool

	workflowFunc := func(ctx tempolite.WorkflowContext) (int, error) {
		flagTriggered.Store(true)
		var life int
		if err := ctx.SideEffect("life", func() int {
			return 42
		}).Get(&life); err != nil {
			return -1, err
		}
		return life, nil
	}

	future := orchestrator.Execute(workflowFunc, nil)
	var life int
	if err := future.Get(&life); err != nil {
		t.Fatal(err)
	}

	if future.WorkflowID() == 0 {
		t.Fatal("workflow ID should not be 0")
	}

	if future.WorkflowExecutionID() == 0 {
		t.Fatal("workflow execution ID should not be 0")
	}

	if life != 42 {
		t.Fatalf("expected life %d, got %d", 42, life)
	}

	if !flagTriggered.Load() {
		t.Fatal("workflowFunc was not triggered")
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

func TestWorkflowExecuteResultsSubWorkflow(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	defer database.SaveAsJSON("./jsons/workflows_basic_execute_results_sub-workflows.json")
	ctx := context.Background()

	orchestrator := tempolite.NewOrchestrator(ctx, database, registry)

	var flagSubTriggered atomic.Bool
	var flagTriggered atomic.Bool

	subworkflowFunc := func(ctx tempolite.WorkflowContext) (int, error) {
		flagSubTriggered.Store(true)
		return 42, nil
	}

	workflowFunc := func(ctx tempolite.WorkflowContext) (int, error) {
		flagTriggered.Store(true)
		var life int
		if err := ctx.Workflow("sub", subworkflowFunc, nil).Get(&life); err != nil {
			return -1, err
		}
		return life, nil
	}

	future := orchestrator.Execute(workflowFunc, nil)
	var life int
	if err := future.Get(&life); err != nil {
		t.Fatal(err)
	}

	if future.WorkflowID() == 0 {
		t.Fatal("workflow ID should not be 0")
	}

	if future.WorkflowExecutionID() == 0 {
		t.Fatal("workflow execution ID should not be 0")
	}

	if life != 42 {
		t.Fatalf("expected life %d, got %d", 42, life)
	}

	if !flagTriggered.Load() {
		t.Fatal("workflowFunc was not triggered")
	}

	if !flagSubTriggered.Load() {
		t.Fatal("subworkflowFunc was not triggered")
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

func TestWorkflowExecuteResultsSubWorkflowSubActivity(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	defer database.SaveAsJSON("./jsons/workflows_basic_execute_results_sub-workflows_sub-activities.json")
	ctx := context.Background()

	orchestrator := tempolite.NewOrchestrator(ctx, database, registry)

	var flagSubTriggered atomic.Bool
	var flagTriggered atomic.Bool

	subactivty := func(ctx tempolite.ActivityContext) (int, error) {
		return 42, nil
	}

	subworkflowFunc := func(ctx tempolite.WorkflowContext) (int, error) {
		flagSubTriggered.Store(true)
		var life int
		if err := ctx.Activity("sub", subactivty, nil).Get(&life); err != nil {
			return -1, err
		}
		return life, nil
	}

	workflowFunc := func(ctx tempolite.WorkflowContext) (int, error) {
		flagTriggered.Store(true)
		var life int
		if err := ctx.Workflow("sub", subworkflowFunc, nil).Get(&life); err != nil {
			return -1, err
		}
		return life, nil
	}

	future := orchestrator.Execute(workflowFunc, nil)
	var life int
	if err := future.Get(&life); err != nil {
		t.Fatal(err)
	}

	if future.WorkflowID() == 0 {
		t.Fatal("workflow ID should not be 0")
	}

	if future.WorkflowExecutionID() == 0 {
		t.Fatal("workflow execution ID should not be 0")
	}

	if life != 42 {
		t.Fatalf("expected life %d, got %d", 42, life)
	}

	if !flagTriggered.Load() {
		t.Fatal("workflowFunc was not triggered")
	}

	if !flagSubTriggered.Load() {
		t.Fatal("subworkflowFunc was not triggered")
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

func TestWorkflowExecuteResultsSagaContext(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	defer database.SaveAsJSON("./jsons/workflows_basic_execute_results_sagas_.json")
	ctx := context.Background()

	orchestrator := tempolite.NewOrchestrator(ctx, database, registry)

	var flagTriggered atomic.Bool

	workflowFunc := func(ctx tempolite.WorkflowContext) (int, error) {
		flagTriggered.Store(true)

		def, err := tempolite.NewSaga().
			Add(
				func(tc tempolite.TransactionContext) error {
					var life int = 42
					if err := tc.Store("life", &life); err != nil {
						return err
					}
					return nil
				},
				func(cc tempolite.CompensationContext) error {
					return nil
				},
			).
			Add(
				func(tc tempolite.TransactionContext) error {
					var life int
					if err := tc.Load("life", &life); err != nil {
						return err
					}
					return nil
				},
				func(cc tempolite.CompensationContext) error {
					return nil
				},
			).
			Build()

		if err != nil {
			return -1, err
		}

		if err := ctx.Saga("saga", def).Get(); err != nil {
			return -1, err
		}

		return 42, nil
	}

	future := orchestrator.Execute(workflowFunc, nil)
	var result int
	if err := future.Get(&result); err != nil {
		t.Fatal(err)
	}

	if future.WorkflowID() == 0 {
		t.Fatal("workflow ID should not be 0")
	}

	if future.WorkflowExecutionID() == 0 {
		t.Fatal("workflow execution ID should not be 0")
	}

	if result != 42 {
		t.Fatalf("expected result %d, got %d", 42, result)
	}

	if !flagTriggered.Load() {
		t.Fatal("workflowFunc was not triggered")
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

	sagaEntities, err := database.GetSagaEntities(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if len(sagaEntities) != 1 {
		t.Fatalf("expected 1 saga entity, got %d", len(sagaEntities))
	}

	sagaEntity := sagaEntities[0]
	if sagaEntity.Status != tempolite.StatusCompleted {
		t.Fatalf("expected saga status %s, got %s", tempolite.StatusCompleted, sagaEntity.Status)
	}

	var lifeBack int
	val, err := database.GetSagaValueByKey(sagaEntity.ID, "life")
	if err != nil {
		t.Fatal(err)
	}
	if len(val) == 0 {
		t.Fatal("expected saga value 'life' to exist")
	}

	if err := tempolite.ConvertBack(val, &lifeBack); err != nil {
		t.Fatal(err)
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

func TestWorkflowExecuteResultsSignals(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	defer database.SaveAsJSON("./jsons/workflows_basic_execute_results_signals.json")
	ctx := context.Background()

	var snmap sync.Map

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

	workflowFunc := func(ctx tempolite.WorkflowContext) (int, error) {
		flagTriggered.Store(true)
		var value string

		if err := ctx.Signal("hello", &value); err != nil {
			return -1, err
		}

		return 42, nil
	}

	future := orchestrator.Execute(workflowFunc, nil)

	go func() {
		<-time.After(time.Second)
		signalFuture, ok := snmap.Load("hello")
		if !ok {
			t.Fatal("signal not found")
		}
		signalFuture.(tempolite.Future).SetResult([]interface{}{"world"})
	}()

	var result int
	if err := future.Get(&result); err != nil {
		t.Fatal(err)
	}

	if future.WorkflowID() == 0 {
		t.Fatal("workflow ID should not be 0")
	}

	if future.WorkflowExecutionID() == 0 {
		t.Fatal("workflow execution ID should not be 0")
	}

	if result != 42 {
		t.Fatalf("expected result %d, got %d", 42, result)
	}

	if !flagTriggered.Load() {
		t.Fatal("workflowFunc was not triggered")
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

	signals, err := database.GetSignalEntities(future.WorkflowID(), tempolite.SignalEntityWithData())
	if err != nil {
		t.Fatal(err)
	}

	var msgBack string
	for _, v := range signals {
		if v.SignalData.Name == "hello" {
			execs, err := database.GetSignalExecutions(v.ID)
			if err != nil {
				t.Fatal(err)
			}

			if len(execs) != 1 {
				t.Fatalf("expected 1 execution, got %d", len(execs))
			}

			sort.Slice(execs, func(i, j int) bool {
				return execs[i].ID < execs[j].ID
			})

			data, err := database.GetSignalExecutionDataByExecutionID(execs[0].ID)
			if err != nil {
				t.Fatal(err)
			}

			if err := tempolite.ConvertBack(data.Value, &msgBack); err != nil {
				t.Fatal(err)
			}
		}
	}

	if msgBack != "world" {
		t.Fatalf("expected message %s, got %s", "world", msgBack)
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
