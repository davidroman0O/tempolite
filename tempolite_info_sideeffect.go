package tempolite

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"runtime"
	"time"

	"github.com/davidroman0O/go-tempolite/ent"
	"github.com/davidroman0O/go-tempolite/ent/sideeffect"
	"github.com/davidroman0O/go-tempolite/ent/sideeffectexecution"
)

type SideEffectInfo[T Identifier] struct {
	tp       *Tempolite[T]
	EntityID SideEffectID
	err      error
}

func (i *SideEffectInfo[T]) Get(output ...interface{}) error {
	if i.err != nil {
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
			return i.tp.ctx.Err()
		case <-ticker.C:
			sideEffectEntity, err := i.tp.client.SideEffect.Query().
				Where(sideeffect.IDEQ(i.EntityID.String())).
				WithExecutions().
				Only(i.tp.ctx)
			if err != nil {
				return err
			}

			if value, ok = i.tp.sideEffects.Load(sideEffectEntity.ID); ok {
				if sideeffectHandlerInfo, ok = value.(SideEffect); !ok {
					log.Printf("scheduler: sideeffect %s is not handler info", i.EntityID.String())
					return errors.New("sideeffect is not handler info")
				}

				switch sideEffectEntity.Status {
				case sideeffect.StatusCompleted:
					// fmt.Println("searching for side effect execution of ", i.EntityID.String())
					latestExec, err := i.tp.client.SideEffectExecution.Query().
						Where(
							sideeffectexecution.HasSideEffect(),
						).
						Order(ent.Desc(sideeffectexecution.FieldStartedAt)).
						First(i.tp.ctx)
					if err != nil {
						if ent.IsNotFound(err) {
							// Handle the case where no execution is found
							log.Printf("No execution found for sideeffect %s", i.EntityID)
							return fmt.Errorf("no execution found for sideeffect %s", i.EntityID)
						}
						log.Printf("Error querying sideeffect execution: %v", err)
						return fmt.Errorf("error querying sideeffect execution: %w", err)
					}

					switch latestExec.Status {
					case sideeffectexecution.StatusCompleted:
						outputs, err := i.tp.convertOuputs(HandlerInfo(sideeffectHandlerInfo), latestExec.Output)
						if err != nil {
							return err
						}

						if len(output) != len(outputs) {
							return fmt.Errorf("output length mismatch: expected %d, got %d", len(outputs), len(output))
						}

						for idx, outPtr := range output {
							outVal := reflect.ValueOf(outPtr).Elem()
							outputVal := reflect.ValueOf(outputs[idx])

							if outVal.Type() != outputVal.Type() {
								return fmt.Errorf("type mismatch at index %d: expected %v, got %v", idx, outVal.Type(), outputVal.Type())
							}

							outVal.Set(outputVal)
						}
						return nil
					case sideeffectexecution.StatusFailed:
						return errors.New(latestExec.Error)
					case sideeffectexecution.StatusPending, sideeffectexecution.StatusRunning:
						// The workflow is still in progress
						// return  errors.New("workflow is still in progress")
						runtime.Gosched()
						continue
					}

				case sideeffect.StatusPending, sideeffect.StatusRunning:
					runtime.Gosched()
					continue
				}
			}
			runtime.Gosched()
		}
	}
}
