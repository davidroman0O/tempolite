package tempolite

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"runtime"
	"time"

	"github.com/davidroman0O/go-tempolite/ent/workflow"
	"github.com/davidroman0O/go-tempolite/ent/workflowexecution"
)

type WorkflowExecutionInfo[T Identifier] struct {
	tp          *Tempolite[T]
	ExecutionID WorkflowExecutionID
	err         error
}

// Try to find the workflow execution until it reaches a final state
func (i *WorkflowExecutionInfo[T]) Get(output ...interface{}) error {

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
	var workflowHandlerInfo Workflow
	for {
		select {
		case <-i.tp.ctx.Done():
			return i.tp.ctx.Err()
		case <-ticker.C:
			workflowExecEntity, err := i.tp.client.WorkflowExecution.Query().Where(workflowexecution.IDEQ(i.ExecutionID.String())).WithWorkflow().Only(i.tp.ctx)
			if err != nil {
				return err
			}
			workflowEntity, err := i.tp.client.Workflow.Query().Where(workflow.IDEQ(workflowExecEntity.Edges.Workflow.ID)).Only(i.tp.ctx)
			if err != nil {
				return err
			}
			if value, ok = i.tp.workflows.Load(HandlerIdentity(workflowEntity.Identity)); ok {
				if workflowHandlerInfo, ok = value.(Workflow); !ok {
					log.Printf("scheduler: workflow %s is not handler info", workflowExecEntity.Edges.Workflow.ID)
					return errors.New("workflow is not handler info")
				}
				switch workflowExecEntity.Status {
				case workflowexecution.StatusCompleted:
					outputs, err := i.tp.convertOuputs(HandlerInfo(workflowHandlerInfo), workflowExecEntity.Output)
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
				case workflowexecution.StatusCancelled:
					return fmt.Errorf("workflow %s was cancelled", workflowExecEntity.Edges.Workflow.ID)
				case workflowexecution.StatusFailed:
					return errors.New(workflowExecEntity.Error)
				case workflowexecution.StatusRetried:
					return errors.New("workflow was retried")
				case workflowexecution.StatusPending, workflowexecution.StatusRunning:
					runtime.Gosched()
					continue
				}
			}
			runtime.Gosched()
		}
	}
}
