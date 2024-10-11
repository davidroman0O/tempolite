package tempolite

import (
	"context"
	"errors"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/davidroman0O/go-tempolite/ent"
	"github.com/davidroman0O/go-tempolite/ent/activity"
	"github.com/davidroman0O/go-tempolite/ent/activityexecution"
	"github.com/davidroman0O/go-tempolite/ent/workflow"
	"github.com/davidroman0O/go-tempolite/ent/workflowexecution"
	"github.com/k0kubun/pp/v3"
)

type WorkflowInfo struct {
	tp         *Tempolite
	WorkflowID WorkflowID
}

// Try to find the latest workflow execution until it reaches a final state
func (i *WorkflowInfo) Get() ([]interface{}, error) {
	ticker := time.NewTicker(time.Second / 16)
	defer ticker.Stop()
	var value any
	var ok bool
	var workflowHandlerInfo Workflow
	for {
		select {
		case <-i.tp.ctx.Done():
			return nil, i.tp.ctx.Err()
		case <-ticker.C:

			workflowEntity, err := i.tp.client.Workflow.Query().Where(workflow.IDEQ(i.WorkflowID.String())).Only(i.tp.ctx)
			if err != nil {
				return nil, err
			}
			if value, ok = i.tp.workflows.Load(HandlerIdentity(workflowEntity.Identity)); ok {
				if workflowHandlerInfo, ok = value.(Workflow); !ok {
					log.Printf("scheduler: workflow %s is not handler info", i.WorkflowID)
					return nil, errors.New("workflow is not handler info")
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
						return nil, err
					}
					switch latestExec.Status {
					case workflowexecution.StatusCompleted:
						return i.tp.convertOuputs(HandlerInfo(workflowHandlerInfo), latestExec.Output)
					case workflowexecution.StatusCancelled:
						return nil, fmt.Errorf("workflow %s was cancelled", i.WorkflowID)
					case workflowexecution.StatusFailed:
						return nil, errors.New(latestExec.Error)
					case workflowexecution.StatusRetried:
						return nil, errors.New("workflow was retried")
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
}

// Try to find the workflow execution until it reaches a final state
func (i *WorkflowExecutionInfo) Get() ([]interface{}, error) {
	ticker := time.NewTicker(time.Second / 16)
	defer ticker.Stop()
	var value any
	var ok bool
	var workflowHandlerInfo Workflow
	for {
		select {
		case <-i.tp.ctx.Done():
			return nil, i.tp.ctx.Err()
		case <-ticker.C:
			workflowExecEntity, err := i.tp.client.WorkflowExecution.Query().Where(workflowexecution.IDEQ(i.ExecutionID.String())).WithWorkflow().Only(i.tp.ctx)
			if err != nil {
				return nil, err
			}
			workflowEntity, err := i.tp.client.Workflow.Query().Where(workflow.IDEQ(workflowExecEntity.Edges.Workflow.ID)).Only(i.tp.ctx)
			if err != nil {
				return nil, err
			}
			if value, ok = i.tp.workflows.Load(HandlerIdentity(workflowEntity.Identity)); ok {
				if workflowHandlerInfo, ok = value.(Workflow); !ok {
					log.Printf("scheduler: workflow %s is not handler info", workflowExecEntity.Edges.Workflow.ID)
					return nil, errors.New("workflow is not handler info")
				}
				switch workflowExecEntity.Status {
				case workflowexecution.StatusCompleted:
					return i.tp.convertOuputs(HandlerInfo(workflowHandlerInfo), workflowExecEntity.Output)
				case workflowexecution.StatusCancelled:
					return nil, fmt.Errorf("workflow %s was cancelled", workflowExecEntity.Edges.Workflow.ID)
				case workflowexecution.StatusFailed:
					return nil, errors.New(workflowExecEntity.Error)
				case workflowexecution.StatusRetried:
					return nil, errors.New("workflow was retried")
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
}

func (i *ActivityInfo) Get() ([]interface{}, error) {
	ticker := time.NewTicker(time.Second / 16)
	defer ticker.Stop()
	var value any
	var ok bool
	var activityHandlerInfo Activity
	fmt.Println("searching id", i.ActivityID.String())
	for {
		select {
		case <-i.tp.ctx.Done():
			return nil, i.tp.ctx.Err()
		case <-ticker.C:
			activityEntity, err := i.tp.client.Activity.Query().Where(activity.IDEQ(i.ActivityID.String())).Only(i.tp.ctx)
			if err != nil {
				fmt.Println("error searching activity", err)
				return nil, err
			}
			if value, ok = i.tp.activities.Load(HandlerIdentity(activityEntity.Identity)); ok {
				if activityHandlerInfo, ok = value.(Activity); !ok {
					log.Printf("scheduler: activity %s is not handler info", i.ActivityID.String())
					return nil, errors.New("activity is not handler info")
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
							return nil, fmt.Errorf("no execution found for activity %s", i.ActivityID)
						}
						log.Printf("Error querying activity execution: %v", err)
						return nil, fmt.Errorf("error querying activity execution: %w", err)
					}

					switch latestExec.Status {
					case activityexecution.StatusCompleted:
						return i.tp.convertOuputs(HandlerInfo(activityHandlerInfo), latestExec.Output)
					case activityexecution.StatusFailed:
						return nil, errors.New(latestExec.Error)
					case activityexecution.StatusRetried:
						return nil, errors.New("activity was retried")
					case activityexecution.StatusPending, activityexecution.StatusRunning:
						// The workflow is still in progress
						// return nil, errors.New("workflow is still in progress")
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
}

// Try to find the activity execution until it reaches a final state
func (i *ActivityExecutionInfo) Get() ([]interface{}, error) {
	ticker := time.NewTicker(time.Second / 16)
	defer ticker.Stop()
	var value any
	var ok bool
	var activityHandlerInfo Activity
	for {
		select {
		case <-i.tp.ctx.Done():
			return nil, i.tp.ctx.Err()
		case <-ticker.C:
			activityExecEntity, err := i.tp.client.ActivityExecution.Query().Where(activityexecution.IDEQ(i.ExecutionID.String())).WithActivity().Only(i.tp.ctx)
			if err != nil {
				return nil, err
			}
			activityEntity, err := i.tp.client.Activity.Query().Where(activity.IDEQ(activityExecEntity.Edges.Activity.ID)).Only(i.tp.ctx)
			if err != nil {
				return nil, err
			}
			if value, ok = i.tp.activities.Load(HandlerIdentity(activityEntity.Identity)); ok {
				if activityHandlerInfo, ok = value.(Activity); !ok {
					log.Printf("scheduler: activity %s is not handler info", activityExecEntity.Edges.Activity.ID)
					return nil, errors.New("activity is not handler info")
				}
				switch activityExecEntity.Status {
				case activityexecution.StatusCompleted:
					pp.Println("inputs:", activityExecEntity.Edges.Activity.Input)
					pp.Println("outputs", activityExecEntity.Output)
					return i.tp.convertOuputs(HandlerInfo(activityHandlerInfo), activityExecEntity.Output)
				case activityexecution.StatusFailed:
					return nil, errors.New(activityExecEntity.Error)
				case activityexecution.StatusRetried:
					return nil, errors.New("activity was retried")
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
	tp          *Tempolite
	workflowID  string
	executionID string
	runID       string
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

// Since I don't want to hide any implementation, when the WorkflowInfo call Pause/Resume, the moment the Yield() is called, the workflow will be paused or resume if called Resume.
func (w WorkflowContext) Yield() error {
	// todo: implement - could use ConsumerSignalChannel and ProducerSignalChannel behind the scene
	return nil
}

func (w WorkflowContext) ContinueAsNew(ctx WorkflowContext, values ...any) error {
	// todo: implement
	return nil
}

func (w WorkflowContext) GetWorkflow(id WorkflowExecutionID) (*WorkflowExecutionInfo, error) {
	return w.tp.getWorkflowExecution(w, id)
}

func (w WorkflowContext) ExecuteWorkflow(handler interface{}, inputs ...any) (*WorkflowExecutionInfo, error) {
	id, err := w.tp.enqueueSubWorkflow(w, handler, inputs...)
	if err != nil {
		return nil, err
	}
	return w.tp.getWorkflowExecution(w, id)
}

func (w WorkflowContext) ExecuteActivity(name HandlerIdentity, inputs ...any) (*ActivityInfo, error) {
	// todo: implement
	return nil, nil
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

func (w ActivityContext) ExecuteWorkflow(handler interface{}, inputs ...any) (*WorkflowExecutionInfo, error) {
	id, err := w.tp.enqueueSubWorkflow(w, handler, inputs...)
	if err != nil {
		return nil, err
	}
	return w.tp.getWorkflowExecution(w, id)
}

func (w ActivityContext) ExecuteSideEffect(name HandlerIdentity, inputs ...any) (*SideEffectInfo, error) {
	// todo: implement
	return nil, nil
}

func (w ActivityContext) ExecuteSaga(name HandlerIdentity, inputs interface{}) (*SagaInfo, error) {
	// todo: implement - should build the saga, then step the instaciated steps, then use reflect to get the methods, assign it to a callback on the TransactionTask or CompensationTask, so what's necessary for the Next/Compensate callbacks, generate a bunch of callbacks based on the rules of the Saga Pattern discuess in the documentation, and just dispatch the first TransactionContext
	return nil, nil
}

func (w ActivityContext) ExecuteActivity(name HandlerIdentity, inputs ...any) (*ActivityExecutionInfo, error) {
	id, err := w.tp.enqueueSubActivityExecution(w, name, inputs...)
	if err != nil {
		return nil, err
	}
	return w.tp.getActivityExecution(w, id)
}

func (w ActivityContext) GetActivity(id ActivityExecutionID) (*ActivityExecutionInfo, error) {
	return w.tp.getActivityExecution(w, id)
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

func (w SideEffectContext) ExecuteActivity(name HandlerIdentity, inputs ...any) (*ActivityInfo, error) {
	// todo: implement
	return nil, nil
}

func (w SideEffectContext) ExecuteSideEffect(name HandlerIdentity, inputs ...any) (*SideEffectInfo, error) {
	// todo: implement
	return nil, nil
}

func (w SideEffectContext) GetActivity(id string) (*ActivityInfo, error) {
	// todo: implement
	return nil, nil
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
