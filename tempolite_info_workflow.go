package tempolite

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"time"

	"github.com/davidroman0O/tempolite/ent"
	"github.com/davidroman0O/tempolite/ent/workflow"
	"github.com/davidroman0O/tempolite/ent/workflowexecution"
)

// HandlerSwitch is a type-safe switch for workflow handlers
type HandlerSwitch struct {
	info    *WorkflowInfo
	cases   []reflect.Value
	actions []reflect.Value
	err     error
}

// Switch creates a new HandlerSwitch for the workflow info
func (i *WorkflowInfo) Switch() *HandlerSwitch {
	return &HandlerSwitch{
		info:    i,
		cases:   make([]reflect.Value, 0),
		actions: make([]reflect.Value, 0),
	}
}

// Case adds a handler case to the switch
func (s *HandlerSwitch) Case(handler interface{}, action func() error) *HandlerSwitch {
	s.cases = append(s.cases, reflect.ValueOf(handler))
	s.actions = append(s.actions, reflect.ValueOf(action))
	return s
}

// Default adds a default case to handle unmatched handlers
func (s *HandlerSwitch) Default(action func() error) *HandlerSwitch {
	s.cases = append(s.cases, reflect.Value{})
	s.actions = append(s.actions, reflect.ValueOf(action))
	return s
}

// End executes the switch and returns any error
func (s *HandlerSwitch) End() error {
	if s.info.err != nil {
		return s.info.err // Don't execute if workflow info has an error
	}

	handlerValue := reflect.ValueOf(s.info.handler)

	// Handle nil handler
	if !handlerValue.IsValid() {
		for i, caseHandler := range s.cases {
			if !caseHandler.IsValid() { // Found default case
				if action := s.actions[i].Interface().(func() error); action != nil {
					return action()
				}
				return nil
			}
		}
		return fmt.Errorf("no handler found and no default case")
	}

	// Compare handler with cases
	for i, caseHandler := range s.cases {
		if !caseHandler.IsValid() { // Skip default case
			continue
		}

		if handlerValue.Pointer() == caseHandler.Pointer() {
			// Call the matching action
			if action := s.actions[i].Interface().(func() error); action != nil {
				return action()
			}
			return nil
		}
	}

	// Try default case if no match found
	for i, caseHandler := range s.cases {
		if !caseHandler.IsValid() { // Found default case
			if action := s.actions[i].Interface().(func() error); action != nil {
				return action()
			}
			return nil
		}
	}

	return fmt.Errorf("no matching handler found and no default case")
}

type WorkflowInfo struct {
	tp          *Tempolite
	WorkflowID  WorkflowID
	err         error
	handler     interface{}
	IsContinued bool
}

func (i *WorkflowInfo) GetContinuation() *WorkflowInfo {
	workflowEntity, err := i.tp.client.Workflow.Get(i.tp.ctx, i.WorkflowID.String())
	if err != nil {
		return &WorkflowInfo{tp: i.tp, err: err}
	}

	nextWorkflow, err := i.tp.client.Workflow.Query().
		Where(workflow.ContinuedFromIDEQ(workflowEntity.ID)).
		First(i.tp.ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return &WorkflowInfo{tp: i.tp, err: err}
		}
		return &WorkflowInfo{tp: i.tp, err: err}
	}

	info := i.tp.getWorkflow(nil, WorkflowID(nextWorkflow.ID), nil)
	return info
}

// Try to find the latest workflow execution until it reaches a final state
func (i *WorkflowInfo) Get(output ...interface{}) error {
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

			// Check for continuation
			_, err = i.tp.client.Workflow.Query().
				Where(workflow.ContinuedFromIDEQ(workflowEntity.ID)).
				First(i.tp.ctx)
			if err == nil {
				i.IsContinued = true
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
						if len(output) == 0 {
							return nil
						}

						deserializedOutput, err := i.tp.convertOutputsFromSerialization(HandlerInfo(workflowHandlerInfo), latestExec.Output)
						if err != nil {
							i.tp.logger.Error(i.tp.ctx, "WorkflowInfo.Get: failed to convert outputs", "error", err)
							return err
						}

						if len(output) != len(deserializedOutput) {
							i.tp.logger.Error(i.tp.ctx, "WorkflowInfo.Get: output length mismatch", "expected", len(deserializedOutput), "got", len(output))
							return fmt.Errorf("output length mismatch: expected %d, got %d", len(deserializedOutput), len(output))
						}

						for idx, outPtr := range output {
							outVal := reflect.ValueOf(outPtr).Elem()
							outputVal := reflect.ValueOf(deserializedOutput[idx])

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
