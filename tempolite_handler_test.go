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

func TestHandlerSimple(t *testing.T) {

	ctx := context.Background()

	tp, err := New(ctx,
		WithDestructive(),
		WithPath("tempolite_test.db"),
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

	tp.Wait(func(ti TempoliteInfo) bool {
		log.Println(ti)
		return ti.IsCompleted()
	}, time.Second)
}
