package tempolite

import (
	"context"
	"fmt"
	"testing"
)

// go test -timeout 30s -v -count=1 -run ^TestWorkflowSimple$ .
func TestWorkflowSimple(t *testing.T) {

	tp, err := New(
		context.Background(),
		WithPath("./db/tempolite-workflow-simple.db"),
		WithDestructive(),
	)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer tp.Close()

	type workflowData struct {
		Message string
	}

	localWrkflw := func(ctx WorkflowContext, input int, msg workflowData) error {
		fmt.Println("localWrkflw: ", input, msg)
		return nil
	}

	if err := tp.RegisterWorkflow(localWrkflw); err != nil {
		t.Fatalf("RegisterWorkflow failed: %v", err)
	}

	tp.workflows.Range(func(key, value any) bool {
		fmt.Println("key: ", key, "value: ", value)
		return true
	})

	if _, err := tp.EnqueueWorkflow(localWrkflw, 1, workflowData{Message: "hello"}); err != nil {
		t.Fatalf("EnqueueActivityFunc failed: %v", err)
	}

	if err := tp.Wait(); err != nil {
		t.Fatalf("Wait failed: %v", err)
	}
}
