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
	return "PaymentCompensated", nil
}

// go test -timeout 30s -v -count=1 -run ^TestSagaSimple$ .
func TestSagaSimple(t *testing.T) {

	tp, err := New[string](
		context.Background(),
		WithPath("./db/tempolite-workflow-sideeffect.db"),
		WithDestructive(),
	)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer tp.Close()

	type workflowData struct {
		Message string
	}

	localWrkflw := func(ctx WorkflowContext[string], input int, msg workflowData) (int, error) {

		sagaBuilder := NewSaga[string]()
		sagaBuilder.AddStep(testOrderSaga{OrderID: "12345"})
		sagaBuilder.AddStep(testPaymentSaga{OrderID: "12345"})
		sagaBuilder.AddStep(testFailureSaga{OrderID: "12345"})
		saga, err := sagaBuilder.Build()
		if err != nil {
			t.Fatalf("Failed to build saga: %v", err)
		}

		if err := ctx.Saga("saga test", saga).Get(); err != nil {
			t.Fatalf("EnqueueActivityFunc failed: %v", err)
		}

		return 69, nil
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
	if number != 69 {
		t.Fatalf("number: %d", number)
	}
}
