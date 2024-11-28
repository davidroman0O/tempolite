package tempolite

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"errors"
)

func TestUnitPrepareRootWorkflowEntity(t *testing.T) {

	ctx := context.Background()
	db := NewMemoryDatabase()
	registry := NewRegistry()

	defer db.SaveAsJSON("./json/workflow_entity.json")

	o := NewOrchestrator(ctx, db, registry)

	wrfl := func(ctx WorkflowContext) error {
		fmt.Println("Hello, World!")
		return nil
	}

	future := o.Execute(wrfl, nil)

	if future == nil {
		t.Fatal("future is nil")
	}

	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

	workflow, err := db.GetWorkflowEntity(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if workflow.Status != StatusCompleted {
		t.Fatalf("expected %s, got %s", StatusCompleted, workflow.Status)
	}

	execs, err := db.GetWorkflowExecutions(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if len(execs) != 1 {
		t.Fatalf("expected 1 execution, got %d", len(execs))
	}

	for _, v := range execs {
		if v.Status != ExecutionStatusCompleted {
			t.Fatalf("expected %s, got %s", ExecutionStatusCompleted, v.Status)
		}
	}

}

func TestUnitPrepareRootWorkflowEntityPanic(t *testing.T) {

	ctx := context.Background()
	db := NewMemoryDatabase()
	defer db.SaveAsJSON("./json/workflow_entity_panic.json")
	registry := NewRegistry()

	o := NewOrchestrator(ctx, db, registry)

	wrfl := func(ctx WorkflowContext) error {
		fmt.Println("Hello, World!")
		panic("panic")
		return nil
	}

	future := o.Execute(wrfl, &WorkflowOptions{
		RetryPolicy: &RetryPolicy{
			MaxAttempts: 1, // which is the default already
		},
	})

	if future == nil {
		t.Fatal("future is nil")
	}

	err := future.Get()

	if err == nil {
		t.Fatalf("error shouldn't be nil")
	}

	workflow, err := db.GetWorkflowEntity(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if workflow.Status != StatusFailed {
		t.Fatalf("expected %s, got %s", StatusFailed, workflow.Status)
	}

	execs, err := db.GetWorkflowExecutions(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if len(execs) != 2 {
		t.Fatalf("expected 2 execution, got %d", len(execs))
	}

	for _, v := range execs {
		if v.Status != ExecutionStatusFailed {
			t.Fatalf("expected %s, got %s", ExecutionStatusFailed, v.Status)
		}
	}

}

func TestUnitPrepareRootWorkflowEntityFailOnce(t *testing.T) {

	ctx := context.Background()
	db := NewMemoryDatabase()
	defer db.SaveAsJSON("./json/workflow_entity_fail_once.json")
	registry := NewRegistry()

	o := NewOrchestrator(ctx, db, registry)

	var atomicFailure atomic.Bool
	atomicFailure.Store(true)

	wrfl := func(ctx WorkflowContext) error {
		fmt.Println("Hello, World!")
		if atomicFailure.Load() {
			atomicFailure.Store(false)
			return fmt.Errorf("on purpose")
		}
		return nil
	}

	future := o.Execute(wrfl, &WorkflowOptions{
		RetryPolicy: &RetryPolicy{
			MaxAttempts: 1, // which is the default already
		},
	})

	if future == nil {
		t.Fatal("future is nil")
	}

	err := future.Get()

	if err != nil {
		t.Fatalf("error be nil")
	}

	executions, err := db.GetWorkflowExecutions(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if len(executions) != 2 {
		t.Fatalf("expected 2 executions, got %d", len(executions))
	}

	found := false
	for _, exec := range executions {
		if exec.ID == 1 {
			found = true
			if exec.Status != ExecutionStatusFailed {
				t.Fatalf("expected %s, got %s", ExecutionStatusFailed, exec.Status)
			}
			if exec.Error == "" {
				t.Fatalf("expected error, got nil")
			}
		}
	}

	if !found {
		t.Fatalf("expected execution with ID 1, got none")
	}

}

func TestUnitPrepareRootWorkflowActivityEntity(t *testing.T) {

	ctx := context.Background()
	db := NewMemoryDatabase()
	registry := NewRegistry()

	o := NewOrchestrator(ctx, db, registry)

	act := func(ctx ActivityContext) error {
		fmt.Println("Activity, World!")
		return nil
	}

	wrfl := func(ctx WorkflowContext) error {
		fmt.Println("Hello, World!")
		if err := ctx.Activity("activity", act, nil).Get(); err != nil {
			return err
		}
		return nil
	}

	future := o.Execute(wrfl, &WorkflowOptions{
		RetryPolicy: &RetryPolicy{
			MaxAttempts: 1, // which is the default already
		},
	})

	if future == nil {
		t.Fatal("future is nil")
	}

	if err := future.Get(); err != nil {
		t.Fatal(err)
	}
}

func TestUnitPrepareRootWorkflowActivityEntityWithOuputs(t *testing.T) {

	ctx := context.Background()
	db := NewMemoryDatabase()
	registry := NewRegistry()

	o := NewOrchestrator(ctx, db, registry)

	act := func(ctx ActivityContext) (int, error) {
		fmt.Println("Activity, World!")
		return 420, nil
	}

	wrfl := func(ctx WorkflowContext) (int, error) {
		fmt.Println("Hello, World!")
		var result int
		if err := ctx.Activity("activity", act, nil).Get(&result); err != nil {
			return -1, err
		}
		return result + 1, nil
	}

	future := o.Execute(wrfl, &WorkflowOptions{
		RetryPolicy: &RetryPolicy{
			MaxAttempts: 1, // which is the default already
		},
	})

	if future == nil {
		t.Fatal("future is nil")
	}

	finalRes := 0

	if err := future.Get(&finalRes); err != nil {
		t.Fatal(err)
	}

	if finalRes != 421 {
		t.Fatalf("expected 421, got %d", finalRes)
	}
}

func TestUnitPrepareRootWorkflowActivityEntityFailureOnce(t *testing.T) {

	ctx := context.Background()
	db := NewMemoryDatabase()
	defer db.SaveAsJSON("./json/workflow_activity_entity_fail_once.json")
	registry := NewRegistry()

	o := NewOrchestrator(ctx, db, registry)

	act := func(ctx ActivityContext) error {
		fmt.Println("Activity, World!")
		return nil
	}

	var atomicFailure atomic.Bool
	atomicFailure.Store(true)

	wrfl := func(ctx WorkflowContext) error {
		fmt.Println("Hello, World!")
		if err := ctx.Activity("activity", act, nil).Get(); err != nil {
			return err
		}
		if atomicFailure.Load() {
			atomicFailure.Store(false)
			return fmt.Errorf("on purpose")
		}
		return nil
	}

	future := o.Execute(wrfl, &WorkflowOptions{
		RetryPolicy: &RetryPolicy{
			MaxAttempts: 1, // which is the default already
		},
	})

	if future == nil {
		t.Fatal("future is nil")
	}

	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

	wexecs, err := db.GetWorkflowExecutions(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if len(wexecs) != 2 {
		t.Fatalf("expected 2 executions, got %d", len(wexecs))
	}

	hierarchies, err := db.GetHierarchiesByParentEntity(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	countActivityChildren := 0

	for _, h := range hierarchies {
		if h.ParentEntityID == future.WorkflowID() && h.ChildType == EntityActivity {
			countActivityChildren++
		}
	}

	if countActivityChildren != 1 {
		t.Fatalf("expected 1 activity children, got %d", countActivityChildren)
	}

}

func TestUnitPrepareRootWorkflowActivityEntityWithOutputFailureOnce(t *testing.T) {

	ctx := context.Background()
	db := NewMemoryDatabase()
	registry := NewRegistry()

	o := NewOrchestrator(ctx, db, registry)

	var counterActivityCalled atomic.Int32

	act := func(ctx ActivityContext) (int, error) {
		fmt.Println("Activity, World!")
		counterActivityCalled.Add(1)
		return 420, nil
	}

	var atomicFailure atomic.Bool
	atomicFailure.Store(true)

	var counterWorkflowCalled atomic.Int32

	wrfl := func(ctx WorkflowContext) (int, error) {
		fmt.Println("Hello, World!")
		counterWorkflowCalled.Add(1)
		if err := ctx.Activity("activity", act, nil).Get(); err != nil {
			return -1, err
		}
		if atomicFailure.Load() {
			atomicFailure.Store(false)
			return -1, fmt.Errorf("on purpose")
		}
		return 420 + 1, nil
	}

	future := o.Execute(wrfl, &WorkflowOptions{
		RetryPolicy: &RetryPolicy{
			MaxAttempts: 1, // which is the default already
		},
	})

	if future == nil {
		t.Fatal("future is nil")
	}

	finalRes := 0

	if err := future.Get(&finalRes); err != nil {
		t.Fatal(err)
	}

	wexecs, err := db.GetWorkflowExecutions(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if len(wexecs) != 2 {
		t.Fatalf("expected 2 executions, got %d", len(wexecs))
	}

	hierarchies, err := db.GetHierarchiesByParentEntity(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	countActivityChildren := 0

	for _, h := range hierarchies {
		if h.ParentEntityID == future.WorkflowID() && h.ChildType == EntityActivity {
			countActivityChildren++
		}
	}

	if countActivityChildren != 1 {
		t.Fatalf("expected 1 activity children, got %d", countActivityChildren)
	}

	if finalRes != 421 {
		t.Fatalf("expected 421, got %d", finalRes)
	}

	if counterActivityCalled.Load() != 1 {
		t.Fatalf("expected 1 activity call, got %d", counterActivityCalled.Load())
	}

	if counterWorkflowCalled.Load() != 2 {
		t.Fatalf("expected 2 workflow call, got %d", counterWorkflowCalled.Load())
	}

}

func TestUnitPrepareRootWorkflowActivityEntityPanic(t *testing.T) {

	ctx := context.Background()
	db := NewMemoryDatabase()
	defer db.SaveAsJSON("./json/workflow_activity_entity_panic.json")
	registry := NewRegistry()

	o := NewOrchestrator(ctx, db, registry)

	act := func(ctx ActivityContext) error {
		fmt.Println("Activity, World!")
		panic("panic on purpose")
		return nil
	}

	wrfl := func(ctx WorkflowContext) error {
		fmt.Println("Hello, World!")
		if err := ctx.Activity("activity", act, nil).Get(); err != nil {
			return err
		}
		return nil
	}

	future := o.Execute(wrfl, nil)

	if future == nil {
		t.Fatal("future is nil")
	}

	if err := future.Get(); err != nil {
		fmt.Println(err)
		if !errors.Is(err, ErrActivityPanicked) {
			t.Fatalf("expected ErrActivityPanicked, got %v", err)
		}
	}
}

func TestUnitPrepareRootWorkflowContinueAsNew(t *testing.T) {

	ctx := context.Background()
	db := NewMemoryDatabase()
	registry := NewRegistry()

	o := NewOrchestrator(ctx, db, registry)

	var counterContinueAsNew atomic.Int32

	wrfl := func(ctx WorkflowContext) (int, error) {
		fmt.Println("Hello, World!", counterContinueAsNew.Load())
		if counterContinueAsNew.Load() == 0 {
			counterContinueAsNew.Add(1)
			return 42, ctx.ContinueAsNew(nil)
		}
		return 48, nil
	}

	future := o.Execute(wrfl, &WorkflowOptions{
		RetryPolicy: &RetryPolicy{
			MaxAttempts: 1, // which is the default already
		},
	})

	if future == nil {
		t.Fatal("future is nil")
	}

	firstRes := 0

	if err := future.Get(&firstRes); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	if !future.ContinuedAsNew() {
		t.Fatalf("expected true, got false")
	}

	if firstRes != 42 {
		t.Fatalf("expected 42, got %d", firstRes)
	}

	// The orchestrator provide a function to allow you to execute any entity
	// A workflow continued as new is just a new workflow to execute
	future, err := o.ExecuteWithEntity(future.ContinuedAs())
	if err != nil {
		t.Fatal(err)
	}

	if future == nil {
		t.Fatal("future is nil")
	}

	secondRes := 0

	if err := future.Get(&secondRes); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	if secondRes != 48 {
		t.Fatalf("expected 48, got %d", secondRes)
	}

	if future.ContinuedAsNew() {
		t.Fatalf("expected false, got true")
	}

	db.SaveAsJSON("./json/continue_as_new.json")
}

func TestUnitPrepareRootWorkflowActivityEntityPauseResume(t *testing.T) {

	ctx := context.Background()
	db := NewMemoryDatabase()
	registry := NewRegistry()
	defer db.SaveAsJSON("./json/workflow_pause_resume.json")

	o := NewOrchestrator(ctx, db, registry)

	act := func(ctx ActivityContext, number int) error {
		fmt.Println("Activity, World!", number)
		<-time.After(1 * time.Second)
		return nil
	}

	wrfl := func(ctx WorkflowContext) error {
		fmt.Println("Hello, World!")
		if err := ctx.Activity("activity1", act, nil, 1).Get(); err != nil {
			return err
		}
		if err := ctx.Activity("activity2", act, nil, 2).Get(); err != nil {
			return err
		}
		if err := ctx.Activity("activity3", act, nil, 3).Get(); err != nil {
			return err
		}
		if err := ctx.Activity("activity4", act, nil, 4).Get(); err != nil {
			return err
		}
		return nil
	}

	future := o.Execute(wrfl, &WorkflowOptions{
		RetryPolicy: &RetryPolicy{
			MaxAttempts: 1, // which is the default already
		},
	})

	if future == nil {
		t.Fatal("future is nil")
	}

	<-time.After(time.Second)

	o.Pause()

	o.WaitActive()

	db.SaveAsJSON("./json/workflow_paused.json")

	future = o.Resume(future.WorkflowID())

	if future == nil {
		t.Fatal("future is nil")
	}

	if err := future.Get(); err != nil {
		t.Fatal(err)
	}
}

func TestUnitPrepareRootWorkflowActivityEntityPauseResumeWithFailure(t *testing.T) {

	ctx := context.Background()
	db := NewMemoryDatabase()
	registry := NewRegistry()
	defer db.SaveAsJSON("./json/workflow_pause_resume_with_failure.json")

	o := NewOrchestrator(ctx, db, registry)

	act := func(ctx ActivityContext, number int) error {
		fmt.Println("Activity, World!", number)
		<-time.After(1 * time.Second)
		return nil
	}

	var failure atomic.Bool
	failure.Store(true)

	wrfl := func(ctx WorkflowContext) error {
		fmt.Println("Hello, World!")
		if err := ctx.Activity("activity1", act, nil, 1).Get(); err != nil {
			return err
		}
		if failure.Load() {
			failure.Store(false)
			return fmt.Errorf("on purpose")
		}
		if err := ctx.Activity("activity2", act, nil, 2).Get(); err != nil {
			return err
		}
		if err := ctx.Activity("activity3", act, nil, 3).Get(); err != nil {
			return err
		}
		if err := ctx.Activity("activity4", act, nil, 4).Get(); err != nil {
			return err
		}
		return nil
	}

	future := o.Execute(wrfl, &WorkflowOptions{
		RetryPolicy: &RetryPolicy{
			MaxAttempts: 1, // which is the default already
		},
	})

	if future == nil {
		t.Fatal("future is nil")
	}

	<-time.After(time.Second)

	o.Pause()

	o.WaitActive()

	db.SaveAsJSON("./json/workflow_paused_with_failure.json")

	future = o.Resume(future.WorkflowID())

	if future == nil {
		t.Fatal("future is nil")
	}

	if err := future.Get(); err != nil {
		t.Fatal(err)
	}
}
