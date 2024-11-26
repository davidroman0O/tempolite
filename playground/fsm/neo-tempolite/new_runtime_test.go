package tempolite

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"errors"
)

func TestUnitPrepareRootWorkflowEntity(t *testing.T) {

	ctx := context.Background()
	db := NewMemoryDatabase()
	registry := NewRegistry()

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

}

func TestUnitPrepareRootWorkflowEntityPanic(t *testing.T) {

	ctx := context.Background()
	db := NewMemoryDatabase()
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

}

func TestUnitPrepareRootWorkflowEntityFailOnce(t *testing.T) {

	ctx := context.Background()
	db := NewMemoryDatabase()
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

func TestUnitPrepareRootWorkflowActivityEntityFailureOnce(t *testing.T) {

	ctx := context.Background()
	db := NewMemoryDatabase()
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

func TestUnitPrepareRootWorkflowActivityEntityPanic(t *testing.T) {

	ctx := context.Background()
	db := NewMemoryDatabase()
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

	future := o.Execute(wrfl, &WorkflowOptions{
		RetryPolicy: &RetryPolicy{
			MaxAttempts: 1, // which is the default already
		},
	})

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
