package tempolite

import (
	"context"
	"fmt"
	"testing"
)

func TestUnitPrepareRootWorkflowEntity(t *testing.T) {

	ctx := context.Background()
	db := NewMemoryDatabase()
	registry := NewRegistry()

	o := NewOrchestrator(ctx, db, registry)

	wrfl := func(ctx WorkflowContext) error {
		fmt.Println("Hello, World!")
		return nil
	}

	future := o.Execute(wrfl, nil)

	if future == nil {
		t.Fatal("future is nil")
	}

	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

}
