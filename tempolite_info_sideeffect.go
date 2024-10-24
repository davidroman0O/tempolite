package tempolite

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"time"

	"github.com/davidroman0O/tempolite/ent"
	"github.com/davidroman0O/tempolite/ent/sideeffect"
	"github.com/davidroman0O/tempolite/ent/sideeffectexecution"
)

type SideEffectInfo struct {
	tp       *Tempolite
	EntityID SideEffectID
	err      error
}

func (i *SideEffectInfo) Get(output ...interface{}) error {
	if i.err != nil {
		i.tp.logger.Error(i.tp.ctx, "SideEffectInfo.Get", "entityID", i.EntityID, "error", i.err)
		return i.err
	}

	ticker := time.NewTicker(time.Second / 16)
	defer ticker.Stop()

	var value any
	var ok bool
	var sideeffectHandlerInfo SideEffect

	for {
		select {
		case <-i.tp.ctx.Done():
			i.tp.logger.Error(i.tp.ctx, "SideEffectInfo.Get: context done", "entityID", i.EntityID)
			return i.tp.ctx.Err()
		case <-ticker.C:
			sideEffectEntity, err := i.tp.client.SideEffect.Query().
				Where(sideeffect.IDEQ(i.EntityID.String())).
				WithExecutions().
				Only(i.tp.ctx)
			if err != nil {
				i.tp.logger.Error(i.tp.ctx, "SideEffectInfo.Get: failed to query side effect", "entityID", i.EntityID, "error", err)
				return err
			}

			if value, ok = i.tp.sideEffects.Load(sideEffectEntity.ID); ok {
				if sideeffectHandlerInfo, ok = value.(SideEffect); !ok {
					i.tp.logger.Error(i.tp.ctx, "SideEffectInfo.Get: sideeffect is not handler info", "entityID", i.EntityID)
					return errors.New("sideeffect is not handler info")
				}

				switch sideEffectEntity.Status {
				case sideeffect.StatusCompleted:
					latestExec, err := i.tp.client.SideEffectExecution.Query().
						Where(
							sideeffectexecution.HasSideEffect(),
						).
						Order(ent.Desc(sideeffectexecution.FieldStartedAt)).
						First(i.tp.ctx)
					if err != nil {
						if ent.IsNotFound(err) {
							i.tp.logger.Error(i.tp.ctx, "SideEffectInfo.Get: no execution found for sideeffect", "entityID", i.EntityID)
							return fmt.Errorf("no execution found for sideeffect %s", i.EntityID)
						}
						i.tp.logger.Error(i.tp.ctx, "Error querying sideeffect execution", "error", err)
						return fmt.Errorf("error querying sideeffect execution: %w", err)
					}

					switch latestExec.Status {
					case sideeffectexecution.StatusCompleted:
						outputs, err := i.tp.convertOuputs(HandlerInfo(sideeffectHandlerInfo), latestExec.Output)
						if err != nil {
							i.tp.logger.Error(i.tp.ctx, "SideEffectInfo.Get: failed to convert outputs", "error", err)
							return err
						}

						if len(output) != len(outputs) {
							i.tp.logger.Error(i.tp.ctx, "SideEffectInfo.Get: output length mismatch", "expected", len(outputs), "got", len(output))
							return fmt.Errorf("output length mismatch: expected %d, got %d", len(outputs), len(output))
						}

						for idx, outPtr := range output {
							outVal := reflect.ValueOf(outPtr).Elem()
							outputVal := reflect.ValueOf(outputs[idx])

							if outVal.Type() != outputVal.Type() {
								i.tp.logger.Error(i.tp.ctx, "SideEffectInfo.Get: type mismatch", "index", idx, "expected", outVal.Type(), "got", outputVal.Type())
								return fmt.Errorf("type mismatch at index %d: expected %v, got %v", idx, outVal.Type(), outputVal.Type())
							}

							outVal.Set(outputVal)
						}
						return nil
					case sideeffectexecution.StatusFailed:
						i.tp.logger.Debug(i.tp.ctx, "SideEffectInfo.Get: sideeffect failed", "entityID", i.EntityID, "error", latestExec.Error)
						return errors.New(latestExec.Error)
					case sideeffectexecution.StatusPending, sideeffectexecution.StatusRunning:
						i.tp.logger.Debug(i.tp.ctx, "SideEffectInfo.Get: sideeffect is still in progress", "entityID", i.EntityID)
						runtime.Gosched()
						continue
					}

				case sideeffect.StatusPending, sideeffect.StatusRunning:
					i.tp.logger.Debug(i.tp.ctx, "SideEffectInfo.Get: sideeffect is still in progress", "entityID", i.EntityID)
					runtime.Gosched()
					continue
				}
			}
			runtime.Gosched()
		}
	}
}
