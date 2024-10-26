package tempolite

import (
	"runtime"

	"github.com/davidroman0O/tempolite/ent"
	"github.com/davidroman0O/tempolite/ent/sideeffect"
	"github.com/davidroman0O/tempolite/ent/sideeffectexecution"
)

func (tp *Tempolite) schedulerExecutionSideEffect() {
	defer close(tp.schedulerSideEffectDone)
	for {
		select {
		case <-tp.ctx.Done():
			return
		default:

			availableSlots := len(tp.ListWorkersSideEffect()) - tp.sideEffectPool.ProcessingCount()
			if availableSlots <= 0 {
				runtime.Gosched()
				continue
			}

			pendingSideEffects, err := tp.client.SideEffectExecution.Query().
				Where(sideeffectexecution.StatusEQ(sideeffectexecution.StatusPending)).
				Order(ent.Asc(sideeffectexecution.FieldStartedAt)).WithSideEffect().
				WithSideEffect().
				Limit(availableSlots).
				All(tp.ctx)
			if err != nil {
				tp.logger.Error(tp.ctx, "Scheduler sideeffect execution: SideEffectExecution.Query failed", "error", err)
				continue
			}

			tp.schedulerSideEffectStarted.Store(true)

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

				if err := tp.sideEffectPool.Dispatch(task); err != nil {
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
