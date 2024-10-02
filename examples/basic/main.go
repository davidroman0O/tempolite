package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/davidroman0O/comfylite3"
	"github.com/davidroman0O/go-tempolite"
	"github.com/google/uuid"
	// _ "github.com/mattn/go-sqlite3"
)

type EmailParams struct {
	Recipient string
	Subject   string
	Body      string
}

func EmailHandler(ctx context.Context, payload EmailParams) ([]byte, error) {

	// Simulate sending email
	log.Printf("Sending email to %s with subject %s\n", payload.Recipient, payload.Subject)
	// ... actual email sending logic

	return []byte("Email sent"), nil
}

func ComplexHandler(ctx context.Context, payload map[string]interface{}) ([]byte, error) {
	tc := ctx.(*tempolite.TaskContext)

	// Record a side effect
	result, err := tc.RecordSideEffect("compute_something", func() (interface{}, error) {
		// Simulate computation
		log.Println("Computing something...")
		time.Sleep(2 * time.Second)
		return "ComputedResult", nil
	})
	if err != nil {
		return nil, err
	}
	log.Printf("Result: %v\n", result)

	// Enqueue another task and wait for its completion
	emailResult, err := tc.EnqueueTaskAndWait(
		EmailHandler,
		EmailParams{
			Recipient: "user@example.com",
			Subject:   "Processing Complete",
			Body:      fmt.Sprintf("Your result is: %v", result),
		})
	if err != nil {
		return nil, err
	}
	log.Printf("Email task result: %s\n", string(emailResult))

	// Simulate failure
	if tc.GetID()%2 == 0 && tc.RetryCount() == 0 {
		log.Println("Simulating failure...")
		return nil, fmt.Errorf("simulated failure")
	}

	// Wait for signals named "update" for up to 15 seconds or until "done" command is received
	log.Println("Waiting for signals named 'update' for up to 15 seconds...")
	signalCh, err := tc.ReceiveSignal("update")
	if err != nil {
		return nil, err
	}
	timeout := time.After(15 * time.Second)
	for {
		select {
		case <-tc.Done():
			log.Println("Context done")
			return nil, tc.Err()
		case <-timeout:
			log.Println("Timeout waiting for signals")
			// Proceed after timeout
			return []byte(fmt.Sprintf("Task %d completed without signals", tc.GetID())), nil
		case payload, ok := <-signalCh:
			if !ok {
				log.Println("Signal channel closed")
				return []byte(fmt.Sprintf("Task %d completed", tc.GetID())), nil
			}
			log.Printf("Received signal payload")
			// Handle signal payload
			var data map[string]interface{}
			json.Unmarshal(payload, &data)
			if data["command"] == "done" {
				log.Println("Received 'done' command, exiting signal loop")
				// Optionally send acknowledgment
				err := tc.SendSignal("ack", map[string]interface{}{"status": "received"})
				if err != nil {
					return nil, err
				}
				return []byte(fmt.Sprintf("Task %d completed after receiving signals", tc.GetID())), nil
			}
			// Optionally send a signal back
			err := tc.SendSignal("ack", map[string]interface{}{"status": "processing"})
			if err != nil {
				return nil, err
			}
		}
	}
}

func init() {
	tempolite.RegisterHandler(EmailHandler)
	tempolite.RegisterHandler(ComplexHandler)
}

func main() {
	// comfy, err := comfylite3.New(comfylite3.WithMemory())
	comfy, err := comfylite3.New(comfylite3.WithPath("tempolite.db"))
	if err != nil {
		panic(err)
	}

	db := comfylite3.OpenDB(comfy, comfylite3.WithForeignKeys())

	defer db.Close()
	defer comfy.Close()

	// db, err := sql.Open("sqlite3", "tasks.db")
	// if err != nil {
	// 	log.Fatalf("Failed to open database: %v", err)
	// }
	// defer db.Close()

	// Create repositories
	taskRepo, err := tempolite.NewSQLiteTaskRepository(db)
	if err != nil {
		log.Fatalf("Failed to create task repository: %v", err)
	}

	sideEffectRepo, err := tempolite.NewSQLiteSideEffectRepository(db)
	if err != nil {
		log.Fatalf("Failed to create side effect repository: %v", err)
	}

	signalRepo, err := tempolite.NewSQLiteSignalRepository(db)
	if err != nil {
		log.Fatalf("Failed to create signal repository: %v", err)
	}

	// Create WorkflowBox
	ctx := context.Background()
	wb, err := tempolite.New(ctx, taskRepo, sideEffectRepo, signalRepo)
	if err != nil {
		log.Fatalf("Failed to initialize WorkflowBox: %v", err)
	}

	// Start processing with 5 workers
	wb.Start(5)

	// Enqueue a complex task
	executionContextID := "workflow-" + uuid.NewString()
	taskID, err := wb.EnqueueTask(ctx, executionContextID, ComplexHandler, map[string]interface{}{})
	if err != nil {
		log.Printf("Failed to enqueue complex task: %v", err)
	}

	// Example of sending a signal to a task and receiving a response
	go func() {
		log.Println("Waiting before sending signal...")
		time.Sleep(5 * time.Second)
		log.Println("Sending signal to task...")
		err := wb.SendSignal(ctx, taskID, "update", map[string]interface{}{"command": "done"})
		if err != nil {
			log.Printf("Failed to send signal: %v", err)
			return
		}

		defer func() {
			data, err := wb.WaitForTaskCompletion(ctx, taskID)
			if err != nil {
				log.Printf("Failed to wait for task completion: %v", err)
				return
			}
			log.Printf("Task completion result: %s\n", string(data))
		}()

		// Wait for acknowledgment
		log.Println("Waiting for acknowledgment from task...")
		ackCh, err := wb.ReceiveSignal(ctx, taskID, "ack")
		if err != nil {
			log.Printf("Failed to set up ReceiveSignal: %v", err)
			return
		}
		timeoutAck := time.After(10 * time.Second)
		select {
		case <-timeoutAck:
			log.Println("Timeout waiting for acknowledgment")
		case payload := <-ackCh:
			log.Printf("Received acknowledgment from task")
			var data map[string]interface{}
			json.Unmarshal(payload, &data)
			log.Printf("Signal data: %v", data)
		}
	}()

	// Wait indefinitely
	select {}
}
