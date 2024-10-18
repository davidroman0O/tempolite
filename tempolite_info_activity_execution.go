package tempolite

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"runtime"
	"time"

	"github.com/davidroman0O/go-tempolite/ent/activity"
	"github.com/davidroman0O/go-tempolite/ent/activityexecution"
)

type ActivityExecutionInfo[T Identifier] struct {
	tp          *Tempolite[T]
	ExecutionID ActivityExecutionID
	err         error
}

// Try to find the activity execution until it reaches a final state
func (i *ActivityExecutionInfo[T]) Get(output ...interface{}) error {
	if i.err != nil {
		return i.err
	}
	for idx, out := range output {
		if reflect.TypeOf(out).Kind() != reflect.Ptr {
			return fmt.Errorf("output parameter at index %d is not a pointer", idx)
		}
	}

	ticker := time.NewTicker(time.Second / 16)
	defer ticker.Stop()
	var value any
	var ok bool
	var activityHandlerInfo Activity
	for {
		select {
		case <-i.tp.ctx.Done():
			return i.tp.ctx.Err()
		case <-ticker.C:
			// fmt.Println("searching id", i.ExecutionID.String())
			activityExecEntity, err := i.tp.client.ActivityExecution.Query().Where(activityexecution.IDEQ(i.ExecutionID.String())).WithActivity().Only(i.tp.ctx)
			if err != nil {
				return err
			}
			activityEntity, err := i.tp.client.Activity.Query().Where(activity.IDEQ(activityExecEntity.Edges.Activity.ID)).Only(i.tp.ctx)
			if err != nil {
				return err
			}
			if value, ok = i.tp.activities.Load(HandlerIdentity(activityEntity.Identity)); ok {
				if activityHandlerInfo, ok = value.(Activity); !ok {
					log.Printf("scheduler: activity %s is not handler info", activityExecEntity.Edges.Activity.ID)
					return errors.New("activity is not handler info")
				}
				switch activityExecEntity.Status {
				case activityexecution.StatusCompleted:
					outputs, err := i.tp.convertOuputs(HandlerInfo(activityHandlerInfo), activityExecEntity.Output)
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
				case activityexecution.StatusFailed:
					return errors.New(activityExecEntity.Error)
				case activityexecution.StatusRetried:
					return errors.New("activity was retried")
				case activityexecution.StatusPending, activityexecution.StatusRunning:
					runtime.Gosched()
					continue
				}
			}
			runtime.Gosched()
		}
	}
}
