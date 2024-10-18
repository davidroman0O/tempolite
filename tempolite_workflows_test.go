package tempolite

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"
)

type testIdentifier string

// go test -timeout 30s -v -count=1 -run ^TestWorkflowSimple$ .
func TestWorkflowSimple(t *testing.T) {

	type workflowData struct {
		Message string
	}

	failed := false

	localWrkflw := func(ctx WorkflowContext[testIdentifier], input int, msg workflowData) error {
		// fmt.Println("localWrkflw: ", input, msg)
		if !failed {
			failed = true
			return fmt.Errorf("localWrkflw: %d, %s", input, msg.Message)
		}
		return nil
	}

	registery := NewRegistry[testIdentifier]().
		Workflow(localWrkflw).
		Build()

	tp, err := New[testIdentifier](
		context.Background(),
		registery,
		WithPath("./db/tempolite-workflow-simple.db"),
		WithDestructive(),
	)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer tp.Close()

	tp.workflows.Range(func(key, value any) bool {
		// fmt.Println("key: ", key, "value: ", value)
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

	// fmt.Println("testSimpleActivity: ", task.Message)

	return 420, "cool", nil
}

// go test -timeout 30s -v -count=1 -run ^TestWorkflowActivitySimple$ .
func TestWorkflowActivitySimple(t *testing.T) {

	type workflowData struct {
		Message string
	}

	failed := false

	localWrkflw := func(ctx WorkflowContext[testIdentifier], input int, msg workflowData) error {
		// fmt.Println("localWrkflw: ", input, msg)

		var number int
		var str string
		err := ctx.ExecuteActivity("test", As[testSimpleActivity, testIdentifier](), testMessageActivitySimple{Message: "hello"}).Get(&number, &str)
		if err != nil {
			return err
		}

		// fmt.Println("number: ", number, "str: ", str)

		if !failed {
			failed = true
			return fmt.Errorf("on purpose error: %d, %s", input, msg.Message)
		}

		return nil
	}

	registery := NewRegistry[testIdentifier]().
		Workflow(localWrkflw).
		Build()

	tp, err := New[testIdentifier](
		context.Background(),
		registery,
		WithPath("./db/tempolite-workflow-activity-simple.db"),
		WithDestructive(),
	)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer tp.Close()

	if err := tp.registerActivity(AsActivity[testSimpleActivity, testIdentifier](testSimpleActivity{SpecialValue: "test"})); err != nil {
		t.Fatalf("RegisterActivity failed: %v", err)
	}
	if err := tp.registerWorkflow(localWrkflw); err != nil {
		t.Fatalf("RegisterWorkflow failed: %v", err)
	}

	tp.workflows.Range(func(key, value any) bool {
		// fmt.Println("key: ", key, "value: ", value)
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

	type workflowData struct {
		Message string
	}

	failed := false

	activtfn := func(ctx ActivityContext[testIdentifier], id int) (int, error) {
		// fmt.Println("activtfn: ", id)

		if !failed {
			failed = true
			return -1, fmt.Errorf("on purpose error")
		}
		return 69, nil
	}

	localWrkflw := func(ctx WorkflowContext[testIdentifier], input int, msg workflowData) error {
		// fmt.Println("localWrkflw: ", input, msg)

		var subnumber int
		if err := ctx.ActivityFunc("first", activtfn, 420).Get(&subnumber); err != nil {
			return err
		}

		// fmt.Println("subnumber: ", subnumber)

		var number int
		var str string
		err := ctx.ExecuteActivity("second", As[testSimpleActivity, testIdentifier](), testMessageActivitySimple{Message: "hello"}).Get(&number, &str)
		if err != nil {
			return err
		}

		// fmt.Println("number: ", number, "str: ", str)

		return nil
	}

	registery := NewRegistry[testIdentifier]().
		Workflow(localWrkflw).
		ActivityFunc(activtfn).
		Build()

	tp, err := New[testIdentifier](
		context.Background(),
		registery,
		WithPath("./db/tempolite-workflow-activity-more.db"),
		WithDestructive(),
	)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer tp.Close()

	if err := tp.registerActivityFunc(activtfn); err != nil {
		t.Fatalf("RegisterActivityFunc failed: %v", err)
	}

	if err := tp.registerActivity(AsActivity[testSimpleActivity, testIdentifier](testSimpleActivity{SpecialValue: "test"})); err != nil {
		t.Fatalf("RegisterActivity failed: %v", err)
	}
	if err := tp.registerWorkflow(localWrkflw); err != nil {
		t.Fatalf("RegisterWorkflow failed: %v", err)
	}

	tp.workflows.Range(func(key, value any) bool {
		// fmt.Println("key: ", key, "value: ", value)
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

	type workflowData struct {
		Message string
	}

	failed := false

	localWrkflw := func(ctx WorkflowContext[testIdentifier], input int, msg workflowData) (int, error) {
		// fmt.Println("localWrkflw: ", input, msg)
		if !failed {
			failed = true
			return -1, fmt.Errorf("localWrkflw: %d, %s", input, msg.Message)
		}
		return 420, nil
	}

	registery := NewRegistry[testIdentifier]().
		Workflow(localWrkflw).
		Build()

	tp, err := New[testIdentifier](
		context.Background(),
		registery,
		WithPath("./db/tempolite-workflow-infoget-simple.db"),
		WithDestructive(),
	)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer tp.Close()

	if err := tp.registerWorkflow(localWrkflw); err != nil {
		t.Fatalf("RegisterWorkflow failed: %v", err)
	}

	tp.workflows.Range(func(key, value any) bool {
		// fmt.Println("key: ", key, "value: ", value)
		return true
	})

	var number int
	if err = tp.Workflow("test", localWrkflw, 1, workflowData{Message: "hello"}).Get(&number); err != nil {
		t.Fatalf("EnqueueActivityFunc failed: %v", err)
	}

	if err := tp.Wait(); err != nil {
		t.Fatalf("Wait failed: %v", err)
	}

	// fmt.Println("data: ", number)
}

// go test -timeout 30s -v -count=1 -run ^TestWorkflowSimpleSubWorkflowInfoGetFailChild$ .
func TestWorkflowSimpleSubWorkflowInfoGetFailChild(t *testing.T) {

	type workflowData struct {
		Message string
	}

	failed := false

	anotherWrk := func(ctx WorkflowContext[testIdentifier]) error {
		// fmt.Println("anotherWrk")
		// If we fail here, then the the info.Get will fail and the parent workflow, will also fail
		// but does that mean, we should be retried?
		if !failed {
			failed = true
			// fmt.Println("failed on purpose: ", failed)
			return fmt.Errorf("on purpose")
		}
		return nil
	}

	localWrkflw := func(ctx WorkflowContext[testIdentifier], input int, msg workflowData) (int, error) {
		// fmt.Println("localWrkflw: ", failed, input, msg)

		err := ctx.Workflow("test", anotherWrk).Get()

		if err != nil {
			// fmt.Println("info.Get failed: ", err)
			return -1, err
		}
		return 420, nil
	}

	registery := NewRegistry[testIdentifier]().
		Workflow(localWrkflw).
		Workflow(anotherWrk).
		Build()

	tp, err := New[testIdentifier](
		context.Background(),
		registery,
		WithPath("./db/tempolite-workflow-sub-info-child-fail.db"),
		WithDestructive(),
	)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer tp.Close()

	if err := tp.registerWorkflow(anotherWrk); err != nil {
		t.Fatalf("RegisterWorkflow failed: %v", err)
	}

	if err := tp.registerWorkflow(localWrkflw); err != nil {
		t.Fatalf("RegisterWorkflow failed: %v", err)
	}

	tp.workflows.Range(func(key, value any) bool {
		// fmt.Println("key: ", key, "value: ", value)
		return true
	})

	var number int
	if err = tp.Workflow("test", localWrkflw, 1, workflowData{Message: "hello"}).Get(&number); err != nil {
		t.Fatalf("EnqueueActivityFunc failed: %v", err)
	}

	if err := tp.Wait(); err != nil {
		t.Fatalf("Wait failed: %v", err)
	}

	// fmt.Println("data: ", number)
}

// go test -timeout 30s -v -count=1 -run ^TestWorkflowSimpleSubWorkflowInfoGetFailParent$ .
func TestWorkflowSimpleSubWorkflowInfoGetFailParent(t *testing.T) {

	type workflowData struct {
		Message string
	}

	failed := false

	anotherWrk := func(ctx WorkflowContext[testIdentifier]) error {
		// fmt.Println("anotherWrk")
		return nil
	}

	localWrkflw := func(ctx WorkflowContext[testIdentifier], input int, msg workflowData) (int, error) {
		// fmt.Println("localWrkflw: ", failed, input, msg)

		err := ctx.Workflow("test", anotherWrk).Get()
		if err != nil {
			// fmt.Println("info.Get failed: ", err)
			return -1, err
		}

		if !failed {
			failed = true
			// fmt.Println("failed on purpose: ", failed)
			return -1, fmt.Errorf("localWrkflw: %d, %s", input, msg.Message)
		}
		return 420, nil
	}

	registery := NewRegistry[testIdentifier]().
		Workflow(localWrkflw).
		Build()

	tp, err := New[testIdentifier](
		context.Background(),
		registery,
		WithPath("./db/tempolite-workflow-sub-info-parent-fail.db"),
		WithDestructive(),
	)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer tp.Close()

	if err := tp.registerWorkflow(anotherWrk); err != nil {
		t.Fatalf("RegisterWorkflow failed: %v", err)
	}

	if err := tp.registerWorkflow(localWrkflw); err != nil {
		t.Fatalf("RegisterWorkflow failed: %v", err)
	}

	tp.workflows.Range(func(key, value any) bool {
		// fmt.Println("key: ", key, "value: ", value)
		return true
	})

	var number int
	if err = tp.Workflow("test", localWrkflw, 1, workflowData{Message: "hello"}).Get(&number); err != nil {
		t.Fatalf("EnqueueActivityFunc failed: %v", err)
	}

	if err := tp.Wait(); err != nil {
		t.Fatalf("Wait failed: %v", err)
	}

	// fmt.Println("data: ", number)
}

// go test -timeout 30s -v -count=1 -run ^TestWorkflowSimpleSideEffect$ .
func TestWorkflowSimpleSideEffect(t *testing.T) {

	type workflowData struct {
		Message string
	}

	failed := false

	localWrkflw := func(ctx WorkflowContext[testIdentifier], input int, msg workflowData) (int, error) {
		// fmt.Println("localWrkflw: ", failed, input, msg)

		var value int
		if err := ctx.SideEffect("eventual switch", func(ctx SideEffectContext[testIdentifier]) int {
			return 69
		}).Get(&value); err != nil {
			return -1, err
		}

		if !failed {
			failed = true
			// fmt.Println("failed on purpose: ", failed)
			return -1, fmt.Errorf("localWrkflw: %d, %s", input, msg.Message)
		}

		return 69, nil
	}

	registery := NewRegistry[testIdentifier]().
		Workflow(localWrkflw).
		Build()

	tp, err := New[testIdentifier](
		context.Background(),
		registery,
		WithPath("./db/tempolite-workflow-sideeffect.db"),
		WithDestructive(),
	)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer tp.Close()

	if err := tp.registerWorkflow(localWrkflw); err != nil {
		t.Fatalf("RegisterWorkflow failed: %v", err)
	}

	tp.workflows.Range(func(key, value any) bool {
		// fmt.Println("key: ", key, "value: ", value)
		return true
	})

	var number int
	if err = tp.Workflow("test", localWrkflw, 1, workflowData{Message: "hello"}).Get(&number); err != nil {
		t.Fatalf("EnqueueActivityFunc failed: %v", err)
	}

	if err := tp.Wait(); err != nil {
		t.Fatalf("Wait failed: %v", err)
	}

	// fmt.Println("data: ", number)
	if number != 69 {
		t.Fatalf("number: %d", number)
	}
}

// go test -timeout 30s -v -count=1 -run ^TestWorkflowSimplePauseResume$ .
func TestWorkflowSimplePauseResume(t *testing.T) {

	type workflowData struct {
		Message string
	}

	activityWork := func(ctx ActivityContext[testIdentifier]) error {
		<-time.After(1 * time.Second)
		return nil
	}

	localWrkflw := func(ctx WorkflowContext[testIdentifier], input int, msg workflowData) (int, error) {

		log.Println("fake work 1")
		<-time.After(1 * time.Second)

		log.Println("pausing1")

		if err := ctx.ActivityFunc("pause1", activityWork).Get(); err != nil {
			return -1, err
		}

		log.Println("pause1 finished")

		log.Println("fake work 2")
		<-time.After(1 * time.Second)

		log.Println("pausing2")

		if err := ctx.ActivityFunc("pause2", activityWork).Get(); err != nil {
			return -1, err
		}

		log.Println("pause2 finished")
		<-time.After(1 * time.Second)

		defer log.Println("workflow finished")

		return 69, nil
	}

	registery := NewRegistry[testIdentifier]().
		Workflow(localWrkflw).
		ActivityFunc(activityWork).
		Build()

	tp, err := New[testIdentifier](
		context.Background(),
		registery,
		WithPath("./db/tempolite-workflow-yield.db"),
		WithDestructive(),
	)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer tp.Close()

	if err := tp.registerWorkflow(localWrkflw); err != nil {
		t.Fatalf("RegisterWorkflow failed: %v", err)
	}

	tp.workflows.Range(func(key, value any) bool {
		log.Println("key: ", key, "value: ", value)
		return true
	})

	var number int
	var workflowInfo *WorkflowInfo[testIdentifier]
	if workflowInfo = tp.Workflow("test", localWrkflw, 1, workflowData{Message: "hello"}); err != nil {
		t.Fatalf("EnqueueActivityFunc failed: %v", err)
	}

	{
		log.Println("\t pause1")
		if err := tp.PauseWorkflow(workflowInfo.WorkflowID); err != nil {
			t.Fatalf("PauseWorkflow failed: %v", err)
		}
		// fmt.Println("\t\t PAUSED (5s) !!!")
		<-time.After(5 * time.Second)
	}

	{
		log.Println("\t RESUME1")
		if err := tp.ResumeWorkflow(workflowInfo.WorkflowID); err != nil {
			t.Fatalf("ResumeWorkflow failed: %v", err)
		}
		<-time.After(time.Second / 2)
		log.Println("\t PAUSE2!!")
		if err := tp.PauseWorkflow(workflowInfo.WorkflowID); err != nil {
			t.Fatalf("Pause failed: %v", err)
		}
		<-time.After(5 * time.Second)
	}

	// fmt.Println("\t\t RESTARTING...")

	{
		tp.Close() // close the DB and start again
		tp, err = New[testIdentifier](
			context.Background(),
			registery,
			WithPath("./db/tempolite-workflow-yield.db"),
		)
		if err != nil {
			t.Fatalf("New failed: %v", err)
		}
		if err := tp.registerWorkflow(localWrkflw); err != nil {
			t.Fatalf("RegisterWorkflow failed: %v", err)
		}
	}
	// fmt.Println("\t\t RESTARTED !!!")

	pauses, err := tp.ListPausedWorkflows()
	if err != nil {
		t.Fatalf("ListPausedWorkflows failed: %v", err)
	}

	for _, pauseworkflow := range pauses {
		fmt.Println("pauseworkflow: ", pauseworkflow.String())
	}

	// fmt.Println("\t\t RESUMING (it will finish)...")
	<-time.After(2 * time.Second)
	log.Println("\t resume2")
	if err := tp.ResumeWorkflow(workflowInfo.WorkflowID); err != nil {
		t.Fatalf("ResumeWorkflow failed: %v", err)
	}

	if err := tp.Wait(); err != nil {
		t.Fatalf("Wait failed: %v", err)
	}

	// We changed the context
	workflowInfo = tp.GetWorkflow(workflowInfo.WorkflowID)

	if err := workflowInfo.Get(&number); err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// fmt.Println("data: ", number)
	if number != 69 {
		t.Fatalf("number: %d", number)
	}
}

// go test -timeout 30s -v -count=1 -run ^TestWorkflowSimpleSignal$ .
func TestWorkflowSimpleSignal(t *testing.T) {

	type workflowData struct {
		Message string
	}

	failure := false

	localWrkflw := func(ctx WorkflowContext[testIdentifier], input int, msg workflowData) (int, error) {

		// fmt.Println("signal..")
		signal := ctx.Signal("waiting data")

		// fmt.Println("waiting signal")
		var value int
		if err := signal.Receive(ctx, &value); err != nil {
			return -1, err
		}
		// fmt.Println("signal received: ", value)

		if !failure {
			failure = true
			return -1, fmt.Errorf("on purpose")
		}

		return 69, nil
	}

	registery := NewRegistry[testIdentifier]().
		Workflow(localWrkflw).
		Build()

	tp, err := New[testIdentifier](
		context.Background(),
		registery,
		WithPath("./db/tempolite-workflow-signal.db"),
		WithDestructive(),
	)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer tp.Close()

	if err := tp.registerWorkflow(localWrkflw); err != nil {
		t.Fatalf("RegisterWorkflow failed: %v", err)
	}

	tp.workflows.Range(func(key, value any) bool {
		// fmt.Println("key: ", key, "value: ", value)
		return true
	})

	var number int
	var workflowInfo *WorkflowInfo[testIdentifier]
	if workflowInfo = tp.Workflow("test", localWrkflw, 1, workflowData{Message: "hello"}); err != nil {
		t.Fatalf("EnqueueActivityFunc failed: %v", err)
	}

	go func() {
		// fmt.Println("waiting 2s")
		<-time.After(2 * time.Second)
		// fmt.Println("sending signal")
		if err := tp.PublishSignal(workflowInfo.WorkflowID, "waiting data", 420); err != nil {
			t.Fatalf("PublishSignal failed: %v", err)
		}
		// fmt.Println("signal sent")
	}()

	if err := workflowInfo.Get(&number); err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if err := tp.Wait(); err != nil {
		t.Fatalf("Wait failed: %v", err)
	}

	// fmt.Println("data: ", number)
	if number != 69 {
		t.Fatalf("number: %d", number)
	}
}

// go test -timeout 30s -v -count=1 -run ^TestWorkflowSimpleCancel$ .
func TestWorkflowSimpleCancel(t *testing.T) {

	type workflowData struct {
		Message string
	}

	activtyLocal := func(ctx ActivityContext[testIdentifier]) error {
		<-time.After(time.Second * 5)
		return nil
	}

	localWrkflw := func(ctx WorkflowContext[testIdentifier], input int, msg workflowData) (int, error) {

		if err := ctx.ActivityFunc("test", activtyLocal).Get(); err != nil {
			return -1, err
		}
		<-time.After(time.Second * 5)
		if err := ctx.ActivityFunc("second time", activtyLocal).Get(); err != nil {
			return -1, err
		}

		return 69, nil
	}

	registery := NewRegistry[testIdentifier]().
		Workflow(localWrkflw).
		ActivityFunc(activtyLocal).
		Build()

	tp, err := New[testIdentifier](
		context.Background(),
		registery,
		WithPath("./db/tempolite-workflow-cancel.db"),
		WithDestructive(),
	)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer tp.Close()

	if err := tp.registerWorkflow(localWrkflw); err != nil {
		t.Fatalf("RegisterWorkflow failed: %v", err)
	}

	tp.workflows.Range(func(key, value any) bool {
		// fmt.Println("key: ", key, "value: ", value)
		return true
	})

	var workflowInfo *WorkflowInfo[testIdentifier]
	if workflowInfo = tp.Workflow("test", localWrkflw, 1, workflowData{Message: "hello"}); err != nil {
		t.Fatalf("EnqueueActivityFunc failed: %v", err)
	}

	// fmt.Println("waiting 2s")
	<-time.After(2 * time.Second)

	if err := tp.CancelWorkflow(workflowInfo.WorkflowID); err != nil {
		t.Fatalf("CancelWorkflow failed: %v", err)
	}

	// fmt.Println("waiting until end")

	if err := tp.Wait(); err != nil {
		t.Fatalf("Wait failed: %v", err)
	}

}
