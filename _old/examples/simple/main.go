package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/davidroman0O/comfylite3"
	"github.com/davidroman0O/go-tempolite"
	"github.com/google/uuid"
	// _ "github.com/mattn/go-sqlite3"
)

type SomeAgent struct {
}

type SomeID struct {
	ID string `json:"id"`
}

func (s *SomeAgent) trackHandler(ctx context.Context, payload SomeID) (interface{}, error) {
	// var identifier *SomeID
	// fmt.Println("TrackHandler", string(payload))
	// if err := json.Unmarshal(payload, identifier); err != nil {
	// 	return nil, err
	// }
	fmt.Println("TrackHandler", payload)
	return payload, nil
}

func handler(ctx context.Context, payload []byte) ([]byte, error) {
	var identifier SomeID
	fmt.Println("TrackHandler", string(payload))
	if err := json.Unmarshal(payload, &identifier); err != nil {
		fmt.Println("TrackHandler failed to unmsharll", err)
		return nil, err
	}
	fmt.Println("TrackHandler", identifier)
	return []byte("Task completed successfully"), nil
}

func main() {
	// comfy, err := comfylite3.New(comfylite3.WithMemory())
	comfy, err := comfylite3.New(comfylite3.WithPath("tempolite.db"))
	if err != nil {
		panic(err)
	}

	db := comfylite3.OpenDB(comfy)

	defer db.Close()
	defer comfy.Close()

	// db, err := sql.Open("sqlite3", "tasks.db")
	// if err != nil {
	// 	log.Fatalf("Failed to open database: %v", err)
	// }
	// defer db.Close()

	// Create WorkflowBox
	ctx := context.Background()
	wb, err := tempolite.New(ctx, db, 3, 3, 3, 3)
	if err != nil {
		log.Fatalf("Failed to initialize WorkflowBox: %v", err)
	}

	agent := SomeAgent{}

	tempolite.RegisterHandler(agent.trackHandler)
	tempolite.RegisterHandler(handler)

	// Enqueue a complex task
	executionContextID := "workflow-" + uuid.NewString()
	taskID, err := wb.EnqueueTask(ctx, executionContextID, agent.trackHandler, SomeID{ID: "123"})
	if err != nil {
		log.Printf("Failed to enqueue complex task: %v", err)
		panic(err)
	}
	var data []byte

	if data, err = wb.WaitForTaskCompletion(ctx, taskID); err != nil {
		log.Printf("Failed to wait for task completion: %v", err)
		panic(err)
	}

	fmt.Println("Data", string(data))

	// Wait indefinitely
	select {}
}
