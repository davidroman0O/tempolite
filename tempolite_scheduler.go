package tempolite

import (
	"fmt"
	"log"
	"time"

	"github.com/davidroman0O/go-tempolite/ent"
	"github.com/davidroman0O/go-tempolite/ent/executionunit"
	"github.com/davidroman0O/go-tempolite/ent/task"
)

type Scheduler struct {
	tp   *Tempolite
	done chan struct{}
}

func NewScheduler(tp *Tempolite) *Scheduler {
	s := &Scheduler{
		tp:   tp,
		done: make(chan struct{}),
	}

	go s.run()

	return s
}

func (s *Scheduler) Close() {
	close(s.done)
}

func (s *Scheduler) run() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
			if err := s.processPendingTasks(); err != nil {
				log.Printf("Error processing pending tasks: %v", err)
			}
		}
	}
}

func (s *Scheduler) processPendingTasks() error {
	ctx := s.tp.ctx

	// Query for pending tasks
	pendingTasks, err := s.tp.client.Task.
		Query().
		Where(task.StatusEQ(task.StatusPending)).
		WithExecutionUnit().
		Order(ent.Asc(task.FieldCreatedAt)).
		Limit(10).
		All(ctx)

	if err != nil {
		return fmt.Errorf("error querying pending tasks: %w", err)
	}

	for _, taskPending := range pendingTasks {
		if taskPending.Edges.ExecutionUnit == nil {
			log.Printf("Task %s has no associated ExecutionUnit", taskPending.ID)
			continue
		}

		// Check if the parent task (if any) has completed
		if taskPending.Edges.ExecutionUnit.Edges.Parent != nil {
			parentStatus := taskPending.Edges.ExecutionUnit.Edges.Parent.Status
			if parentStatus != executionunit.StatusCompleted {
				log.Printf("Parent task for %s is not completed. Current status: %s", taskPending.ID, parentStatus)
				continue
			}
		}

		// Dispatch the task to the appropriate pool
		var dispatchErr error
		switch taskPending.Type {
		case task.TypeHandler:
			dispatchErr = s.tp.handlerTaskPool.Dispatch(taskPending)
		case task.TypeSideEffect:
			dispatchErr = s.tp.sideEffectTaskPool.Dispatch(taskPending)
		case task.TypeSagaTransaction, task.TypeSagaCompensation:
			dispatchErr = s.tp.sagaTaskPool.Dispatch(taskPending)
		default:
			log.Printf("Unknown task type for task %v: %v", taskPending.ID, taskPending.Type)
			continue
		}

		if dispatchErr != nil {
			log.Printf("Error dispatching task %v: %v", taskPending.ID, dispatchErr)
			continue
		}

		// Update the task status to in progress
		_, err := s.tp.client.Task.
			UpdateOne(taskPending).
			SetStatus(task.StatusInProgress).
			Save(ctx)

		if err != nil {
			log.Printf("Error updating status for task %v: %v", taskPending.ID, err)
		}
	}

	return nil
}
