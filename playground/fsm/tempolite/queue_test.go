package tempolite

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/davidroman0O/retrypool"
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

func TestQueueWorkflow(t *testing.T) {

	database := NewDefaultDatabase()
	register := NewRegistry()
	ctx := context.Background()

	workflow := func(ctx WorkflowContext) error {
		t.Log("Workflow called")
		return nil
	}

	register.RegisterWorkflow(workflow)

	qm := newQueueManager(ctx, "default", 1, register, database)

	future := qm.ExecuteRuntimeWorkflow(workflow, nil)

	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

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

	processed := retrypool.NewProcessedNotification()
	if err := qm.ExecuteDatabaseWorkflow(id, processed); err != nil {
		t.Fatal(err)
	}

	t.Log("Waiting for being processed")
	<-processed
	t.Log("Processed")

	future := NewDatabaseFuture(ctx, database)
	future.setEntityID(id)

	t.Log("Waiting for future")
	if err := future.Get(); err != nil { // it is stuck there
		t.Fatal(err)
	}

	t.Log("Waiting for queue manager")
	if err := qm.Wait(); err != nil {
		t.Fatal(err)
	}
}

func TestQueueWorkflowActivity(t *testing.T) {

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

	future := qm.ExecuteRuntimeWorkflow(workflow, nil)

	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

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

	processed := retrypool.NewProcessedNotification()
	if err := qm.ExecuteDatabaseWorkflow(id, processed); err != nil {
		t.Fatal(err)
	}

	t.Log("Waiting for being processed")
	<-processed
	t.Log("Processed")

	future := NewDatabaseFuture(ctx, database)
	future.setEntityID(id)

	t.Log("Waiting for future")
	if err := future.Get(); err != nil { // it is stuck there
		t.Fatal(err)
	}

	t.Log("Waiting for queue manager")
	if err := qm.Wait(); err != nil {
		t.Fatal(err)
	}
}

// TODO: make test when we pause/resume

func TestQueueWorkflowDbPauseResume(t *testing.T) {

	database := NewDefaultDatabase()
	register := NewRegistry()
	ctx := context.Background()

	executed := 0

	subWorkflow := func(ctx WorkflowContext) error {
		<-time.After(1 * time.Second)
		return nil
	}

	workflow := func(ctx WorkflowContext) error {
		t.Log("Executing workflows")
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

	qm := newQueueManager(ctx, "default", 1, register, database)

	id, err := qm.CreateWorkflow(workflow, nil)
	if err != nil {
		t.Fatal(err)
	}

	processed := retrypool.NewProcessedNotification()
	if err := qm.ExecuteDatabaseWorkflow(id, processed); err != nil {
		t.Fatal(err)
	}

	t.Log("Waiting for being processed")
	<-processed
	t.Log("Processed")

	future := NewDatabaseFuture(ctx, database)
	future.setEntityID(id)

	go func() {
		<-time.After(1 * time.Second)
		if _, err := qm.Pause(id).Wait(context.Background()); err != nil {
			t.Fatal(err)
		}
		fmt.Println("Pausing orchestrator, 2s")
		<-time.After(2 * time.Second)
		fmt.Println("Resuming orchestrator")
		if _, err := qm.Resume(id).Wait(context.Background()); err != nil {
			t.Fatal(err)
		}
	}()

	t.Log("Waiting for future")
	if err := future.Get(); err != nil { // it is stuck there
		if !errors.Is(err, ErrPaused) {
			t.Fatal(err)
		}
	}

	<-time.After(2 * time.Second)

	t.Log("Waiting for future post wait")
	if err := future.Get(); err != nil { // it is stuck there
		fmt.Println("post wait", err)
		t.Fatal(err)
	}

	t.Log("Waiting for queue manager")
	if err := qm.Wait(); err != nil {
		t.Fatal(err)
	}

}

// TODO: make test when we restart Tempolite
