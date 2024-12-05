package tempolite

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"errors"

	"github.com/k0kubun/pp/v3"
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
	defer db.SaveAsJSON("./json/workflow_activity_entity.json")
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

	workflowEntity, err := db.GetWorkflowEntity(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if workflowEntity.Status != StatusCompleted {
		t.Fatalf("expected %s, got %s", StatusCompleted, workflowEntity.Status)
	}

	activities, err := db.GetActivityEntities(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if len(activities) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(activities))
	}

}

func TestUnitPrepareRootWorkflowActivityEntityWithOuputs(t *testing.T) {

	ctx := context.Background()
	db := NewMemoryDatabase()
	defer db.SaveAsJSON("./json/workflow_activity_entity_with_outputs.json")
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

	workflowEntity, err := db.GetWorkflowEntity(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if workflowEntity.Status != StatusCompleted {
		t.Fatalf("expected %s, got %s", StatusCompleted, workflowEntity.Status)
	}

	activities, err := db.GetActivityEntities(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if len(activities) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(activities))
	}

	exects, err := db.GetActivityExecutions(activities[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	data, err := db.GetActivityExecutionData(exects[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	pp.Println(data)

	handler, ok := o.registry.GetActivityFunc(act)
	if !ok {
		t.Fatal("activity not found")
	}

	output, err := convertOutputsFromSerialization(handler, data.Outputs)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("output", output)

	if len(output) != 1 {
		t.Fatalf("expected 1 output, got %d", len(output))
	}

	if output[0].(int) != 420 {
		t.Fatalf("expected 420, got %d", output[0].(int))
	}

	workflowData, err := db.GetWorkflowExecutionData(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if len(workflowData.Outputs) != 1 {
		t.Fatalf("expected 1 workflow data, got %d", len(workflowData.Outputs))
	}

	handlerWork, ok := o.registry.GetWorkflow(getFunctionName(wrfl))
	if !ok {
		t.Fatal("workflow not found")
	}

	workflowoutput, err := convertOutputsFromSerialization(handlerWork, workflowData.Outputs)
	if err != nil {
		t.Fatal(err)
	}

	if len(workflowoutput) != 1 {
		t.Fatalf("expected 1 workflow output, got %d", len(workflowoutput))
	}

	if workflowoutput[0].(int) != 421 {
		t.Fatalf("expected 421, got %d", workflowoutput[0].(int))
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

	workflowEntity, err := db.GetWorkflowEntity(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if workflowEntity.Status != StatusCompleted {
		t.Fatalf("expected %s, got %s", StatusCompleted, workflowEntity.Status)
	}

	activities, err := db.GetActivityEntities(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if len(activities) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(activities))
	}

	exects, err := db.GetActivityExecutions(activities[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	data, err := db.GetActivityExecutionData(exects[0].ID)
	if err != nil {
		t.Fatal(err)
	}

	handler, ok := o.registry.GetActivityFunc(act)
	if !ok {
		t.Fatal("activity not found")
	}

	outputBytes, err := convertOutputsFromSerialization(handler, data.Outputs)
	if err != nil {
		t.Fatal(err)
	}

	if len(outputBytes) != 1 {
		t.Fatalf("expected 1 output, got %d", len(outputBytes))
	}

	if outputBytes[0].(int) != 420 {
		t.Fatalf("expected 420, got %d", outputBytes[0].(int))
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

	workflowEntity, err := db.GetWorkflowEntity(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if workflowEntity.Status != StatusFailed {
		t.Fatalf("expected %s, got %s", StatusFailed, workflowEntity.Status)
	}

	activities, err := db.GetActivityEntities(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if len(activities) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(activities))
	}

	for _, a := range activities {
		if a.Status != StatusFailed {
			t.Fatalf("expected %s, got %s", StatusFailed, a.Status)
		}
	}
}

func TestUnitPrepareRootWorkflowContinueAsNew(t *testing.T) {

	ctx := context.Background()
	db := NewMemoryDatabase()
	defer db.SaveAsJSON("./json/continue_as_new.json")
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

	futureFirst := o.Execute(wrfl, &WorkflowOptions{
		RetryPolicy: &RetryPolicy{
			MaxAttempts: 1, // which is the default already
		},
	})

	if futureFirst == nil {
		t.Fatal("futureFirst is nil")
	}

	firstRes := 0

	if err := futureFirst.Get(&firstRes); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	if firstRes != 42 {
		t.Fatalf("expected 42, got %d", firstRes)
	}

	// The orchestrator provide a function to allow you to execute any entity
	// A workflow continued as new is just a new workflow to execute
	futureSecond, err := o.ExecuteWithEntity(2) // TODO: should have a function to get pending workflows
	if err != nil {
		t.Fatal(err)
	}

	if futureSecond == nil {
		t.Fatal("futureSecond is nil")
	}

	secondRes := 0

	if err := futureSecond.Get(&secondRes); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	if secondRes != 48 {
		t.Fatalf("expected 48, got %d", secondRes)
	}

	first, err := db.GetWorkflowEntity(futureFirst.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}
	second, err := db.GetWorkflowEntity(futureSecond.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if first.Status != StatusCompleted {
		t.Fatalf("expected %s, got %s", StatusCompleted, first.Status)
	}

	if second.Status != StatusCompleted {
		t.Fatalf("expected %s, got %s", StatusCompleted, second.Status)
	}

	if first.ID == second.ID {
		t.Fatalf("expected different IDs, got the same")
	}

	execFirst, err := db.GetWorkflowExecutions(first.ID)
	if err != nil {
		t.Fatal(err)
	}
	execSecond, err := db.GetWorkflowExecutions(second.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(execFirst) != 1 {
		t.Fatalf("expected 1 execution, got %d", len(execFirst))
	}

	if len(execSecond) != 1 {
		t.Fatalf("expected 1 execution, got %d", len(execSecond))
	}

	if execFirst[0].Status != ExecutionStatusCompleted {
		t.Fatalf("expected %s, got %s", ExecutionStatusCompleted, execFirst[0].Status)
	}

	if execSecond[0].Status != ExecutionStatusCompleted {
		t.Fatalf("expected %s, got %s", ExecutionStatusCompleted, execSecond[0].Status)
	}
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

	if err := o.WaitFor(future.WorkflowID(), StatusPaused); err != nil {
		t.Fatal(err)
	}

	db.SaveAsJSON("./json/workflow_paused.json")

	future = o.Resume(future.WorkflowID())

	if future == nil {
		t.Fatal("future is nil")
	}

	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

	workflowEntity, err := db.GetWorkflowEntity(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if workflowEntity.Status != StatusCompleted {
		t.Fatalf("expected %s, got %s", StatusCompleted, workflowEntity.Status)
	}

	activities, err := db.GetActivityEntities(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if len(activities) != 4 {
		t.Fatalf("expected 4 activities, got %d", len(activities))
	}

	for _, a := range activities {
		if a.Status != StatusCompleted {
			t.Fatalf("expected %s, got %s", StatusCompleted, a.Status)
		}
		execs, err := db.GetActivityExecutions(a.ID)
		if err != nil {
			t.Fatal(err)
		}

		if len(execs) != 1 {
			t.Fatalf("expected 1 execution, got %d", len(execs))
		}
	}

	workflowExecs, err := db.GetWorkflowExecutions(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if len(workflowExecs) != 2 {
		t.Fatalf("expected 2 executions, got %d", len(workflowExecs))
	}

}

func TestUnitPrepareRootWorkflowActivityEntityDetectContextCancellation(t *testing.T) {

	ctx := context.Background()
	db := NewMemoryDatabase()
	registry := NewRegistry()
	defer db.SaveAsJSON("./json/workflow_context_cancel.json")

	o := NewOrchestrator(ctx, db, registry)

	var contextCancelledDetected atomic.Bool
	var pauseTriggered atomic.Bool

	act := func(ctx ActivityContext, number int) error {
		fmt.Println("Activity, World!", number)
		duration := 10 * time.Second
		if pauseTriggered.Load() {
			duration = 1 * time.Second
		}
		select {
		case <-ctx.Done():
			fmt.Println("activity detected context cancellation")
			contextCancelledDetected.Store(true)
			return ctx.Err()
		case <-time.After(duration):
			return nil
		}
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

	pauseTriggered.Store(true)
	o.Cancel()

	if err := o.WaitFor(future.WorkflowID(), StatusCancelled); err != nil {
		t.Fatal(err)
	}

	if err := future.Get(); err != nil {
		if !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	}

	if !contextCancelledDetected.Load() {
		t.Fatalf("expected true, got false")
	}

	workflowEntity, err := db.GetWorkflowEntity(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if workflowEntity.Status != StatusCancelled {
		t.Fatalf("expected %s, got %s", StatusCancelled, workflowEntity.Status)
	}

	activities, err := db.GetActivityEntities(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	// at least one
	if len(activities) == 0 {
		t.Fatalf("expected at least 1 activity, got 0")
	}

	for _, a := range activities {
		if a.Status != StatusCancelled {
			t.Fatalf("expected %s, got %s", StatusCancelled, a.Status)
		}
		execs, err := db.GetActivityExecutions(a.ID)
		if err != nil {
			t.Fatal(err)
		}

		if len(execs) != 1 {
			t.Fatalf("expected 1 execution, got %d", len(execs))
		}

		if execs[0].Status != ExecutionStatusCancelled {
			t.Fatalf("expected %s, got %s", ExecutionStatusCancelled, execs[0].Status)
		}
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

	if err := o.WaitFor(future.WorkflowID(), StatusPaused); err != nil {
		t.Fatal(err)
	}

	db.SaveAsJSON("./json/workflow_paused_with_failure.json")

	future = o.Resume(future.WorkflowID())

	if future == nil {
		t.Fatal("future is nil")
	}

	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

	workflowEntity, err := db.GetWorkflowEntity(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if workflowEntity.Status != StatusCompleted {
		t.Fatalf("expected %s, got %s", StatusCompleted, workflowEntity.Status)
	}

	activities, err := db.GetActivityEntities(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if len(activities) != 4 {
		t.Fatalf("expected 4 activities, got %d", len(activities))
	}

	for _, a := range activities {
		if a.Status != StatusCompleted {
			t.Fatalf("expected %s, got %s", StatusCompleted, a.Status)
		}
		execs, err := db.GetActivityExecutions(a.ID)
		if err != nil {
			t.Fatal(err)
		}

		if len(execs) != 1 {
			t.Fatalf("expected 1 execution, got %d", len(execs))
		}

		for _, e := range execs {
			if e.Status != ExecutionStatusCompleted {
				t.Fatalf("expected %s, got %s", ExecutionStatusCompleted, e.Status)
			}
		}
	}

	workflowExecs, err := db.GetWorkflowExecutions(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if len(workflowExecs) != 3 {
		t.Fatalf("expected 3 executions, got %d", len(workflowExecs))
	}

	for _, e := range workflowExecs {
		if e.ID == 1 {
			if e.Status != ExecutionStatusFailed {
				t.Fatalf("expected %s, got %s", ExecutionStatusFailed, e.Status)
			}
		}
		if e.ID == 2 {
			if e.Status != ExecutionStatusPaused {
				t.Fatalf("expected %s, got %s", ExecutionStatusPaused, e.Status)
			}
		}
		if e.ID > 2 {
			if e.Status != ExecutionStatusCompleted {
				t.Fatalf("expected %s, got %s", ExecutionStatusCompleted, e.Status)
			}
		}
	}
}

func TestUnitPrepareRootWorkflowSideEffect(t *testing.T) {

	ctx := context.Background()
	db := NewMemoryDatabase()
	registry := NewRegistry()

	defer db.SaveAsJSON("./json/workflow_sideeffect.json")

	o := NewOrchestrator(ctx, db, registry)

	wrfl := func(ctx WorkflowContext) error {

		var value int
		if err := ctx.SideEffect("sideeffect", func() int {
			fmt.Println("SideEffect, World!")
			return 42
		}).Get(&value); err != nil {
			return err
		}

		fmt.Println("Hello, World!", value)

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

func TestUnitPrepareRootWorkflowSideEffectPanic(t *testing.T) {
	ctx := context.Background()
	db := NewMemoryDatabase()
	registry := NewRegistry()

	defer db.SaveAsJSON("./json/workflow_sideeffect_panic.json")

	o := NewOrchestrator(ctx, db, registry)

	wrfl := func(ctx WorkflowContext) error {
		var value int
		if err := ctx.SideEffect("sideeffect", func() int {
			fmt.Println("SideEffect, World!")
			panic("on purpose")
			return 42
		}).Get(&value); err != nil {
			return err
		}

		fmt.Println("Hello, World!", value)
		return nil
	}

	future := o.Execute(wrfl, &WorkflowOptions{
		RetryPolicy: &RetryPolicy{
			MaxAttempts: 1, // Add one retry
			MaxInterval: 100 * time.Millisecond,
		},
	})

	if future == nil {
		t.Fatal("future is nil")
	}

	if err := future.Get(); err != nil {
		if !errors.Is(err, ErrWorkflowFailed) {
			t.Fatalf("expected %v, got %v", ErrWorkflowFailed, err)
		}
	}

	workflow, err := db.GetWorkflowEntity(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if workflow.Status != StatusFailed {
		t.Fatalf("expected %s, got %s", StatusFailed, workflow.Status)
	}

	if workflow.RetryState.Attempts != 2 { // Initial attempt + 1 retry
		t.Fatalf("expected 2 attempts, got %d", workflow.RetryState.Attempts)
	}

	execs, err := db.GetWorkflowExecutions(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if len(execs) != 2 {
		t.Fatalf("expected 2 executions, got %d", len(execs))
	}

	// Sort executions by creation time to ensure correct order
	sort.Slice(execs, func(i, j int) bool {
		return execs[i].CreatedAt.Before(execs[j].CreatedAt)
	})

	for _, exec := range execs {
		if exec.Status != ExecutionStatusFailed {
			t.Fatalf("expected %s, got %s", ExecutionStatusFailed, exec.Status)
		}
		if !strings.Contains(exec.Error, "side effect panicked") {
			t.Fatalf("expected error containing 'side effect panicked', got '%s'", exec.Error)
		}
	}

	sideeffects, err := db.GetSideEffectEntities(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if len(sideeffects) != 1 {
		t.Fatalf("expected 1 sideeffect entity, got %d", len(sideeffects))
	}

	for _, se := range sideeffects {
		if se.Status != StatusFailed {
			t.Fatalf("expected %s, got %s", StatusFailed, se.Status)
		}

		seExecs, err := db.GetSideEffectExecutions(se.ID)
		if err != nil {
			t.Fatal(err)
		}

		if len(seExecs) != 2 {
			t.Fatalf("expected 2 side effect executions, got %d", len(seExecs))
		}

		// Sort side effect executions
		sort.Slice(seExecs, func(i, j int) bool {
			return seExecs[i].CreatedAt.Before(seExecs[j].CreatedAt)
		})

		for _, seExec := range seExecs {
			if seExec.Status != ExecutionStatusFailed {
				t.Fatalf("expected %s, got %s", ExecutionStatusFailed, seExec.Status)
			}
			if !strings.Contains(seExec.Error, "side effect panicked") {
				t.Fatalf("expected error containing 'side effect panicked', got '%s'", seExec.Error)
			}
		}
	}
}

func TestUnitPrepareRootWorkflowPanicSideEffect(t *testing.T) {

	ctx := context.Background()
	db := NewMemoryDatabase()
	registry := NewRegistry()

	defer db.SaveAsJSON("./json/workflow_panic_sideeffect.json")

	o := NewOrchestrator(ctx, db, registry)

	wrfl := func(ctx WorkflowContext) error {

		var value int
		if err := ctx.SideEffect("sideeffect", func() int {
			fmt.Println("SideEffect, World!")
			return 42
		}).Get(&value); err != nil {
			return err
		}
		panic("on purpose")

		fmt.Println("Hello, World!", value)

		return nil
	}

	future := o.Execute(wrfl, nil)

	if future == nil {
		t.Fatal("future is nil")
	}

	if err := future.Get(); err != nil {
		if !errors.Is(err, ErrWorkflowFailed) {
			t.Fatalf("expected %v, got %v", ErrWorkflowFailed, err)
		}
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

	// Expect only 1 execution since we don't want retries by default
	if len(execs) != 1 {
		t.Fatalf("expected 1 execution, got %d", len(execs))
	}

	// Single execution should be failed
	if execs[0].Status != ExecutionStatusFailed {
		t.Fatalf("expected %s, got %s", ExecutionStatusFailed, execs[0].Status)
	}

	sideeffects, err := db.GetSideEffectEntities(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if len(sideeffects) != 1 {
		t.Fatalf("expected 1 sideeffect, got %d", len(sideeffects))
	}

	for _, v := range sideeffects {
		if v.Status != StatusCompleted {
			t.Fatalf("expected %s, got %s", StatusCompleted, v.Status)
		}

		execs, err := db.GetSideEffectExecutions(v.ID)
		if err != nil {
			t.Fatal(err)
		}

		// Side effect should also only have 1 execution
		if len(execs) != 1 {
			t.Fatalf("expected 1 execution, got %d", len(execs))
		}

		if execs[0].Status != ExecutionStatusCompleted {
			t.Fatalf("expected %s, got %s", ExecutionStatusCompleted, execs[0].Status)
		}
	}
}

func TestUnitPrepareRootWorkflowPanicWithRetry(t *testing.T) {
	ctx := context.Background()
	db := NewMemoryDatabase()
	registry := NewRegistry()

	defer db.SaveAsJSON("./json/workflow_panic_sideeffect_with_retry.json")

	o := NewOrchestrator(ctx, db, registry)

	hasPanicked := atomic.Bool{}
	sideEffectCount := atomic.Int32{}

	wrfl := func(ctx WorkflowContext) error {
		var value int
		// The side effect should be using workflow context to track its result
		// and avoid re-execution during retry
		if err := ctx.SideEffect("sideeffect", func() int {
			newCount := sideEffectCount.Add(1)
			fmt.Printf("SideEffect executed %d time(s)\n", newCount)
			if newCount > 1 {
				t.Error("Side effect executed more than once")
			}
			return 42
		}).Get(&value); err != nil {
			return err
		}

		if !hasPanicked.Swap(true) {
			panic("on purpose")
		}

		fmt.Println("Hello, World!", value)
		return nil
	}

	future := o.Execute(wrfl, &WorkflowOptions{
		RetryPolicy: &RetryPolicy{
			MaxAttempts: 1,
			MaxInterval: 100 * time.Millisecond,
		},
	})

	if err := future.Get(); err != nil {
		t.Fatal("workflow should succeed on retry")
	}

	workflow, err := db.GetWorkflowEntity(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	// The RetryState.Attempts should be 1 (one retry)
	if workflow.RetryState.Attempts != 1 {
		t.Fatalf("expected 1 retry attempt, got %d", workflow.RetryState.Attempts)
	}

	execs, err := db.GetWorkflowExecutions(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	// Should have 2 executions - initial + 1 retry
	if len(execs) != 2 {
		t.Fatalf("expected 2 executions, got %d", len(execs))
	}

	// Sort executions by creation time to ensure correct order
	sort.Slice(execs, func(i, j int) bool {
		return execs[i].CreatedAt.Before(execs[j].CreatedAt)
	})

	// First execution should be failed due to panic
	if execs[0].Status != ExecutionStatusFailed {
		t.Fatalf("expected first execution %s, got %s", ExecutionStatusFailed, execs[0].Status)
	}

	if strings.Contains(execs[0].Error, `failed to run workflow instance
        workflow panicked`) {
		t.Fatalf("expected error 'workflow panicked', got '%s'", execs[0].Error)
	}

	// Second execution should be completed
	if execs[1].Status != ExecutionStatusCompleted {
		t.Fatalf("expected second execution %s, got %s", ExecutionStatusCompleted, execs[1].Status)
	}

	// Verify side effect was executed exactly once via the atomic counter
	if count := sideEffectCount.Load(); count != 1 {
		t.Fatalf("expected side effect to execute once, got %d executions", count)
	}

	// Check side effect entities and executions
	sideeffects, err := db.GetSideEffectEntities(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if len(sideeffects) != 1 {
		t.Fatalf("expected 1 sideeffect entity, got %d", len(sideeffects))
	}
}

type sagaStep struct {
	fail bool
}

func (s *sagaStep) Transaction(ctx TransactionContext) error {
	if s.fail {
		return fmt.Errorf("on purpose")
	}
	fmt.Println("Transaction, World!")
	return nil
}

func (s *sagaStep) Compensation(ctx CompensationContext) error {
	fmt.Println("Compensation, World!")
	return nil
}

func TestUnitPrepareRootWorkflowEntitySaga(t *testing.T) {

	ctx := context.Background()
	db := NewMemoryDatabase()
	registry := NewRegistry()

	defer db.SaveAsJSON("./json/workflow_entity_saga.json")

	o := NewOrchestrator(ctx, db, registry)

	wrfl := func(ctx WorkflowContext) error {
		fmt.Println("Hello, World!")

		def, err := NewSaga().
			Add(
				func(ctx TransactionContext) error {
					fmt.Println("Transaction, World!")
					return nil
				},
				func(ctx CompensationContext) error {
					fmt.Println("Compensation, World!")
					return nil
				},
			).
			AddStep(&sagaStep{}).
			Build()
		if err != nil {
			return err
		}

		if err := ctx.Saga("saga", def).Get(); err != nil {
			return err
		}

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

func TestUnitPrepareRootWorkflowEntitySagaCompensate(t *testing.T) {
	ctx := context.Background()
	db := NewMemoryDatabase()
	registry := NewRegistry()

	defer db.SaveAsJSON("./json/workflow_entity_saga_compensate.json")

	o := NewOrchestrator(ctx, db, registry)

	wrfl := func(ctx WorkflowContext) error {
		fmt.Println("Hello, World!")

		def, err := NewSaga().
			Add(
				func(ctx TransactionContext) error {
					fmt.Println("Transaction, World!")
					return nil
				},
				func(ctx CompensationContext) error {
					fmt.Println("Compensation, World!")
					return nil
				},
			).
			Add(
				func(tc TransactionContext) error {
					fmt.Println("Second Transaction, World!")
					return nil
				},
				func(cc CompensationContext) error {
					fmt.Println("Second Compensation, World!")
					return nil
				},
			).
			Add(
				func(ctx TransactionContext) error {
					val := 42
					if err := ctx.Store("something", &val); err != nil {
						return err
					}
					val = 420
					if err := ctx.Store("something2", &val); err != nil {
						return err
					}
					fmt.Println("Third Transaction, World!")
					return nil
				},
				func(ctx CompensationContext) error {
					var value int
					if err := ctx.Load("something", &value); err != nil {
						return err
					}
					fmt.Println("Third Compensation, World!", value)
					return nil
				},
			).
			AddStep(&sagaStep{fail: true}).
			Add(
				func(tc TransactionContext) error {
					fmt.Println("Fifth Transaction, World!")
					return nil
				},
				func(cc CompensationContext) error {
					fmt.Println("Fifth Compensation, World!")
					return nil
				},
			).
			Build()
		if err != nil {
			return err
		}

		if err := ctx.Saga("saga", def).Get(); err != nil {
			return err
		}

		return nil
	}

	future := o.Execute(wrfl, nil)

	if future == nil {
		t.Fatal("future is nil")
	}

	if err := future.Get(); err != nil {
		if !errors.Is(err, ErrSagaFailed) && !errors.Is(err, ErrSagaCompensated) {
			t.Fatal(err)
		}
	}

	// Check workflow status
	workflow, err := db.GetWorkflowEntity(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if workflow.Status != StatusFailed {
		t.Fatalf("expected workflow status %s, got %s", StatusFailed, workflow.Status)
	}

	// Get saga entity
	sagaEntities, err := db.GetSagaEntities(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if len(sagaEntities) != 1 {
		t.Fatalf("expected 1 saga entity, got %d", len(sagaEntities))
	}

	sagaEntity := sagaEntities[0]
	if sagaEntity.Status != StatusCompensated {
		t.Fatalf("expected saga status %s, got %s", StatusCompensated, sagaEntity.Status)
	}

	// Check saga executions
	sagaExecutions, err := db.GetSagaExecutions(sagaEntity.ID)
	if err != nil {
		t.Fatal(err)
	}

	// Should have 7 executions total: 4 transactions (last one failed) + 3 compensations
	if len(sagaExecutions) != 7 {
		t.Fatalf("expected 7 saga executions, got %d", len(sagaExecutions))
	}

	// Check transaction executions
	transactionExecutions := make([]*SagaExecution, 0)
	compensationExecutions := make([]*SagaExecution, 0)
	for _, exec := range sagaExecutions {
		if exec.ExecutionType == ExecutionTypeTransaction {
			transactionExecutions = append(transactionExecutions, exec)
		} else {
			compensationExecutions = append(compensationExecutions, exec)
		}
	}

	// Should have 4 transaction executions
	if len(transactionExecutions) != 4 {
		t.Fatalf("expected 4 transaction executions, got %d", len(transactionExecutions))
	}

	// Check failed transaction
	failedTx := transactionExecutions[3] // The fourth transaction should be the failed one
	if failedTx.Status != ExecutionStatusFailed || failedTx.Error != "on purpose" {
		t.Fatalf("expected failed transaction with 'on purpose' error, got status %s with error %s", failedTx.Status, failedTx.Error)
	}

	// Should have 3 compensation executions (for the first 3 successful transactions)
	if len(compensationExecutions) != 3 {
		t.Fatalf("expected 3 compensation executions, got %d", len(compensationExecutions))
	}

	// All compensations should be completed
	for i, comp := range compensationExecutions {
		if comp.Status != ExecutionStatusCompleted {
			t.Fatalf("compensation execution %d: expected status %s, got %s", i, ExecutionStatusCompleted, comp.Status)
		}
	}

	// Check saga values
	val, err := db.GetSagaValueByExecutionID(3, "something")
	if err != nil {
		t.Fatal(err)
	}
	if len(val) == 0 {
		t.Fatal("expected saga value 'something' to exist")
	}

	val, err = db.GetSagaValueByExecutionID(3, "something2")
	if err != nil {
		t.Fatal(err)
	}
	if len(val) == 0 {
		t.Fatal("expected saga value 'something2' to exist")
	}
}

func TestUnitSagaTransactionPanic(t *testing.T) {
	ctx := context.Background()
	db := NewMemoryDatabase()
	registry := NewRegistry()

	defer db.SaveAsJSON("./json/workflow_entity_saga_transaction_panic.json")

	o := NewOrchestrator(ctx, db, registry)

	wrfl := func(ctx WorkflowContext) error {
		def, err := NewSaga().
			Add(
				func(ctx TransactionContext) error {
					fmt.Println("First Transaction")
					return nil
				},
				func(ctx CompensationContext) error {
					fmt.Println("First Compensation")
					return nil
				},
			).
			Add(
				func(ctx TransactionContext) error {
					fmt.Println("Second Transaction")
					return nil
				},
				func(ctx CompensationContext) error {
					fmt.Println("Second Compensation")
					return nil
				},
			).
			Add(
				func(ctx TransactionContext) error {
					panic("panic during transaction")
				},
				func(ctx CompensationContext) error {
					fmt.Println("Third Compensation")
					return nil
				},
			).Build()
		if err != nil {
			return err
		}

		if err := ctx.Saga("saga", def).Get(); err != nil {
			return err
		}

		return nil
	}

	future := o.Execute(wrfl, nil)

	if future == nil {
		t.Fatal("future is nil")
	}

	if err := future.Get(); err != nil {
		if !errors.Is(err, ErrSagaFailed) && !errors.Is(err, ErrSagaCompensated) {
			t.Fatal(err)
		}
	}
}

func TestUnitSagaCompensationPanic(t *testing.T) {
	ctx := context.Background()
	db := NewMemoryDatabase()
	registry := NewRegistry()

	defer db.SaveAsJSON("./json/workflow_entity_saga_compensation_panic.json")

	o := NewOrchestrator(ctx, db, registry)

	wrfl := func(ctx WorkflowContext) error {
		def, err := NewSaga().
			Add(
				func(ctx TransactionContext) error {
					fmt.Println("First Transaction")
					return nil
				},
				func(ctx CompensationContext) error {
					fmt.Println("First Compensation")
					return nil
				},
			).
			Add(
				func(ctx TransactionContext) error {
					fmt.Println("Second Transaction")
					return nil
				},
				func(ctx CompensationContext) error {
					fmt.Println("Second Compensation")
					panic("panic during compensation")
				},
			).
			Add(
				func(ctx TransactionContext) error {
					fmt.Println("third transaction error")
					return fmt.Errorf("trigger compensation")
				},
				func(ctx CompensationContext) error {
					fmt.Println("Third Compensation")
					return nil
				},
			).Build()
		if err != nil {
			return err
		}

		if err := ctx.Saga("saga", def).Get(); err != nil {
			return err
		}

		return nil
	}

	future := o.Execute(wrfl, nil)

	if future == nil {
		t.Fatal("future is nil")
	}

	// Check error propagation
	if err := future.Get(); err != nil {
		if !errors.Is(err, ErrSagaFailed) {
			t.Fatal(err)
		}
	}

	// Check workflow status
	workflow, err := db.GetWorkflowEntity(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}
	if workflow.Status != StatusFailed {
		t.Errorf("expected workflow status Failed, got %s", workflow.Status)
	}

	// Get saga entity
	sagaEntities, err := db.GetSagaEntities(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}
	if len(sagaEntities) != 1 {
		t.Fatalf("expected 1 saga entity, got %d", len(sagaEntities))
	}
	saga := sagaEntities[0]
	if saga.Status != StatusFailed {
		t.Errorf("expected saga status Failed, got %s", saga.Status)
	}

	// Check saga executions
	sagaExecutions, err := db.GetSagaExecutions(saga.ID)
	if err != nil {
		t.Fatal(err)
	}

	// Should have 4 executions total:
	// - 3 transactions (last one failed)
	// - 1 compensation (panicked)
	if len(sagaExecutions) != 4 {
		t.Fatalf("expected 4 saga executions, got %d", len(sagaExecutions))
	}

	// Check transaction executions
	var transactionExecutions []*SagaExecution
	var compensationExecutions []*SagaExecution
	for _, exec := range sagaExecutions {
		if exec.ExecutionType == ExecutionTypeTransaction {
			transactionExecutions = append(transactionExecutions, exec)
		} else {
			compensationExecutions = append(compensationExecutions, exec)
		}
	}

	// Verify transactions
	if len(transactionExecutions) != 3 {
		t.Fatalf("expected 3 transaction executions, got %d", len(transactionExecutions))
	}
	// First two transactions should be completed
	for i := 0; i < 2; i++ {
		if transactionExecutions[i].Status != ExecutionStatusCompleted {
			t.Errorf("transaction %d: expected status Completed, got %s", i, transactionExecutions[i].Status)
		}
	}
	// Last transaction should be failed with our trigger error
	lastTx := transactionExecutions[2]
	if lastTx.Status != ExecutionStatusFailed || lastTx.Error != "trigger compensation" {
		t.Errorf("last transaction: expected Failed status with 'trigger compensation' error, got %s with error %s", lastTx.Status, lastTx.Error)
	}

	// Verify compensation
	if len(compensationExecutions) != 1 {
		t.Fatalf("expected 1 compensation execution, got %d", len(compensationExecutions))
	}
	compExec := compensationExecutions[0]
	if compExec.Status != ExecutionStatusFailed {
		t.Errorf("compensation: expected status Failed, got %s", compExec.Status)
	}
	if compExec.StackTrace == nil {
		t.Error("compensation: expected stack trace for panic, got nil")
	}
	if !strings.Contains(compExec.Error, "panic: panic during compensation") {
		t.Errorf("compensation: expected panic error message, got %s", compExec.Error)
	}
}

func TestUnitSequentialSagasWithManualCompensation(t *testing.T) {
	ctx := context.Background()
	db := NewMemoryDatabase()
	registry := NewRegistry()

	defer db.SaveAsJSON("./json/workflow_entity_sequential_sagas_manual_compensation.json")

	o := NewOrchestrator(ctx, db, registry)

	var saga1ID int

	wrfl := func(ctx WorkflowContext) error {
		// First saga - should succeed
		saga1, err := NewSaga().
			Add(
				func(tc TransactionContext) error {
					fmt.Println("Saga1: First Transaction")
					val := 42
					if err := tc.Store("saga1_value", &val); err != nil {
						return err
					}
					return nil
				},
				func(cc CompensationContext) error {
					var value int
					if err := cc.Load("saga1_value", &value); err != nil {
						return err
					}
					fmt.Println("Saga1: First Compensation", value)
					return nil
				},
			).
			Add(
				func(tc TransactionContext) error {
					fmt.Println("Saga1: Second Transaction")
					return nil
				},
				func(cc CompensationContext) error {
					fmt.Println("Saga1: Second Compensation")
					return nil
				},
			).Build()
		if err != nil {
			return err
		}

		future := ctx.Saga("saga1", saga1)
		saga1ID = future.WorkflowID()
		if err := future.Get(); err != nil {
			return err
		}

		// Second saga - should fail
		saga2, err := NewSaga().
			Add(
				func(tc TransactionContext) error {
					fmt.Println("Saga2: First Transaction")
					return nil
				},
				func(cc CompensationContext) error {
					fmt.Println("Saga2: First Compensation")
					return nil
				},
			).
			Add(
				func(tc TransactionContext) error {
					err := fmt.Errorf("saga2 deliberate failure")
					// Make sure this error propagates up
					fmt.Printf("Saga2: Failed with error: %v\n", err)
					return err
				},
				func(cc CompensationContext) error {
					fmt.Println("Saga2: Second Compensation")
					return nil
				},
			).Build()
		if err != nil {
			return err
		}

		// Run saga2 and handle its error
		saga2Future := ctx.Saga("saga2", saga2)
		if saga2Err := saga2Future.Get(); saga2Err != nil {
			// When saga2 fails, manually compensate saga1
			fmt.Printf("Saga2 failed, compensating saga1. Error: %v\n", saga2Err)
			if compErr := ctx.CompensateSaga("saga1"); compErr != nil {
				return fmt.Errorf("compensation chain failed: %w (original error: %w)", compErr, saga2Err)
			}
			// Make sure we return the original saga2 error to fail the workflow
			return saga2Err
		}

		return nil
	}

	future := o.Execute(wrfl, nil)
	if future == nil {
		t.Fatal("future is nil")
	}

	err := future.Get()
	if err == nil {
		t.Fatal("expected error from saga2 failure")
	}

	if !strings.Contains(err.Error(), "saga2 deliberate failure") {
		t.Errorf("unexpected error: %v", err)
	}

	// Verify workflow final state is Failed
	workflow, err := db.GetWorkflowEntity(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}
	if workflow.Status != StatusFailed {
		t.Errorf("expected workflow status Failed, got %s", workflow.Status)
	}

	// Verify saga1 final state
	saga1, err := db.GetSagaEntity(saga1ID)
	if err != nil {
		t.Fatal(err)
	}
	if saga1.Status != StatusCompensated {
		t.Errorf("saga1: expected status %s, got %s", StatusCompensated, saga1.Status)
	}

	// Get saga2
	sagaEntities, err := db.GetSagaEntities(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}
	var saga2 *SagaEntity
	for _, s := range sagaEntities {
		if s.StepID == "saga2" {
			saga2 = s
			break
		}
	}
	if saga2 == nil {
		t.Fatal("saga2 not found")
	}
	if saga2.Status != StatusCompensated {
		t.Errorf("saga2: expected status %s, got %s", StatusCompensated, saga2.Status)
	}

	// Verify execution counts
	saga1Execs, err := db.GetSagaExecutions(saga1.ID)
	if err != nil {
		t.Fatal(err)
	}

	var txCount, compCount int
	for _, exec := range saga1Execs {
		if exec.ExecutionType == ExecutionTypeTransaction {
			txCount++
			if exec.Status != ExecutionStatusCompleted {
				t.Errorf("saga1 transaction %d: expected status %s, got %s", exec.ID, ExecutionStatusCompleted, exec.Status)
			}
		} else {
			compCount++
			if exec.Status != ExecutionStatusCompleted {
				t.Errorf("saga1 compensation %d: expected status %s, got %s", exec.ID, ExecutionStatusCompleted, exec.Status)
			}
		}
	}

	if txCount != 2 {
		t.Errorf("saga1: expected 2 transactions, got %d", txCount)
	}
	if compCount != 2 {
		t.Errorf("saga1: expected 2 compensations, got %d", compCount)
	}

	// Verify saga value was accessible
	val, err := db.GetSagaValueByExecutionID(saga1.ID, "saga1_value")
	if err != nil {
		t.Fatal(err)
	}
	if len(val) == 0 {
		t.Error("saga1_value not found")
	}
}

func TestUnitSagaSkipOnWorkflowRetry(t *testing.T) {
	ctx := context.Background()
	db := NewMemoryDatabase()
	registry := NewRegistry()

	defer db.SaveAsJSON("./json/workflow_entity_saga_skip_on_retry.json")

	o := NewOrchestrator(ctx, db, registry)

	// Use atomic values to track calls
	var sagaCallCount atomic.Int64
	var workflowAttempts atomic.Int64

	wrfl := func(ctx WorkflowContext) error {
		workflowAttempts.Add(1)

		// First run the saga
		saga, err := NewSaga().
			Add(
				func(tc TransactionContext) error {
					sagaCallCount.Add(1)
					fmt.Println("Saga Transaction called, count:", sagaCallCount.Load())
					val := 42
					if err := tc.Store("test_value", &val); err != nil {
						return err
					}
					return nil
				},
				func(cc CompensationContext) error {
					fmt.Println("Saga Compensation")
					return nil
				},
			).Build()
		if err != nil {
			return err
		}

		// Execute saga - should succeed
		sagaFuture := ctx.Saga("test_saga", saga)
		if err := sagaFuture.Get(); err != nil {
			return err
		}

		// After saga succeeds, fail the workflow on first attempt
		attempt := workflowAttempts.Load()
		if attempt == 1 {
			return fmt.Errorf("deliberate workflow failure on attempt %d", attempt)
		}

		// Second attempt should succeed
		return nil
	}

	// Configure workflow with retry policy
	options := &WorkflowOptions{
		RetryPolicy: &RetryPolicy{
			MaxAttempts: 1, // Allow one retry
			MaxInterval: 100 * time.Millisecond,
		},
	}

	future := o.Execute(wrfl, options)
	if future == nil {
		t.Fatal("future is nil")
	}

	// Wait for completion
	err := future.Get()
	if err != nil {
		t.Fatalf("unexpected workflow error: %v", err)
	}

	// Verify workflow attempts
	attempts := workflowAttempts.Load()
	if attempts != 2 {
		t.Errorf("expected 2 workflow attempts, got %d", attempts)
	}

	// Verify saga was called exactly once
	sagaCalls := sagaCallCount.Load()
	if sagaCalls != 1 {
		t.Errorf("expected saga to be called exactly once, got %d calls", sagaCalls)
	}

	// Verify workflow final state
	workflow, err := db.GetWorkflowEntity(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}
	if workflow.Status != StatusCompleted {
		t.Errorf("expected workflow status Completed, got %s", workflow.Status)
	}

	// Get saga entity
	sagaEntities, err := db.GetSagaEntities(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}
	if len(sagaEntities) != 1 {
		t.Fatalf("expected 1 saga entity, got %d", len(sagaEntities))
	}

	saga := sagaEntities[0]
	if saga.Status != StatusCompleted {
		t.Errorf("expected saga status Completed, got %s", saga.Status)
	}

	// Verify saga executions - should have exactly 1 transaction
	sagaExecutions, err := db.GetSagaExecutions(saga.ID)
	if err != nil {
		t.Fatal(err)
	}

	var txCount int
	for _, exec := range sagaExecutions {
		if exec.ExecutionType == ExecutionTypeTransaction {
			txCount++
			if exec.Status != ExecutionStatusCompleted {
				t.Errorf("saga transaction: expected status %s, got %s",
					ExecutionStatusCompleted, exec.Status)
			}
		}
	}

	if txCount != 1 {
		t.Errorf("expected exactly 1 saga transaction, got %d", txCount)
	}

	// Verify saga value was stored and preserved across retry
	val, err := db.GetSagaValueByExecutionID(saga.ID, "test_value")
	if err != nil {
		t.Fatal(err)
	}
	if len(val) == 0 {
		t.Error("test_value not found")
	}
}

func TestUnitPrepareRootWorkflowSubWorkflowEntity(t *testing.T) {
	ctx := context.Background()
	db := NewMemoryDatabase()
	registry := NewRegistry()

	defer db.SaveAsJSON("./json/workflow_subworkflow_entity.json")

	o := NewOrchestrator(ctx, db, registry)

	subwrfl := func(ctx WorkflowContext) error {
		fmt.Println("Hello,sub-World!")
		return nil
	}

	wrfl := func(ctx WorkflowContext) error {
		fmt.Println("Hello, World!")
		if err := ctx.Workflow("sub", subwrfl, nil).Get(); err != nil {
			return err
		}
		return nil
	}

	future := o.Execute(wrfl, nil)

	if future == nil {
		t.Fatal("future is nil")
	}

	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

	// Get the root workflow entity
	rootWorkflow, err := db.GetWorkflowEntity(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	// Verify root workflow properties
	if rootWorkflow.Status != StatusCompleted {
		t.Fatalf("expected root workflow status %s, got %s", StatusCompleted, rootWorkflow.Status)
	}
	if !rootWorkflow.WorkflowData.IsRoot {
		t.Fatal("expected root workflow IsRoot to be true")
	}

	// Get root workflow executions
	rootExecs, err := db.GetWorkflowExecutions(rootWorkflow.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(rootExecs) != 1 {
		t.Fatalf("expected 1 root execution, got %d", len(rootExecs))
	}
	if rootExecs[0].Status != ExecutionStatusCompleted {
		t.Fatalf("expected root execution status %s, got %s", ExecutionStatusCompleted, rootExecs[0].Status)
	}

	// Get hierarchies to find subworkflow
	hierarchies, err := db.GetHierarchiesByParentEntityAndStep(rootWorkflow.ID, "sub", EntityWorkflow)
	if err != nil {
		t.Fatal(err)
	}
	if len(hierarchies) != 1 {
		t.Fatalf("expected 1 hierarchy, got %d", len(hierarchies))
	}

	hierarchy := hierarchies[0]
	// Verify hierarchy properties
	if hierarchy.ParentType != EntityWorkflow {
		t.Fatalf("expected parent type %s, got %s", EntityWorkflow, hierarchy.ParentType)
	}
	if hierarchy.ChildType != EntityWorkflow {
		t.Fatalf("expected child type %s, got %s", EntityWorkflow, hierarchy.ChildType)
	}
	if hierarchy.ParentEntityID != rootWorkflow.ID {
		t.Fatalf("expected parent entity ID %d, got %d", rootWorkflow.ID, hierarchy.ParentEntityID)
	}

	// Get and verify subworkflow
	subWorkflow, err := db.GetWorkflowEntity(hierarchy.ChildEntityID)
	if err != nil {
		t.Fatal(err)
	}
	if subWorkflow.Status != StatusCompleted {
		t.Fatalf("expected sub-workflow status %s, got %s", StatusCompleted, subWorkflow.Status)
	}
	if subWorkflow.WorkflowData.IsRoot {
		t.Fatal("expected sub-workflow IsRoot to be false")
	}
	if subWorkflow.StepID != "sub" {
		t.Fatalf("expected sub-workflow stepID 'sub', got %s", subWorkflow.StepID)
	}

	// Get and verify subworkflow executions
	subExecs, err := db.GetWorkflowExecutions(subWorkflow.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(subExecs) != 1 {
		t.Fatalf("expected 1 sub-workflow execution, got %d", len(subExecs))
	}
	if subExecs[0].Status != ExecutionStatusCompleted {
		t.Fatalf("expected sub-workflow execution status %s, got %s", ExecutionStatusCompleted, subExecs[0].Status)
	}

	// Verify execution IDs in hierarchy
	if hierarchy.ParentExecutionID != rootExecs[0].ID {
		t.Fatalf("expected hierarchy parent execution ID %d, got %d", rootExecs[0].ID, hierarchy.ParentExecutionID)
	}
	if hierarchy.ChildExecutionID != subExecs[0].ID {
		t.Fatalf("expected hierarchy child execution ID %d, got %d", subExecs[0].ID, hierarchy.ChildExecutionID)
	}
}

func TestUnitSubWorkflowError(t *testing.T) {
	ctx := context.Background()
	db := NewMemoryDatabase()
	registry := NewRegistry()

	defer db.SaveAsJSON("./json/workflow_subworkflow_error.json")

	o := NewOrchestrator(ctx, db, registry)

	expectedErr := errors.New("sub-workflow error")

	subwrfl := func(ctx WorkflowContext) error {
		fmt.Println("Hello from failing sub-workflow!")
		return expectedErr
	}

	wrfl := func(ctx WorkflowContext) error {
		fmt.Println("Hello from root workflow!")
		if err := ctx.Workflow("sub", subwrfl, nil).Get(); err != nil {
			// Should propagate error up
			return fmt.Errorf("sub-workflow failed: %w", err)
		}
		return nil
	}

	future := o.Execute(wrfl, nil)

	if future == nil {
		t.Fatal("future is nil")
	}

	// Get() should return error from root workflow
	err := future.Get()
	if err == nil {
		t.Fatal("expected error from future.Get(), got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error to contain %v, got %v", expectedErr, err)
	}

	// Get the root workflow entity
	rootWorkflow, err := db.GetWorkflowEntity(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	// Root workflow should be failed since sub-workflow failed
	if rootWorkflow.Status != StatusFailed {
		t.Fatalf("expected root workflow status %s, got %s", StatusFailed, rootWorkflow.Status)
	}
	if !rootWorkflow.WorkflowData.IsRoot {
		t.Fatal("expected root workflow IsRoot to be true")
	}

	// Get root workflow executions
	rootExecs, err := db.GetWorkflowExecutions(rootWorkflow.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(rootExecs) != 1 {
		t.Fatalf("expected 1 root execution, got %d", len(rootExecs))
	}
	if rootExecs[0].Status != ExecutionStatusFailed {
		t.Fatalf("expected root execution status %s, got %s", ExecutionStatusFailed, rootExecs[0].Status)
	}
	// Root execution should contain error message
	if rootExecs[0].Error == "" {
		t.Fatal("expected root execution to have error message")
	}

	// Get hierarchies to find subworkflow
	hierarchies, err := db.GetHierarchiesByParentEntityAndStep(rootWorkflow.ID, "sub", EntityWorkflow)
	if err != nil {
		t.Fatal(err)
	}
	if len(hierarchies) != 1 {
		t.Fatalf("expected 1 hierarchy, got %d", len(hierarchies))
	}

	hierarchy := hierarchies[0]
	// Verify hierarchy still maintains correct types
	if hierarchy.ParentType != EntityWorkflow {
		t.Fatalf("expected parent type %s, got %s", EntityWorkflow, hierarchy.ParentType)
	}
	if hierarchy.ChildType != EntityWorkflow {
		t.Fatalf("expected child type %s, got %s", EntityWorkflow, hierarchy.ChildType)
	}

	// Get and verify subworkflow status
	subWorkflow, err := db.GetWorkflowEntity(hierarchy.ChildEntityID)
	if err != nil {
		t.Fatal(err)
	}
	if subWorkflow.Status != StatusFailed {
		t.Fatalf("expected sub-workflow status %s, got %s", StatusFailed, subWorkflow.Status)
	}
	if subWorkflow.WorkflowData.IsRoot {
		t.Fatal("expected sub-workflow IsRoot to be false")
	}
	if subWorkflow.StepID != "sub" {
		t.Fatalf("expected sub-workflow stepID 'sub', got %s", subWorkflow.StepID)
	}

	// Get and verify subworkflow executions
	subExecs, err := db.GetWorkflowExecutions(subWorkflow.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(subExecs) != 1 {
		t.Fatalf("expected 1 sub-workflow execution, got %d", len(subExecs))
	}
	if subExecs[0].Status != ExecutionStatusFailed {
		t.Fatalf("expected sub-workflow execution status %s, got %s", ExecutionStatusFailed, subExecs[0].Status)
	}
	// Sub-workflow execution should contain original error
	if subExecs[0].Error != expectedErr.Error() {
		t.Fatalf("expected sub-workflow execution error %q, got %q", expectedErr.Error(), subExecs[0].Error)
	}

	// Verify execution IDs in hierarchy
	if hierarchy.ParentExecutionID != rootExecs[0].ID {
		t.Fatalf("expected hierarchy parent execution ID %d, got %d", rootExecs[0].ID, hierarchy.ParentExecutionID)
	}
	if hierarchy.ChildExecutionID != subExecs[0].ID {
		t.Fatalf("expected hierarchy child execution ID %d, got %d", subExecs[0].ID, hierarchy.ChildExecutionID)
	}
}

func TestUnitSubWorkflowRetrySuccess(t *testing.T) {
	ctx := context.Background()
	db := NewMemoryDatabase()
	registry := NewRegistry()

	defer db.SaveAsJSON("./json/workflow_subworkflow_retry_success.json")

	o := NewOrchestrator(ctx, db, registry)

	// Counter to track retry attempts
	attempts := 0
	subwrfl := func(ctx WorkflowContext) error {
		attempts++
		fmt.Printf("Sub-workflow attempt %d\n", attempts)
		if attempts == 1 {
			return errors.New("first attempt fails")
		}
		// Second attempt succeeds
		return nil
	}

	wrfl := func(ctx WorkflowContext) error {
		fmt.Println("Hello from root workflow!")
		// Configure retry policy for the subworkflow
		retryPolicy := &RetryPolicy{
			MaxAttempts: 3,               // Allow up to 3 attempts
			MaxInterval: time.Second * 1, // 1 second between retries
		}
		if err := ctx.Workflow("sub", subwrfl, &WorkflowOptions{
			RetryPolicy: retryPolicy,
		}).Get(); err != nil {
			return fmt.Errorf("sub-workflow failed: %w", err)
		}
		return nil
	}

	future := o.Execute(wrfl, nil)

	if future == nil {
		t.Fatal("future is nil")
	}

	// Should succeed after retry
	if err := future.Get(); err != nil {
		t.Fatalf("expected success after retry, got error: %v", err)
	}

	// Verify attempts
	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}

	// Get the root workflow entity
	rootWorkflow, err := db.GetWorkflowEntity(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	// Root workflow should be completed
	if rootWorkflow.Status != StatusCompleted {
		t.Fatalf("expected root workflow status %s, got %s", StatusCompleted, rootWorkflow.Status)
	}

	// Get hierarchies to find subworkflow
	hierarchies, err := db.GetHierarchiesByParentEntityAndStep(rootWorkflow.ID, "sub", EntityWorkflow)
	if err != nil {
		t.Fatal(err)
	}
	if len(hierarchies) != 1 {
		t.Fatalf("expected 1 hierarchy, got %d", len(hierarchies))
	}

	hierarchy := hierarchies[0]

	// Get and verify subworkflow
	subWorkflow, err := db.GetWorkflowEntity(hierarchy.ChildEntityID)
	if err != nil {
		t.Fatal(err)
	}

	// Final state should be completed
	if subWorkflow.Status != StatusCompleted {
		t.Fatalf("expected sub-workflow status %s, got %s", StatusCompleted, subWorkflow.Status)
	}

	// Verify retry state
	if subWorkflow.RetryState.Attempts != 1 { // Attempts counter starts at 0
		t.Fatalf("expected retry attempts 1, got %d", subWorkflow.RetryState.Attempts)
	}

	// Get and verify all subworkflow executions
	subExecs, err := db.GetWorkflowExecutions(subWorkflow.ID)
	if err != nil {
		t.Fatal(err)
	}

	// Should have 2 executions (first failed, second succeeded)
	if len(subExecs) != 2 {
		t.Fatalf("expected 2 sub-workflow executions, got %d", len(subExecs))
	}

	// Sort executions by creation time to verify sequence
	sort.Slice(subExecs, func(i, j int) bool {
		return subExecs[i].CreatedAt.Before(subExecs[j].CreatedAt)
	})

	// Verify first execution failed
	if subExecs[0].Status != ExecutionStatusFailed {
		t.Fatalf("expected first execution status %s, got %s", ExecutionStatusFailed, subExecs[0].Status)
	}
	if subExecs[0].Error == "" {
		t.Fatal("expected first execution to have error message")
	}

	// Verify second execution succeeded
	if subExecs[1].Status != ExecutionStatusCompleted {
		t.Fatalf("expected second execution status %s, got %s", ExecutionStatusCompleted, subExecs[1].Status)
	}
	if subExecs[1].Error != "" {
		t.Fatalf("expected second execution to have no error, got: %s", subExecs[1].Error)
	}

	// Verify final execution ID in hierarchy matches successful execution
	if hierarchy.ChildExecutionID != subExecs[1].ID {
		t.Fatalf("expected hierarchy to reference successful execution ID %d, got %d", subExecs[1].ID, hierarchy.ChildExecutionID)
	}

	// Get root executions to verify they're still properly linked
	rootExecs, err := db.GetWorkflowExecutions(rootWorkflow.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(rootExecs) != 1 {
		t.Fatalf("expected 1 root execution, got %d", len(rootExecs))
	}
	if rootExecs[0].Status != ExecutionStatusCompleted {
		t.Fatalf("expected root execution status %s, got %s", ExecutionStatusCompleted, rootExecs[0].Status)
	}

	// Verify parent/child execution relationship maintained
	if hierarchy.ParentExecutionID != rootExecs[0].ID {
		t.Fatalf("expected hierarchy parent execution ID %d, got %d", rootExecs[0].ID, hierarchy.ParentExecutionID)
	}
}

func TestUnitSubWorkflowPanic(t *testing.T) {
	ctx := context.Background()
	db := NewMemoryDatabase()
	registry := NewRegistry()

	defer db.SaveAsJSON("./json/workflow_subworkflow_panic.json")

	o := NewOrchestrator(ctx, db, registry)

	subwrfl := func(ctx WorkflowContext) error {
		fmt.Println("About to panic in sub-workflow!")
		panic("sub-workflow panic")
	}

	wrfl := func(ctx WorkflowContext) error {
		fmt.Println("Hello from root workflow!")
		if err := ctx.Workflow("sub", subwrfl, nil).Get(); err != nil {
			return fmt.Errorf("sub-workflow failed: %w", err)
		}
		return nil
	}

	future := o.Execute(wrfl, nil)

	if future == nil {
		t.Fatal("future is nil")
	}

	// Get() should return panic as error
	err := future.Get()
	if err == nil {
		t.Fatal("expected error from future.Get(), got nil")
	}
	if !errors.Is(err, ErrWorkflowPanicked) {
		t.Fatalf("expected error to be ErrWorkflowPanicked, got %v", err)
	}

	// Get the root workflow entity
	rootWorkflow, err := db.GetWorkflowEntity(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	// Root workflow should be failed
	if rootWorkflow.Status != StatusFailed {
		t.Fatalf("expected root workflow status %s, got %s", StatusFailed, rootWorkflow.Status)
	}

	// Get hierarchies to find subworkflow
	hierarchies, err := db.GetHierarchiesByParentEntityAndStep(rootWorkflow.ID, "sub", EntityWorkflow)
	if err != nil {
		t.Fatal(err)
	}
	if len(hierarchies) != 1 {
		t.Fatalf("expected 1 hierarchy, got %d", len(hierarchies))
	}

	// Get and verify subworkflow
	subWorkflow, err := db.GetWorkflowEntity(hierarchies[0].ChildEntityID)
	if err != nil {
		t.Fatal(err)
	}

	if subWorkflow.Status != StatusFailed {
		t.Fatalf("expected sub-workflow status %s, got %s", StatusFailed, subWorkflow.Status)
	}

	// Get and verify subworkflow executions
	subExecs, err := db.GetWorkflowExecutions(subWorkflow.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(subExecs) != 1 {
		t.Fatalf("expected 1 sub-workflow execution, got %d", len(subExecs))
	}

	// Verify execution details
	if subExecs[0].Status != ExecutionStatusFailed {
		t.Fatalf("expected execution status %s, got %s", ExecutionStatusFailed, subExecs[0].Status)
	}
	if subExecs[0].StackTrace == nil {
		t.Fatal("expected stack trace to be present")
	}
}

func TestUnitSubWorkflowPanicWithRetrySuccess(t *testing.T) {
	ctx := context.Background()
	db := NewMemoryDatabase()
	registry := NewRegistry()

	defer db.SaveAsJSON("./json/workflow_subworkflow_panic_retry.json")

	o := NewOrchestrator(ctx, db, registry)

	attempts := 0
	subwrfl := func(ctx WorkflowContext) error {
		attempts++
		fmt.Printf("Sub-workflow attempt %d\n", attempts)
		if attempts == 1 {
			panic("first attempt panics")
		}
		// Second attempt succeeds
		return nil
	}

	wrfl := func(ctx WorkflowContext) error {
		fmt.Println("Hello from root workflow!")
		retryPolicy := &RetryPolicy{
			MaxAttempts: 3,
			MaxInterval: time.Second * 1,
		}
		if err := ctx.Workflow("sub", subwrfl, &WorkflowOptions{
			RetryPolicy: retryPolicy,
		}).Get(); err != nil {
			return fmt.Errorf("sub-workflow failed: %w", err)
		}
		return nil
	}

	future := o.Execute(wrfl, nil)

	if future == nil {
		t.Fatal("future is nil")
	}

	// Should succeed after retry
	if err := future.Get(); err != nil {
		t.Fatalf("expected success after retry, got error: %v", err)
	}

	// Verify attempts
	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}

	// Get the root workflow entity
	rootWorkflow, err := db.GetWorkflowEntity(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	// Root workflow should be completed
	if rootWorkflow.Status != StatusCompleted {
		t.Fatalf("expected root workflow status %s, got %s", StatusCompleted, rootWorkflow.Status)
	}

	// Get hierarchies to find subworkflow
	hierarchies, err := db.GetHierarchiesByParentEntityAndStep(rootWorkflow.ID, "sub", EntityWorkflow)
	if err != nil {
		t.Fatal(err)
	}

	// Get and verify subworkflow
	subWorkflow, err := db.GetWorkflowEntity(hierarchies[0].ChildEntityID)
	if err != nil {
		t.Fatal(err)
	}

	if subWorkflow.Status != StatusCompleted {
		t.Fatalf("expected sub-workflow status %s, got %s", StatusCompleted, subWorkflow.Status)
	}

	// Verify retry state
	if subWorkflow.RetryState.Attempts != 1 {
		t.Fatalf("expected retry attempts 1, got %d", subWorkflow.RetryState.Attempts)
	}

	// Get and verify all executions
	subExecs, err := db.GetWorkflowExecutions(subWorkflow.ID)
	if err != nil {
		t.Fatal(err)
	}

	if len(subExecs) != 2 {
		t.Fatalf("expected 2 sub-workflow executions, got %d", len(subExecs))
	}

	// Sort executions by creation time
	sort.Slice(subExecs, func(i, j int) bool {
		return subExecs[i].CreatedAt.Before(subExecs[j].CreatedAt)
	})

	// Verify first execution (panic)
	firstExec := subExecs[0]
	if firstExec.Status != ExecutionStatusFailed {
		t.Fatalf("expected first execution status %s, got %s", ExecutionStatusFailed, firstExec.Status)
	}
	if firstExec.StackTrace == nil {
		t.Fatal("expected stack trace in first execution")
	}

	// Verify second execution (success)
	secondExec := subExecs[1]
	if secondExec.Status != ExecutionStatusCompleted {
		t.Fatalf("expected second execution status %s, got %s", ExecutionStatusCompleted, secondExec.Status)
	}
	if secondExec.Error != "" {
		t.Fatalf("expected no error in successful execution, got: %s", secondExec.Error)
	}
	if secondExec.StackTrace != nil {
		t.Fatal("expected no stack trace in successful execution")
	}

	// Verify hierarchy points to successful execution
	if hierarchies[0].ChildExecutionID != secondExec.ID {
		t.Fatalf("expected hierarchy to reference successful execution ID %d, got %d",
			secondExec.ID, hierarchies[0].ChildExecutionID)
	}
}

func TestWorkflowSignal(t *testing.T) {

	ctx := context.Background()
	db := NewMemoryDatabase()
	registry := NewRegistry()

	defer db.SaveAsJSON("./json/workflow_signal.json")

	o := NewOrchestrator(
		ctx,
		db,
		registry,
		WithSignalNew(func(workflowID int, signal string) Future {
			fmt.Println("Signal received:", signal)
			<-time.After(100 * time.Millisecond)
			future := NewRuntimeFuture()
			future.setResult([]interface{}{42})
			fmt.Println("Signal processed")
			return future
		}))

	wrfl := func(ctx WorkflowContext) error {
		fmt.Println("Hello, World!")
		var life int
		if err := ctx.Signal("life", &life); err != nil {
			return err
		}
		fmt.Println("Life, the universe, and everything:", life)
		if err := ctx.Signal("life", &life); err != nil {
			return err
		}
		fmt.Println("again:", life)
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

func TestWorkflowSignalPanic(t *testing.T) {

	ctx := context.Background()
	db := NewMemoryDatabase()
	registry := NewRegistry()

	defer db.SaveAsJSON("./json/workflow_signal_panic.json")

	o := NewOrchestrator(
		ctx,
		db,
		registry,
		WithSignalNew(func(workflowID int, signal string) Future {
			fmt.Println("Signal received:", signal)
			<-time.After(100 * time.Millisecond)
			future := NewRuntimeFuture()
			future.setResult([]interface{}{42})
			fmt.Println("Signal processed")
			return future
		}))

	wrfl := func(ctx WorkflowContext) error {
		fmt.Println("Hello, World!")
		var life int
		if err := ctx.Signal("life", &life); err != nil {
			return err
		}
		panic("deliberate panic")
		fmt.Println("Life, the universe, and everything:", life)
		if err := ctx.Signal("life", &life); err != nil {
			return err
		}
		fmt.Println("again:", life)
		return nil
	}

	future := o.Execute(wrfl, nil)

	if future == nil {
		t.Fatal("future is nil")
	}

	if err := future.Get(); err != nil {
		if !errors.Is(err, ErrWorkflowPanicked) {
			t.Fatal(err)
		}
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

	if len(execs) != 1 {
		t.Fatalf("expected 1 execution, got %d", len(execs))
	}

	for _, v := range execs {
		if v.Status != ExecutionStatusFailed {
			t.Fatalf("expected %s, got %s", ExecutionStatusFailed, v.Status)
		}
	}

}

func TestVersion(t *testing.T) {
	ctx := context.Background()
	db := NewMemoryDatabase()
	registry := NewRegistry()

	defer db.SaveAsJSON("./json/workflow_version.json")

	o := NewOrchestrator(ctx, db, registry)

	versioningA := atomic.Int32{}
	versioningB := atomic.Int32{}

	wrfl := func(ctx WorkflowContext) error {
		fmt.Println("Hello, World!")
		featureAVersion, err := ctx.GetVersion("featureA", DefaultVersion, int(versioningA.Load()))
		if err != nil {
			return err
		}
		fmt.Println("Feature A version:", featureAVersion)
		switch versioningB.Load() {
		case 0:
			fmt.Println("Default real version")
		default:
			version, err := ctx.GetVersion("test", DefaultVersion, int(versioningB.Load()))
			if err != nil {
				return err
			}
			fmt.Println("Version:", version)
			switch version {
			case DefaultVersion:
				fmt.Println("Default version")
			case 1:
				fmt.Println("Version 1")
			case 2:
				fmt.Println("Version 2")
			default:
				return fmt.Errorf("unexpected version: %d", version)
			}
		}
		return nil
	}

	for i := 0; i < 3; i++ {
		future := o.Execute(wrfl, &WorkflowOptions{})
		if future == nil {
			t.Fatal("future is nil")
		}
		if err := future.Get(); err != nil {
			t.Fatal(err)
		}
		versioningB.Add(1)
		if i == 1 {
			versioningA.Add(1)
		}
	}

}

// Test the capabilities of the versioning system using the pause and resume feature
func TestVersionPauseResume(t *testing.T) {
	ctx := context.Background()
	db := NewMemoryDatabase()
	registry := NewRegistry()

	defer db.SaveAsJSON("./json/workflow_version_pause_resume.json")

	o := NewOrchestrator(ctx, db, registry)

	versioningA := atomic.Int32{}

	wrfl := func(ctx WorkflowContext) error {
		fmt.Println("Hello, World!")
		featureAVersion, err := ctx.GetVersion("featureA", DefaultVersion, int(versioningA.Load()))
		if err != nil {
			return err
		}
		<-time.After(1 * time.Second) // In theory, we should have the version 0
		if err := ctx.SideEffect("test", func() int {
			return 1
		}).Get(); err != nil {
			return err
		}
		fmt.Println("Feature A version:", featureAVersion)
		return nil
	}

	future := o.Execute(wrfl, &WorkflowOptions{})
	if future == nil {
		t.Fatal("future is nil")
	}

	o.Pause()

	<-time.After(2 * time.Second)
	versioningA.Add(1)
	fmt.Println("Resuming")

	future = o.Resume(future.WorkflowID())
	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

	version, err := db.GetVersionsByWorkflowID(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if len(version) != 1 {
		t.Fatalf("expected 1 version, got %d", len(version))
	}

	if version[0].Version != 0 {
		t.Fatalf("expected version 0, got %d", version[0].Version)
	}
}

func TestVersionPauseResumeOverride(t *testing.T) {
	ctx := context.Background()
	db := NewMemoryDatabase()
	registry := NewRegistry()

	defer db.SaveAsJSON("./json/workflow_version_pause_resume.json")

	o := NewOrchestrator(ctx, db, registry)

	versioningA := atomic.Int32{}

	wrfl := func(ctx WorkflowContext) error {
		fmt.Println("Hello, World!")
		featureAVersion, err := ctx.GetVersion("featureA", DefaultVersion, 2)
		if err != nil {
			return err
		}
		<-time.After(1 * time.Second) // In theory, we should have the version 0
		if err := ctx.SideEffect("test", func() int {
			return 1
		}).Get(); err != nil {
			return err
		}
		fmt.Println("Feature A version:", featureAVersion)
		return nil
	}

	future := o.Execute(wrfl, &WorkflowOptions{
		VersionOverrides: map[string]int{
			"featureA": 1,
		},
	})
	if future == nil {
		t.Fatal("future is nil")
	}

	o.Pause()

	<-time.After(2 * time.Second)
	versioningA.Add(1)
	fmt.Println("Resuming")

	future = o.Resume(future.WorkflowID())
	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

	version, err := db.GetVersionsByWorkflowID(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if len(version) != 1 {
		t.Fatalf("expected 1 version, got %d", len(version))
	}

	if version[0].Version != 1 {
		t.Fatalf("expected version 1, got %d", version[0].Version)
	}
}
