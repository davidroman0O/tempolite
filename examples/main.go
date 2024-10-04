package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/davidroman0O/go-tempolite"
	_ "github.com/mattn/go-sqlite3"
)

// SimpleTask represents a simple task with a message
type SimpleTask struct {
	Message string
}

// SagaTask represents a task within a saga
type SagaTask struct {
	StepMessage string
}

// SimpleSideEffect implements the SideEffect interface
type SimpleSideEffect struct {
	Message string
}

func (s SimpleSideEffect) Run(ctx tempolite.SideEffectContext) (interface{}, error) {
	return "Side effect result for " + s.Message, nil
}

// SimpleHandler handles SimpleTask
func SimpleHandler(ctx tempolite.HandlerContext, task SimpleTask) error {
	log.Printf("Executing simple task: %s", task.Message)

	// Simulate some work
	time.Sleep(1 * time.Second)

	// Use a side effect
	result, err := ctx.SideEffect("simple-task-side-effect", SimpleSideEffect{Message: task.Message})
	if err != nil {
		return fmt.Errorf("side effect failed: %v", err)
	}

	log.Printf("Side effect result: %v", result)

	return nil
}

// SagaStep1 implements the SagaStep interface for step 1
type SagaStep1 struct {
	Message string
}

func (s SagaStep1) Transaction(ctx tempolite.TransactionContext) error {
	log.Printf("Executing saga step 1: %s", s.Message)
	time.Sleep(1 * time.Second)
	return nil
}

func (s SagaStep1) Compensation(ctx tempolite.CompensationContext) error {
	log.Printf("Compensating saga step 1: %s", s.Message)
	return nil
}

// SagaStep2 implements the SagaStep interface for step 2
type SagaStep2 struct {
	Message string
}

func (s SagaStep2) Transaction(ctx tempolite.TransactionContext) error {
	log.Printf("Executing saga step 2: %s", s.Message)
	time.Sleep(1 * time.Second)
	// Simulate a failure in step 2
	return fmt.Errorf("simulated failure in step 2")
}

func (s SagaStep2) Compensation(ctx tempolite.CompensationContext) error {
	log.Printf("Compensating saga step 2: %s", s.Message)
	return nil
}

// SagaHandler handles a saga with multiple steps
func SagaHandler(ctx tempolite.HandlerSagaContext, task SagaTask) error {
	log.Printf("Starting saga: %s", task.StepMessage)

	// Step 1
	err := ctx.Step(SagaStep1{Message: task.StepMessage + " - Step 1"})
	if err != nil {
		return err
	}

	// Step 2
	err = ctx.Step(SagaStep2{Message: task.StepMessage + " - Step 2"})
	if err != nil {
		return err
	}

	return nil
}

func main() {
	// Open SQLite database
	db, err := sql.Open("sqlite3", "tempolite.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create Tempolite instance
	tp, err := tempolite.New(context.Background(), db,
		tempolite.WithHandlerWorkers(2),
		tempolite.WithSagaWorkers(2),
		tempolite.WithCompensationWorkers(2),
		tempolite.WithSideEffectWorkers(2),
	)
	if err != nil {
		log.Fatalf("Failed to create Tempolite instance: %v", err)
	}
	defer tp.Close()

	// Register handlers
	tp.RegisterHandler(SimpleHandler)
	tp.RegisterHandler(SagaHandler)

	// Create a simple task
	simpleTaskID, err := tp.Enqueue(context.Background(), SimpleHandler, SimpleTask{Message: "Hello, Tempolite!"})
	if err != nil {
		log.Fatalf("Failed to enqueue simple task: %v", err)
	}
	log.Printf("Enqueued simple task with ID: %s", simpleTaskID)

	// // Create a saga
	// sagaTaskID, err := tp.EnqueueSaga(context.Background(), SagaHandler, SagaTask{StepMessage: "Saga steps"})
	// if err != nil {
	// 	log.Fatalf("Failed to enqueue saga: %v", err)
	// }
	// log.Printf("Enqueued saga with ID: %s", sagaTaskID)

	value, err := tp.WaitForTaskCompletion(context.Background(), simpleTaskID, time.Second)
	if err != nil {
		log.Fatalf("Failed to wait for task completion: %v", err)
	}
	log.Printf("Task completed with value: %v", value)

	// Wait for tasks to complete
	err = tp.Wait(func(info tempolite.TempoliteInfo) bool {
		return info.Tasks == 0 && info.SagaTasks == 0
	}, 500*time.Millisecond)
	if err != nil {
		log.Fatalf("Error waiting for tasks to complete: %v", err)
	}

	// Get results
	simpleTaskResult, err := tp.GetInfo(context.Background(), simpleTaskID)
	if err != nil {
		log.Fatalf("Failed to get simple task result: %v", err)
	}
	log.Printf("Simple task result: %+v", simpleTaskResult)

	// sagaTaskResult, err := tp.GetInfo(context.Background(), sagaTaskID)
	// if err != nil {
	// 	log.Fatalf("Failed to get saga task result: %v", err)
	// }
	// log.Printf("Saga task result: %+v", sagaTaskResult)

	// Print execution tree
	executionTree, err := tp.GetExecutionTree(context.Background(), simpleTaskID)
	if err != nil {
		log.Fatalf("Failed to get execution tree: %v", err)
	}
	log.Printf("Execution tree for simple task:\n%s", executionTree.String())

	// executionTree, err = tp.GetExecutionTree(context.Background(), sagaTaskID)
	// if err != nil {
	// 	log.Fatalf("Failed to get execution tree: %v", err)
	// }
	// log.Printf("Execution tree for saga task:\n%s", executionTree.String())
}
