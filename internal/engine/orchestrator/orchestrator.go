package orchestrator

import (
	"context"
	"sync"

	"github.com/davidroman0O/tempolite/internal/engine/registry"
	"github.com/davidroman0O/tempolite/internal/persistence/repository"
	"github.com/davidroman0O/tempolite/internal/types"
	"github.com/qmuntal/stateless"
)

type WorkflowState struct {
}

type Orchestrator struct {
	ctx context.Context

	db       repository.Repository
	registry *registry.Registry

	fsm *stateless.StateMachine

	state   *WorkflowState
	muState sync.RWMutex

	done chan struct{}
}

type State string

const (
	StatePrepare  State = "Prepare"
	StateWorkflow State = "Workflow"
	StateChildren State = "Children"
	StateActivity State = "Activity"
)

type SubStateWorkflow string

const (
	SubStateWorkflowPrepareExecution       SubStateWorkflow = "PrepareExecution"
	SubStateWorkflowCallExecution          SubStateWorkflow = "CallExecution"
	SubStateWorkflowWaitChildrenOrNewState SubStateWorkflow = "WaitChildrenOrNewState"
	SubStateWorkflowChildren               SubStateWorkflow = "Children"
	SubStateWorkflowCompleted              SubStateWorkflow = "Completed"
	SubStateWorkflowFailed                 SubStateWorkflow = "Failed"
	SubStateWorkflowRetryable              SubStateWorkflow = "Retryable"
)

func New(
	ctx context.Context,
	db repository.Repository,
	registry *registry.Registry,
) (*Orchestrator, error) {
	o := &Orchestrator{
		ctx:      ctx,
		db:       db,
		registry: registry,
		done:     make(chan struct{}),
	}

	o.fsm = stateless.NewStateMachineWithExternalStorage(o.getState, o.setState, stateless.FiringQueued)

	o.fsm.Configure(StatePrepare)

	o.fsm.CanFireCtx(ctx, StatePrepare)

	return o, nil
}

func (o *Orchestrator) getState(ctx context.Context) (any, error) {
	o.muState.RLock()
	defer o.muState.RUnlock()
	return o.state, nil
}

func (o *Orchestrator) setState(ctx context.Context, data any) error {
	o.muState.Lock()
	defer o.muState.Unlock()
	o.state = data.(*WorkflowState)
	return nil
}

func (o *Orchestrator) Wait() {
	<-o.done
}

func (o *Orchestrator) Workflow(workflowFunc interface{}, opts types.WorkflowOptions, params ...any) (types.WorkflowID, error) {
	// return o.commands.CommandWorkflow(workflowFunc, opts, params...)
	return types.NoWorkflowID, nil
}
