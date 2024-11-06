package tempolite

import (
	"context"
	"errors"
	"runtime"

	"github.com/davidroman0O/tempolite/ent"
	"github.com/davidroman0O/tempolite/ent/sideeffect"
	"github.com/davidroman0O/tempolite/ent/sideeffectexecution"
)

func (tp *Tempolite) schedulerExecutionSideEffectForQueue(queueName string, done chan struct{}) {

	queue, err := tp.getSideEffectPoolQueue(queueName)
	if err != nil {
		tp.logger.Error(tp.ctx, "Scheduler side effect execution: getSideEffectPoolQueue failed", "error", err)
		return
	}

	for {
		select {
		case <-tp.ctx.Done():
			tp.logger.Debug(tp.ctx, "scheduler side effect execution: context done", "queue", queueName)
			return
		case <-done:
			tp.logger.Debug(tp.ctx, "scheduler side effect execution: done signal", "queue", queueName)
			return
		default:

			queueWorkersRaw, ok := tp.queues.Load(queueName)
			if !ok {
				continue
			}
			queueWorkers := queueWorkersRaw.(*QueueWorkers)

			workerIDs := queueWorkers.SideEffects.GetWorkerIDs()
			availableSlots := len(workerIDs) - queueWorkers.SideEffects.ProcessingCount()
			if availableSlots <= 0 {
				runtime.Gosched()
				continue
			}

			pendingSideEffects, err := tp.client.SideEffectExecution.Query().
				Where(
					sideeffectexecution.StatusEQ(sideeffectexecution.StatusPending),
					sideeffectexecution.HasSideEffectWith(sideeffect.QueueNameEQ(queueName)),
				).
				Order(ent.Asc(sideeffectexecution.FieldStartedAt)).
				WithSideEffect().
				Limit(availableSlots).
				All(tp.ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					tp.logger.Debug(tp.ctx, "scheduler sideeffect execution: context canceled", "queue", queueName)
					return
				}
				tp.logger.Error(tp.ctx, "Scheduler sideeffect execution: SideEffectExecution.Query failed", "error", err)
				continue
			}

			if len(pendingSideEffects) == 0 {
				continue
			}

			for _, se := range pendingSideEffects {
				sideEffectInfo, ok := tp.sideEffects.Load(se.Edges.SideEffect.ID)
				if !ok {
					tp.logger.Error(tp.ctx, "Scheduler sideeffect execution: SideEffect not found", "sideEffectID", se.Edges.SideEffect.ID)
					continue
				}

				sideEffect := sideEffectInfo.(SideEffect)

				contextSideEffect := SideEffectContext{
					tp:           tp,
					sideEffectID: se.Edges.SideEffect.ID,
					executionID:  se.ID,
					stepID:       se.Edges.SideEffect.StepID,
				}

				task := &sideEffectTask{
					ctx:         contextSideEffect,
					handlerName: sideEffect.HandlerLongName,
					handler:     sideEffect.Handler,
				}

				tp.logger.Debug(tp.ctx, "Scheduler sideeffect execution: Dispatching side effect", "sideEffectHandler", se.Edges.SideEffect.HandlerName)

				if err := queue.Submit(task); err != nil {
					tp.logger.Error(tp.ctx, "Scheduler sideeffect execution: Dispatch failed", "error", err)

					tx, err := tp.client.Tx(tp.ctx)
					if err != nil {
						tp.logger.Error(tp.ctx, "Scheduler sideeffect execution: Failed to start transaction", "error", err)
						continue
					}

					if _, err = tx.SideEffectExecution.UpdateOneID(se.ID).SetStatus(sideeffectexecution.StatusFailed).SetError(err.Error()).Save(tp.ctx); err != nil {
						tp.logger.Error(tp.ctx, "Scheduler sideeffect execution: SideEffectExecution.UpdateOneID failed", "error", err)
						if rollbackErr := tx.Rollback(); rollbackErr != nil {
							tp.logger.Error(tp.ctx, "Scheduler sideeffect execution: Failed to rollback transaction", "error", rollbackErr)
						}
						continue
					}

					if _, err = tx.SideEffect.UpdateOneID(se.Edges.SideEffect.ID).SetStatus(sideeffect.StatusFailed).Save(tp.ctx); err != nil {
						tp.logger.Error(tp.ctx, "Scheduler sideeffect execution: SideEffect.UpdateOneID failed", "error", err)
						if rollbackErr := tx.Rollback(); rollbackErr != nil {
							tp.logger.Error(tp.ctx, "Scheduler sideeffect execution: Failed to rollback transaction", "error", rollbackErr)
						}
						continue
					}

					if err = tx.Commit(); err != nil {
						tp.logger.Error(tp.ctx, "Scheduler sideeffect execution: Failed to commit transaction", "error", err)
					}
					continue
				}

				tx, err := tp.client.Tx(tp.ctx)
				if err != nil {
					tp.logger.Error(tp.ctx, "Scheduler sideeffect execution: Failed to start transaction", "error", err)
					continue
				}

				if _, err = tx.SideEffectExecution.UpdateOneID(se.ID).SetStatus(sideeffectexecution.StatusRunning).Save(tp.ctx); err != nil {
					tp.logger.Error(tp.ctx, "Scheduler sideeffect execution: SideEffectExecution.UpdateOneID failed", "error", err)
					if rollbackErr := tx.Rollback(); rollbackErr != nil {
						tp.logger.Error(tp.ctx, "Scheduler sideeffect execution: Failed to rollback transaction", "error", rollbackErr)
					}
					continue
				}

				if _, err = tx.SideEffect.UpdateOneID(se.Edges.SideEffect.ID).SetStatus(sideeffect.StatusRunning).Save(tp.ctx); err != nil {
					tp.logger.Error(tp.ctx, "Scheduler sideeffect execution: SideEffect.UpdateOneID failed", "error", err)
					if rollbackErr := tx.Rollback(); rollbackErr != nil {
						tp.logger.Error(tp.ctx, "Scheduler sideeffect execution: Failed to rollback transaction", "error", rollbackErr)
					}
					continue
				}

				if err = tx.Commit(); err != nil {
					tp.logger.Error(tp.ctx, "Scheduler sideeffect execution: Failed to commit transaction", "error", err)
				}
			}

			runtime.Gosched()
		}
	}
}
