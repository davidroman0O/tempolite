package tempolite

import (
	"context"
	"errors"

	"github.com/davidroman0O/tempolite/ent/workflow"
)

// One shot function at startup to resume all running workflows
// Basically, we treat workflow with "Running false false" as paused!
func (tp *Tempolite) resumeRunningWorkflows(queue string) error {
	// get all workflows with status Running, ispaused as false and isready as false
	workflows, err := tp.client.Workflow.Query().
		Where(workflow.And(
			workflow.StatusEQ(workflow.StatusRunning),
			workflow.IsPausedEQ(false),
			workflow.IsReadyEQ(false),
			workflow.QueueName(queue),
		)).
		All(tp.ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			tp.logger.Debug(tp.ctx, "resume workflow execution: context canceled", "queue", queue)
			return err
		}
		return err
	}

	for _, w := range workflows {
		// let's update as resumable
		_, err := tp.client.Workflow.UpdateOneID(w.ID).
			SetStatus(workflow.StatusPaused).
			SetIsPaused(false).
			SetIsReady(true).
			Save(tp.ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				tp.logger.Debug(tp.ctx, "resume workflow execution: context canceled", "queue", queue)
				return err
			}
			return err
		}
	}

	return nil
}
