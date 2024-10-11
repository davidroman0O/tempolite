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

	failed := false

	localWrkflw := func(ctx WorkflowContext, input int, msg workflowData) error {
		fmt.Println("localWrkflw: ", input, msg)
		if !failed {
			failed = true
			return fmt.Errorf("localWrkflw: %d, %s", input, msg.Message)
		}
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

// go test -timeout 30s -v -count=1 -run ^TestWorkflowSimpleInfoGet$ .
func TestWorkflowSimpleInfoGet(t *testing.T) {

	tp, err := New(
		context.Background(),
		WithPath("./db/tempolite-workflow-infoget-simple.db"),
		WithDestructive(),
	)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer tp.Close()

	type workflowData struct {
		Message string
	}

	failed := false

	localWrkflw := func(ctx WorkflowContext, input int, msg workflowData) (int, error) {
		fmt.Println("localWrkflw: ", input, msg)
		if !failed {
			failed = true
			return -1, fmt.Errorf("localWrkflw: %d, %s", input, msg.Message)
		}
		return 420, nil
	}

	if err := tp.RegisterWorkflow(localWrkflw); err != nil {
		t.Fatalf("RegisterWorkflow failed: %v", err)
	}

	tp.workflows.Range(func(key, value any) bool {
		fmt.Println("key: ", key, "value: ", value)
		return true
	})

	var id WorkflowID
	if id, err = tp.EnqueueWorkflow(localWrkflw, 1, workflowData{Message: "hello"}); err != nil {
		t.Fatalf("EnqueueActivityFunc failed: %v", err)
	}

	if err := tp.Wait(); err != nil {
		t.Fatalf("Wait failed: %v", err)
	}

	info, err := tp.GetWorkflow(id)
	if err != nil {
		t.Fatalf("GetWorkflow failed: %v", err)
	}

	fmt.Println("info: ", info)

	data, err := info.Get()
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	fmt.Println("data: ", data)

}

// go test -timeout 30s -v -count=1 -run ^TestWorkflowSimpleSubWorkflowInfoGetFailChild$ .
func TestWorkflowSimpleSubWorkflowInfoGetFailChild(t *testing.T) {

	tp, err := New(
		context.Background(),
		WithPath("./db/tempolite-workflow-sub-info-child-fail.db"),
		WithDestructive(),
	)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer tp.Close()

	type workflowData struct {
		Message string
	}

	failed := false

	anotherWrk := func(ctx WorkflowContext) error {
		fmt.Println("anotherWrk")
		// If we fail here, then the the info.Get will fail and the parent workflow, will also fail
		// but does that mean, we should be retried?
		if !failed {
			failed = true
			fmt.Println("failed on purpose: ", failed)
			return fmt.Errorf("on purpose")
		}
		return nil
	}

	localWrkflw := func(ctx WorkflowContext, input int, msg workflowData) (int, error) {
		fmt.Println("localWrkflw: ", failed, input, msg)

		info, err := ctx.ExecuteWorkflow(anotherWrk)
		if err != nil {
			fmt.Println("ExecuteWorkflow failed: ", err)
			return -1, err
		}

		_, err = info.Get()
		if err != nil {
			fmt.Println("info.Get failed: ", err)
			return -1, err
		}
		return 420, nil
	}

	if err := tp.RegisterWorkflow(anotherWrk); err != nil {
		t.Fatalf("RegisterWorkflow failed: %v", err)
	}

	if err := tp.RegisterWorkflow(localWrkflw); err != nil {
		t.Fatalf("RegisterWorkflow failed: %v", err)
	}

	tp.workflows.Range(func(key, value any) bool {
		fmt.Println("key: ", key, "value: ", value)
		return true
	})

	var id WorkflowID
	if id, err = tp.EnqueueWorkflow(localWrkflw, 1, workflowData{Message: "hello"}); err != nil {
		t.Fatalf("EnqueueActivityFunc failed: %v", err)
	}

	if err := tp.Wait(); err != nil {
		t.Fatalf("Wait failed: %v", err)
	}

	info, err := tp.GetWorkflow(id)
	if err != nil {
		t.Fatalf("GetWorkflow failed: %v", err)
	}

	fmt.Println("info: ", info)

	data, err := info.Get()
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	fmt.Println("data: ", data)
}

// go test -timeout 30s -v -count=1 -run ^TestWorkflowSimpleSubWorkflowInfoGetFailParent$ .
func TestWorkflowSimpleSubWorkflowInfoGetFailParent(t *testing.T) {

	tp, err := New(
		context.Background(),
		WithPath("./db/tempolite-workflow-sub-info-parent-fail.db"),
		WithDestructive(),
	)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer tp.Close()

	type workflowData struct {
		Message string
	}

	failed := false

	anotherWrk := func(ctx WorkflowContext) error {
		fmt.Println("anotherWrk")
		return nil
	}

	localWrkflw := func(ctx WorkflowContext, input int, msg workflowData) (int, error) {
		fmt.Println("localWrkflw: ", failed, input, msg)

		info, err := ctx.ExecuteWorkflow(anotherWrk)
		if err != nil {
			fmt.Println("ExecuteWorkflow failed: ", err)
			return -1, err
		}

		_, err = info.Get()
		if err != nil {
			fmt.Println("info.Get failed: ", err)
			return -1, err
		}

		if !failed {
			failed = true
			fmt.Println("failed on purpose: ", failed)
			return -1, fmt.Errorf("localWrkflw: %d, %s", input, msg.Message)
		}
		return 420, nil
	}

	if err := tp.RegisterWorkflow(anotherWrk); err != nil {
		t.Fatalf("RegisterWorkflow failed: %v", err)
	}

	if err := tp.RegisterWorkflow(localWrkflw); err != nil {
		t.Fatalf("RegisterWorkflow failed: %v", err)
	}

	tp.workflows.Range(func(key, value any) bool {
		fmt.Println("key: ", key, "value: ", value)
		return true
	})

	var id WorkflowID
	if id, err = tp.EnqueueWorkflow(localWrkflw, 1, workflowData{Message: "hello"}); err != nil {
		t.Fatalf("EnqueueActivityFunc failed: %v", err)
	}

	if err := tp.Wait(); err != nil {
		t.Fatalf("Wait failed: %v", err)
	}

	info, err := tp.GetWorkflow(id)
	if err != nil {
		t.Fatalf("GetWorkflow failed: %v", err)
	}

	fmt.Println("info: ", info)

	data, err := info.Get()
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	fmt.Println("data: ", data)
}
