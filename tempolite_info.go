package tempolite

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"runtime"
	"time"

	"github.com/davidroman0O/go-tempolite/ent"
	"github.com/davidroman0O/go-tempolite/ent/activity"
	"github.com/davidroman0O/go-tempolite/ent/activityexecution"
	"github.com/davidroman0O/go-tempolite/ent/workflow"
	"github.com/davidroman0O/go-tempolite/ent/workflowexecution"
)

type WorkflowInfo struct {
	tp         *Tempolite
	WorkflowID WorkflowID
	err        error
}

// Try to find the latest workflow execution until it reaches a final state
func (i *WorkflowInfo) Get(output ...interface{}) error {
	if i.err != nil {
		return i.err
	}
	// Check if all output parameters are pointers
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
			workflowEntity, err := i.tp.client.Workflow.Query().Where(workflow.IDEQ(i.WorkflowID.String())).Only(i.tp.ctx)
			if err != nil {
				return err
			}
			if value, ok = i.tp.workflows.Load(HandlerIdentity(workflowEntity.Identity)); ok {
				if workflowHandlerInfo, ok = value.(Workflow); !ok {
					log.Printf("scheduler: workflow %s is not handler info", i.WorkflowID)
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
						return err
					}
					switch latestExec.Status {
					case workflowexecution.StatusCompleted:
						outputs, err := i.tp.convertOuputs(HandlerInfo(workflowHandlerInfo), latestExec.Output)
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
						return fmt.Errorf("workflow %s was cancelled", i.WorkflowID)
					case workflowexecution.StatusFailed:
						return errors.New(latestExec.Error)
					case workflowexecution.StatusRetried:
						return errors.New("workflow was retried")
					case workflowexecution.StatusPending, workflowexecution.StatusRunning:
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

type WorkflowExecutionInfo struct {
	tp          *Tempolite
	ExecutionID WorkflowExecutionID
	err         error
}

// Try to find the workflow execution until it reaches a final state
func (i *WorkflowExecutionInfo) Get(output ...interface{}) error {

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

func (i *WorkflowExecutionInfo) Cancel() error {
	// todo: implement
	return nil
}

func (i *WorkflowExecutionInfo) Pause() error {
	// todo: implement
	return nil
}

func (i *WorkflowExecutionInfo) Resume() error {
	// todo: implement
	return nil
}

type ActivityInfo struct {
	tp         *Tempolite
	ActivityID ActivityID
	err        error
}

func (i *ActivityInfo) Get(output ...interface{}) error {
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
	fmt.Println("searching id", i.ActivityID.String())
	for {
		select {
		case <-i.tp.ctx.Done():
			return i.tp.ctx.Err()
		case <-ticker.C:
			activityEntity, err := i.tp.client.Activity.Query().Where(activity.IDEQ(i.ActivityID.String())).Only(i.tp.ctx)
			if err != nil {
				fmt.Println("error searching activity", err)
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
					fmt.Println("searching for activity execution of ", i.ActivityID.String())
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

type ActivityExecutionInfo struct {
	tp          *Tempolite
	ExecutionID ActivityExecutionID
	err         error
}

// Try to find the activity execution until it reaches a final state
func (i *ActivityExecutionInfo) Get(output ...interface{}) error {
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
			fmt.Println("searching id", i.ExecutionID.String())
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

type SideEffectInfo struct {
	tp    *Tempolite
	ID    string
	RunID string
}

func (i *SideEffectInfo) Get() ([]interface{}, error) {
	// todo: implement
	return nil, nil
}

type SagaInfo struct {
	tp    *Tempolite
	ID    string
	RunID string
}

func (i *SagaInfo) Get() ([]interface{}, error) {
	// todo: implement
	return nil, nil
}

type TempoliteContext interface {
	context.Context
	RunID() string
	EntityID() string
	ExecutionID() string
	EntityType() string
}
type WorkflowContext struct {
	TempoliteContext
	tp              *Tempolite
	workflowID      string
	executionID     string
	runID           string
	workflowType    string
	activityCounter int
}

func (ctx *WorkflowContext) NextActivityID() string {
	ctx.activityCounter++
	return fmt.Sprintf("%s-%d", ctx.runID, ctx.activityCounter)
}

func (w WorkflowContext) RunID() string {
	return w.runID
}

func (w WorkflowContext) EntityID() string {
	return w.workflowID
}

func (w WorkflowContext) ExecutionID() string {
	return w.executionID
}

func (w WorkflowContext) EntityType() string {
	return "workflow"
}

func (w WorkflowContext) SideEffect(uniqueID string, handler interface{}, inputs ...any) (*SideEffectInfo, error) {
	return nil, nil
}

func (w WorkflowContext) GetVersion(changeID string, minSupported, maxSupported int) int {
	log.Printf("GetVersion called for workflowType: %s, workflowID: %s, changeID: %s, min: %d, max: %d", w.workflowType, w.workflowID, changeID, minSupported, maxSupported)
	version, err := w.tp.getOrCreateVersion(w.workflowType, w.workflowID, changeID, minSupported, maxSupported)
	if err != nil {
		log.Printf("Error getting version for workflowType %s, changeID %s: %v", w.workflowType, changeID, err)
		return minSupported
	}
	log.Printf("GetVersion returned %d for workflowType: %s, changeID: %s", version, w.workflowType, changeID)
	return version
}

// Since I don't want to hide any implementation, when the WorkflowInfo call Pause/Resume, the moment the Yield() is called, the workflow will be paused or resume if called Resume.
func (w WorkflowContext) Yield() error {
	// todo: implement - could use ConsumerSignalChannel and ProducerSignalChannel behind the scene
	return nil
}

func (w WorkflowContext) ContinueAsNew(ctx WorkflowContext, values ...any) error {
	// todo: implement
	return nil
}

func (w WorkflowContext) GetWorkflow(id WorkflowExecutionID) *WorkflowExecutionInfo {
	return w.tp.getWorkflowExecution(w, id, nil)
}

func (w WorkflowContext) ExecuteWorkflow(handler interface{}, inputs ...any) *WorkflowExecutionInfo {
	id, err := w.tp.enqueueSubWorkflow(w, handler, inputs...)
	// if err != nil {
	// 	return nil, err
	// }
	return w.tp.getWorkflowExecution(w, id, err)
}

func (w WorkflowContext) ExecuteActivity(name HandlerIdentity, inputs ...any) *ActivityExecutionInfo {
	id, err := w.tp.enqueueSubActivityExecution(w, name, inputs...)
	if err != nil {
		log.Printf("Error enqueuing activity execution: %v", err)
	}
	fmt.Println("\t \t activity execution id", id, err)
	return w.tp.getActivityExecution(w, id, err)
}

func (w WorkflowContext) GetActivity(id string) (*ActivityInfo, error) {
	// todo: implement
	return nil, nil
}

func (w WorkflowContext) ExecuteSideEffect(name HandlerIdentity, inputs ...any) (*SideEffectInfo, error) {
	// todo: implement
	return nil, nil
}

type ActivityContext struct {
	TempoliteContext
	tp          *Tempolite
	activityID  string
	executionID string
	runID       string
}

func (w ActivityContext) RunID() string {
	return w.runID
}

func (w ActivityContext) EntityID() string {
	return w.activityID
}

func (w ActivityContext) ExecutionID() string {
	return w.executionID
}

func (w ActivityContext) EntityType() string {
	return "activity"
}

// func (w ActivityContext) ExecuteActivity(name HandlerIdentity, inputs ...any) *ActivityExecutionInfo {
// 	id, err := w.tp.enqueueSubActivityExecution(w, name, inputs...)

// 	return w.tp.getActivityExecution(w, id, err)
// }

func (w ActivityContext) GetActivity(id ActivityExecutionID) *ActivityExecutionInfo {
	return w.tp.getActivityExecution(w, id, nil)
}

type SideEffectContext struct {
	TempoliteContext
	tp           *Tempolite
	sideEffectID string
	executionID  string
	runID        string
}

func (w SideEffectContext) RunID() string {
	return w.runID
}

func (w SideEffectContext) EntityID() string {
	return w.sideEffectID
}

func (w SideEffectContext) ExecutionID() string {
	return w.executionID
}

func (w SideEffectContext) EntityType() string {
	return "sideEffect"
}

type TransactionContext struct {
	TempoliteContext
	tp *Tempolite
}

func (w TransactionContext) GetActivity(id string) (*ActivityInfo, error) {
	// todo: implement
	return nil, nil
}

func (w TransactionContext) ExecuteActivity(name HandlerIdentity, inputs ...any) (*ActivityInfo, error) {
	// todo: implement
	return nil, nil
}

func (w TransactionContext) ExecuteSideEffect(name HandlerIdentity, inputs ...any) (*SideEffectInfo, error) {
	// todo: implement
	return nil, nil
}

func (w TransactionContext) EntityType() string {
	return "transaction"
}

type CompensationContext struct {
	TempoliteContext
	tp *Tempolite
}

func (w CompensationContext) GetActivity(id string) (*ActivityInfo, error) {
	// todo: implement
	return nil, nil
}

func (w CompensationContext) ExecuteActivity(name HandlerIdentity, inputs ...any) (*ActivityInfo, error) {
	// todo: implement
	return nil, nil
}

func (w CompensationContext) EntityType() string {
	return "compensation"
}
