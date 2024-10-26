package tempolite

import (
	"runtime"

	"github.com/davidroman0O/tempolite/ent"
	"github.com/davidroman0O/tempolite/ent/activity"
	"github.com/davidroman0O/tempolite/ent/activityexecution"
	"github.com/google/uuid"
)

func (tp *Tempolite) schedulerExecutionActivity() {
	defer close(tp.schedulerExecutionActivityDone)
	for {
		select {
		case <-tp.ctx.Done():
			return
		default:

			// Get number of available worker slots
			availableSlots := len(tp.ListWorkersActivity()) - tp.activityPool.ProcessingCount()
			if availableSlots <= 0 {
				// All workers are busy, wait and continue
				runtime.Gosched()
				continue
			}

			pendingActivities, err := tp.client.ActivityExecution.Query().
				Where(activityexecution.StatusEQ(activityexecution.StatusPending)).
				Order(ent.Asc(activityexecution.FieldStartedAt)).WithActivity().
				Limit(availableSlots).
				All(tp.ctx)
			if err != nil {
				tp.logger.Error(tp.ctx, "scheduler activity execution: ActivityExecution.Query failed", "error", err)
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
					tp.logger.Error(tp.ctx, "scheduler activity execution: Activity.Get failed", "error", err)
					continue
				}

				if value, ok = tp.activities.Load(HandlerIdentity(activityEntity.Identity)); ok {
					var activityHandlerInfo Activity
					if activityHandlerInfo, ok = value.(Activity); !ok {
						// could be development bug
						tp.logger.Error(tp.ctx, "scheduler activity execution: activity is missing handler info", "activity", activityEntity.HandlerName)
						continue
					}

					// inputs := []interface{}{}

					// for idx, rawInput := range activityEntity.Input {
					// 	inputType := activityHandlerInfo.ParamTypes[idx]

					// 	inputKind := activityHandlerInfo.ParamsKinds[idx]

					// 	realInput, err := convertIO(rawInput, inputType, inputKind)
					// 	if err != nil {
					// 		tp.logger.Error(tp.ctx, "scheduler activity execution: convertInput failed", "error", err)
					// 		continue
					// 	}

					// 	inputs = append(inputs, realInput)
					// }

					inputs, err := tp.convertInputsFromSerialization(HandlerInfo(activityHandlerInfo), activityEntity.Input)
					if err != nil {
						tp.logger.Error(tp.ctx, "Scheduler workflow execution: convertInputsFromSerialization failed", "error", err)
						continue
					}

					contextActivity := ActivityContext{
						tp:          tp,
						activityID:  activityEntity.ID,
						executionID: act.ID,
						runID:       act.RunID,
						stepID:      activityEntity.StepID,
					}

					task := &activityTask{
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
							tp.logger.Error(tp.ctx, "scheduler activity execution: ActivityExecution.Create failed", "error", err)
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
						tp.logger.Error(tp.ctx, "scheduler activity execution: ActivityExecution.Query failed", "error", err)
						continue
					}

					// If it's not me
					if total > 1 {
						task.retryCount = total
					}

					tp.logger.Debug(tp.ctx, "scheduler: Dispatching activity", "activity", activityEntity.HandlerName, "params", activityEntity.Input)

					if err := tp.activityPool.Dispatch(task); err != nil {
						tp.logger.Error(tp.ctx, "scheduler activity execution: Dispatch failed", "error", err)

						// Start transaction for status updates
						tx, err := tp.client.Tx(tp.ctx)
						if err != nil {
							tp.logger.Error(tp.ctx, "Failed to start transaction for status updates", "error", err)
							continue
						}

						if _, err = tx.ActivityExecution.UpdateOneID(act.ID).SetStatus(activityexecution.StatusFailed).SetError(err.Error()).Save(tp.ctx); err != nil {
							tp.logger.Error(tp.ctx, "scheduler activity execution: ActivityExecution.UpdateOneID failed when dispatched", "error", err)
							if rerr := tx.Rollback(); rerr != nil {
								tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
							}
							continue
						}

						if _, err = tx.Activity.UpdateOneID(activityEntity.ID).SetStatus(activity.StatusFailed).Save(tp.ctx); err != nil {
							tp.logger.Error(tp.ctx, "scheduler activity execution: Activity.UpdateOne failed when dispatched", "error", err)
							if rerr := tx.Rollback(); rerr != nil {
								tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
							}
							continue
						}

						if err = tx.Commit(); err != nil {
							tp.logger.Error(tp.ctx, "Failed to commit transaction for status updates", "error", err)
						}
						continue
					}

					// Start transaction for status updates
					tx, err := tp.client.Tx(tp.ctx)
					if err != nil {
						tp.logger.Error(tp.ctx, "Failed to start transaction for status updates", "error", err)
						continue
					}

					if _, err = tx.ActivityExecution.UpdateOneID(act.ID).SetStatus(activityexecution.StatusRunning).Save(tp.ctx); err != nil {
						tp.logger.Error(tp.ctx, "scheduler activity execution: ActivityExecution.UpdateOneID failed", "error", err)
						if rerr := tx.Rollback(); rerr != nil {
							tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
						}
						continue
					}

					if _, err = tx.Activity.UpdateOneID(activityEntity.ID).SetStatus(activity.StatusRunning).Save(tp.ctx); err != nil {
						tp.logger.Error(tp.ctx, "scheduler activity execution: Activity.UpdateOne failed", "error", err)
						if rerr := tx.Rollback(); rerr != nil {
							tp.logger.Error(tp.ctx, "Failed to rollback transaction", "error", rerr)
						}
						continue
					}

					if err = tx.Commit(); err != nil {
						tp.logger.Error(tp.ctx, "Failed to commit transaction for status updates", "error", err)
					}

				} else {
					tp.logger.Error(tp.ctx, "scheduler activity execution: Activity not found", "activity", act.Edges.Activity.HandlerName)
					continue
				}
			}

			runtime.Gosched()
		}
	}
}
