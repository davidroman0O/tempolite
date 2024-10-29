package tests

import (
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/davidroman0O/tempolite"
)

type TestTransactionStep struct {
	ID                string
	ShouldFail        bool
	TransactionCount  *atomic.Int32
	CompensationCount *atomic.Int32
}

func (s *TestTransactionStep) Transaction(ctx tempolite.TransactionContext) (interface{}, error) {
	s.TransactionCount.Add(1)
	if s.ShouldFail {
		return nil, fmt.Errorf("transaction %s failed", s.ID)
	}
	return fmt.Sprintf("transaction %s completed", s.ID), nil
}

func (s *TestTransactionStep) Compensation(ctx tempolite.CompensationContext) (interface{}, error) {
	s.CompensationCount.Add(1)
	return fmt.Sprintf("compensation %s completed", s.ID), nil
}

// go test -timeout 30s -run ^TestSaga$ ./tests
func TestSaga(t *testing.T) {

	// go test -v -count=1 -timeout 30s -run ^TestSaga$/^successful_saga_execution$ ./tests
	t.Run("successful saga execution", func(t *testing.T) {
		step1Counter := &atomic.Int32{}
		step2Counter := &atomic.Int32{}
		step1Comp := &atomic.Int32{}
		step2Comp := &atomic.Int32{}

		step1 := &TestTransactionStep{
			ID:                "step1",
			TransactionCount:  step1Counter,
			CompensationCount: step1Comp,
		}
		step2 := &TestTransactionStep{
			ID:                "step2",
			TransactionCount:  step2Counter,
			CompensationCount: step2Comp,
		}

		workflowFn := func(ctx tempolite.WorkflowContext) error {
			sagaBuilder := tempolite.NewSaga()
			sagaBuilder.AddStep(step1)
			sagaBuilder.AddStep(step2)
			saga, err := sagaBuilder.Build()
			if err != nil {
				fmt.Println("Error building saga", err)
				return err
			}

			return ctx.Saga("test-saga", saga).Get()
		}

		cfg := NewTestConfig(t, "successful-saga")
		cfg.Registry.Workflow(workflowFn)

		tp := NewTestTempolite(t, cfg)
		defer tp.Close()

		err := tp.Workflow(workflowFn, nil).Get()
		if err != nil {
			t.Fatalf("Workflow execution failed: %v", err)
		}

		if step1Counter.Load() != 1 {
			t.Errorf("Step 1 transaction count: got %d, want 1", step1Counter.Load())
		}
		if step2Counter.Load() != 1 {
			t.Errorf("Step 2 transaction count: got %d, want 1", step2Counter.Load())
		}
		if step1Comp.Load() != 0 {
			t.Errorf("Step 1 compensation count: got %d, want 0", step1Comp.Load())
		}
		if step2Comp.Load() != 0 {
			t.Errorf("Step 2 compensation count: got %d, want 0", step2Comp.Load())
		}
	})

	// go test -v -count=1 -timeout 30s -run ^TestSaga$/^saga_with_compensation$ ./tests
	t.Run("saga with compensation", func(t *testing.T) {
		step1Counter := &atomic.Int32{}
		step2Counter := &atomic.Int32{}
		step1Comp := &atomic.Int32{}
		step2Comp := &atomic.Int32{}

		step1 := &TestTransactionStep{
			ID:                "step1",
			TransactionCount:  step1Counter,
			CompensationCount: step1Comp,
		}
		step2 := &TestTransactionStep{
			ID:                "step2",
			ShouldFail:        true,
			TransactionCount:  step2Counter,
			CompensationCount: step2Comp,
		}

		workflowFn := func(ctx tempolite.WorkflowContext) error {
			sagaBuilder := tempolite.NewSaga()
			sagaBuilder.AddStep(step1)
			sagaBuilder.AddStep(step2)
			saga, err := sagaBuilder.Build()
			if err != nil {
				fmt.Println("Error building saga", err)
				return err
			}

			return ctx.Saga("test-saga", saga).Get()
		}

		cfg := NewTestConfig(t, "saga-with-compensation")
		cfg.Registry.Workflow(workflowFn)

		tp := NewTestTempolite(t, cfg)
		defer tp.Close()

		// If we want the Saga to fail for real with the workflow, we need to set the retry policy to 0, otherwise it will be retried again with the retry of the workflow.
		err := tp.Workflow(workflowFn, tempolite.WorkflowConfig(tempolite.WithWorkflowRetryMaximumAttempts(0))).Get()
		if err == nil {
			t.Fatal("Expected workflow to fail")
		}

		if step1Counter.Load() != 1 {
			t.Errorf("Step 1 transaction count: got %d, want 1", step1Counter.Load())
		}
		if step2Counter.Load() != 1 {
			t.Errorf("Step 2 transaction count: got %d, want 1", step2Counter.Load())
		}
		if step1Comp.Load() != 1 {
			t.Errorf("Step 1 compensation count: got %d, want 1", step1Comp.Load())
		}
		if step2Comp.Load() != 0 {
			t.Errorf("Step 2 compensation count: got %d, want 0", step2Comp.Load())
		}
	})
}
