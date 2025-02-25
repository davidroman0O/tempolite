package tests

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"math/rand"

	tempolite "github.com/davidroman0O/tempolite"
	"github.com/k0kubun/pp/v3"
)

func TestTempoliteWorkflowsExecute(t *testing.T) {

	database := tempolite.NewMemoryDatabase()
	defer database.SaveAsJSON("./jsons/tempolite_workflows_execute.json")
	ctx := context.Background()

	tp, err := tempolite.New(
		ctx,
		database,
	)
	if err != nil {
		t.Fatal(err)
	}

	var flagTriggered atomic.Bool
	workflowFunc := func(ctx tempolite.WorkflowContext) error {
		flagTriggered.Store(true)
		return nil
	}

	future, err := tp.ExecuteDefault(workflowFunc, nil)
	if err != nil {
		t.Fatal(err)
	}

	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

	if !flagTriggered.Load() {
		t.Fatal("workflowFunc was not triggered")
	}

	if future.WorkflowID() == 0 {
		t.Fatal("workflow ID should not be 0")
	}

	if future.WorkflowExecutionID() == 0 {
		t.Fatal("workflow execution ID should not be 0")
	}

	// Get workflow entity directly using ID from future
	workflow, err := database.GetWorkflowEntity(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	// Get workflow execution directly using ID from future
	execution, err := database.GetWorkflowExecution(future.WorkflowExecutionID())
	if err != nil {
		t.Fatal(err)
	}

	// Get the run directly
	run, err := database.GetRun(workflow.RunID)
	if err != nil {
		t.Fatal(err)
	}

	// Get the queue directly
	queue, err := database.GetQueue(workflow.QueueID)
	if err != nil {
		t.Fatal(err)
	}

	// Validate Run
	if run.Status != tempolite.RunStatusCompleted {
		t.Errorf("expected run status %s, got %s", tempolite.RunStatusCompleted, run.Status)
	}

	// Validate Queue
	if queue.Name != tempolite.DefaultQueue {
		t.Errorf("expected default queue name %s, got %s", tempolite.DefaultQueue, queue.Name)
	}

	// Check workflow entity properties
	if workflow.Status != tempolite.StatusCompleted {
		t.Errorf("expected workflow status %s, got %s", tempolite.StatusCompleted, workflow.Status)
	}
	if workflow.Type != tempolite.EntityWorkflow {
		t.Errorf("expected entity type %s, got %s", tempolite.EntityWorkflow, workflow.Type)
	}
	if workflow.StepID != "root" {
		t.Errorf("expected step ID 'root', got %s", workflow.StepID)
	}
	if workflow.QueueID != queue.ID {
		t.Errorf("expected queue ID %d, got %d", queue.ID, workflow.QueueID)
	}
	if workflow.RunID != run.ID {
		t.Errorf("expected run ID %d, got %d", run.ID, workflow.RunID)
	}

	// Validate Workflow Data
	if workflow.WorkflowData == nil {
		t.Fatal("workflow data should not be nil")
	}
	if !workflow.WorkflowData.IsRoot {
		t.Error("workflow should be marked as root")
	}
	if workflow.WorkflowData.Paused {
		t.Error("workflow should not be paused")
	}
	if workflow.WorkflowData.Resumable {
		t.Error("workflow should not be resumable")
	}

	// Validate Workflow Execution
	if execution.Status != tempolite.ExecutionStatusCompleted {
		t.Errorf("expected execution status %s, got %s", tempolite.ExecutionStatusCompleted, execution.Status)
	}
	if execution.Error != "" {
		t.Errorf("expected no execution error, got: %s", execution.Error)
	}
	if execution.WorkflowEntityID != workflow.ID {
		t.Errorf("expected workflow entity ID %d, got %d", workflow.ID, execution.WorkflowEntityID)
	}
	if execution.CompletedAt == nil {
		t.Error("execution completed timestamp should be set")
	}

	// Verify no hierarchies exist for this workflow by checking directly
	hierarchies, err := database.GetHierarchiesByParentEntity(int(future.WorkflowID()))
	if err != nil && !errors.Is(err, tempolite.ErrHierarchyNotFound) {
		t.Fatal(err)
	}
	if len(hierarchies) != 0 {
		t.Errorf("expected 0 hierarchies, got %d", len(hierarchies))
	}

	// Validate RetryPolicy defaults
	defaultPolicy := tempolite.DefaultRetryPolicyInternal()
	if workflow.RetryPolicy.MaxAttempts != defaultPolicy.MaxAttempts {
		t.Errorf("expected max attempts %d, got %d", defaultPolicy.MaxAttempts, workflow.RetryPolicy.MaxAttempts)
	}
	if workflow.RetryPolicy.MaxInterval != defaultPolicy.MaxInterval {
		t.Errorf("expected max interval %d, got %d", defaultPolicy.MaxInterval, workflow.RetryPolicy.MaxInterval)
	}

	// Validate no retry attempts were made
	if workflow.RetryState.Attempts != 0 {
		t.Errorf("expected 0 retry attempts, got %d", workflow.RetryState.Attempts)
	}

	tp.Close()
}

func TestTempoliteWorkflowsExecuteSaveLoad(t *testing.T) {

	database := tempolite.NewMemoryDatabase()

	ctx := context.Background()

	tp, err := tempolite.New(
		ctx,
		database,
	)
	if err != nil {
		t.Fatal(err)
	}

	someActivity := func(ctx tempolite.ActivityContext) error {
		<-time.After(1 * time.Second)
		return nil
	}

	var flagTriggered atomic.Bool
	workflowFunc := func(ctx tempolite.WorkflowContext) error {
		flagTriggered.Store(true)
		<-time.After(1 * time.Second)
		if err := ctx.Activity("someActivity", someActivity, nil).Get(); err != nil {
			return err
		}
		return nil
	}

	future, err := tp.ExecuteDefault(workflowFunc, nil)
	if err != nil {
		t.Fatal(err)
	}

	<-time.After(time.Second / 2)

	if err := tp.Pause("default", future.WorkflowID()); err != nil {
		t.Fatal(err)
	}

	if err := tp.Wait(); err != nil {
		t.Fatal(err)
	}

	tp.Close()

	database.SaveAsJSON("./jsons/tempolite_workflows_execute_save.json")

	database = tempolite.NewMemoryDatabase()
	if err := database.LoadFromJSON("./jsons/tempolite_workflows_execute_save.json"); err != nil {
		t.Fatal(err)
	}

	tp, err = tempolite.New(
		ctx,
		database,
	)
	if err != nil {
		t.Fatal(err)
	}

	tp.RegisterWorkflow(workflowFunc)

	future, err = tp.Resume(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

	database.SaveAsJSON("./jsons/tempolite_workflows_execute_load.json")
}

func TestTempoliteWorkflowsExecuteGet(t *testing.T) {

	database := tempolite.NewMemoryDatabase()
	defer database.SaveAsJSON("./jsons/tempolite_workflows_execute_get.json")
	ctx := context.Background()

	tp, err := tempolite.New(
		ctx,
		database,
	)
	if err != nil {
		t.Fatal(err)
	}

	var flagTriggered atomic.Bool
	workflowFunc := func(ctx tempolite.WorkflowContext) (int, float64, error) {
		flagTriggered.Store(true)
		return 42, 3.14, nil
	}

	future, err := tp.ExecuteDefault(workflowFunc, nil)
	if err != nil {
		t.Fatal(err)
	}

	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

	newFuture, err := tp.Get(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	var resultInt int
	var resultFloat float64
	if err := newFuture.Get(&resultInt, &resultFloat); err != nil {
		t.Fatal(err)
	}

	if resultInt != 42 {
		t.Errorf("expected result int 42, got %d", resultInt)
	}

	if resultFloat != 3.14 {
		t.Errorf("expected result float 3.14, got %f", resultFloat)
	}

	t.Logf("result int: %d, result float: %f", resultInt, resultFloat)

}

// TODO: When a workflow created many subworkflows and only return their IDS, even if we pause them all, we might need also to remove the task from the task queue of the pool!
// otherwise we might end up a taskqueue that will still process the workflow
// Also, we might stop while a workflow is still Running, we might need to wait for it to reach a pause state or completed state
func TestTempoliteWorkflowsExecuteSubWorkflowsTaskQueue(t *testing.T) {
	database := tempolite.NewMemoryDatabase()
	defer database.SaveAsJSON("./jsons/tempolite_workflows_execute_subworkflows_taskqueue.json")
	ctx := context.Background()

	tp, err := tempolite.New(
		ctx,
		database,
	)
	if err != nil {
		t.Fatal(err)
	}

	subWorkflowFunc := func(ctx tempolite.WorkflowContext) (int, float64, error) {
		// no way to pause here since we have the workflow directly
		// we have some kind of situation here
		<-time.After(1 * time.Second)
		return 42, 3.14, nil
	}

	var flagTriggered atomic.Bool
	workflowFunc := func(ctx tempolite.WorkflowContext) ([]tempolite.WorkflowEntityID, error) {
		flagTriggered.Store(true)
		wids := []tempolite.WorkflowEntityID{}
		for i := 0; i < 100; i++ {
			// that workflow will be executed on a goroutine
			future := ctx.Workflow(fmt.Sprintf("sub%d", i), subWorkflowFunc, &tempolite.WorkflowOptions{
				DeferExecution: true,
			})
			wids = append(wids, future.WorkflowID())
		}
		return wids, nil
	}

	future, err := tp.ExecuteDefault(workflowFunc, nil)
	if err != nil {
		t.Fatal(err)
	}

	metrics := tp.Metrics()
	pp.Println("before list", metrics)

	var wids []tempolite.WorkflowEntityID
	if err := future.Get(&wids); err != nil {
		t.Fatal(err)
	}

	metrics = tp.Metrics()
	pp.Println("before enqueue", metrics)

	// We manually enqueue the defered workflow
	future, err = tp.Enqueue(wids[0])
	if err != nil {
		t.Fatal(err)
	}

	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

	wrkf, err := tp.GetWorkflow(wids[0])
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(wrkf)
}

func TestTempoliteWorkflowsExecuteSubWorkflowsTaskQueueWait(t *testing.T) {
	database := tempolite.NewMemoryDatabase()
	defer database.SaveAsJSON("./jsons/tempolite_workflows_execute_subworkflows_taskqueue_wait.json")
	ctx := context.Background()

	tp, err := tempolite.New(
		ctx,
		database,
		tempolite.WithDefaultQueueWorkers(1, 5),
	)
	if err != nil {
		t.Fatal(err)
	}

	subWorkflowFunc := func(ctx tempolite.WorkflowContext, id int) (int, float64, error) {
		// no way to pause here since we have the workflow directly
		// we have some kind of situation here
		seconds := time.Duration(rand.Intn(5) + 1)
		fmt.Println(id, "sleeping for", seconds)
		defer fmt.Println(id, "done sleeping")
		<-time.After(seconds * time.Second)
		return 42, 3.14, nil
	}

	var flagTriggered atomic.Bool
	workflowFunc := func(ctx tempolite.WorkflowContext) ([]tempolite.WorkflowEntityID, error) {
		flagTriggered.Store(true)
		wids := []tempolite.WorkflowEntityID{}
		for i := 0; i < 10; i++ {
			// that workflow will be executed on a goroutine
			future := ctx.Workflow(
				fmt.Sprintf("sub%d", i),
				subWorkflowFunc,
				&tempolite.WorkflowOptions{
					DeferExecution: true,
				},
				i,
			)
			wids = append(wids, future.WorkflowID())
		}
		return wids, nil
	}

	future, err := tp.ExecuteDefault(workflowFunc, nil)
	if err != nil {
		t.Fatal(err)
	}

	metrics := tp.Metrics()
	pp.Println("before list", metrics)

	var list []tempolite.WorkflowEntityID
	if err := future.Get(&list); err != nil {
		t.Fatal(err)
	}

	metrics = tp.Metrics()
	pp.Println("before enqueue", metrics)

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				metrics := tp.Metrics()
				pp.Println("metrics", metrics)
			case <-ctx.Done():
				return
			}
		}
	}()

	if len(list) > 0 {
		for _, id := range list {
			wrk, err := tp.GetWorkflow(id)
			if err != nil {
				fmt.Println(err)
				continue
			}
			if wrk.Status == tempolite.StatusPaused {
				if _, err := tp.Resume(id); err != nil {
					panic(err)
				}
			} else if wrk.Status == tempolite.StatusPending {
				fmt.Println("\tenqueuing", id)
				if _, err := tp.Enqueue(id); err != nil {
					panic(err)
				}
			} else {
				fmt.Printf("Workflow %d is %s\n", id, wrk.Status)
			}
		}
	} else {
		t.Fatal("no workflows to wait for")
	}

	metrics = tp.Metrics()
	pp.Println("before wait", metrics) // it's empty but we see that we enqueued workflows wtf

	if err := tp.Wait(); err != nil {
		t.Fatal(err)
	}

	fmt.Println("====== done ======")

}
