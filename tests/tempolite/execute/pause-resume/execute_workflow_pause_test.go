package main

import (
	"context"
	"testing"

	"github.com/davidroman0O/tempolite"
)

func TestTempoliteWorkflowsExecutePause(t *testing.T) {
	database := tempolite.NewMemoryDatabase()
	defer database.SaveAsJSON("./jsons/tempolite_workflows_execute_pause.json")
	ctx := context.Background()

	tp, err := tempolite.New(
		ctx,
		database,
		tempolite.WithDefaultQueueWorkers(1, 5),
	)
	if err != nil {
		t.Fatal(err)
	}

	activityFun := func(ctx tempolite.ActivityContext) error {
		return nil
	}

	workflowFunc := func(ctx tempolite.WorkflowContext) error {
		if err := ctx.Activity("a", activityFun, nil).Get(); err != nil {
			return err
		}
		if err := ctx.Activity("b", activityFun, nil).Get(); err != nil {
			return err
		}
		return nil
	}

	future, err := tp.ExecuteDefault(workflowFunc, nil)
	if err != nil {
		t.Fatal(err)
	}

	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

}
