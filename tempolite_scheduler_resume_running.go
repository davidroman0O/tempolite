package tempolite

import (
	"runtime"
	"time"

	"github.com/davidroman0O/tempolite/ent/workflow"
)

func (tp *Tempolite) schedulerResumeRunningWorkflows(queueName string, done chan struct{}) {
	ticker := time.NewTicker(time.Second / 16)
	defer ticker.Stop()

	for {
		select {
		case <-tp.ctx.Done():
			tp.logger.Debug(tp.ctx, "Resume running workflows scheduler stopped due to context done")
			return
		case <-done:
			tp.logger.Debug(tp.ctx, "Resume running workflows scheduler stopped due to done signal")
			return
		case <-ticker.C:
			// Get distinct queue names that have workflows to resume
			workflows, err := tp.client.Workflow.Query().
				Where(
					workflow.StatusEQ(workflow.StatusRunning),
					workflow.IsPausedEQ(false),
					workflow.IsReadyEQ(false),
					workflow.QueueNameEQ(queueName),
				).
				All(tp.ctx)
			if err != nil {
				tp.logger.Error(tp.ctx, "Error querying workflows to resume", "error", err)
				continue
			}

			for _, wf := range workflows {
				tp.logger.Debug(tp.ctx, "Found running workflow to resume", "workflowID", wf.ID)

				tx, err := tp.client.Tx(tp.ctx)
				if err != nil {
					tp.logger.Error(tp.ctx, "Error starting transaction for updating workflow", "workflowID", wf.ID, "error", err)
					continue
				}

				// Mark as ready to be picked up
				_, err = tx.Workflow.UpdateOne(wf).
					SetIsReady(true).
					Save(tp.ctx)
				if err != nil {
					tp.logger.Error(tp.ctx, "Error updating workflow", "workflowID", wf.ID, "error", err)
					if rerr := tx.Rollback(); rerr != nil {
						tp.logger.Error(tp.ctx, "Error rolling back transaction", "error", rerr)
					}
					continue
				}

				if err := tx.Commit(); err != nil {
					tp.logger.Error(tp.ctx, "Error committing transaction", "error", err)
					continue
				}

				if err := tp.redispatchWorkflow(WorkflowID(wf.ID)); err != nil {
					tp.logger.Error(tp.ctx, "Error redispatching workflow", "workflowID", wf.ID, "error", err)
				}
			}
			runtime.Gosched()
		}
	}
}
