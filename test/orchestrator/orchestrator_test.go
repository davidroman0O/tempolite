package test

import (
	"context"
	"testing"

	"github.com/davidroman0O/tempolite"
	tempoliteContext "github.com/davidroman0O/tempolite/internal/engine/context"
	"github.com/davidroman0O/tempolite/internal/engine/registry"
	"github.com/davidroman0O/tempolite/pkg/logs"
)

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

	_, err = tp.Workflow(wrk, nil)
	if err != nil {
		t.Fatal(err)
	}

	// var realValue int

	// fmt.Println("Info", info.Get(&realValue))

	// if realValue != 42 {
	// 	t.Fatal("Real value should be 42")
	// }

	if err = tp.Shutdown(); err != nil {
		if err != context.Canceled {
			t.Fatal(err)
		}
	}
}

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

	_, err = tp.Workflow(wrk, nil)
	if err != nil {
		t.Fatal(err)
	}

	if err = tp.Shutdown(); err != nil {
		if err != context.Canceled {
			t.Fatal(err)
		}
	}
}
