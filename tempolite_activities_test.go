package tempolite

import (
	"context"
	"fmt"
	"testing"
)

type testMessageActivitySimple struct {
	Message string
}

type testSimpleActivity struct {
	SpecialValue string
}

func (h testSimpleActivity) Run(ctx ActivityContext, task testMessageActivitySimple) (int, string, error) {

	fmt.Println("testSimpleActivity: ", task.Message)

	return 420, "cool", nil
}

// go test -timeout 30s -v -count=1 -run ^TestActivitySimple$ .
func TestActivitySimple(t *testing.T) {

	tp, err := New(
		context.Background(),
		WithPath("./db/tempolite-activity-simple.db"),
		WithDestructive(),
	)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer tp.Close()

	if err := tp.RegisterActivity(AsActivity[testSimpleActivity](testSimpleActivity{SpecialValue: "test"})); err != nil {
		t.Fatalf("RegisterActivity failed: %v", err)
	}

	tp.workflows.Range(func(key, value any) bool {
		fmt.Println("key: ", key, "value: ", value)
		return true
	})

	if _, err := tp.EnqueueActivity(As[testSimpleActivity](), testMessageActivitySimple{"hello"}); err != nil {
		t.Fatalf("EnqueueActivity failed: %v", err)
	}

	if err := tp.Wait(); err != nil {
		t.Fatalf("Wait failed: %v", err)
	}
}

type testSimpleParentActivity struct{}

func (h testSimpleParentActivity) Run(ctx ActivityContext) (int, string, error) {

	execID, err := ctx.ExecuteActivity(As[testSimpleActivity](), testMessageActivitySimple{"parent"})
	if err != nil {
		return 0, "", err
	}

	res, err := execID.Get()
	if err != nil {
		return 0, "", err
	}

	fmt.Println("res: ", res)

	if res == nil {
		return 0, "", fmt.Errorf("invalid result: %v", res)
	}
	if len(res) != 2 {
		return 0, "", fmt.Errorf("invalid result length: %d", len(res))
	}

	return res[0].(int), res[1].(string), nil
}

// go test -timeout 30s -v -count=1 -run ^TestActivityChildrenSimple$ .
func TestActivityChildrenSimple(t *testing.T) {

	tp, err := New(
		context.Background(),
		WithPath("./db/tempolite-activity-children-simple.db"),
		WithDestructive(),
	)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer tp.Close()

	if err := tp.RegisterActivity(AsActivity[testSimpleActivity](testSimpleActivity{SpecialValue: "children"})); err != nil {
		t.Fatalf("RegisterActivity failed: %v", err)
	}
	if err := tp.RegisterActivity(AsActivity[testSimpleParentActivity]()); err != nil {
		t.Fatalf("RegisterActivity failed: %v", err)
	}

	tp.workflows.Range(func(key, value any) bool {
		fmt.Println("key: ", key, "value: ", value)
		return true
	})

	var id ActivityID
	if id, err = tp.EnqueueActivity(As[testSimpleParentActivity]()); err != nil {
		t.Fatalf("EnqueueActivity failed: %v", err)
	}

	if err := tp.Wait(); err != nil {
		t.Fatalf("Wait failed: %v", err)
	}

	info, err := tp.GetActivity(id)
	if err != nil {
		t.Fatalf("GetActivity failed: %v", err)
	}

	fmt.Println("info: ", info)

	data, err := info.Get()
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	fmt.Println("data: ", data)
}

type testActivityParentWorkflow struct{}

func (h testActivityParentWorkflow) Run(ctx ActivityContext) (int, error) {

	execID, err := ctx.ExecuteWorkflow(localWrkflw, 12, "your parent is an activity")
	if err != nil {
		return 0, err
	}

	res, err := execID.Get()
	if err != nil {
		return 0, err
	}

	fmt.Println("res: ", res)

	if res == nil {
		return 0, fmt.Errorf("invalid result: %v", res)
	}
	if len(res) != 1 {
		return 0, fmt.Errorf("invalid result length: %d", len(res))
	}

	return res[0].(int), nil
}

func localWrkflw(ctx WorkflowContext, input int, msg string) (int, error) {
	fmt.Println("localWrkflw: ", input, msg)
	return 420, nil
}

// go test -timeout 30s -v -count=1 -run ^TestActivityChildrenWorkflow$ .
func TestActivityChildrenWorkflow(t *testing.T) {

	tp, err := New(
		context.Background(),
		WithPath("./db/tempolite-activity-parent-workflow-simple.db"),
		WithDestructive(),
	)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer tp.Close()

	if err := tp.RegisterWorkflow(localWrkflw); err != nil {
		t.Fatalf("RegisterWorkflow failed: %v", err)
	}

	if err := tp.RegisterActivity(AsActivity[testActivityParentWorkflow]()); err != nil {
		t.Fatalf("RegisterActivity failed: %v", err)
	}

	tp.workflows.Range(func(key, value any) bool {
		fmt.Println("key: ", key, "value: ", value)
		return true
	})

	var id ActivityID
	if id, err = tp.EnqueueActivity(As[testActivityParentWorkflow]()); err != nil {
		t.Fatalf("EnqueueActivity failed: %v", err)
	}

	if err := tp.Wait(); err != nil {
		t.Fatalf("Wait failed: %v", err)
	}

	info, err := tp.GetActivity(id)
	if err != nil {
		t.Fatalf("GetActivity failed: %v", err)
	}

	fmt.Println("info: ", info)

	data, err := info.Get()
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	fmt.Println("data: ", data)
}
