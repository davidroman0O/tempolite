package tests

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/davidroman0O/tempolite"
)

// go test -v -count=1 -timeout 30s -run ^TestQueues$ ./tests
func TestQueues(t *testing.T) {
	executionCounts := struct {
		defaultQueue atomic.Int32
		customQueue  atomic.Int32
		sync.Mutex
	}{}

	simpleActivity := func(ctx tempolite.ActivityContext, queueName string) error {
		switch queueName {
		case "default":
			executionCounts.defaultQueue.Add(1)
		case "custom":
			executionCounts.customQueue.Add(1)
		}
		return nil
	}

	workflowFn := func(ctx tempolite.WorkflowContext, useCustomQueue bool) error {
		if useCustomQueue {
			return ctx.Activity("custom-queue-step", simpleActivity, nil, "custom").Get()
		}
		return ctx.Activity("default-queue-step", simpleActivity, nil, "default").Get()
	}

	t.Run("multiple queues", func(t *testing.T) {
		cfg := NewTestConfig(t, "queues")
		cfg.Registry.
			Workflow(workflowFn).
			Activity(simpleActivity)

		// Create Tempolite with custom queue configuration
		tp, err := tempolite.New(
			context.Background(),
			cfg.Registry.Build(),
			tempolite.WithPath(cfg.DBPath),
			tempolite.WithDestructive(),
			tempolite.WithQueueConfig(
				tempolite.NewQueue("custom", tempolite.WithWorkflowWorkers(2), tempolite.WithActivityWorkers(2)),
			),
		)
		if err != nil {
			t.Fatalf("Failed to create Tempolite instance: %v", err)
		}
		defer tp.Close()

		// Execute workflows targeting different queues
		var wg sync.WaitGroup
		wg.Add(4)

		// Two workflows using default queue
		for i := 0; i < 2; i++ {
			go func() {
				defer wg.Done()
				err := tp.Workflow(workflowFn, nil, false).Get()
				if err != nil {
					t.Errorf("Default queue workflow failed: %v", err)
				}
			}()
		}

		// Two workflows using custom queue
		for i := 0; i < 2; i++ {
			go func() {
				defer wg.Done()
				err := tp.Workflow(workflowFn, nil, true).Get()
				if err != nil {
					t.Errorf("Custom queue workflow failed: %v", err)
				}
			}()
		}

		wg.Wait()

		if err := WaitWithTimeout(t, tp, 5*time.Second); err != nil {
			t.Fatalf("Failed to wait for workflows completion: %v", err)
		}

		// Verify execution counts
		if count := executionCounts.defaultQueue.Load(); count != 2 {
			t.Errorf("Default queue executions: got %d, want 2", count)
		}
		if count := executionCounts.customQueue.Load(); count != 2 {
			t.Errorf("Custom queue executions: got %d, want 2", count)
		}
	})

	t.Run("queue worker scaling", func(t *testing.T) {
		cfg := NewTestConfig(t, "queue-scaling")
		cfg.Registry.
			Workflow(workflowFn).
			Activity(simpleActivity)

		tp := NewTestTempolite(t, cfg)
		defer tp.Close()

		// Add workers to default queue
		for i := 0; i < 3; i++ {
			if err := tp.AddWorkerActivity("default"); err != nil {
				t.Fatalf("Failed to add worker: %v", err)
			}
		}

		workers, err := tp.ListWorkersActivity("default")
		if err != nil {
			t.Fatalf("Failed to list workers: %v", err)
		}

		originalCount := len(workers)

		// Remove a worker
		if err := tp.RemoveWorkerActivity("default", workers[0]); err != nil {
			t.Fatalf("Failed to remove worker: %v", err)
		}

		workers, err = tp.ListWorkersActivity("default")
		if err != nil {
			t.Fatalf("Failed to list workers: %v", err)
		}

		if len(workers) != originalCount-1 {
			t.Errorf("Worker count after removal: got %d, want %d", len(workers), originalCount-1)
		}
	})
}
