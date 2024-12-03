package tempolite

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
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

func TestTempoliteBasicCross(t *testing.T) {

	ctx := context.Background()
	db := NewMemoryDatabase()

	tp, err := New(
		ctx,
		db,
		WithDefaultQueueWorkers(10),
		WithQueue(QueueConfig{
			Name:        "second",
			WorkerCount: 10,
		}))
	if err != nil {
		t.Fatal(err)
	}

	counter := atomic.Int32{}

	var subWork func(ctx WorkflowContext) error

	subWork = func(ctx WorkflowContext) error {
		fmt.Println("Hello, second world!")
		<-time.After(1 * time.Second)
		if counter.Load() < 5 {
			counter.Store(counter.Load() + 1)
			fmt.Println("second ", counter.Load())
			return ctx.ContinueAsNew(&WorkflowOptions{
				Queue: "second",
			})
		}
		fmt.Println("second done!")
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
		fmt.Println("finished!!")
		return nil
	}

	future, err := tp.ExecuteDefault(wrkfl, nil)
	if err != nil {
		t.Fatal(err)
	}

	if err := future.Get(); err != nil {
		t.Fatal(err)
	}

	if err := tp.Wait(); err != nil {
		t.Fatal(err)
	}

	if err := tp.Close(); err != nil {
		t.Fatal(err)
	}
}
