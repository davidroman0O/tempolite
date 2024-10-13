package tempolite

import (
	"context"
	"fmt"
	"testing"
)

type testIdentifier string

// go test -timeout 30s -v -count=1 -run ^TestWorkflowSimple$ .
func TestWorkflowSimple(t *testing.T) {

	tp, err := New[testIdentifier](
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

	localWrkflw := func(ctx WorkflowContext[testIdentifier], input int, msg workflowData) error {
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

	if err := tp.Workflow("test", localWrkflw, 1, workflowData{Message: "hello"}).Get(); err != nil {
		t.Fatalf("EnqueueActivityFunc failed: %v", err)
	}

	if err := tp.Wait(); err != nil {
		t.Fatalf("Wait failed: %v", err)
	}
}

type testMessageActivitySimple struct {
	Message string
}

type testSimpleActivity struct {
	SpecialValue string
}

func (h testSimpleActivity) Run(ctx ActivityContext[testIdentifier], task testMessageActivitySimple) (int, string, error) {

	fmt.Println("testSimpleActivity: ", task.Message)

	return 420, "cool", nil
}

// go test -timeout 30s -v -count=1 -run ^TestWorkflowActivitySimple$ .
func TestWorkflowActivitySimple(t *testing.T) {

	tp, err := New[testIdentifier](
		context.Background(),
		WithPath("./db/tempolite-workflow-activity-simple.db"),
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

	localWrkflw := func(ctx WorkflowContext[testIdentifier], input int, msg workflowData) error {
		fmt.Println("localWrkflw: ", input, msg)

		var number int
		var str string
		err := ctx.ExecuteActivity("test", As[testSimpleActivity, testIdentifier](), testMessageActivitySimple{Message: "hello"}).Get(&number, &str)
		if err != nil {
			return err
		}

		fmt.Println("number: ", number, "str: ", str)

		if !failed {
			failed = true
			return fmt.Errorf("on purpose error: %d, %s", input, msg.Message)
		}

		return nil
	}

	if err := tp.RegisterActivity(AsActivity[testSimpleActivity, testIdentifier](testSimpleActivity{SpecialValue: "test"})); err != nil {
		t.Fatalf("RegisterActivity failed: %v", err)
	}
	if err := tp.RegisterWorkflow(localWrkflw); err != nil {
		t.Fatalf("RegisterWorkflow failed: %v", err)
	}

	tp.workflows.Range(func(key, value any) bool {
		fmt.Println("key: ", key, "value: ", value)
		return true
	})

	if err := tp.Workflow("test", localWrkflw, 1, workflowData{Message: "hello"}).Get(); err != nil {
		t.Fatalf("EnqueueActivityFunc failed: %v", err)
	}

	if err := tp.Wait(); err != nil {
		t.Fatalf("Wait failed: %v", err)
	}
}

// go test -timeout 30s -v -count=1 -run ^TestWorkflowActivityMore$ .
func TestWorkflowActivityMore(t *testing.T) {

	tp, err := New[testIdentifier](
		context.Background(),
		WithPath("./db/tempolite-workflow-activity-more.db"),
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

	activtfn := func(ctx ActivityContext[testIdentifier], id int) (int, error) {
		fmt.Println("activtfn: ", id)

		if !failed {
			failed = true
			return -1, fmt.Errorf("on purpose error")
		}
		return 69, nil
	}

	localWrkflw := func(ctx WorkflowContext[testIdentifier], input int, msg workflowData) error {
		fmt.Println("localWrkflw: ", input, msg)

		var subnumber int
		if err := ctx.ActivityFunc("first", activtfn, 420).Get(&subnumber); err != nil {
			return err
		}

		fmt.Println("subnumber: ", subnumber)

		var number int
		var str string
		err := ctx.ExecuteActivity("second", As[testSimpleActivity, testIdentifier](), testMessageActivitySimple{Message: "hello"}).Get(&number, &str)
		if err != nil {
			return err
		}

		fmt.Println("number: ", number, "str: ", str)

		return nil
	}

	if err := tp.RegisterActivityFunc(activtfn); err != nil {
		t.Fatalf("RegisterActivityFunc failed: %v", err)
	}

	if err := tp.RegisterActivity(AsActivity[testSimpleActivity, testIdentifier](testSimpleActivity{SpecialValue: "test"})); err != nil {
		t.Fatalf("RegisterActivity failed: %v", err)
	}
	if err := tp.RegisterWorkflow(localWrkflw); err != nil {
		t.Fatalf("RegisterWorkflow failed: %v", err)
	}

	tp.workflows.Range(func(key, value any) bool {
		fmt.Println("key: ", key, "value: ", value)
		return true
	})

	if err := tp.Workflow("test", localWrkflw, 1, workflowData{Message: "hello"}).Get(); err != nil {
		t.Fatalf("EnqueueActivityFunc failed: %v", err)
	}

	if err := tp.Wait(); err != nil {
		t.Fatalf("Wait failed: %v", err)
	}
}

// go test -timeout 30s -v -count=1 -run ^TestWorkflowSimpleInfoGet$ .
func TestWorkflowSimpleInfoGet(t *testing.T) {

	tp, err := New[testIdentifier](
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

	localWrkflw := func(ctx WorkflowContext[testIdentifier], input int, msg workflowData) (int, error) {
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

	var number int
	if err = tp.Workflow("test", localWrkflw, 1, workflowData{Message: "hello"}).Get(&number); err != nil {
		t.Fatalf("EnqueueActivityFunc failed: %v", err)
	}

	if err := tp.Wait(); err != nil {
		t.Fatalf("Wait failed: %v", err)
	}

	fmt.Println("data: ", number)
}

// go test -timeout 30s -v -count=1 -run ^TestWorkflowSimpleSubWorkflowInfoGetFailChild$ .
func TestWorkflowSimpleSubWorkflowInfoGetFailChild(t *testing.T) {

	tp, err := New[testIdentifier](
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

	anotherWrk := func(ctx WorkflowContext[testIdentifier]) error {
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

	localWrkflw := func(ctx WorkflowContext[testIdentifier], input int, msg workflowData) (int, error) {
		fmt.Println("localWrkflw: ", failed, input, msg)

		err := ctx.Workflow("test", anotherWrk).Get()

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

	var number int
	if err = tp.Workflow("test", localWrkflw, 1, workflowData{Message: "hello"}).Get(&number); err != nil {
		t.Fatalf("EnqueueActivityFunc failed: %v", err)
	}

	if err := tp.Wait(); err != nil {
		t.Fatalf("Wait failed: %v", err)
	}

	fmt.Println("data: ", number)
}

// go test -timeout 30s -v -count=1 -run ^TestWorkflowSimpleSubWorkflowInfoGetFailParent$ .
func TestWorkflowSimpleSubWorkflowInfoGetFailParent(t *testing.T) {

	tp, err := New[testIdentifier](
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

	anotherWrk := func(ctx WorkflowContext[testIdentifier]) error {
		fmt.Println("anotherWrk")
		return nil
	}

	localWrkflw := func(ctx WorkflowContext[testIdentifier], input int, msg workflowData) (int, error) {
		fmt.Println("localWrkflw: ", failed, input, msg)

		err := ctx.Workflow("test", anotherWrk).Get()
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

	var number int
	if err = tp.Workflow("test", localWrkflw, 1, workflowData{Message: "hello"}).Get(&number); err != nil {
		t.Fatalf("EnqueueActivityFunc failed: %v", err)
	}

	if err := tp.Wait(); err != nil {
		t.Fatalf("Wait failed: %v", err)
	}

	fmt.Println("data: ", number)
}
