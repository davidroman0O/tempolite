package tempolite

import (
	"log"
	"runtime"

	"github.com/davidroman0O/go-tempolite/ent"
	"github.com/davidroman0O/go-tempolite/ent/activity"
	"github.com/davidroman0O/go-tempolite/ent/activityexecution"
	"github.com/google/uuid"
)

func (tp *Tempolite[T]) schedulerExecutionActivity() {
	defer close(tp.schedulerExecutionActivityDone)
	for {
		select {
		case <-tp.ctx.Done():
			return
		default:

			pendingActivities, err := tp.client.ActivityExecution.Query().
				Where(activityexecution.StatusEQ(activityexecution.StatusPending)).
				Order(ent.Asc(activityexecution.FieldStartedAt)).WithActivity().
				Limit(1).All(tp.ctx)
			if err != nil {
				log.Printf("scheduler: ActivityExecution.Query failed: %v", err)
				continue
			}

			tp.schedulerActivityStarted.Store(true)

			if len(pendingActivities) == 0 {
				continue
			}

			var value any
			var ok bool

			for _, act := range pendingActivities {

				var activityEntity *ent.Activity
				if activityEntity, err = tp.client.Activity.Get(tp.ctx, act.Edges.Activity.ID); err != nil {

					log.Printf("scheduler: Activity.Get failed: %v", err)
					continue
				}

				if value, ok = tp.activities.Load(HandlerIdentity(activityEntity.Identity)); ok {
					var activityHandlerInfo Activity
					if activityHandlerInfo, ok = value.(Activity); !ok {
						// could be development bug
						log.Printf("scheduler: activity %s is not handler info", activityEntity.HandlerName)
						continue
					}

					inputs := []interface{}{}

					for idx, rawInput := range activityEntity.Input {
						inputType := activityHandlerInfo.ParamTypes[idx]

						inputKind := activityHandlerInfo.ParamsKinds[idx]

						realInput, err := convertIO(rawInput, inputType, inputKind)
						if err != nil {
							log.Printf("scheduler: convertInput failed: %v", err)
							continue
						}

						inputs = append(inputs, realInput)
					}

					contextActivity := ActivityContext[T]{
						tp:          tp,
						activityID:  activityEntity.ID,
						executionID: act.ID,
						runID:       act.RunID,
						stepID:      activityEntity.StepID,
					}

					task := &activityTask[T]{
						ctx:         contextActivity,
						handlerName: activityHandlerInfo.HandlerLongName,
						handler:     activityHandlerInfo.Handler,
						params:      inputs,
						maxRetry:    activityEntity.RetryPolicy.MaximumAttempts,
						retryCount:  0,
					}

					retryIt := func() error {

						// create a new execution for the same activity
						var activityExecution *ent.ActivityExecution
						if activityExecution, err = tp.client.ActivityExecution.
							Create().
							SetID(uuid.NewString()).
							SetRunID(act.RunID).
							SetActivity(activityEntity).
							Save(tp.ctx); err != nil {
							log.Printf("ERROR scheduler: ActivityExecution.Create failed: %v", err)
							return err
						}

						// update the current execution id
						task.ctx.executionID = activityExecution.ID
						task.retryCount++

						return nil
					}

					task.retry = retryIt

					// query the count of how many activity execution exists related to the activityEntity
					// > but but why are you getting the count?!?!
					// well maybe if we crashed, then when re-enqueueing the activity, we can prepare the retry count and continue our work
					total, err := tp.client.ActivityExecution.Query().Where(activityexecution.HasActivityWith(activity.IDEQ(activityEntity.ID))).Count(tp.ctx)
					if err != nil {
						log.Printf("scheduler: ActivityExecution.Query failed: %v", err)
						continue
					}

					log.Printf("scheduler: total: %d", total)

					// If it's not me
					if total > 1 {
						task.retryCount = total
					}

					log.Printf("scheduler: Dispatching activity %s with params: %v", activityEntity.HandlerName, activityEntity.Input)

					if err := tp.activityPool.Dispatch(task); err != nil {
						log.Printf("scheduler: Dispatch failed: %v", err)
						continue
					}

					if _, err = tp.client.ActivityExecution.UpdateOneID(act.ID).SetStatus(activityexecution.StatusRunning).Save(tp.ctx); err != nil {
						log.Printf("scheduler: ActivityExecution.UpdateOneID failed: %v", err)
						continue
					}

				} else {
					log.Printf("scheduler: Activity %s not found", act.Edges.Activity.HandlerName)
					continue
				}
			}

			runtime.Gosched()
		}
	}
}
