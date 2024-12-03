package tempolite

import (
	"context"
	"fmt"
	"testing"
)

func TestQueueBasic(t *testing.T) {

	ctx := context.Background()
	db := NewMemoryDatabase()
	registry := NewRegistry()
	q := NewQueueInstance(ctx, db, registry, "default", 1)

	wrkfl := func(ctx WorkflowContext) error {
		fmt.Println("Hello, world!")
		return nil
	}

	future, _, err := q.Submit(wrkfl, nil)
	if err != nil {
		t.Fatal(err)
	}

	if err := future.Get(); err != nil {
		t.Fatal(err)
	}
}
