package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/davidroman0O/tempolite"
)

type ExampleID string

type StorageActivity struct{}

var storage = StorageActivity{}

func (s StorageActivity) Run(ctx tempolite.ActivityContext[ExampleID], iteration int) (int, error) {
	return iteration * 2, nil
}

func simpleWorkflow(ctx tempolite.WorkflowContext[ExampleID], iteration int) (int, error) {
	var result int
	if err := ctx.Activity("process", storage.Run, iteration).Get(&result); err != nil {
		return 0, err
	}
	return result, nil // This will be stored in workflow execution output
}

func main() {
	dir := "./db/tempolite-pool-demo"
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}
	log.Printf("Database directory: %s", dir)

	registry := tempolite.NewRegistry[ExampleID]().
		Workflow(simpleWorkflow).
		Activity(storage.Run).
		Build()

	poolOpts := tempolite.TempolitePoolOptions{
		MaxFileSize:  1024 * 50,
		MaxPageCount: 50,
		BaseFolder:   dir,
		BaseName:     "demo",
	}

	pool, err := tempolite.NewTempolitePool[ExampleID](
		context.Background(),
		registry,
		poolOpts,
		tempolite.WithDestructive(),
	)
	if err != nil {
		log.Fatalf("Failed to create pool: %v", err)
	}
	defer pool.Close()

	// Start multiple workflows
	for i := 1; i <= 1000; i++ {
		var result int
		if err := pool.Workflow(
			ExampleID(fmt.Sprintf("step-%d", i)),
			simpleWorkflow,
			nil,
			i,
		).Get(&result); err != nil {
			log.Printf("Workflow %d failed: %v", i, err)
			continue
		}

		log.Printf("Workflow %d completed with result: %d", i, result)

		// Check current database files
		files, err := filepath.Glob(filepath.Join(dir, "demo_*.db"))
		if err != nil {
			log.Printf("Failed to list database files: %v", err)
			continue
		}

		log.Printf("Current database files:")
		for _, file := range files {
			info, err := os.Stat(file)
			if err != nil {
				log.Printf("  Failed to stat %s: %v", file, err)
				continue
			}
			log.Printf("  - %s (size: %d bytes)", filepath.Base(file), info.Size())
		}
	}

	// Wait for all workflows to complete
	if err := pool.Wait(); err != nil {
		log.Printf("Wait failed: %v", err)
	}
}
