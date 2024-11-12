package schedulers

import (
	"context"
	"errors"
	"sync"

	"github.com/davidroman0O/retrypool"
	"github.com/davidroman0O/tempolite/internal/engine/execution"
	"github.com/davidroman0O/tempolite/internal/persistence/ent"
	"github.com/davidroman0O/tempolite/internal/persistence/repository"
	"github.com/davidroman0O/tempolite/pkg/logs"
)

type SchedulerWorkflowsPending struct {
	db  repository.Repository
	ctx context.Context

	queueName string

	getSlot func() int
	submit  func(task *retrypool.RequestResponse[execution.WorkflowRequest, execution.WorkflowReponse]) (chan struct{}, error)

	mu                   sync.Mutex
	currentNotifications map[int]chan struct{}
}

func NewSchedulerWorkflowPending(
	ctx context.Context,
	db repository.Repository,
	queueName string,
	getSlot func() int,
	submit func(task *retrypool.RequestResponse[execution.WorkflowRequest, execution.WorkflowReponse]) (chan struct{}, error),
) *SchedulerWorkflowsPending {
	s := &SchedulerWorkflowsPending{
		db:                   db,
		ctx:                  ctx,
		queueName:            queueName,
		getSlot:              getSlot,
		submit:               submit,
		currentNotifications: make(map[int]chan struct{}),
	}

	return s
}

func (s *SchedulerWorkflowsPending) Tick(ctx context.Context) error {

	var err error

	select {
	case <-ctx.Done():
		logs.Debug(s.ctx, "Scheduler Workflows Pending context canceled", "error", ctx.Err())
		return ctx.Err()
	default:

		var txQueueGetList *ent.Tx
		if txQueueGetList, err = s.db.Tx(); err != nil {
			logs.Error(s.ctx, "Scheduler Workflows Pending error creating transaction", "error", err)
			return err
		}

		logs.Debug(s.ctx, "Scheduler Workflows Pending checking queue", "queue", s.queueName)

		slots := s.getSlot()
		if slots <= 0 {
			return nil
		}

		s.mu.Lock()
		slots = slots - len(s.currentNotifications) // We have to subtract the current notifications
		s.mu.Unlock()
		if slots <= 0 {
			return nil
		}

		workflows, err := s.db.Workflows().ListExecutionsPending(txQueueGetList, s.queueName, slots)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				logs.Debug(s.ctx, "Scheduler Workflows Pending list execution pending context canceled", "error", err, "queue", s.queueName)
				return nil
			}
			if err := txQueueGetList.Rollback(); err != nil {
				logs.Error(s.ctx, "Scheduler Workflows Pending error rolling back transaction", "error", err, "queue", s.queueName)
				return err
			}
			return err
		}

		logs.Debug(s.ctx, "Scheduler Workflows Pending found workflows", "queue", s.queueName, "workflows", len(workflows), "workers", slots)

		if ctx.Err() != nil {
			logs.Debug(s.ctx, "Scheduler Workflows Pending context canceled", "error", ctx.Err())
			txQueueGetList.Rollback()
			return ctx.Err()
		}

		if err := txQueueGetList.Commit(); err != nil {
			logs.Error(s.ctx, "Scheduler Workflows Pending error committing transaction", "error", err)
			return err
		}

		for _, w := range workflows {
			var notification chan struct{}

			logs.Debug(s.ctx, "Scheduler Workflows Pending processing workflow", "queue", s.queueName, "workflowID", w.ID, "stepID", w.StepID, "queueID", w.QueueID, "runID", w.RunID)

			task := retrypool.
				NewRequestResponse[execution.WorkflowRequest, execution.WorkflowReponse](
				execution.WorkflowRequest{
					WorkflowInfo: w,
					Retry: func() error {
						var rtx *ent.Tx
						if rtx, err = s.db.Tx(); err != nil {
							logs.Error(s.ctx, "Scheduler Workflows Pending error creating retry transaction", "error", err, "queue", s.queueName, "workflowID", w.ID, "stepID", w.StepID, "queueID", w.QueueID, "runID", w.RunID)
							return err
						}
						_, err := s.db.Workflows().CreateRetry(rtx, w.ID)
						if err != nil {
							if err := rtx.Rollback(); err != nil {
								logs.Error(s.ctx, "Scheduler Workflows Pending error rolling back retry transaction", "error", err, "queue", s.queueName, "workflowID", w.ID, "stepID", w.StepID, "queueID", w.QueueID, "runID", w.RunID)
								return err
							}
							logs.Error(s.ctx, "Scheduler Workflows Pending error creating retry", "error", err, "queue", s.queueName, "workflowID", w.ID, "stepID", w.StepID, "queueID", w.QueueID, "runID", w.RunID)
							return err
						}
						if err := rtx.Commit(); err != nil {
							logs.Error(s.ctx, "Scheduler Workflows Pending error committing retry transaction", "error", err, "queue", s.queueName, "workflowID", w.ID, "stepID", w.StepID, "queueID", w.QueueID, "runID", w.RunID)
							return err
						}
						return nil
					},
				})

			logs.Debug(s.ctx, "Scheduler Workflows Pending submitting workflow", "queue", s.queueName, "workflowID", w.ID, "stepID", w.StepID, "queueID", w.QueueID, "runID", w.RunID)
			// TODO: if we fail, then we should retry while updating the workflow somehow cause we need to fail
			if notification, err = s.submit(task); err != nil {
				if errors.Is(err, context.Canceled) {
					logs.Debug(s.ctx, "Scheduler Workflows Pending submit workflow context canceled", "error", err)
					return nil
				}
				logs.Error(s.ctx, "Scheduler Workflows Pending error submitting workflow", "error", err)
				return err
			}

			s.mu.Lock()
			s.currentNotifications[w.Execution.ID] = notification

			go func(workflowID int, executionID int, nono chan struct{}) {
				select {
				case <-nono:
					logs.Debug(s.ctx, "Scheduler Workflows Pending workflow finished", "queue", s.queueName, "workflowID", workflowID, "executionID", executionID)
					s.mu.Lock()
					delete(s.currentNotifications, w.Execution.ID)
					s.mu.Unlock()
					return
				case <-ctx.Done():
					logs.Debug(s.ctx, "Scheduler Workflows Pending workflow context canceled", "error", ctx.Err(), "queue", s.queueName, "workflowID", workflowID, "executionID", executionID)
					s.mu.Lock()
					delete(s.currentNotifications, w.Execution.ID)
					s.mu.Unlock()
					return
				}
			}(w.ID, w.Execution.ID, notification)

			s.mu.Unlock()

			logs.Debug(s.ctx, "Scheduler Workflows Pending updating added for listening", "queue", s.queueName, "workflowID", w.ID, "stepID", w.StepID, "queueID", w.QueueID, "runID", w.RunID)
		}

		// // Now we look at the queued workflows
		// for _, queued := range waitNotifications {
		// 	// We have to wait for the workflow to be ackowledged before do the next one
		// 	// TODO: we might want to batch many workflows and wait for all of them to be ackowledged
		// 	// pendingWorkflows = append(pendingWorkflows, queued)
		// 	// VERY IMPORTANT to check the context
		// 	select {
		// 	case <-queued:
		// 	case <-ctx.Done():
		// 		logs.Debug(s.ctx, "Scheduler Workflows Pending context canceled", "error", ctx.Err())
		// 		return ctx.Err()
		// 	}
		// }

	}

	return nil
}
