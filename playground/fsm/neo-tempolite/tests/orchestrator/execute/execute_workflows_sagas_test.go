package tests

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync/atomic"
	"testing"

	tempolite "github.com/davidroman0O/tempolite/playground/fsm/neo-tempolite"
)

func TestWorkflowExecuteSagas(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	defer database.SaveAsJSON("./jsons/workflows_basic_saga_execute_structs.json")
	ctx := context.Background()

	orchestrator := tempolite.NewOrchestrator(ctx, database, registry)

	var flagTriggered atomic.Bool
	var transaction1Triggered atomic.Bool
	var transaction2Triggered atomic.Bool
	var transaction3Triggered atomic.Bool
	var compensation1Triggered atomic.Bool
	var compensation2Triggered atomic.Bool
	var compensation3Triggered atomic.Bool

	saga := tempolite.NewSaga().
		AddStep(
			NewTestSaga(
				func() error {
					transaction1Triggered.Store(true)
					return nil
				},
				func() error {
					compensation1Triggered.Store(true)
					return nil
				},
				false,
				false,
			),
		).
		Add(
			func(ctx tempolite.TransactionContext) error {
				transaction2Triggered.Store(true)
				return nil
			},
			func(ctx tempolite.CompensationContext) error {
				compensation2Triggered.Store(true)
				return nil
			},
		).
		AddStep(
			NewTestSaga(
				func() error {
					transaction3Triggered.Store(true)
					return nil
				},
				func() error {
					compensation3Triggered.Store(true)
					return nil
				},
				false,
				false,
			),
		)

	def, err := saga.Build()
	if err != nil {
		t.Fatal(err)
	}

	workflowFunc := func(ctx tempolite.WorkflowContext) error {
		flagTriggered.Store(true)
		return ctx.Saga("sub", def).Get()
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

	// Verify workflow entity
	we, err := database.GetWorkflowEntity(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if we.Status != tempolite.StatusCompleted {
		t.Fatalf("expected status %s, got %s", tempolite.StatusCompleted, we.Status)
	}

	// Get saga entity
	sagas, err := database.GetSagaEntities(we.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(sagas) != 1 {
		t.Fatalf("expected 1 saga, got %d", len(sagas))
	}

	sort.Slice(sagas, func(i, j int) bool {
		return sagas[i].ID < sagas[j].ID
	})

	sagaEntity := sagas[0]
	if sagaEntity.Status != tempolite.StatusCompleted {
		t.Fatalf("expected saga status %s, got %s",
			tempolite.StatusCompleted, sagaEntity.Status)
	}

	// Verify saga executions
	sagaExecs, err := database.GetSagaExecutions(sagaEntity.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(sagaExecs) != 3 {
		t.Fatalf("expected 3 saga executions, got %d", len(sagaExecs))
	}

	sort.Slice(sagaExecs, func(i, j int) bool {
		return sagaExecs[i].ID < sagaExecs[j].ID
	})

	// Verify each transaction execution
	for i := 0; i < 3; i++ {
		exec := sagaExecs[i]
		if exec.Status != tempolite.ExecutionStatusCompleted {
			t.Fatalf("expected saga execution %d status %s, got %s",
				i, tempolite.ExecutionStatusCompleted, exec.Status)
		}
		if exec.ExecutionType != tempolite.ExecutionTypeTransaction {
			t.Fatalf("expected execution type %s, got %s",
				tempolite.ExecutionTypeTransaction, exec.ExecutionType)
		}
		if exec.Error != "" {
			t.Fatalf("expected no error, got %s", exec.Error)
		}
	}

	// Verify workflow execution
	workflowExecs, err := database.GetWorkflowExecutions(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if len(workflowExecs) != 1 {
		t.Fatalf("expected 1 workflow execution, got %d", len(workflowExecs))
	}

	if workflowExecs[0].Status != tempolite.ExecutionStatusCompleted {
		t.Fatalf("expected workflow execution status %s, got %s",
			tempolite.ExecutionStatusCompleted, workflowExecs[0].Status)
	}

	if workflowExecs[0].Error != "" {
		t.Fatal("expected no workflow execution error")
	}

	// Verify flags
	if !flagTriggered.Load() {
		t.Fatal("workflowFunc was not triggered")
	}
	if !transaction1Triggered.Load() {
		t.Fatal("transaction1 was not triggered")
	}
	if !transaction2Triggered.Load() {
		t.Fatal("transaction2 was not triggered")
	}
	if !transaction3Triggered.Load() {
		t.Fatal("transaction3 was not triggered")
	}
	if compensation1Triggered.Load() {
		t.Fatal("compensation1 was triggered when it shouldn't")
	}
	if compensation2Triggered.Load() {
		t.Fatal("compensation2 was triggered when it shouldn't")
	}
	if compensation3Triggered.Load() {
		t.Fatal("compensation3 was triggered when it shouldn't")
	}
}

func TestWorkflowExecuteSagasFailure(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	defer database.SaveAsJSON("./jsons/workflows_basic_saga_execute_structs_failure.json")
	ctx := context.Background()

	orchestrator := tempolite.NewOrchestrator(ctx, database, registry)

	var flagTriggered atomic.Bool
	var transaction1Triggered atomic.Bool
	var transaction2Triggered atomic.Bool
	var transaction3Triggered atomic.Bool
	var compensation1Triggered atomic.Bool
	var compensation2Triggered atomic.Bool
	var compensation3Triggered atomic.Bool

	saga := tempolite.NewSaga().
		AddStep(
			NewTestSaga(
				func() error {
					transaction1Triggered.Store(true)
					return nil
				},
				func() error {
					compensation1Triggered.Store(true)
					return nil
				},
				false,
				false,
			),
		).
		Add(
			func(ctx tempolite.TransactionContext) error {
				transaction2Triggered.Store(true)
				return fmt.Errorf("on purpose")
			},
			func(ctx tempolite.CompensationContext) error {
				compensation2Triggered.Store(true)
				return nil
			},
		).
		AddStep(
			NewTestSaga(
				func() error {
					transaction3Triggered.Store(true)
					return nil
				},
				func() error {
					compensation3Triggered.Store(true)
					return nil
				},
				false,
				false,
			),
		)

	def, err := saga.Build()
	if err != nil {
		t.Fatal(err)
	}

	workflowFunc := func(ctx tempolite.WorkflowContext) error {
		flagTriggered.Store(true)
		return ctx.Saga("sub", def).Get()
	}

	future := orchestrator.Execute(workflowFunc, nil)
	if err := future.Get(); err == nil {
		t.Fatal("expected error")
	}

	if future.WorkflowID() == 0 {
		t.Fatal("workflow ID should not be 0")
	}

	if future.WorkflowExecutionID() == 0 {
		t.Fatal("workflow execution ID should not be 0")
	}

	// Verify workflow entity
	we, err := database.GetWorkflowEntity(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if we.Status != tempolite.StatusFailed {
		t.Fatalf("expected status %s, got %s", tempolite.StatusFailed, we.Status)
	}

	// Get saga entity
	sagas, err := database.GetSagaEntities(we.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(sagas) != 1 {
		t.Fatalf("expected 1 saga, got %d", len(sagas))
	}

	sort.Slice(sagas, func(i, j int) bool {
		return sagas[i].ID < sagas[j].ID
	})

	sagaEntity := sagas[0]
	if sagaEntity.Status != tempolite.StatusCompensated {
		t.Fatalf("expected saga status %s, got %s",
			tempolite.StatusCompensated, sagaEntity.Status)
	}

	// Verify saga executions
	sagaExecs, err := database.GetSagaExecutions(sagaEntity.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(sagaExecs) != 3 { // 2 transactions (1 success, 1 fail) + 1 compensation
		t.Fatalf("expected 3 saga executions, got %d", len(sagaExecs))
	}

	sort.Slice(sagaExecs, func(i, j int) bool {
		return sagaExecs[i].ID < sagaExecs[j].ID
	})

	// First transaction succeeds
	if sagaExecs[0].Status != tempolite.ExecutionStatusCompleted {
		t.Fatalf("expected first transaction status %s, got %s",
			tempolite.ExecutionStatusCompleted, sagaExecs[0].Status)
	}
	if sagaExecs[0].ExecutionType != tempolite.ExecutionTypeTransaction {
		t.Fatalf("expected execution type %s, got %s",
			tempolite.ExecutionTypeTransaction, sagaExecs[0].ExecutionType)
	}

	// Second transaction fails
	if sagaExecs[1].Status != tempolite.ExecutionStatusFailed {
		t.Fatalf("expected second transaction status %s, got %s",
			tempolite.ExecutionStatusFailed, sagaExecs[1].Status)
	}
	if sagaExecs[1].ExecutionType != tempolite.ExecutionTypeTransaction {
		t.Fatalf("expected execution type %s, got %s",
			tempolite.ExecutionTypeTransaction, sagaExecs[1].ExecutionType)
	}
	if !strings.Contains(sagaExecs[1].Error, "on purpose") {
		t.Fatalf("expected error to contain 'on purpose', got %s", sagaExecs[1].Error)
	}

	// Third execution is compensation for first transaction
	if sagaExecs[2].Status != tempolite.ExecutionStatusCompleted {
		t.Fatalf("expected compensation status %s, got %s",
			tempolite.ExecutionStatusCompleted, sagaExecs[2].Status)
	}
	if sagaExecs[2].ExecutionType != tempolite.ExecutionTypeCompensation {
		t.Fatalf("expected execution type %s, got %s",
			tempolite.ExecutionTypeCompensation, sagaExecs[2].ExecutionType)
	}
	// Verify compensation step index is for first transaction
	if sagaExecs[2].SagaExecutionData.StepIndex != 0 {
		t.Fatalf("expected compensation step index 0, got %d",
			sagaExecs[2].SagaExecutionData.StepIndex)
	}

	// Verify workflow execution
	workflowExecs, err := database.GetWorkflowExecutions(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if len(workflowExecs) != 1 {
		t.Fatalf("expected 1 workflow execution, got %d", len(workflowExecs))
	}

	if workflowExecs[0].Status != tempolite.ExecutionStatusFailed {
		t.Fatalf("expected workflow execution status %s, got %s",
			tempolite.ExecutionStatusFailed, workflowExecs[0].Status)
	}

	if !strings.Contains(workflowExecs[0].Error, "on purpose") {
		t.Fatalf("expected error to contain 'on purpose', got %s",
			workflowExecs[0].Error)
	}

	// Verify flags
	if !flagTriggered.Load() {
		t.Fatal("workflowFunc was not triggered")
	}
	if !transaction1Triggered.Load() {
		t.Fatal("transaction1 was not triggered")
	}
	if !transaction2Triggered.Load() {
		t.Fatal("transaction2 was not triggered")
	}
	if transaction3Triggered.Load() {
		t.Fatal("transaction3 was triggered when it shouldn't")
	}
	if !compensation1Triggered.Load() {
		t.Fatal("compensation1 was not triggered")
	}
	if compensation2Triggered.Load() {
		t.Fatal("compensation2 was triggered when it shouldn't")
	}
	if compensation3Triggered.Load() {
		t.Fatal("compensation3 was triggered when it shouldn't")
	}
}

func TestWorkflowExecuteSagasPanic(t *testing.T) {
	registry := tempolite.NewRegistry()
	database := tempolite.NewMemoryDatabase()
	defer database.SaveAsJSON("./jsons/workflows_basic_saga_execute_structs_panic.json")
	ctx := context.Background()

	orchestrator := tempolite.NewOrchestrator(ctx, database, registry)

	var flagTriggered atomic.Bool
	var transaction1Triggered atomic.Bool
	var transaction2Triggered atomic.Bool
	var transaction3Triggered atomic.Bool
	var compensation1Triggered atomic.Bool
	var compensation2Triggered atomic.Bool
	var compensation3Triggered atomic.Bool

	saga := tempolite.NewSaga().
		AddStep(
			NewTestSaga(
				func() error {
					transaction1Triggered.Store(true)
					return nil
				},
				func() error {
					compensation1Triggered.Store(true)
					return nil
				},
				false,
				false,
			),
		).
		Add(
			func(ctx tempolite.TransactionContext) error {
				transaction2Triggered.Store(true)
				panic("on purpose")
			},
			func(ctx tempolite.CompensationContext) error {
				compensation2Triggered.Store(true)
				return nil
			},
		).
		AddStep(
			NewTestSaga(
				func() error {
					transaction3Triggered.Store(true)
					return nil
				},
				func() error {
					compensation3Triggered.Store(true)
					return nil
				},
				false,
				false,
			),
		)

	def, err := saga.Build()
	if err != nil {
		t.Fatal(err)
	}

	workflowFunc := func(ctx tempolite.WorkflowContext) error {
		flagTriggered.Store(true)
		return ctx.Saga("sub", def).Get()
	}

	future := orchestrator.Execute(workflowFunc, nil)
	if err := future.Get(); err == nil {
		t.Fatal("expected error")
	}

	if future.WorkflowID() == 0 {
		t.Fatal("workflow ID should not be 0")
	}

	if future.WorkflowExecutionID() == 0 {
		t.Fatal("workflow execution ID should not be 0")
	}

	// Verify workflow entity
	we, err := database.GetWorkflowEntity(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if we.Status != tempolite.StatusFailed {
		t.Fatalf("expected status %s, got %s", tempolite.StatusFailed, we.Status)
	}

	// Get saga entity
	sagas, err := database.GetSagaEntities(we.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(sagas) != 1 {
		t.Fatalf("expected 1 saga, got %d", len(sagas))
	}

	sort.Slice(sagas, func(i, j int) bool {
		return sagas[i].ID < sagas[j].ID
	})

	sagaEntity := sagas[0]
	if sagaEntity.Status != tempolite.StatusCompensated {
		t.Fatalf("expected saga status %s, got %s",
			tempolite.StatusCompensated, sagaEntity.Status)
	}

	// Verify saga executions
	sagaExecs, err := database.GetSagaExecutions(sagaEntity.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(sagaExecs) != 3 { // 2 transactions (1 success, 1 panic) + 1 compensation
		t.Fatalf("expected 3 saga executions, got %d", len(sagaExecs))
	}

	sort.Slice(sagaExecs, func(i, j int) bool {
		return sagaExecs[i].ID < sagaExecs[j].ID
	})

	// First transaction succeeds
	if sagaExecs[0].Status != tempolite.ExecutionStatusCompleted {
		t.Fatalf("expected first transaction status %s, got %s",
			tempolite.ExecutionStatusCompleted, sagaExecs[0].Status)
	}
	if sagaExecs[0].ExecutionType != tempolite.ExecutionTypeTransaction {
		t.Fatalf("expected execution type %s, got %s",
			tempolite.ExecutionTypeTransaction, sagaExecs[0].ExecutionType)
	}

	// Second transaction panics
	if sagaExecs[1].Status != tempolite.ExecutionStatusFailed {
		t.Fatalf("expected second transaction status %s, got %s",
			tempolite.ExecutionStatusFailed, sagaExecs[1].Status)
	}
	if sagaExecs[1].ExecutionType != tempolite.ExecutionTypeTransaction {
		t.Fatalf("expected execution type %s, got %s",
			tempolite.ExecutionTypeTransaction, sagaExecs[1].ExecutionType)
	}
	if !strings.Contains(sagaExecs[1].Error, "on purpose") {
		t.Fatalf("expected error to contain 'on purpose', got %s", sagaExecs[1].Error)
	}
	if sagaExecs[1].StackTrace == nil {
		t.Fatal("expected stack trace to be present")
	}

	// Third execution is compensation for first transaction
	if sagaExecs[2].Status != tempolite.ExecutionStatusCompleted {
		t.Fatalf("expected compensation status %s, got %s",
			tempolite.ExecutionStatusCompleted, sagaExecs[2].Status)
	}
	if sagaExecs[2].ExecutionType != tempolite.ExecutionTypeCompensation {
		t.Fatalf("expected execution type %s, got %s",
			tempolite.ExecutionTypeCompensation, sagaExecs[2].ExecutionType)
	}
	// Verify compensation step index is for first transaction
	if sagaExecs[2].SagaExecutionData.StepIndex != 0 {
		t.Fatalf("expected compensation step index 0, got %d",
			sagaExecs[2].SagaExecutionData.StepIndex)
	}

	// Verify workflow execution
	workflowExecs, err := database.GetWorkflowExecutions(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if len(workflowExecs) != 1 {
		t.Fatalf("expected 1 workflow execution, got %d", len(workflowExecs))
	}

	if workflowExecs[0].Status != tempolite.ExecutionStatusFailed {
		t.Fatalf("expected workflow execution status %s, got %s",
			tempolite.ExecutionStatusFailed, workflowExecs[0].Status)
	}

	if !strings.Contains(workflowExecs[0].Error, "step panicked: on purpose") {
		t.Fatalf("expected error to contain 'step panicked: on purpose', got %s",
			workflowExecs[0].Error)
	}

	// Verify flags
	if !flagTriggered.Load() {
		t.Fatal("workflowFunc was not triggered")
	}
	if !transaction1Triggered.Load() {
		t.Fatal("transaction1 was not triggered")
	}
	if !transaction2Triggered.Load() {
		t.Fatal("transaction2 was not triggered")
	}
	if transaction3Triggered.Load() {
		t.Fatal("transaction3 was triggered when it shouldn't")
	}
	if !compensation1Triggered.Load() {
		t.Fatal("compensation1 was not triggered")
	}
	if compensation2Triggered.Load() {
		t.Fatal("compensation2 was triggered when it shouldn't")
	}
	if compensation3Triggered.Load() {
		t.Fatal("compensation3 was triggered when it shouldn't")
	}
}
