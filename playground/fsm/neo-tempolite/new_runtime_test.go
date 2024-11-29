package tempolite

import (
	"context"
	"fmt"
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

	if !futureFirst.ContinuedAsNew() {
		t.Fatalf("expected true, got false")
	}

	if firstRes != 42 {
		t.Fatalf("expected 42, got %d", firstRes)
	}

	// The orchestrator provide a function to allow you to execute any entity
	// A workflow continued as new is just a new workflow to execute
	futureSecond, err := o.ExecuteWithEntity(futureFirst.ContinuedAs())
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

	if futureSecond.ContinuedAsNew() {
		t.Fatalf("expected false, got true")
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

	o.WaitActive()

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

	o.WaitActive()

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

	o.WaitActive()

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

	if len(execs) != 2 {
		t.Fatalf("expected 2 execution, got %d", len(execs))
	}

	for _, v := range execs {
		if v.Status != ExecutionStatusFailed {
			t.Fatalf("expected %s, got %s", ExecutionStatusFailed, v.Status)
		}
	}

	sideeffects, err := db.GetSideEffectEntities(future.WorkflowID())
	if err != nil {
		t.Fatal(err)
	}

	if len(sideeffects) != 1 {
		t.Fatalf("expected 1 sideeffect, got %d", len(sideeffects))
	}

	for _, v := range sideeffects {
		if v.Status != StatusFailed {
			t.Fatalf("expected %s, got %s", StatusFailed, v.Status)
		}

		execs, err := db.GetSideEffectExecutions(v.ID)
		if err != nil {
			t.Fatal(err)
		}

		if len(execs) != 2 {
			t.Fatalf("expected 2 executions, got %d", len(execs))
		}

		for _, e := range execs {
			if e.Status != ExecutionStatusFailed {
				t.Fatalf("expected %s, got %s", ExecutionStatusFailed, e.Status)
			}
		}
	}

}
