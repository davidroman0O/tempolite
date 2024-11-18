package tempolite

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestOrchestratorWorkflow(t *testing.T) {

	database := NewDefaultDatabase()
	register := NewRegistry()
	ctx := context.Background()

	executed := false

	workflow := func(ctx WorkflowContext) error {
		executed = true
		return nil
	}

	register.RegisterWorkflow(workflow)

	o := NewOrchestrator(ctx, database, register)

	future := o.Execute(workflow, nil)

	t.Log("Waiting for workflow to complete")
	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

	if !executed {
		t.Fatal("Workflow was not executed")
	}

	t.Log("Orchestrator workflow completed successfully")
	o.Wait()

	info, err := o.GetWorkflow(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if info.Status != StatusCompleted {
		t.Fatalf("Expected workflow to be completed, got %s", info.Status)
	}
}

func TestOrchestratorSubWorkflow(t *testing.T) {

	database := NewDefaultDatabase()
	register := NewRegistry()
	ctx := context.Background()

	executed := false

	subWorkflow := func(ctx WorkflowContext) error {
		executed = true
		return nil
	}

	workflow := func(ctx WorkflowContext) error {
		if err := ctx.Workflow("sub", subWorkflow, nil).Get(); err != nil {
			return err
		}
		return nil
	}

	register.RegisterWorkflow(workflow)

	o := NewOrchestrator(ctx, database, register)

	future := o.Execute(workflow, nil)

	t.Log("Waiting for workflow to complete")
	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

	if !executed {
		t.Fatal("Workflow was not executed")
	}

	t.Log("Orchestrator workflow completed successfully")
	o.Wait()
}

func TestOrchestratorWorkflowActivity(t *testing.T) {

	database := NewDefaultDatabase()
	register := NewRegistry()
	ctx := context.Background()

	executed := false

	activity := func(ctx ActivityContext) error {
		executed = true
		return nil
	}

	workflow := func(ctx WorkflowContext) error {
		if err := ctx.Activity("sub", activity, nil).Get(); err != nil {
			return err
		}
		return nil
	}

	register.RegisterWorkflow(workflow)

	o := NewOrchestrator(ctx, database, register)

	future := o.Execute(workflow, nil)

	t.Log("Waiting for workflow to complete")
	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

	if !executed {
		t.Fatal("Workflow was not executed")
	}

	t.Log("Orchestrator workflow completed successfully")
	o.Wait()
}

type sagaA struct {
	executed           bool
	onCompensationFunc func()
}

func (a *sagaA) Transaction(ctx TransactionContext) (interface{}, error) {
	a.executed = true

	return nil, nil
}

func (a *sagaA) Compensation(ctx CompensationContext) (interface{}, error) {

	if a.onCompensationFunc != nil {
		a.onCompensationFunc()
	}
	return nil, nil
}

type sagaB struct {
	executed           bool
	onCompensationFunc func()
}

func (b *sagaB) Transaction(ctx TransactionContext) (interface{}, error) {
	b.executed = true
	return nil, nil
}

func (b *sagaB) Compensation(ctx CompensationContext) (interface{}, error) {
	if b.onCompensationFunc != nil {
		b.onCompensationFunc()
	}
	return nil, nil
}

type sagaC struct {
	executed           bool
	onCompensationFunc func()
	failTransaction    bool
}

func (c *sagaC) Transaction(ctx TransactionContext) (interface{}, error) {
	c.executed = true
	if c.failTransaction {
		return nil, errors.New("intentional failure")
	}
	return nil, nil
}

func (c *sagaC) Compensation(ctx CompensationContext) (interface{}, error) {
	if c.onCompensationFunc != nil {
		c.onCompensationFunc()
	}
	return nil, nil
}

func TestOrchestratorWorkflowSaga(t *testing.T) {

	database := NewDefaultDatabase()
	register := NewRegistry()
	ctx := context.Background()

	// Create instances of saga steps
	sagaStepA := &sagaA{}
	sagaStepB := &sagaB{}
	sagaStepC := &sagaC{}

	workflow := func(ctx WorkflowContext) error {

		def, err := NewSaga().
			AddStep(sagaStepA).
			AddStep(sagaStepB).
			AddStep(sagaStepC).
			Build()
		if err != nil {
			return err
		}

		if err := ctx.Saga("saga-sub", def).Get(); err != nil {
			return err
		}

		return nil
	}

	register.RegisterWorkflow(workflow)

	o := NewOrchestrator(ctx, database, register)

	future := o.Execute(workflow, nil)

	t.Log("Waiting for workflow to complete")
	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

	// Check if all transactions have been executed
	if !sagaStepA.executed {
		t.Fatal("Transaction for sagaA was not executed")
	}
	if !sagaStepB.executed {
		t.Fatal("Transaction for sagaB was not executed")
	}
	if !sagaStepC.executed {
		t.Fatal("Transaction for sagaC was not executed")
	}

	t.Log("Orchestrator workflow completed successfully")
	o.Wait()
}

func TestOrchestratorSagaCompensation(t *testing.T) {

	database := NewDefaultDatabase()
	register := NewRegistry()
	ctx := context.Background()

	// Slice to track compensation order
	var compensationOrder []string

	// Create instances of saga steps with compensation functions
	sagaStepA := &sagaA{
		onCompensationFunc: func() {
			compensationOrder = append(compensationOrder, "A")
		},
	}
	sagaStepB := &sagaB{
		onCompensationFunc: func() {
			compensationOrder = append(compensationOrder, "B")
		},
	}
	sagaStepC := &sagaC{
		failTransaction: true,
	}

	workflow := func(ctx WorkflowContext) error {

		def, err := NewSaga().
			AddStep(sagaStepA).
			AddStep(sagaStepB).
			AddStep(sagaStepC).
			Build()
		if err != nil {
			return err
		}

		if err := ctx.Saga("saga-sub", def).Get(); err != nil {
			return err
		}

		return nil
	}

	register.RegisterWorkflow(workflow)

	o := NewOrchestrator(ctx, database, register)

	future := o.Execute(workflow, nil)

	t.Log("Waiting for workflow to complete")
	err := future.Get()
	if err == nil {
		t.Fatal("Expected error but got nil")
	}

	fmt.Println("Saga failure", err)

	// Check that compensations were called in the correct order
	if len(compensationOrder) != 2 {
		t.Fatalf("Expected 2 compensations, got %d", len(compensationOrder))
	}
	if compensationOrder[0] != "B" || compensationOrder[1] != "A" {
		t.Fatalf("Compensations were not called in correct order: %v", compensationOrder)
	}

	t.Log("Orchestrator saga compensation test completed successfully")
	o.Wait()
}

func TestOrchestratorWorkflowSideEffect(t *testing.T) {

	database := NewDefaultDatabase()
	register := NewRegistry()
	ctx := context.Background()

	executed := false
	var a int

	workflow := func(ctx WorkflowContext) error {

		if err := ctx.SideEffect("switch-a", func() int {
			executed = true
			return rand.Intn(100)
		}).Get(&a); err != nil {
			return err
		}

		return nil
	}

	register.RegisterWorkflow(workflow)

	o := NewOrchestrator(ctx, database, register)

	future := o.Execute(workflow, nil)

	t.Log("Waiting for workflow to complete")
	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

	if !executed {
		t.Fatal("Workflow was not executed")
	}

	if a == 0 {
		t.Fatal("Side effect did not return a value")
	}

	t.Log("Orchestrator workflow completed successfully")
	o.Wait()
}

func TestOrchestratorRetryFailureWorkflow(t *testing.T) {

	database := NewDefaultDatabase()
	register := NewRegistry()
	ctx := context.Background()

	executed := 0

	workflow := func(ctx WorkflowContext) error {
		executed++
		if executed < 3 {
			return errors.New("intentional failure")
		}
		return nil
	}

	register.RegisterWorkflow(workflow)

	o := NewOrchestrator(ctx, database, register)

	future := o.Execute(workflow, nil)

	t.Log("Waiting for workflow to complete")
	shouldHaveFailed := false
	if err := future.Get(); err != nil {
		shouldHaveFailed = true
	}

	if !shouldHaveFailed {
		t.Fatal("Expected workflow to fail")
	}

	shouldHaveFailed = false

	future = o.Retry(future.WorkflowID())
	if err := future.Get(); err != nil {
		shouldHaveFailed = true
	}

	if !shouldHaveFailed {
		t.Fatal("Expected workflow to fail")
	}

	future = o.Retry(future.WorkflowID())
	if err := future.Get(); err != nil {
		shouldHaveFailed = true
	}

	if !shouldHaveFailed {
		t.Fatal("Expected workflow to fail")
	}
	shouldHaveFailed = false

	future = o.Retry(future.WorkflowID())
	if err := future.Get(); err != nil {
		shouldHaveFailed = false
	}

	if shouldHaveFailed {
		t.Fatal("Expected workflow to succeed")
	}

	if executed != 4 {
		t.Fatalf("Expected 4 executions, got %d", executed)
	}

	t.Log("Orchestrator workflow completed successfully")
	o.Wait()
}

func TestOrchestratorWorkflowRetryPolicy(t *testing.T) {

	database := NewDefaultDatabase()
	register := NewRegistry()
	ctx := context.Background()

	executed := 0

	workflow := func(ctx WorkflowContext) error {
		executed++
		if executed < 3 {
			return errors.New("intentional failure")
		}
		return nil
	}

	register.RegisterWorkflow(workflow)

	o := NewOrchestrator(ctx, database, register)

	future := o.Execute(workflow, &WorkflowOptions{
		RetryPolicy: &RetryPolicy{
			MaxAttempts: 4,
		},
	})

	t.Log("Waiting for workflow to complete")
	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

	if executed != 3 {
		t.Fatalf("Expected 3 executions, got %d", executed)
	}

	t.Log("Orchestrator workflow completed successfully")
	o.Wait()
}

func TestOrchestratorWorkflowRetryPolicyFailure(t *testing.T) {

	database := NewDefaultDatabase()
	register := NewRegistry()
	ctx := context.Background()

	executed := 0

	workflow := func(ctx WorkflowContext) error {
		executed++
		return errors.New("intentional failure")
	}

	register.RegisterWorkflow(workflow)

	o := NewOrchestrator(ctx, database, register)

	future := o.Execute(workflow, &WorkflowOptions{
		RetryPolicy: &RetryPolicy{
			MaxAttempts:     3,
			InitialInterval: 1 * time.Second,
		},
	})

	t.Log("Waiting for workflow to complete")
	if err := future.Get(); err == nil {
		t.Fatalf("Expected error, got nil")
	}

	if executed != 3 {
		t.Fatalf("Expected 3 executions, got %d", executed)
	}

	t.Log("Orchestrator workflow completed successfully")
	o.Wait()
}

func TestOrchestratorWorkflowContinueAsNew(t *testing.T) {

	database := NewDefaultDatabase()
	register := NewRegistry()
	ctx := context.Background()

	executed := 0

	workflow := func(ctx WorkflowContext) error {
		executed++
		if executed < 3 {
			fmt.Println("Continue as new", executed)
			return ctx.ContinueAsNew(nil)
		}
		return nil
	}

	register.RegisterWorkflow(workflow)

	o := NewOrchestrator(ctx, database, register)

	future := o.Execute(workflow, nil)

	t.Log("Waiting for workflow to complete")
	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

	o.Wait() // you have to wait before checking the executed count (duuuh)

	if executed != 3 {
		t.Fatalf("Expected 3 executions, got %d", executed)
	}

	t.Log("Orchestrator workflow completed successfully")
}

func TestOrchestratorWorkflowPauseResume(t *testing.T) {

	database := NewDefaultDatabase()
	register := NewRegistry()
	ctx := context.Background()

	executed := 0

	subWorkflow := func(ctx WorkflowContext) error {
		<-time.After(1 * time.Second)
		return nil
	}

	workflow := func(ctx WorkflowContext) error {
		executed++
		if err := ctx.Workflow("sub1", subWorkflow, nil).Get(); err != nil {
			return err
		}
		if err := ctx.Workflow("sub2", subWorkflow, nil).Get(); err != nil {
			return err
		}
		if err := ctx.Workflow("sub3", subWorkflow, nil).Get(); err != nil {
			return err
		}
		return nil
	}

	register.RegisterWorkflow(workflow)

	o := NewOrchestrator(ctx, database, register)

	future := o.Execute(workflow, nil)

	go func() {
		<-time.After(2 * time.Second)
		o.Pause()
		fmt.Println("Pausing orchestrator, 2s")
		<-time.After(2 * time.Second)
		fmt.Println("Resuming orchestrator")
		o.Resume(future.WorkflowID())
	}()

	t.Log("Waiting for workflow to complete")
	if err := future.Get(); err != nil {
		if errors.Is(err, ErrPaused) {
			fmt.Println("Workflow paused")
		} else {
			t.Fatal(err)
		}
	}
	t.Log("Workflow future came back after pause")

	<-time.After(2 * time.Second)

	if err := o.Wait(); err != nil {
		t.Fatal(err)
	} // you have to wait before checking the executed count (duuuh)

	if executed != 2 {
		t.Fatalf("Expected 2 executions, got %d", executed)
	}

	t.Log("Orchestrator workflow completed successfully")

	info, err := o.GetWorkflow(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}
	if info.Status != StatusCompleted {
		t.Fatalf("Expected workflow to be completed, got %s", info.Status)
	}
	fmt.Println("Workflow info", info)
}
