package tempolite

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"runtime"
	"time"

	"github.com/davidroman0O/go-tempolite/ent"
	"github.com/davidroman0O/go-tempolite/ent/activity"
	"github.com/davidroman0O/go-tempolite/ent/activityexecution"
)

type ActivityInfo[T Identifier] struct {
	tp         *Tempolite[T]
	ActivityID ActivityID
	err        error
}

func (i *ActivityInfo[T]) Get(output ...interface{}) error {
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
	// fmt.Println("searching id", i.ActivityID.String())
	for {
		select {
		case <-i.tp.ctx.Done():
			return i.tp.ctx.Err()
		case <-ticker.C:
			activityEntity, err := i.tp.client.Activity.Query().Where(activity.IDEQ(i.ActivityID.String())).Only(i.tp.ctx)
			if err != nil {
				// fmt.Println("error searching activity", err)
				return err
			}
			if value, ok = i.tp.activities.Load(HandlerIdentity(activityEntity.Identity)); ok {
				if activityHandlerInfo, ok = value.(Activity); !ok {
					log.Printf("scheduler: activity %s is not handler info", i.ActivityID.String())
					return errors.New("activity is not handler info")
				}

				switch activityEntity.Status {
				// wait for the confirmation that the workflow entity reached a final state
				case activity.StatusCompleted, activity.StatusFailed, activity.StatusCancelled:
					// fmt.Println("searching for activity execution of ", i.ActivityID.String())
					// Then only get the latest activity execution
					// Simply because eventually my children can have retries and i need to let them finish
					latestExec, err := i.tp.client.ActivityExecution.Query().
						Where(
							activityexecution.HasActivityWith(activity.IDEQ(i.ActivityID.String())),
						).
						Order(ent.Desc(activityexecution.FieldStartedAt)).
						First(i.tp.ctx)

					if err != nil {
						if ent.IsNotFound(err) {
							// Handle the case where no execution is found
							log.Printf("No execution found for activity %s", i.ActivityID)
							return fmt.Errorf("no execution found for activity %s", i.ActivityID)
						}
						log.Printf("Error querying activity execution: %v", err)
						return fmt.Errorf("error querying activity execution: %w", err)
					}

					switch latestExec.Status {
					case activityexecution.StatusCompleted:
						outputs, err := i.tp.convertOuputs(HandlerInfo(activityHandlerInfo), latestExec.Output)
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
						return errors.New(latestExec.Error)
					case activityexecution.StatusRetried:
						return errors.New("activity was retried")
					case activityexecution.StatusPending, activityexecution.StatusRunning:
						// The workflow is still in progress
						// return  errors.New("workflow is still in progress")
						runtime.Gosched()
						continue
					}
				default:
					runtime.Gosched()
					continue
				}
			}
			runtime.Gosched()
		}
	}
}
