package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/davidroman0O/tempolite"
	tempoliteContext "github.com/davidroman0O/tempolite/internal/engine/context"
	"github.com/davidroman0O/tempolite/internal/engine/registry"
	"github.com/davidroman0O/tempolite/internal/types"
	"github.com/davidroman0O/tempolite/pkg/logs"
)

// TODO: test content of the DB
func TestOneWorkflow(t *testing.T) {

	// Feature tested: we should be able to have workflows and return their results
	wrk := func(ctx tempoliteContext.WorkflowContext) (int, error) {
		return 42, nil
	}

	ctx := context.Background()

	tp, err := tempolite.New(
		ctx,
		registry.
			New().
			Workflow(wrk).
			Build(),
		tempolite.WithPath("./dbs/tempolite-one-workflow.db"),
		tempolite.WithDestructive(),
		tempolite.WithDefaultLogLevel(logs.LevelDebug),
	)

	if err != nil {
		t.Fatal(err)
	}

	info := tp.Workflow(wrk, nil)

	var realValue int

	fmt.Println("Info", info.Get(&realValue))

	if realValue != 42 {
		t.Fatal("Real value should be 42")
	}

	if err = tp.Shutdown(); err != nil {
		if err != context.Canceled {
			t.Fatal(err)
		}
	}
}

// TODO: test content of the DB
func TestSubWorkflowSuccess(t *testing.T) {
	subwrk := func(ctx tempoliteContext.WorkflowContext) (int, error) {
		return 69, nil
	}

	wrk := func(ctx tempoliteContext.WorkflowContext) (int, error) {
		var result int
		// Feature tested: we should be able to have a sub-workflow
		if err := ctx.Workflow("subworkflow", subwrk, nil).Get(&result); err != nil {
			return 0, err
		}

		return 420, nil
	}

	ctx := context.Background()

	tp, err := tempolite.New(
		ctx,
		registry.
			New().
			Workflow(subwrk).
			Workflow(wrk).
			Build(),
		tempolite.WithPath("./dbs/tempolite-subworkflow-success.db"),
		tempolite.WithDestructive(),
		tempolite.WithDefaultLogLevel(logs.LevelDebug),
	)

	if err != nil {
		t.Fatal(err)
	}

	// Because we have a sub-workflow
	tp.Scale("default", map[string]int{
		"workflows":   2,
		"activities":  2,
		"sideEffects": 2,
		"sagas":       2,
	})

	info := tp.Workflow(wrk, nil)

	var realValue int

	fmt.Println("Info", info.Get(&realValue))

	if realValue != 420 {
		t.Fatal("Real value should be 420")
	}

	if err = tp.Shutdown(); err != nil {
		if err != context.Canceled {
			t.Fatal(err)
		}
	}
}

// TODO: test content of the DB
func TestSubWorkflowWithParentFailureBeforeSubWorkflow(t *testing.T) {
	failure := true

	subwrk := func(ctx tempoliteContext.WorkflowContext) (int, error) {
		return 69, nil
	}

	wrk := func(ctx tempoliteContext.WorkflowContext) (int, error) {
		// Feature tested: we should be able to restart the workflow automatically (we will test the restart scheduler on another test)
		if failure {
			failure = false
			return 0, fmt.Errorf("error on purpose")
		}

		var result int
		if err := ctx.Workflow("subworkflow", subwrk, nil).Get(&result); err != nil {
			return 0, err
		}

		return 420, nil
	}

	ctx := context.Background()

	tp, err := tempolite.New(
		ctx,
		registry.
			New().
			Workflow(subwrk).
			Workflow(wrk).
			Build(),
		tempolite.WithPath("./dbs/tempolite-subworkflow-before-failure.db"),
		tempolite.WithDestructive(),
		tempolite.WithDefaultLogLevel(logs.LevelDebug),
	)

	if err != nil {
		t.Fatal(err)
	}

	// Because we have a sub-workflow
	tp.Scale("default", map[string]int{
		"workflows":   2,
		"activities":  2,
		"sideEffects": 2,
		"sagas":       2,
	})

	info := tp.Workflow(wrk, nil)

	var realValue int

	fmt.Println("Info", info.Get(&realValue))

	if realValue != 420 {
		t.Fatal("Real value should be 420")
	}

	if err = tp.Shutdown(); err != nil {
		if err != context.Canceled {
			t.Fatal(err)
		}
	}
}

// TODO: test content of the DB
func TestSubWorkflowWithParentFailureAfterSubWorkflow(t *testing.T) {
	failure := true

	// We should have 3 executions:
	// - main one
	// - subworkflow
	// - main one again while only getting the result of subworkflow because it was a success

	subwrk := func(ctx tempoliteContext.WorkflowContext) (int, error) {
		return 69, nil
	}

	wrk := func(ctx tempoliteContext.WorkflowContext) (int, error) {
		var result int
		if err := ctx.Workflow("subworkflow", subwrk, nil).Get(&result); err != nil {
			return 0, err
		}

		// Feature tested: we should be able to directly get back the result of `subworkflow` and not re-execute it
		if failure {
			failure = false
			return 0, fmt.Errorf("error on purpose")
		}

		return 420, nil
	}

	ctx := context.Background()

	tp, err := tempolite.New(
		ctx,
		registry.
			New().
			Workflow(subwrk).
			Workflow(wrk).
			Build(),
		tempolite.WithPath("./dbs/tempolite-subworkflow-after-failure.db"),
		tempolite.WithDestructive(),
		tempolite.WithDefaultLogLevel(logs.LevelDebug),
	)

	if err != nil {
		t.Fatal(err)
	}

	// Because we have a sub-workflow
	tp.Scale("default", map[string]int{
		"workflows":   2,
		"activities":  2,
		"sideEffects": 2,
		"sagas":       2,
	})

	info := tp.Workflow(wrk, nil)

	var realValue int

	fmt.Println("Info", info.Get(&realValue))

	if realValue != 420 {
		t.Fatal("Real value should be 420")
	}

	if err = tp.Shutdown(); err != nil {
		if err != context.Canceled {
			t.Fatal(err)
		}
	}
}

// We need to create multiple workflows with ONE worker with ZERO sub-workflows, quit tempolite and restart it
func TestWorkflowRestart(t *testing.T) {

	wrk1 := func(ctx tempoliteContext.WorkflowContext) (int, error) {
		// <-time.After(1 * time.Second)
		return 1, nil
	}

	ctx := context.Background()

	tp, err := tempolite.New(
		ctx,
		registry.
			New().
			Workflow(wrk1).
			Build(),
		tempolite.WithPath("./dbs/tempolite-restart.db"),
		tempolite.WithDestructive(),
		tempolite.WithDefaultLogLevel(logs.LevelDebug),
	)

	if err != nil {
		t.Fatal(err)
	}

	infos := []types.WorkflowID{}

	for i := 0; i < 100; i++ {
		info := tp.Workflow(wrk1, nil)
		infos = append(infos, info.WorkflowID())
	}

	// we stop immediately tempolite
	if err = tp.Shutdown(); err != nil {
		if err != context.Canceled {
			t.Fatal(err)
		}
	}

	// in theory, we have at least ONE of the 3 workflows that should be restarted
	<-time.After(1 * time.Second)
	fmt.Println("Restarting Tempolite")

	ctx = context.Background()
	tp, err = tempolite.New(
		ctx,
		registry.
			New().
			Workflow(wrk1).
			Build(),
		tempolite.WithPath("./dbs/tempolite-restart.db"),
	)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Tempolite restarted")

	var realValue int

	for i := 0; i < len(infos); i++ {
		fmt.Println("Info", i, infos[i])
		info := tp.GetWorkflow(infos[i])
		if err = info.Get(&realValue); err != nil {
			t.Fatal(err)
		}
		fmt.Println("Info", i, realValue)
		if realValue != 1 {
			t.Fatal("Real value should be 1")
		}
	}

	fmt.Println("Ready to relaly shutdown")
	if err = tp.Shutdown(); err != nil {
		if err != context.Canceled {
			t.Fatal(err)
		}
	}
}
