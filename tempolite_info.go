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
)

type WorkflowInfo struct {
	tp    *Tempolite
	ID    string
	RunID string
}

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
			workflowEntity, err := i.tp.client.Workflow.Query().Where(workflow.IDEQ(i.ID)).Only(i.tp.ctx)
			if err != nil {
				return nil, err
			}
			if value, ok = i.tp.workflows.Load(HandlerIdentity(workflowEntity.Identity)); ok {
				if workflowHandlerInfo, ok = value.(Workflow); !ok {
					log.Printf("scheduler: workflow %s is not handler info", i.ID)
					return nil, errors.New("workflow is not handler info")
				}

				switch workflowEntity.Status {
				// wait for the confirmation that the workflow entity reached a final state
				case workflow.StatusCompleted, workflow.StatusFailed, workflow.StatusCancelled:
					// Then only get the latest workflow execution
					// Simply because eventually my children can have retries and i need to let them finish
					latestExec, err := i.tp.client.WorkflowExecution.Query().
						Where(workflowexecution.HasWorkflowWith(workflow.IDEQ(i.ID))).
						Order(ent.Desc(workflowexecution.FieldStartedAt)).
						First(i.tp.ctx)
					if err != nil {
						return nil, err
					}
					switch latestExec.Status {
					case workflowexecution.StatusCompleted:
						return i.tp.convertBackResults(HandlerInfo(workflowHandlerInfo), latestExec.Output)
					case workflowexecution.StatusCancelled:
						return nil, fmt.Errorf("workflow %s was cancelled", i.ID)
					case workflowexecution.StatusFailed:
						return nil, errors.New(latestExec.Error)
					case workflowexecution.StatusPending, workflowexecution.StatusRunning, workflowexecution.StatusRetried:
						// The workflow is still in progress
						return nil, errors.New("workflow is still in progress")
					}
				default:
					continue
				}
			}
			runtime.Gosched()
		}
	}
}

func (i *WorkflowInfo) Cancel() error {
	// todo: implement
	return nil
}

func (i *WorkflowInfo) Pause() error {
	// todo: implement
	return nil
}

func (i *WorkflowInfo) Resume() error {
	// todo: implement
	return nil
}

type ActivityInfo struct {
	tp    *Tempolite
	ID    string
	RunID string
}

func (i *ActivityInfo) Get() ([]interface{}, error) {
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
			activityEntity, err := i.tp.client.Activity.Query().Where(activity.IDEQ(i.ID)).Only(i.tp.ctx)
			if err != nil {
				return nil, err
			}
			if value, ok = i.tp.activities.Load(HandlerIdentity(activityEntity.Identity)); ok {
				if activityHandlerInfo, ok = value.(Activity); !ok {
					log.Printf("scheduler: activity %s is not handler info", i.ID)
					return nil, errors.New("activity is not handler info")
				}

				switch activityEntity.Status {
				// wait for the confirmation that the workflow entity reached a final state
				case activity.StatusCompleted, activity.StatusFailed, activity.StatusCancelled:
					// Then only get the latest activity execution
					// Simply because eventually my children can have retries and i need to let them finish
					latestExec, err := i.tp.client.ActivityExecution.Query().
						Where(activityexecution.HasActivityWith(activity.IDEQ(i.ID))).
						Order(ent.Desc(activityexecution.FieldStartedAt)).
						First(i.tp.ctx)
					if err != nil {
						return nil, err
					}
					switch latestExec.Status {
					case activityexecution.StatusCompleted:
						return i.tp.convertBackResults(HandlerInfo(activityHandlerInfo), latestExec.Output)
					case activityexecution.StatusFailed:
						return nil, errors.New(latestExec.Error)
					case activityexecution.StatusPending, activityexecution.StatusRunning, activityexecution.StatusRetried:
						// The workflow is still in progress
						return nil, errors.New("workflow is still in progress")
					}
				default:
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

func (w WorkflowContext) GetWorkflow(id string) (*WorkflowInfo, error) {
	return w.tp.GetWorkflow(id)
}

func (w WorkflowContext) ExecuteWorkflow(handler interface{}, inputs ...any) (*WorkflowInfo, error) {
	id, err := w.tp.enqueueSubWorkflow(w, handler, inputs...)
	if err != nil {
		return nil, err
	}
	return w.tp.GetWorkflow(id)
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

func (w ActivityContext) ExecuteSideEffect(name HandlerIdentity, inputs ...any) (*SideEffectInfo, error) {
	// todo: implement
	return nil, nil
}

func (w ActivityContext) ExecuteSaga(name HandlerIdentity, inputs interface{}) (*SagaInfo, error) {
	// todo: implement - should build the saga, then step the instaciated steps, then use reflect to get the methods, assign it to a callback on the TransactionTask or CompensationTask, so what's necessary for the Next/Compensate callbacks, generate a bunch of callbacks based on the rules of the Saga Pattern discuess in the documentation, and just dispatch the first TransactionContext
	return nil, nil
}

func (w ActivityContext) ExecuteActivity(name HandlerIdentity, inputs ...any) (*ActivityInfo, error) {
	id, err := w.tp.enqueueSubActivty(w, name, inputs...)
	if err != nil {
		return nil, err
	}
	return w.tp.GetActivity(id)
}

func (w ActivityContext) GetActivity(id string) (*ActivityInfo, error) {
	return w.tp.GetActivity(id)
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
