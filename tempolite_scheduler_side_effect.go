package tempolite

import (
	"log"
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
				log.Printf("scheduler: SideEffectExecution.Query failed: %v", err)
				continue
			}

			tp.schedulerSideEffectStarted.Store(true)

			if len(pendingSideEffects) == 0 {
				continue
			}

			for _, se := range pendingSideEffects {
				sideEffectInfo, ok := tp.sideEffects.Load(se.Edges.SideEffect.ID)
				if !ok {
					log.Printf("scheduler: SideEffect %s not found", se.Edges.SideEffect.ID)
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

				log.Printf("scheduler: Dispatching side effect %s", se.Edges.SideEffect.HandlerName)

				if err := tp.sideEffectPool.Dispatch(task); err != nil {
					log.Printf("scheduler: Dispatch failed: %v", err)
					continue
				}

				if _, err = tp.client.SideEffectExecution.UpdateOneID(se.ID).SetStatus(sideeffectexecution.StatusRunning).Save(tp.ctx); err != nil {
					log.Printf("scheduler: SideEffectExecution.UpdateOneID failed: %v", err)
					continue
				}
			}

			runtime.Gosched()
		}
	}
}
