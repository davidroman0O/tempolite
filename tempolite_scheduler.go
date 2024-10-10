package tempolite

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/davidroman0O/go-tempolite/ent"
	"github.com/davidroman0O/go-tempolite/ent/handlerexecution"
	"github.com/davidroman0O/go-tempolite/ent/handlertask"
	"github.com/davidroman0O/retrypool"
)

// Scheduler is responsible for pulling tasks from the database and dispatching them to the handler pool
type Scheduler struct {
	tp   *Tempolite
	done chan struct{}
}

// NewScheduler creates a new Scheduler instance
func NewScheduler(tp *Tempolite) *Scheduler {
	s := &Scheduler{
		tp:   tp,
		done: make(chan struct{}),
	}

	go s.run()

	return s
}

// Close stops the scheduler
func (s *Scheduler) Close() {
	close(s.done)
}

// run is the main loop of the scheduler
func (s *Scheduler) run() {
	for {
		select {
		case <-s.done:
			return
		default:
			// Fetch and process pending tasks
			if err := s.processPendingTasks(); err != nil {
				log.Printf("Error processing pending tasks: %v", err)
			}
			runtime.Gosched() // Allow other goroutines to run
		}
	}
}

// processPendingTasks fetches pending tasks and dispatches them to the handler pool
func (s *Scheduler) processPendingTasks() error {
	// Query for pending tasks
	pendingTasks, err := s.tp.client.HandlerTask.
		Query().
		Where(handlertask.StatusEQ(handlertask.StatusPending)).
		WithHandlerExecution().
		Order(ent.Asc(handlertask.FieldCreatedAt)).
		Limit(1).
		All(s.tp.ctx)

	if err != nil {
		return fmt.Errorf("error querying pending tasks: %v", err)
	}

	for _, task := range pendingTasks {
		if task.Edges.HandlerExecution == nil {
			log.Printf("Task %s has no associated HandlerExecution", task.ID)
			continue
		}

		// Check if the parent task (if any) has completed
		if task.Edges.HandlerExecution.Edges.Parent != nil {
			parentStatus := task.Edges.HandlerExecution.Edges.Parent.Status
			if parentStatus != handlerexecution.StatusCompleted {
				log.Printf("Parent task for %s is not completed. Current status: %s", task.ID, parentStatus)
				continue
			}
		}

		// Dispatch the task to the handler pool
		if err := s.tp.handlerTaskPool.pool.
			Dispatch(
				task,
				retrypool.WithMaxDuration[*ent.HandlerTask](time.Minute*10), // 10m
				retrypool.WithTimeLimit[*ent.HandlerTask](time.Minute*30),   // 30m
				retrypool.WithImmediateRetry[*ent.HandlerTask](),            // failed track jobs will be retried immediately
			); err != nil {
			log.Printf("Error dispatching task %s: %v", task.ID, err)
			continue
		}

		// Update the task status to in progress
		_, err := s.tp.client.HandlerTask.
			UpdateOne(task).
			SetStatus(handlertask.StatusInProgress).
			Save(s.tp.ctx)

		if err != nil {
			log.Printf("Error updating status for task %s: %v", task.ID, err)
		}
	}

	// If no tasks were processed, sleep for a short duration to avoid tight looping
	if len(pendingTasks) == 0 {
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}
