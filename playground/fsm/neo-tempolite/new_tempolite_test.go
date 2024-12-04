package tempolite

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/davidroman0O/retrypool"
)

func TestQueueCrossBasic(t *testing.T) {

	ctx := context.Background()
	db := NewMemoryDatabase()
	registry := NewRegistry()

	var defaultQ *QueueInstance
	var secondQ *QueueInstance

	var onCross crossWorkflow = func(queueName string, workflowID int, workflowFunc interface{}, options *WorkflowOptions, args ...interface{}) Future {

		queue, err := db.GetQueueByName(queueName)
		if err != nil {
			future := NewRuntimeFuture()
			future.setError(err)
			return future
		}

		future := NewRuntimeFuture()

		if queue.Name == "default" {
			if err := defaultQ.orchestrators.Submit(
				retrypool.NewRequestResponse[*WorkflowRequest, *WorkflowResponse](&WorkflowRequest{
					workflowFunc: workflowFunc,
					options:      options,
					workflowID:   workflowID,
					args:         args,
					future:       future,
					queueName:    queueName,
					continued:    true,
				})); err != nil {
				future.setError(err)
			}
		} else if queue.Name == "second" {
			if err := secondQ.orchestrators.Submit(
				retrypool.NewRequestResponse[*WorkflowRequest, *WorkflowResponse](&WorkflowRequest{
					workflowFunc: workflowFunc,
					options:      options,
					workflowID:   workflowID,
					args:         args,
					future:       future,
					queueName:    queueName,
					continued:    true,
				})); err != nil {
				future.setError(err)
			}
		} else {
			future := NewRuntimeFuture()
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

	future, _, err := defaultQ.Submit(wrkfl, nil, nil)
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

	defer db.SaveAsJSON("./json/tempolite_cross_continueasnew.json")

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

	once := atomic.Bool{}
	once.Store(false)

	var subWork func(ctx WorkflowContext) error

	subWork = func(ctx WorkflowContext) error {
		fmt.Println("Hello, second world!", once.Load())
		<-time.After(1 * time.Second)
		if !once.Load() {
			once.Store(true)
			fmt.Println("continue second world")
			return ctx.ContinueAsNew(&WorkflowOptions{})
		}
		fmt.Println("second done!", once.Load())
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

func TestTempoliteBasicSubWorkflow(t *testing.T) {

	ctx := context.Background()
	db := NewMemoryDatabase()

	tp, err := New(
		ctx,
		db,
		WithDefaultQueueWorkers(10))
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
			return ctx.ContinueAsNew(&WorkflowOptions{})
		}
		fmt.Println("second done!")
		return nil
	}

	wrkfl := func(ctx WorkflowContext) error {
		fmt.Println("Hello, world!")
		if err := ctx.Workflow(
			"next",
			subWork,
			&WorkflowOptions{}).Get(); err != nil {
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

func TestTempoliteBasicSecondQueueSubWorkflow(t *testing.T) {

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
			if err := ctx.Workflow("second", subWork, &WorkflowOptions{}).Get(); err != nil {
				return err
			}
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

func TestFreeFlow(t *testing.T) {

	ctx := context.Background()
	db := NewMemoryDatabase()

	// defer db.SaveAsJSON("./json/tempolite_freeflow.json")

	tp, err := New(
		ctx,
		db,
		WithDefaultQueueWorkers(10),
		WithQueue(QueueConfig{
			Name:        "second",
			WorkerCount: 10,
		}),
	)
	if err != nil {
		t.Fatal(err)
	}

	counter := atomic.Int32{}

	subWorkflowFunc := func(ctx WorkflowContext) error {
		fmt.Println("Hello, second world!")
		if counter.Load() < 5 {
			counter.Store(counter.Load() + 1)
			fmt.Println("second ", counter.Load())
			return ctx.ContinueAsNew(nil)
		}
		fmt.Println("second done!")
		return nil
	}

	workflowFunc := func(ctx WorkflowContext) error {
		fmt.Println("Hello, world!")

		if err := ctx.Workflow(
			"next",
			subWorkflowFunc,
			&WorkflowOptions{
				Queue: "second",
			}).Get(); err != nil {
			return err
		}

		fmt.Println("finished!!")
		return nil
	}

	future, err := tp.ExecuteDefault(workflowFunc, nil)
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

func TestTempoliteSignal(t *testing.T) {
	ctx := context.Background()
	db := NewMemoryDatabase()

	// defer db.SaveAsJSON("./json/tempolite_freeflow.json")

	tp, err := New(
		ctx,
		db,
		WithDefaultQueueWorkers(1),
	)
	if err != nil {
		t.Fatal(err)
	}

	workflowFunc := func(ctx WorkflowContext) error {
		fmt.Println("Hello, world!")
		var life int
		if err := ctx.Signal("life", &life); err != nil {
			return err
		}
		fmt.Println("finished!!", life)
		return nil
	}

	future, err := tp.ExecuteDefault(workflowFunc, nil)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		<-time.After(1 * time.Second)
		tp.PublishSignal(future.WorkflowID(), "life", 42)
	}()

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
