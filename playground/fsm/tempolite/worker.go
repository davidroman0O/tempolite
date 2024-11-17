package tempolite

import (
	"context"
	"log"
)

// QueueWorker represents a worker in a queue
type QueueWorker struct {
	orchestrator *Orchestrator
	ctx          context.Context
}

func (w *QueueWorker) Run(ctx context.Context, task *QueueTask) error {
	log.Printf("Worker executing workflow in queue %s", task.queueName)

	// Register workflow if not already registered
	handler, err := task.registry.RegisterWorkflow(task.workflowFunc)
	if err != nil {
		task.future.setError(err)
		return err
	}

	// Execute the workflow using the orchestrator
	future := w.orchestrator.Workflow(handler.Handler, task.options, task.args...)

	// Watch for context cancellation
	done := make(chan struct{})
	go func() {
		defer close(done)
		// Get expects a pointer to a value
		var result int // Match the type that the workflow returns
		if err := future.Get(&result); err != nil {
			task.future.setError(err)
			return
		}
		// Set single result
		task.future.setResult([]interface{}{result})
	}()

	w.orchestrator.Wait()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}
