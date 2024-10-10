package tempolite

import (
	"context"
	"errors"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/davidroman0O/go-tempolite/ent/workflowexecution"
)

type WorkflowInfo struct {
	tp          *Tempolite
	ID          string
	RunID       string
	ExecutionID string
}

func (i *WorkflowInfo) Get(ctx TempoliteContext) ([]interface{}, error) {
	// TODO: make a utilities
	ticker := time.NewTicker(time.Second / 16)
	defer ticker.Stop()
	var value any
	var ok bool
	var workflowHandlerInfo Workflow
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			exec, err := i.tp.client.WorkflowExecution.Query().Where(workflowexecution.IDEQ(i.ExecutionID)).WithWorkflow().Only(ctx)
			if err != nil {
				return nil, err
			}
			if value, ok = i.tp.workflows.Load(HandlerIdentity(exec.Edges.Workflow.Identity)); ok {
				if workflowHandlerInfo, ok = value.(Workflow); !ok {
					log.Printf("scheduler: workflow %s is not handler info", i.ExecutionID)
					return nil, errors.New("workflow is not handler info")
				}
				switch exec.Status {
				case workflowexecution.StatusCompleted:
					return i.tp.convertBackResults(HandlerInfo(workflowHandlerInfo), exec.Output)

				case workflowexecution.StatusCancelled:
					return nil, fmt.Errorf("workflow %s was cancelled", i.ID)
				case workflowexecution.StatusFailed:
					return nil, errors.New(exec.Error)
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

type ActivityInfo struct{}

func (i *ActivityInfo) Get(ctx TempoliteContext) ([]interface{}, error) {
	// todo: implement
	return nil, nil
}

type SideEffectInfo struct{}

func (i *SideEffectInfo) Get(ctx TempoliteContext) ([]interface{}, error) {
	// todo: implement
	return nil, nil
}

type SagaInfo struct{}

func (i *SagaInfo) Get(ctx TempoliteContext) ([]interface{}, error) {
	// todo: implement
	return nil, nil
}

type TempoliteContext interface {
	context.Context
}

type WorkflowContext struct {
	TempoliteContext
	tp          *Tempolite
	workflowID  string
	executionID string
	runID       string
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

func (w WorkflowContext) ExecuteWorkflow(name HandlerIdentity, inputs ...any) (*WorkflowInfo, error) {
	// todo: implement
	return nil, nil
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

func (w ActivityContext) ExecuteSideEffect(name HandlerIdentity, inputs ...any) (*SideEffectInfo, error) {
	// todo: implement
	return nil, nil
}

func (w ActivityContext) ExecuteSaga(name HandlerIdentity, inputs interface{}) (*SagaInfo, error) {
	// todo: implement - should build the saga, then step the instaciated steps, then use reflect to get the methods, assign it to a callback on the TransactionTask or CompensationTask, so what's necessary for the Next/Compensate callbacks, generate a bunch of callbacks based on the rules of the Saga Pattern discuess in the documentation, and just dispatch the first TransactionContext
	return nil, nil
}

func (w ActivityContext) ExecuteActivity(name HandlerIdentity, inputs ...any) (*ActivityInfo, error) {
	// todo: implement
	return nil, nil
}

func (w ActivityContext) GetActivity(id string) (*ActivityInfo, error) {
	// todo: implement
	return nil, nil
}

type SideEffectContext struct {
	TempoliteContext
	tp *Tempolite
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
