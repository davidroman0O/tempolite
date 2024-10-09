package tempolite

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"testing"
	"time"
)

type TaskInt struct {
	Data int
}

func simpleTestHandler(ctx HandlerContext, task TaskInt) (interface{}, error) {
	log.Printf("simpleTestHandler Task %d", task)
	return nil, nil
}

// go test -v -timeout 30s -run ^TestHandlerSimple$ .
func TestHandlerSimple(t *testing.T) {

	ctx := context.Background()

	tp, err := New(ctx,
		WithDestructive(),
		WithPath("./db/tempolite_handler_simple_test.db"),
	)
	if err != nil {
		t.Fatalf("Error creating Tempolite: %v", err)
	}
	defer tp.Close()

	if err := tp.RegisterHandler(simpleTestHandler); err != nil {
		t.Fatalf("Error registering handler: %v", err)
	}

	var id string
	if id, err = tp.Enqueue(ctx, simpleTestHandler, TaskInt{1}); err != nil {
		t.Fatalf("Error enqueuing task: %v", err)
	}

	log.Printf("ID: %s", id)

	ctxTimeout := context.Background()
	ctxTimeout, cancel := context.WithTimeout(ctxTimeout, 5*time.Second)
	defer cancel()

	data, err := tp.WaitFor(ctxTimeout, id)
	if err != nil {
		t.Fatalf("Error waiting for task: %v", err)
	}

	log.Printf("Data: %v", data)

	tp.Wait(func(ti TempoliteInfo) bool {
		log.Println(ti)
		return ti.IsCompleted()
	}, time.Second)
}

/////////////////////////////////////////////////////////////////////////////////////////////////

func parentHandler(ctx HandlerContext, task TaskInt) (interface{}, error) {
	var err error
	var id string

	if id, err = ctx.Enqueue(simpleTestHandler, TaskInt{task.Data + 2}); err != nil {
		return nil, err
	}

	log.Printf("parentHandler Task %d", task)

	return ctx.WaitFor(id)
}

// go test -v -timeout 30s -run ^TestHandlerChildren$ .
func TestHandlerChildren(t *testing.T) {

	ctx := context.Background()

	tp, err := New(ctx,
		WithDestructive(),
		WithPath("./db/tempolite_handler_children_test.db"),
	)
	if err != nil {
		t.Fatalf("Error creating Tempolite: %v", err)
	}
	defer tp.Close()

	if err := tp.RegisterHandler(simpleTestHandler); err != nil {
		t.Fatalf("Error registering handler: %v", err)
	}

	if err := tp.RegisterHandler(parentHandler); err != nil {
		t.Fatalf("Error registering handler: %v", err)
	}

	var id string
	if id, err = tp.Enqueue(ctx, parentHandler, TaskInt{1}); err != nil {
		t.Fatalf("Error enqueuing task: %v", err)
	}

	log.Printf("ID: %s", id)

	ctxTimeout := context.Background()
	ctxTimeout, cancel := context.WithTimeout(ctxTimeout, 5*time.Second)
	defer cancel()

	data, err := tp.WaitFor(ctxTimeout, id)
	if err != nil {
		t.Fatalf("Error waiting for task: %v", err)
	}

	log.Printf("Data: %v", data)

	tp.Wait(func(ti TempoliteInfo) bool {
		log.Println(ti)
		return ti.IsCompleted()
	}, time.Second)
}

/////////////////////////////////////////////////////////////////////////////////////////////////

func simpleRetry(ctx HandlerContext, task TaskInt) (interface{}, error) {
	if retryFlag.Load() {
		retryFlag.Store(false)
		return nil, fmt.Errorf("failed on purpose")
	}

	log.Printf("simpleRetry Task %d", task)

	return nil, nil
}

// go test -v -timeout 30s -run ^TestHandlerSimpleRetry$ .
func TestHandlerSimpleRetry(t *testing.T) {

	ctx := context.Background()

	tp, err := New(ctx,
		WithDestructive(),
		WithPath("./db/tempolite_handler_retry_test.db"),
	)
	if err != nil {
		t.Fatalf("Error creating Tempolite: %v", err)
	}
	defer tp.Close()

	retryFlag.Store(true)

	if err := tp.RegisterHandler(simpleTestHandler); err != nil {
		t.Fatalf("Error registering handler: %v", err)
	}

	if err := tp.RegisterHandler(simpleRetry); err != nil {
		t.Fatalf("Error registering handler: %v", err)
	}

	var id string
	if id, err = tp.Enqueue(ctx, simpleRetry, TaskInt{1}); err != nil {
		t.Fatalf("Error enqueuing task: %v", err)
	}

	log.Printf("ID: %s", id)

	ctxTimeout := context.Background()
	ctxTimeout, cancel := context.WithTimeout(ctxTimeout, 5*time.Second)
	defer cancel()

	data, err := tp.WaitFor(ctxTimeout, id)
	if err != nil {
		t.Fatalf("Error waiting for task: %v", err)
	}

	log.Printf("Data: %v", data)

	tp.Wait(func(ti TempoliteInfo) bool {
		log.Println(ti)
		return ti.IsCompleted()
	}, time.Second)
}

/////////////////////////////////////////////////////////////////////////////////////////////////

var retryFlag atomic.Bool

func simpleHandlerChildrenRetry(ctx HandlerContext, task TaskInt) (interface{}, error) {
	log.Printf("simpleHandlerChildrenRetry Task %d", task)
	if retryFlag.Load() {
		retryFlag.Store(false)
		fmt.Println("retryFlag set to false")
		return nil, fmt.Errorf("failied on purpose")
	}
	return nil, nil
}

func parentHandlerRetry(ctx HandlerContext, task TaskInt) (interface{}, error) {
	var err error
	var id string

	if id, err = ctx.Enqueue(simpleHandlerChildrenRetry, TaskInt{task.Data + 2}); err != nil {
		return nil, err
	}

	log.Printf("parentHandler Task %d", task)

	return ctx.WaitFor(id)
}

// go test -v -timeout 30s -run ^TestHandlerRetriesChildren$ .
func TestHandlerRetriesChildren(t *testing.T) {

	ctx := context.Background()

	tp, err := New(ctx,
		WithDestructive(),
		WithPath("./db/tempolite_handler_children_retries_test.db"),
	)
	if err != nil {
		t.Fatalf("Error creating Tempolite: %v", err)
	}
	defer tp.Close()

	retryFlag.Store(true)

	if err := tp.RegisterHandler(simpleHandlerChildrenRetry); err != nil {
		t.Fatalf("Error registering handler: %v", err)
	}

	if err := tp.RegisterHandler(parentHandlerRetry); err != nil {
		t.Fatalf("Error registering handler: %v", err)
	}

	var id string
	if id, err = tp.Enqueue(ctx, parentHandlerRetry, TaskInt{1}); err != nil {
		t.Fatalf("Error enqueuing task: %v", err)
	}

	log.Printf("ID: %s", id)

	ctxTimeout := context.Background()
	ctxTimeout, cancel := context.WithTimeout(ctxTimeout, 5*time.Second)
	defer cancel()

	data, err := tp.WaitFor(ctxTimeout, id)
	if err != nil {
		t.Fatalf("Error waiting for task: %v", err)
	}

	log.Printf("Data: %v", data)

	tp.Wait(func(ti TempoliteInfo) bool {
		log.Println(ti)
		return ti.IsCompleted()
	}, time.Second)
}

func handlerChildrenHandler(ctx HandlerContext, task TaskInt) (interface{}, error) {
	log.Printf("handlerChildrenHandler Task %d", task)
	return nil, nil
}

func handlerParentRetry(ctx HandlerContext, task TaskInt) (interface{}, error) {
	var err error
	var id string

	if id, err = ctx.Enqueue(handlerChildrenHandler, TaskInt{task.Data + 2}); err != nil {
		return nil, err
	}

	if retryFlag.Load() {
		retryFlag.Store(false)
		fmt.Println("retryFlag set to false")
		return nil, fmt.Errorf("failied on purpose")
	}

	log.Printf("parentHandler Task %d", task)

	return ctx.WaitFor(id)
}

// go test -v -timeout 30s -run ^TestHandlerRetriesParent$ .
func TestHandlerRetriesParent(t *testing.T) {

	ctx := context.Background()

	// defer func() {
	// 	if r := recover(); r != nil {
	// 		log.Printf("Recovered from panic: %v", r)
	// 		fmt.Println(ctx.Err())
	// 	}
	// }()

	tp, err := New(ctx,
		WithDestructive(),
		WithPath("./db/tempolite_handler_parent_retries_test.db"),
	)
	if err != nil {
		t.Fatalf("Error creating Tempolite: %v", err)
	}
	defer tp.Close()

	retryFlag.Store(true)

	if err := tp.RegisterHandler(handlerChildrenHandler); err != nil {
		t.Fatalf("Error registering handler: %v", err)
	}

	if err := tp.RegisterHandler(handlerParentRetry); err != nil {
		t.Fatalf("Error registering handler: %v", err)
	}

	var id string
	if id, err = tp.Enqueue(ctx, handlerParentRetry, TaskInt{1}); err != nil {
		t.Fatalf("Error enqueuing task: %v", err)
	}

	log.Printf("ID: %s", id)

	ctxTimeout := context.Background()
	ctxTimeout, cancel := context.WithTimeout(ctxTimeout, 15*time.Second)
	defer cancel()

	data, err := tp.WaitFor(ctxTimeout, id)
	if err != nil {
		t.Fatalf("Error waiting for task: %v", err)
	}

	log.Printf("Data: %v", data)

	tp.Wait(func(ti TempoliteInfo) bool {
		log.Println(ti)
		return ti.IsCompleted()
	}, time.Second)
	fmt.Println("end wait")
}

func dynamicHandlerChildrenHandler(ctx HandlerContext, task TaskInt) (interface{}, error) {
	log.Printf("handlerChildrenHandler Task %d", task)
	return nil, nil
}

func dynamicHandlerSecondChildrenHandler(ctx HandlerContext, task TaskInt) (interface{}, error) {
	log.Printf("handlerChildrenHandler Task %d", task)
	return nil, nil
}

func dynamicHandlerParentRetry(ctx HandlerContext, task TaskInt) (interface{}, error) {
	var err error
	var id string

	if retryFlag.Load() {
		if id, err = ctx.Enqueue(dynamicHandlerSecondChildrenHandler, TaskInt{task.Data + 2}); err != nil {
			return nil, err
		}
		if _, err = ctx.WaitFor(id); err != nil {
			return nil, err
		}
	}

	if id, err = ctx.Enqueue(dynamicHandlerChildrenHandler, TaskInt{task.Data + 2}); err != nil {
		return nil, err
	}

	if !retryFlag.Load() {
		retryFlag.Store(true)
		fmt.Println("retryFlag set to false")
		return nil, fmt.Errorf("failied on purpose")
	}

	log.Printf("parentHandler Task %d", task)

	return ctx.WaitFor(id)
}

// go test -v -timeout 30s -run ^TestHandlerDynamicRetriesParent$ .
func TestHandlerDynamicRetriesParent(t *testing.T) {

	ctx := context.Background()

	tp, err := New(ctx,
		WithDestructive(),
		WithPath("./db/tempolite_dynamic_handler_parent_retries_test.db"),
	)
	if err != nil {
		t.Fatalf("Error creating Tempolite: %v", err)
	}
	defer tp.Close()

	retryFlag.Store(false)

	if err := tp.RegisterHandler(dynamicHandlerChildrenHandler); err != nil {
		t.Fatalf("Error registering handler: %v", err)
	}
	if err := tp.RegisterHandler(dynamicHandlerSecondChildrenHandler); err != nil {
		t.Fatalf("Error registering handler: %v", err)
	}

	if err := tp.RegisterHandler(dynamicHandlerParentRetry); err != nil {
		t.Fatalf("Error registering handler: %v", err)
	}

	var id string
	if id, err = tp.Enqueue(ctx, dynamicHandlerParentRetry, TaskInt{1}); err != nil {
		t.Fatalf("Error enqueuing task: %v", err)
	}

	log.Printf("ID: %s", id)

	ctxTimeout := context.Background()
	ctxTimeout, cancel := context.WithTimeout(ctxTimeout, 5*time.Second)
	defer cancel()

	data, err := tp.WaitFor(ctxTimeout, id)
	if err != nil {
		t.Fatalf("Error waiting for task: %v", err)
	}

	log.Printf("Data: %v", data)

	tp.Wait(func(ti TempoliteInfo) bool {
		log.Println(ti)
		return ti.IsCompleted()
	}, time.Second)
}
