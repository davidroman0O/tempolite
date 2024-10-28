package main

import (
	"context"
	"fmt"
	"log"

	"github.com/davidroman0O/tempolite"
)

// Example workflow response types
type TrackResponse struct {
	TrackID     string
	Status      string
	Destination string
}

type ProcessResponse struct {
	OrderID string
	Result  string
}

// Example workflow handlers
func TrackWorkflow(ctx tempolite.WorkflowContext, trackID string) (TrackResponse, error) {
	return TrackResponse{
		TrackID:     trackID,
		Status:      "In Transit",
		Destination: "New York",
	}, nil
}

func ProcessWorkflow(ctx tempolite.WorkflowContext, orderID string) (ProcessResponse, error) {
	return ProcessResponse{
		OrderID: orderID,
		Result:  "Completed",
	}, nil
}

func main() {
	// Initialize Tempolite
	tp, err := tempolite.New(
		context.Background(),
		tempolite.NewRegistry().
			Workflow(TrackWorkflow).
			Workflow(ProcessWorkflow).
			Build(),
		tempolite.WithPath("./db/tempolite-switch.db"),
		tempolite.WithDestructive(),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer tp.Close()

	// Start a tracking workflow
	info := tp.Workflow(TrackWorkflow, nil, "TRK123")

	// Handle the workflow response using the switch pattern
	if err := info.Switch().
		Case(TrackWorkflow, func() error {
			var response TrackResponse
			if err := info.Get(&response); err != nil {
				return fmt.Errorf("failed to get track response: %w", err)
			}
			fmt.Printf("Track Status: %s, Destination: %s\n", response.Status, response.Destination)
			return nil
		}).
		Case(ProcessWorkflow, func() error {
			var response ProcessResponse
			if err := info.Get(&response); err != nil {
				return fmt.Errorf("failed to get process response: %w", err)
			}
			fmt.Printf("Process Result: %s\n", response.Result)
			return nil
		}).
		Default(func() error {
			return fmt.Errorf("unknown workflow type")
		}).
		End(); err != nil {
		log.Printf("Error handling workflow: %v", err)
	}

	if err := tp.Wait(); err != nil {
		log.Fatal(err)
	}
}
