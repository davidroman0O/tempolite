package tempolite

import (
	"context"
	"testing"
)

// go test -timeout 30s -v -count=1 -run ^TestWorkflowVersioning$ .

func TestWorkflowVersioning(t *testing.T) {
	dbPath := "./db/test-versioning.db"

	type testWorkflowInput struct {
		ChangeFlag1 int32
		ChangeFlag2 bool
	}

	testWorkflow := func(ctx WorkflowContext, input testWorkflowInput) (string, error) {
		var result string

		changeFlag1Value := int(input.ChangeFlag1)
		v1 := ctx.GetVersion("Change1", DefaultVersion, changeFlag1Value)
		if v1 == DefaultVersion {
			result += "A"
		} else if v1 == 1 {
			result += "B"
		} else {
			result += "C"
		}

		var v2 int
		if input.ChangeFlag2 {
			// Change2 is activated; set min and max to 1
			v2 = ctx.GetVersion("Change2", 1, 1)
		} else {
			// Change2 is not activated; set min and max to DefaultVersion
			v2 = ctx.GetVersion("Change2", DefaultVersion, DefaultVersion)
		}
		if v2 == DefaultVersion {
			result += "X"
		} else {
			result += "Y"
		}

		return result, nil
	}

	var changeFlag1 int32
	var changeFlag2 bool

	// First run with WithDestructive
	tp, err := New(context.Background(), WithPath(dbPath), WithDestructive())
	if err != nil {
		t.Fatalf("Failed to create Tempolite instance: %v", err)
	}

	if err := tp.RegisterWorkflow(testWorkflow); err != nil {
		t.Fatalf("Failed to register workflow: %v", err)
	}

	id, err := tp.Workflow(testWorkflow, testWorkflowInput{changeFlag1, changeFlag2})
	if err != nil {
		t.Fatalf("EnqueueWorkflow failed: %v", err)
	}

	if err := tp.Wait(); err != nil {
		t.Fatalf("Wait failed: %v", err)
	}

	var result string
	err = tp.GetWorkflow(id).Get(&result)
	if err != nil {
		t.Fatalf("info.Get failed: %v", err)
	}
	if result != "AX" {
		t.Errorf("Expected AX, got %s", result)
	}

	tp.Close()

	// Second run, Change1 to version 1
	tp, err = New(context.Background(), WithPath(dbPath))
	if err != nil {
		t.Fatalf("Failed to create Tempolite instance: %v", err)
	}

	if err := tp.RegisterWorkflow(testWorkflow); err != nil {
		t.Fatalf("Failed to register workflow: %v", err)
	}

	changeFlag1 = 1

	id, err = tp.Workflow(testWorkflow, testWorkflowInput{changeFlag1, changeFlag2})
	if err != nil {
		t.Fatalf("EnqueueWorkflow failed: %v", err)
	}

	if err := tp.Wait(); err != nil {
		t.Fatalf("Wait failed: %v", err)
	}

	err = tp.GetWorkflow(id).Get(&result)
	if err != nil {
		t.Fatalf("info.Get failed: %v", err)
	}
	if result != "BX" {
		t.Errorf("Expected BX, got %s", result)
	}

	tp.Close()

	// Third run, Change1 to version 2
	tp, err = New(context.Background(), WithPath(dbPath))
	if err != nil {
		t.Fatalf("Failed to create Tempolite instance: %v", err)
	}

	if err := tp.RegisterWorkflow(testWorkflow); err != nil {
		t.Fatalf("Failed to register workflow: %v", err)
	}

	changeFlag1 = 2

	id, err = tp.Workflow(testWorkflow, testWorkflowInput{changeFlag1, changeFlag2})
	if err != nil {
		t.Fatalf("EnqueueWorkflow failed: %v", err)
	}

	if err := tp.Wait(); err != nil {
		t.Fatalf("Wait failed: %v", err)
	}

	err = tp.GetWorkflow(id).Get(&result)
	if err != nil {
		t.Fatalf("info.Get failed: %v", err)
	}
	if result != "CX" {
		t.Errorf("Expected CX, got %s", result)
	}

	tp.Close()

	// Fourth run, activate Change2
	tp, err = New(context.Background(), WithPath(dbPath))
	if err != nil {
		t.Fatalf("Failed to create Tempolite instance: %v", err)
	}

	if err := tp.RegisterWorkflow(testWorkflow); err != nil {
		t.Fatalf("Failed to register workflow: %v", err)
	}

	changeFlag1 = 0
	changeFlag2 = true

	id, err = tp.Workflow(testWorkflow, testWorkflowInput{changeFlag1, changeFlag2})
	if err != nil {
		t.Fatalf("EnqueueWorkflow failed: %v", err)
	}

	if err := tp.Wait(); err != nil {
		t.Fatalf("Wait failed: %v", err)
	}

	err = tp.GetWorkflow(id).Get(&result)
	if err != nil {
		t.Fatalf("info.Get failed: %v", err)
	}
	if result != "CY" {
		t.Errorf("Expected CY, got %s", result)
	}

	tp.Close()
}
