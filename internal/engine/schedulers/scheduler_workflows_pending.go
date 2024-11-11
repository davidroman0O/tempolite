package schedulers

import (
	"context"
	"errors"

	"github.com/davidroman0O/retrypool"
	"github.com/davidroman0O/tempolite/internal/engine/execution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent"
	"github.com/davidroman0O/tempolite/internal/persistence/repository"
	"github.com/davidroman0O/tempolite/pkg/logs"
)

type SchedulerWorkflowsPending struct {
	*Scheduler
}

func (s SchedulerWorkflowsPending) Tick() error {
	var tx *ent.Tx
	var err error
	var queues []*repository.QueueInfo

	if tx, err = s.db.Tx(); err != nil {
		logs.Error(s.ctx, "Schduler Workflows Pending error creating transaction", "error", err)
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback() // Rollback in case of panic
			panic(p)      // Re-throw panic after rollback
		} else if err != nil {
			tx.Rollback() // Rollback on error
		} else {
			err = tx.Commit() // Commit on success
			if err != nil {
				logs.Error(s.ctx, "Schduler Workflows Pending error committing transaction", "error", err)
				return
			}
		}
	}()

	if queues, err = s.db.Queues().List(tx); err != nil {
		if err := tx.Rollback(); err != nil {
			logs.Error(s.ctx, "Schduler Workflows Pending error rolling back transaction", "error", err)
			return err
		}
		logs.Error(s.ctx, "Schduler Workflows Pending error listing queues", "error", err)
		return err
	}

	for _, q := range queues {
		queue := s.Scheduler.getQueue(q.Name)

		if queue.AvailableWorkflowWorkers() <= 0 {
			continue
		}

		workflows, err := s.db.Workflows().ListExecutionsPending(tx, q.Name)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				logs.Debug(s.ctx, "Schduler Workflows Pending list execution pending context canceled", "error", err)
				return nil
			}
			if err := tx.Rollback(); err != nil {
				logs.Error(s.ctx, "Schduler Workflows Pending error rolling back transaction", "error", err)
				return err
			}
			return err
		}

		pendingWorkflows := []chan struct{}{}

		for _, w := range workflows {
			var processed chan struct{}

			logs.Debug(s.ctx, "Schduler Workflows Pending processing workflow", "queue", q.Name, "workflowID", w.ID, "stepID", w.StepID, "queueID", w.QueueID, "runID", w.RunID)

			task := retrypool.
				NewRequestResponse[execution.WorkflowRequest, execution.WorkflowReponse](
				execution.WorkflowRequest{
					WorkflowInfo: w,
					Retry: func() error {
						var rtx *ent.Tx
						if rtx, err = s.db.Tx(); err != nil {
							logs.Error(s.ctx, "Schduler Workflows Pending error creating retry transaction", "error", err)
							return err
						}
						_, err := s.db.Workflows().CreateRetry(rtx, w.ID)
						if err != nil {
							if err := rtx.Rollback(); err != nil {
								logs.Error(s.ctx, "Schduler Workflows Pending error rolling back retry transaction", "error", err)
								return err
							}
							logs.Error(s.ctx, "Schduler Workflows Pending error creating retry", "error", err)
							return err
						}
						if err := rtx.Commit(); err != nil {
							logs.Error(s.ctx, "Schduler Workflows Pending error committing retry transaction", "error", err)
							return err
						}
						return nil
					},
				})

			logs.Debug(s.ctx, "Schduler Workflows Pending updating execution pending to running", "workflowID", w.ID, "stepID", w.StepID, "queueID", w.QueueID, "runID", w.RunID)
			if err := s.db.Workflows().UpdateExecutionPendingToRunning(tx, w.Execution.ID); err != nil {
				if errors.Is(err, context.Canceled) {
					logs.Debug(s.ctx, "Schduler Workflows Pending update execution pending to running context canceled", "error", err)
					return nil
				}
				if err := tx.Rollback(); err != nil {
					logs.Error(s.ctx, "Schduler Workflows Pending error rolling back transaction", "error", err)
					return err
				}
				logs.Error(s.ctx, "Schduler Workflows Pending error updating execution pending to running", "error", err)
				return err
			}

			logs.Debug(s.ctx, "Schduler Workflows Pending submitting workflow", "workflowID", w.ID, "stepID", w.StepID, "queueID", w.QueueID, "runID", w.RunID)
			// TODO: if we fail, then we should retry while updating the workflow somehow cause we need to fail
			if processed, err = queue.SubmitWorkflow(task); err != nil {
				if errors.Is(err, context.Canceled) {
					logs.Debug(s.ctx, "Schduler Workflows Pending submit workflow context canceled", "error", err)
					return nil
				}
				if err := tx.Rollback(); err != nil {
					logs.Error(s.ctx, "Schduler Workflows Pending error rolling back transaction", "error", err)
					return err
				}
			}

			// We have to wait for the workflow to be ackowledged before do the next one
			// TODO: we might want to batch many workflows and wait for all of them to be ackowledged
			pendingWorkflows = append(pendingWorkflows, processed)
			logs.Debug(s.ctx, "Schduler Workflows Pending updating added for listening", "workflowID", w.ID, "stepID", w.StepID, "queueID", w.QueueID, "runID", w.RunID)
		}

		for _, v := range pendingWorkflows {
			<-v
		}
	}

	return nil
}
