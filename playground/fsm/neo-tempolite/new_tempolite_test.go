package tempolite

import (
	"context"
	"fmt"
	"testing"
)

func TestQueueCrossBasic(t *testing.T) {

	ctx := context.Background()
	db := NewMemoryDatabase()
	registry := NewRegistry()

	var defaultQ *QueueInstance
	var secondQ *QueueInstance

	var onCross crossWorkflow = func(queueName string, workflowFunc interface{}, options *WorkflowOptions, args ...interface{}) Future {

		queue, err := db.GetQueueByName(queueName)
		if err != nil {
			future := NewDatabaseFuture(ctx, db, registry)
			future.setError(err)
			return future
		}

		var future Future
		if queue.Name == "default" {
			future, _, err = defaultQ.Submit(workflowFunc, options, args...)
		} else if queue.Name == "second" {
			future, _, err = secondQ.Submit(workflowFunc, options, args...)
		} else {
			future := NewDatabaseFuture(ctx, db, registry)
			future.setError(fmt.Errorf("queue %s does not exist", queueName))
		}
		if err != nil {
			future.setError(err)
		}

		return future
	}

	var err error

	defaultQ, err = NewQueueInstance(ctx, db, registry, "default", 1, WithCrossWorkflowHandler(onCross))
	if err != nil {
		t.Fatal(err)
	}

	secondQ, err = NewQueueInstance(ctx, db, registry, "second", 1, WithCrossWorkflowHandler(onCross))
	if err != nil {
		t.Fatal(err)
	}

	subWork := func(ctx WorkflowContext) error {
		fmt.Println("Hello, second world!")
		return nil
	}

	wrkfl := func(ctx WorkflowContext) error {
		fmt.Println("Hello, world!")
		if err := ctx.Workflow(
			"next",
			subWork,
			&WorkflowOptions{
				Queue: "second",
			}).Get(); err != nil {
			return err
		}
		return nil
	}

	future, _, err := defaultQ.Submit(wrkfl, nil)
	if err != nil {
		t.Fatal(err)
	}

	if err := future.Get(); err != nil {
		t.Fatal(err)
	}
}
