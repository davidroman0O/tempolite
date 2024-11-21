package tempolite

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestOrchestratorWorkflow(t *testing.T) {

	database := NewDefaultDatabase()
	register := NewRegistry()
	ctx := context.Background()

	var executed uint32 // Use atomic variable

	workflow := func(ctx WorkflowContext) error {
		atomic.StoreUint32(&executed, 1) // Atomically set executed to true
		return nil
	}

	register.RegisterWorkflow(workflow)

	o := NewOrchestrator(ctx, database, register)

	future := o.Execute(workflow, nil)

	t.Log("Waiting for workflow to complete")
	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

	if atomic.LoadUint32(&executed) == 0 {
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

	var executed uint32 // Use atomic variable

	subWorkflow := func(ctx WorkflowContext) error {
		atomic.StoreUint32(&executed, 1) // Atomically set executed to true
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

	if atomic.LoadUint32(&executed) == 0 {
		t.Fatal("Workflow was not executed")
	}

	t.Log("Orchestrator workflow completed successfully")
	o.Wait()
}

func TestOrchestratorWorkflowActivity(t *testing.T) {

	database := NewDefaultDatabase()
	register := NewRegistry()
	ctx := context.Background()

	var executed uint32 // Use atomic variable

	activity := func(ctx ActivityContext) error {
		atomic.StoreUint32(&executed, 1) // Atomically set executed to true
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

	if atomic.LoadUint32(&executed) == 0 {
		t.Fatal("Workflow was not executed")
	}

	t.Log("Orchestrator workflow completed successfully")
	o.Wait()
}

type sagaA struct {
	executed           uint32 // Use atomic variable
	onCompensationFunc func()
}

func (a *sagaA) Transaction(ctx TransactionContext) (interface{}, error) {
	atomic.StoreUint32(&a.executed, 1) // Atomically set executed to true
	return nil, nil
}

func (a *sagaA) Compensation(ctx CompensationContext) (interface{}, error) {

	if a.onCompensationFunc != nil {
		a.onCompensationFunc()
	}
	return nil, nil
}

type sagaB struct {
	executed           uint32 // Use atomic variable
	onCompensationFunc func()
}

func (b *sagaB) Transaction(ctx TransactionContext) (interface{}, error) {
	atomic.StoreUint32(&b.executed, 1) // Atomically set executed to true
	return nil, nil
}

func (b *sagaB) Compensation(ctx CompensationContext) (interface{}, error) {
	if b.onCompensationFunc != nil {
		b.onCompensationFunc()
	}
	return nil, nil
}

type sagaC struct {
	executed           uint32 // Use atomic variable
	onCompensationFunc func()
	failTransaction    bool
}

func (c *sagaC) Transaction(ctx TransactionContext) (interface{}, error) {
	atomic.StoreUint32(&c.executed, 1) // Atomically set executed to true
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
	if atomic.LoadUint32(&sagaStepA.executed) == 0 {
		t.Fatal("Transaction for sagaA was not executed")
	}
	if atomic.LoadUint32(&sagaStepB.executed) == 0 {
		t.Fatal("Transaction for sagaB was not executed")
	}
	if atomic.LoadUint32(&sagaStepC.executed) == 0 {
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
	var compensationOrderMu sync.Mutex // Mutex to protect compensationOrder

	// Create instances of saga steps with compensation functions
	sagaStepA := &sagaA{
		onCompensationFunc: func() {
			compensationOrderMu.Lock()
			defer compensationOrderMu.Unlock()
			compensationOrder = append(compensationOrder, "A")
		},
	}
	sagaStepB := &sagaB{
		onCompensationFunc: func() {
			compensationOrderMu.Lock()
			defer compensationOrderMu.Unlock()
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

	// Lock the mutex before accessing compensationOrder
	compensationOrderMu.Lock()
	orderCopy := make([]string, len(compensationOrder))
	copy(orderCopy, compensationOrder)
	compensationOrderMu.Unlock()

	// Check that compensations were called in the correct order
	if len(orderCopy) != 2 {
		t.Fatalf("Expected 2 compensations, got %d", len(orderCopy))
	}
	if orderCopy[0] != "B" || orderCopy[1] != "A" {
		t.Fatalf("Compensations were not called in correct order: %v", orderCopy)
	}

	t.Log("Orchestrator saga compensation test completed successfully")
	o.Wait()
}

func TestOrchestratorWorkflowSideEffectZeroed(t *testing.T) {

	database := NewDefaultDatabase()
	register := NewRegistry()
	ctx := context.Background()

	var executed uint32 // Use atomic variable
	var a int = -1

	workflow := func(ctx WorkflowContext) error {

		if err := ctx.SideEffect("switch-a", func() int {
			atomic.StoreUint32(&executed, 1) // Atomically set executed to true
			return 0
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

	if atomic.LoadUint32(&executed) == 0 {
		t.Fatal("Workflow was not executed")
	}

	if a < -1 {
		t.Fatal("Side effect did not return a value")
	}

	t.Log("Orchestrator workflow completed successfully")
	o.Wait()
}

func TestOrchestratorWorkflowSideEffect(t *testing.T) {

	database := NewDefaultDatabase()
	register := NewRegistry()
	ctx := context.Background()

	var executed uint32 // Use atomic variable
	var a int

	workflow := func(ctx WorkflowContext) error {

		if err := ctx.SideEffect("switch-a", func() int {
			atomic.StoreUint32(&executed, 1) // Atomically set executed to true
			return rand.Intn(99) + 1         // Returns a number between 1 and 99
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

	if atomic.LoadUint32(&executed) == 0 {
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

	var executed int32 // Use atomic variable

	workflow := func(ctx WorkflowContext) error {
		atomic.AddInt32(&executed, 1) // Atomically increment executed
		if atomic.LoadInt32(&executed) < 3 {
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

	if atomic.LoadInt32(&executed) != 4 {
		t.Fatalf("Expected 4 executions, got %d", atomic.LoadInt32(&executed))
	}

	t.Log("Orchestrator workflow completed successfully")
	o.Wait()
}

func TestOrchestratorWorkflowRetryPolicy(t *testing.T) {

	database := NewDefaultDatabase()
	register := NewRegistry()
	ctx := context.Background()

	var executed int32 // Use atomic variable

	workflow := func(ctx WorkflowContext) error {
		atomic.AddInt32(&executed, 1) // Atomically increment executed
		if atomic.LoadInt32(&executed) < 3 {
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

	if atomic.LoadInt32(&executed) != 3 {
		t.Fatalf("Expected 3 executions, got %d", atomic.LoadInt32(&executed))
	}

	t.Log("Orchestrator workflow completed successfully")
	o.Wait()
}

func TestOrchestratorWorkflowRetryPolicyFailure(t *testing.T) {

	database := NewDefaultDatabase()
	register := NewRegistry()
	ctx := context.Background()

	var executed int32 // Use atomic variable

	workflow := func(ctx WorkflowContext) error {
		atomic.AddInt32(&executed, 1) // Atomically increment executed
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

	if atomic.LoadInt32(&executed) != 3 {
		t.Fatalf("Expected 3 executions, got %d", atomic.LoadInt32(&executed))
	}

	t.Log("Orchestrator workflow completed successfully")
	o.Wait()
}

func TestOrchestratorWorkflowContinueAsNew(t *testing.T) {

	database := NewDefaultDatabase()
	register := NewRegistry()
	ctx := context.Background()

	var executed int32 // Use atomic variable

	workflow := func(ctx WorkflowContext) error {
		count := atomic.AddInt32(&executed, 1)
		if count < 3 {
			fmt.Println("Continue as new", count)
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

	if err := o.WaitForContinuations(future.WorkflowID()); err != nil {
		t.Fatal(err)
	}

	if err := o.Wait(); err != nil {
		t.Fatal(err)
	}

	finalCount := atomic.LoadInt32(&executed)
	if finalCount != 3 {
		t.Fatalf("Expected 3 executions, got %d", finalCount)
	}

	t.Log("Orchestrator workflow completed successfully")
}

func TestOrchestratorWorkflowPauseResume(t *testing.T) {

	database := NewDefaultDatabase()
	register := NewRegistry()
	ctx := context.Background()

	var executed int32 // Use atomic variable

	subWorkflow := func(ctx WorkflowContext) error {
		time.Sleep(1 * time.Second)
		return nil
	}

	workflow := func(ctx WorkflowContext) error {
		t.Log("Executing workflows")
		atomic.AddInt32(&executed, 1) // Atomically increment executed
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
		time.Sleep(2 * time.Second)
		o.Pause()
		fmt.Println("Pausing orchestrator, 2s")
		time.Sleep(2 * time.Second)
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

	time.Sleep(2 * time.Second)

	if err := o.Wait(); err != nil {
		t.Fatal(err)
	}

	if atomic.LoadInt32(&executed) != 2 {
		t.Fatalf("Expected 2 executions, got %d", atomic.LoadInt32(&executed))
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
