package tempolite

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"time"

	"github.com/davidroman0O/tempolite/ent/activity"
	"github.com/davidroman0O/tempolite/ent/activityexecution"
)

type ActivityExecutionInfo struct {
	tp          *Tempolite
	ExecutionID ActivityExecutionID
	err         error
}

// Try to find the activity execution until it reaches a final state
func (i *ActivityExecutionInfo) Get(output ...interface{}) error {
	if i.err != nil {
		i.tp.logger.Error(i.tp.ctx, "ActivityExecutionInfo.Get", "executionID", i.ExecutionID, "error", i.err)
		return i.err
	}
	for idx, out := range output {
		if reflect.TypeOf(out).Kind() != reflect.Ptr {
			i.tp.logger.Error(i.tp.ctx, "ActivityExecutionInfo.Get: output parameter is not a pointer", "index", idx)
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
			i.tp.logger.Error(i.tp.ctx, "ActivityExecutionInfo.Get: context done", "executionID", i.ExecutionID)
			return i.tp.ctx.Err()
		case <-ticker.C:
			activityExecEntity, err := i.tp.client.ActivityExecution.Query().Where(activityexecution.IDEQ(i.ExecutionID.String())).WithActivity().Only(i.tp.ctx)
			if err != nil {
				i.tp.logger.Error(i.tp.ctx, "ActivityExecutionInfo.Get: failed to query activity execution", "executionID", i.ExecutionID, "error", err)
				return err
			}
			activityEntity, err := i.tp.client.Activity.Query().Where(activity.IDEQ(activityExecEntity.Edges.Activity.ID)).Only(i.tp.ctx)
			if err != nil {
				i.tp.logger.Error(i.tp.ctx, "ActivityExecutionInfo.Get: failed to query activity", "activityID", activityExecEntity.Edges.Activity.ID, "error", err)
				return err
			}
			if value, ok = i.tp.activities.Load(HandlerIdentity(activityEntity.Identity)); ok {
				if activityHandlerInfo, ok = value.(Activity); !ok {
					i.tp.logger.Error(i.tp.ctx, "ActivityExecutionInfo.Get: activity is not handler info", "activityID", activityExecEntity.Edges.Activity.ID)
					return errors.New("activity is not handler info")
				}
				switch activityExecEntity.Status {
				case activityexecution.StatusCompleted:
					// outputs, err := i.tp.convertOutputs(HandlerInfo(activityHandlerInfo), activityExecEntity.Output)
					deserializedOutput, err := i.tp.convertOutputsFromSerialization(HandlerInfo(activityHandlerInfo), activityExecEntity.Output)

					if err != nil {
						i.tp.logger.Error(i.tp.ctx, "ActivityExecutionInfo.Get: failed to convert outputs", "activityID", activityExecEntity.Edges.Activity.ID, "error", err)
						return err
					}
					if len(output) != len(deserializedOutput) {
						i.tp.logger.Error(i.tp.ctx, "ActivityExecutionInfo.Get: output length mismatch", "activityID", activityExecEntity.Edges.Activity.ID, "expected", len(deserializedOutput), "got", len(output))
						return fmt.Errorf("output length mismatch: expected %d, got %d", len(deserializedOutput), len(output))
					}

					for idx, outPtr := range output {
						outVal := reflect.ValueOf(outPtr).Elem()
						outputVal := reflect.ValueOf(deserializedOutput[idx])

						if outVal.Type() != outputVal.Type() {
							i.tp.logger.Error(i.tp.ctx, "ActivityExecutionInfo.Get: type mismatch", "activityID", activityExecEntity.Edges.Activity.ID, "index", idx, "expected", outVal.Type(), "got", outputVal.Type())
							return fmt.Errorf("type mismatch at index %d: expected %v, got %v", idx, outVal.Type(), outputVal.Type())
						}

						outVal.Set(outputVal)
					}
					return nil
				case activityexecution.StatusFailed:
					i.tp.logger.Error(i.tp.ctx, "ActivityExecutionInfo.Get: activity execution failed", "activityID", activityExecEntity.Edges.Activity.ID, "error", activityExecEntity.Error)
					return errors.New(activityExecEntity.Error)
				case activityexecution.StatusRetried:
					i.tp.logger.Error(i.tp.ctx, "ActivityExecutionInfo.Get: activity was retried", "activityID", activityExecEntity.Edges.Activity.ID)
					return errors.New("activity was retried")
				case activityexecution.StatusPending, activityexecution.StatusRunning:
					i.tp.logger.Debug(i.tp.ctx, "ActivityExecutionInfo.Get: activity execution is still running", "activityID", activityExecEntity.Edges.Activity.ID)
					runtime.Gosched()
					continue
				}
			}
			runtime.Gosched()
		}
	}
}
