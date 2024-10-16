package tempolite

import (
	"context"
	"fmt"
	"log"
	"testing"
)

type testFailureSaga struct {
	OrderID string
}

func (p testFailureSaga) Transaction(ctx TransactionContext[string]) (interface{}, error) {
	log.Println("Starting transaction for testFailureSaga")
	return nil, fmt.Errorf("testFailureSaga failed")
}

func (p testFailureSaga) Compensation(ctx CompensationContext[string]) (interface{}, error) {
	log.Println("Compensating for testFailureSaga")
	return "FailureSaga", nil
}

// go test -timeout 30s -v -count=1 -run ^TestSagaSimple$ .
func TestSagaSimple(t *testing.T) {

	type workflowData struct {
		Message string
	}

	failure := false

	localWrkflw := func(ctx WorkflowContext[string], input int, msg workflowData) (int, error) {

		sagaBuilder := NewSaga[string]()
		sagaBuilder.AddStep(testOrderSaga{OrderID: "12345"})
		sagaBuilder.AddStep(testPaymentSaga{OrderID: "12345"})
		saga, err := sagaBuilder.Build()
		if err != nil {
			return 0, err
		}

		if err := ctx.Saga("saga test", saga).Get(); err != nil {
			return 0, err
		}

		if !failure {
			failure = true
			return 0, fmt.Errorf("testFailureSaga failed")
		}

		return 69, nil
	}

	tp, err := New[string](
		context.Background(),
		NewRegistry[string]().
			Workflow(localWrkflw).
			Build(),
		WithPath("./db/tempolite-workflow-saga.db"),
		WithDestructive(),
	)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer tp.Close()

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
	if number != 69 {
		t.Fatalf("number: %d", number)
	}
}
