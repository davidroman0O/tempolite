package tempolite

import (
	"context"
	"log"
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
