package tempolite

import (
	"runtime"

	"github.com/davidroman0O/go-tempolite/ent"
	"github.com/davidroman0O/go-tempolite/ent/sideeffectexecution"
)

func (tp *Tempolite[T]) schedulerExecutionSideEffect() {
	defer close(tp.schedulerSideEffectDone)
	for {
		select {
		case <-tp.ctx.Done():
			return
		default:
			pendingSideEffects, err := tp.client.SideEffectExecution.Query().
				Where(sideeffectexecution.StatusEQ(sideeffectexecution.StatusPending)).
				Order(ent.Asc(sideeffectexecution.FieldStartedAt)).WithSideEffect().
				WithSideEffect().
				Limit(1).All(tp.ctx)
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
					tp.logger.Error(tp.ctx, "Scheduler sideeffect execution: SideEffect %s not found", "sideEffectID", se.Edges.SideEffect.ID)
					continue
				}

				sideEffect := sideEffectInfo.(SideEffect)

				contextSideEffect := SideEffectContext[T]{
					tp:           tp,
					sideEffectID: se.Edges.SideEffect.ID,
					executionID:  se.ID,
					// runID:        se.RunID,
					stepID: se.Edges.SideEffect.StepID,
				}

				task := &sideEffectTask[T]{
					ctx:         contextSideEffect,
					handlerName: sideEffect.HandlerLongName,
					handler:     sideEffect.Handler,
				}

				tp.logger.Debug(tp.ctx, "Scheduler sideeffect execution: Dispatching side effect %s", "sideEffectHandler", se.Edges.SideEffect.HandlerName)

				if err := tp.sideEffectPool.Dispatch(task); err != nil {
					tp.logger.Error(tp.ctx, "Scheduler sideeffect execution: Dispatch failed", "error", err)
					continue
				}

				if _, err = tp.client.SideEffectExecution.UpdateOneID(se.ID).SetStatus(sideeffectexecution.StatusRunning).Save(tp.ctx); err != nil {
					tp.logger.Error(tp.ctx, "Scheduler sideeffect execution: SideEffectExecution.UpdateOneID failed", "error", err)
					continue
				}
			}

			runtime.Gosched()
		}
	}
}
