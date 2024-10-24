package tempolite

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"time"

	"github.com/davidroman0O/tempolite/ent/workflow"
	"github.com/davidroman0O/tempolite/ent/workflowexecution"
)

type WorkflowExecutionInfo struct {
	tp          *Tempolite
	ExecutionID WorkflowExecutionID
	err         error
}

// Try to find the workflow execution until it reaches a final state
func (i *WorkflowExecutionInfo) Get(output ...interface{}) error {

	if i.err != nil {
		i.tp.logger.Error(i.tp.ctx, "WorkflowExecutionInfo.Get", "executionID", i.ExecutionID, "error", i.err)
		return i.err
	}

	for idx, out := range output {
		if reflect.TypeOf(out).Kind() != reflect.Ptr {
			i.tp.logger.Error(i.tp.ctx, "WorkflowExecutionInfo.Get: output parameter is not a pointer", "index", idx)
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
			i.tp.logger.Error(i.tp.ctx, "WorkflowExecutionInfo.Get: context done", "executionID", i.ExecutionID)
			return i.tp.ctx.Err()
		case <-ticker.C:
			workflowExecEntity, err := i.tp.client.WorkflowExecution.Query().Where(workflowexecution.IDEQ(i.ExecutionID.String())).WithWorkflow().Only(i.tp.ctx)
			if err != nil {
				i.tp.logger.Error(i.tp.ctx, "WorkflowExecutionInfo.Get: failed to query workflow execution", "executionID", i.ExecutionID, "error", err)
				return err
			}
			workflowEntity, err := i.tp.client.Workflow.Query().Where(workflow.IDEQ(workflowExecEntity.Edges.Workflow.ID)).Only(i.tp.ctx)
			if err != nil {
				i.tp.logger.Error(i.tp.ctx, "WorkflowExecutionInfo.Get: failed to query workflow", "workflowID", workflowExecEntity.Edges.Workflow.ID, "error", err)
				return err
			}
			if value, ok = i.tp.workflows.Load(HandlerIdentity(workflowEntity.Identity)); ok {
				if workflowHandlerInfo, ok = value.(Workflow); !ok {
					i.tp.logger.Error(i.tp.ctx, "WorkflowExecutionInfo.Get: workflow is not handler info", "workflowID", workflowExecEntity.Edges.Workflow.ID)
					return errors.New("workflow is not handler info")
				}
				switch workflowExecEntity.Status {
				case workflowexecution.StatusCompleted:
					outputs, err := i.tp.convertOuputs(HandlerInfo(workflowHandlerInfo), workflowExecEntity.Output)
					if err != nil {
						i.tp.logger.Error(i.tp.ctx, "WorkflowExecutionInfo.Get: failed to convert outputs", "error", err)
						return err
					}
					if len(output) != len(outputs) {
						i.tp.logger.Error(i.tp.ctx, "WorkflowExecutionInfo.Get: output length mismatch", "expected", len(outputs), "got", len(output))
						return fmt.Errorf("output length mismatch: expected %d, got %d", len(outputs), len(output))
					}

					for idx, outPtr := range output {
						outVal := reflect.ValueOf(outPtr).Elem()
						outputVal := reflect.ValueOf(outputs[idx])

						if outVal.Type() != outputVal.Type() {
							i.tp.logger.Error(i.tp.ctx, "WorkflowExecutionInfo.Get: type mismatch", "index", idx, "expected", outVal.Type(), "got", outputVal.Type())
							return fmt.Errorf("type mismatch at index %d: expected %v, got %v", idx, outVal.Type(), outputVal.Type())
						}

						outVal.Set(outputVal)
					}
					return nil
				case workflowexecution.StatusCancelled:
					i.tp.logger.Debug(i.tp.ctx, "WorkflowExecutionInfo.Get: workflow was cancelled", "workflowID", workflowExecEntity.Edges.Workflow.ID)
					return fmt.Errorf("workflow %s was cancelled", workflowExecEntity.Edges.Workflow.ID)
				case workflowexecution.StatusFailed:
					i.tp.logger.Debug(i.tp.ctx, "WorkflowExecutionInfo.Get: workflow failed", "workflowID", workflowExecEntity.Edges.Workflow.ID, "error", workflowExecEntity.Error)
					return errors.New(workflowExecEntity.Error)
				case workflowexecution.StatusRetried:
					i.tp.logger.Debug(i.tp.ctx, "WorkflowExecutionInfo.Get: workflow was retried", "workflowID", workflowExecEntity.Edges.Workflow.ID)
					return errors.New("workflow was retried")
				case workflowexecution.StatusPending, workflowexecution.StatusRunning:
					i.tp.logger.Debug(i.tp.ctx, "WorkflowExecutionInfo.Get: workflow is still running", "workflowID", workflowExecEntity.Edges.Workflow.ID)
					runtime.Gosched()
					continue
				}
			}
			runtime.Gosched()
		}
	}
}
