package tempolite

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"time"

	"github.com/davidroman0O/go-tempolite/ent"
	"github.com/davidroman0O/go-tempolite/ent/workflow"
	"github.com/davidroman0O/go-tempolite/ent/workflowexecution"
)

type WorkflowInfo[T Identifier] struct {
	tp         *Tempolite[T]
	WorkflowID WorkflowID
	err        error
}

// Try to find the latest workflow execution until it reaches a final state
func (i *WorkflowInfo[T]) Get(output ...interface{}) error {
	defer func() {
		i.tp.logger.Debug(i.tp.ctx, "WorkflowInfo.Get", "workflowID", i.WorkflowID, "error", i.err)
	}()
	if i.err != nil {
		return i.err
	}
	// Check if all output parameters are pointers
	for idx, out := range output {
		if reflect.TypeOf(out).Kind() != reflect.Ptr {
			i.tp.logger.Error(i.tp.ctx, "WorkflowInfo.Get: output parameter is not a pointer", "index", idx)
			return fmt.Errorf("output parameter at index %d is not a pointer", idx)
		}
	}
	ticker := time.NewTicker(time.Second / 16)
	defer ticker.Stop()
	var value any
	var ok bool
	var workflowHandlerInfo Workflow
	for {
		select {
		case <-i.tp.ctx.Done():
			i.tp.logger.Error(i.tp.ctx, "WorkflowInfo.Get: context done", "workflowID", i.WorkflowID)
			return i.tp.ctx.Err()
		case <-ticker.C:
			workflowEntity, err := i.tp.client.Workflow.Query().Where(workflow.IDEQ(i.WorkflowID.String())).Only(i.tp.ctx)
			if err != nil {
				return err
			}
			if value, ok = i.tp.workflows.Load(HandlerIdentity(workflowEntity.Identity)); ok {
				if workflowHandlerInfo, ok = value.(Workflow); !ok {
					i.tp.logger.Error(i.tp.ctx, "WorkflowInfo.Get: workflow is not handler info", "workflowID", i.WorkflowID)
					return errors.New("workflow is not handler info")
				}
				switch workflowEntity.Status {
				// wait for the confirmation that the workflow entity reached a final state
				case workflow.StatusCompleted, workflow.StatusFailed, workflow.StatusCancelled:
					// Then only get the latest workflow execution
					// Simply because eventually my children can have retries and i need to let them finish
					latestExec, err := i.tp.client.WorkflowExecution.Query().
						Where(workflowexecution.HasWorkflowWith(workflow.IDEQ(i.WorkflowID.String()))).
						Order(ent.Desc(workflowexecution.FieldStartedAt)).
						First(i.tp.ctx)
					if err != nil {
						i.tp.logger.Error(i.tp.ctx, "WorkflowInfo.Get: failed to get latest workflow execution", "workflowID", i.WorkflowID)
						return err
					}
					switch latestExec.Status {
					case workflowexecution.StatusCompleted:
						outputs, err := i.tp.convertOuputs(HandlerInfo(workflowHandlerInfo), latestExec.Output)
						if err != nil {
							i.tp.logger.Error(i.tp.ctx, "WorkflowInfo.Get: failed to convert outputs", "error", err)
							return err
						}
						if len(output) != len(outputs) {
							i.tp.logger.Error(i.tp.ctx, "WorkflowInfo.Get: output length mismatch", "expected", len(outputs), "got", len(output))
							return fmt.Errorf("output length mismatch: expected %d, got %d", len(outputs), len(output))
						}

						for idx, outPtr := range output {
							outVal := reflect.ValueOf(outPtr).Elem()
							outputVal := reflect.ValueOf(outputs[idx])

							if outVal.Type() != outputVal.Type() {
								i.tp.logger.Error(i.tp.ctx, "WorkflowInfo.Get: type mismatch", "index", idx, "expected", outVal.Type(), "got", outputVal.Type())
								return fmt.Errorf("type mismatch at index %d: expected %v, got %v", idx, outVal.Type(), outputVal.Type())
							}

							outVal.Set(outputVal)
						}
						return nil
					case workflowexecution.StatusCancelled:
						i.tp.logger.Debug(i.tp.ctx, "WorkflowInfo.Get: workflow was cancelled", "workflowID", i.WorkflowID)
						return fmt.Errorf("workflow %s was cancelled", i.WorkflowID)
					case workflowexecution.StatusFailed:
						i.tp.logger.Debug(i.tp.ctx, "WorkflowInfo.Get: workflow failed", "workflowID", i.WorkflowID, "error", latestExec.Error)
						return errors.New(latestExec.Error)
					case workflowexecution.StatusRetried:
						i.tp.logger.Debug(i.tp.ctx, "WorkflowInfo.Get: workflow was retried", "workflowID", i.WorkflowID)
						return errors.New("workflow was retried")
					case workflowexecution.StatusPending, workflowexecution.StatusRunning:
						i.tp.logger.Debug(i.tp.ctx, "WorkflowInfo.Get: workflow is still running", "workflowID", i.WorkflowID)
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
