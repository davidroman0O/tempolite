package tempolite

import (
	"context"
	"testing"
)

func TestOrchestratorVersionWorkflows(t *testing.T) {

	database := NewDefaultDatabase()
	register := NewRegistry()
	ctx := context.Background()

	executed := false

	workflow := func(ctx WorkflowContext) error {

		v, e := ctx.GetVersion("featureA", DefaultVersion, 2)
		if e != nil {
			return e
		}
		switch v {
		case DefaultVersion:
			t.Log("Version original version")
		case 1:
			t.Log("Version 1")
		case 2:
			t.Log("Version 2")
		default:
			t.Log("Unknown version")
		}

		executed = true
		return nil
	}

	register.RegisterWorkflow(workflow)

	o := NewOrchestrator(ctx, database, register)

	future := o.Execute(workflow, &WorkflowOptions{
		VersionOverrides: map[string]int{
			"featureA": 1,
		},
	})

	t.Log("Waiting for workflow to complete")
	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

	if !executed {
		t.Fatal("Workflow was not executed")
	}

	t.Log("Orchestrator workflow completed successfully")
	o.Wait()
}
