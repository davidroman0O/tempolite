package tempolite

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestQueue(t *testing.T) {

	database := NewDefaultDatabase()
	register := NewRegistry()
	ctx := context.Background()

	qm := newQueueManager(ctx, "default", 1, register, database)

	if err := qm.Wait(); err != nil {
		t.Fatal(err)
	}
}

func TestQueueWorkflowDb(t *testing.T) {

	database := NewDefaultDatabase()
	register := NewRegistry()
	ctx := context.Background()

	workflow := func(ctx WorkflowContext) error {
		t.Log("Workflow called")
		return nil
	}

	register.RegisterWorkflow(workflow)

	qm := newQueueManager(ctx, "default", 1, register, database)

	id, err := qm.CreateWorkflow(workflow, nil)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := qm.ExecuteWorkflow(id).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}

	future := NewDatabaseFuture(ctx, database)
	future.setEntityID(id)

	t.Log("Waiting for future")
	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

	t.Log("Waiting for queue manager")
	if err := qm.Wait(); err != nil {
		t.Fatal(err)
	}
}

func TestQueueWorkflowDbActivity(t *testing.T) {

	database := NewDefaultDatabase()
	register := NewRegistry()
	ctx := context.Background()

	activity := func(ctx ActivityContext) error {
		t.Log("Activity called")
		return nil
	}

	workflow := func(ctx WorkflowContext) error {
		t.Log("Workflow called")
		if err := ctx.Activity("sub", activity, nil).Get(); err != nil {
			return err
		}
		return nil
	}

	register.RegisterWorkflow(workflow)

	qm := newQueueManager(ctx, "default", 1, register, database)

	id, err := qm.CreateWorkflow(workflow, nil)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := qm.ExecuteWorkflow(id).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}

	future := NewDatabaseFuture(ctx, database)
	future.setEntityID(id)

	t.Log("Waiting for future")
	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

	t.Log("Waiting for queue manager")
	if err := qm.Wait(); err != nil {
		t.Fatal(err)
	}
}

func TestQueueWorkflowDbPauseResume(t *testing.T) {

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

	qm := newQueueManager(ctx, "default", 1, register, database)

	time.Sleep(2 * time.Second)

	id, err := qm.CreateWorkflow(workflow, nil)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := qm.ExecuteWorkflow(id).Wait(context.Background()); err != nil {
		t.Fatal(err)
	}

	future := NewDatabaseFuture(ctx, database)
	future.setEntityID(id)

	var goroutineError error
	var goroutineErrorMu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		time.Sleep(1 * time.Second)
		if _, err := qm.Pause(id).Wait(context.Background()); err != nil {
			goroutineErrorMu.Lock()
			goroutineError = err
			goroutineErrorMu.Unlock()
			return
		}
		fmt.Println("Pausing orchestrator, 2s")
		time.Sleep(2 * time.Second)
		fmt.Println("Resuming orchestrator")
		if _, err := qm.Resume(id).Wait(context.Background()); err != nil {
			goroutineErrorMu.Lock()
			goroutineError = err
			goroutineErrorMu.Unlock()
			return
		}
	}()

	t.Log("Waiting for future")
	if err := future.Get(); err != nil {
		if !errors.Is(err, ErrPaused) {
			t.Fatal(err)
		}
	}

	time.Sleep(2 * time.Second)

	t.Log("Waiting for future post wait")
	if err := future.Get(); err != nil {
		fmt.Println("post wait", err)
		t.Fatal(err)
	}

	t.Log("Waiting for queue manager")
	if err := qm.Wait(); err != nil {
		t.Fatal(err)
	}

	// Wait for the goroutine to finish
	wg.Wait()

	// Check for errors from goroutine
	goroutineErrorMu.Lock()
	if goroutineError != nil {
		t.Fatal(goroutineError)
	}
	goroutineErrorMu.Unlock()

	if atomic.LoadInt32(&executed) != 2 {
		t.Fatalf("Executed not 2 but %d", atomic.LoadInt32(&executed))
	}
}

func TestQueueWorkflowDbRestart(t *testing.T) {

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

	qm := newQueueManager(ctx, "default", 1, register, database)

	id, err := qm.CreateWorkflow(workflow, nil)
	if err != nil {
		t.Fatal(err)
	}

	execCtx := context.Background()
	// Set timeout less than the total execution time of the workflow
	execCtx, cancel := context.WithTimeout(execCtx, 500*time.Millisecond)
	defer cancel()
	if _, err := qm.ExecuteWorkflow(id).Wait(execCtx); err != nil {
		t.Logf("Expected error on initial execution: %v", err)
	}

	if err := qm.Wait(); err != nil {
		t.Fatal(err)
	}

	if err := qm.Close(); err != nil {
		t.Fatal(err)
	}

	t.Log("Restarting queue manager")

	ctx = context.Background()
	qm = newQueueManager(ctx, "default", 1, register, database)

	// Now, when we resume, the RetryState.Attempts should be reset
	execCtx = context.Background()
	execCtx, cancel = context.WithTimeout(execCtx, time.Second*5)
	defer cancel()
	if _, err := qm.Resume(id).Wait(execCtx); err != nil {
		t.Fatal(err)
	}

	future := NewDatabaseFuture(ctx, database)
	future.setEntityID(id)

	t.Log("Waiting for future")
	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

	if err := qm.Wait(); err != nil {
		t.Fatal(err)
	}

	if atomic.LoadInt32(&executed) != 2 {
		fmt.Printf("Executed not 2 but %d\n", atomic.LoadInt32(&executed))
		t.Fatalf("Executed not 2 but %d", atomic.LoadInt32(&executed))
	}
}
