package data

import (
	"sync/atomic"
	"testing"
)

func TestBuilder(t *testing.T) {

	var transacCalled atomic.Bool
	var compenCalled atomic.Bool

	saga, err := NewSaga().
		Add(func(tc TransactionContext) error {
			transacCalled.Store(true)
			return nil
		}, func(cc CompensationContext) error {
			compenCalled.Store(true)
			return nil
		}).Build()

	if err != nil {
		t.Fatal(err)
	}

	if len(saga.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(saga.Steps))
	}

	if err := saga.Steps[0].Transaction(nil); err != nil {
		t.Fatal(err)
	}

	if err := saga.Steps[0].Compensation(nil); err != nil {
		t.Fatal(err)
	}

	if !transacCalled.Load() {
		t.Fatal("transaction function not called")
	}

	if !compenCalled.Load() {
		t.Fatal("compensation function not called")
	}
}
