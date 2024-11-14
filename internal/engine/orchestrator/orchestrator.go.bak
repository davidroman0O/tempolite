package orchestrator

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/davidroman0O/retrypool"
	"github.com/davidroman0O/tempolite/internal/dag"
	"github.com/davidroman0O/tempolite/pkg/logs"

	tempoliteContext "github.com/davidroman0O/tempolite/internal/engine/context"
	"github.com/davidroman0O/tempolite/internal/engine/cq/commands"
	"github.com/davidroman0O/tempolite/internal/engine/cq/queries"
	"github.com/davidroman0O/tempolite/internal/engine/io"
	"github.com/davidroman0O/tempolite/internal/engine/registry"
	"github.com/davidroman0O/tempolite/internal/persistence/ent"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/entity"
	"github.com/davidroman0O/tempolite/internal/persistence/repository"
	"github.com/davidroman0O/tempolite/internal/types"
)

/// Orchestrator manage one workflow using retrypool
/// - pull a new workflow in charge
/// - One workflow at a time
/// - One DAG maintained by hooks and fetches
/// - One retrypool with [ComputeRequest, ComputeResponse] that is used for all components
/// - Scale up and down workers executed a root Workflow and encounter a sub-workflow which create one more worker
/// - Virtualized queues due to dynamic scaling
/// - use ent hooks on updates to feedback the orchestrator

// We can't store the input
type VertexEntity struct {
	RunID    int
	StepID   string
	EntityID int
	Queue    string
	Type     entity.Type
}

func (v VertexEntity) Name() string {
	return fmt.Sprintf("Entity_%s_%d", v.Type.String(), v.EntityID)
}

// We can't store the ouput
type VertexExecution struct {
	ExecutionID int
	Type        entity.Type
	Task        *retrypool.RequestResponse[ComputeRequest, ComputeResponse]
}

func (v VertexExecution) Name() string {
	return fmt.Sprintf("Execution_%s_%d", v.Type.String(), v.ExecutionID)
}

type State string

const (
	Searching      State = "searching"
	BeforeRun      State = "searching_to_run"
	Running        State = "running"
	SwitchBranch   State = "switch_branch"
	NextBranch     State = "next_branch"
	Wait           State = "wait"
	RunningToErr   State = "running_to_err"
	Completed      State = "completed"
	CompletedToEnd State = "completed_to_end"
	End            State = "end"
)

type ComputeRequest struct {
	entityType entity.Type
	workflow   *repository.WorkflowInfo // optional
	context    types.SharedContext
	inputs     []interface{}
}

type ComputeResponse struct {
	ouputs []interface{}
}

// We assume on start the db has been reconciled with a workable state which mean that e.g. Running/Running cases managed to Pending/Cancelled/Pending
type Orchestrator struct {
	ctx                context.Context
	db                 repository.Repository
	currentWorkflowID  int
	currentExecutionID int

	commands *commands.Commands
	queries  *queries.Queries

	registry *registry.Registry
	state    State
	pool     *retrypool.Pool[*retrypool.RequestResponse[ComputeRequest, ComputeResponse]]
	wg       sync.WaitGroup

	graph                   dag.AcyclicGraph
	workflow                *repository.WorkflowInfo
	workflowEntityVertex    dag.Vertex
	workflowExecutionVertex dag.Vertex

	currentEntityType      entity.Type
	currentEntityVertex    dag.Vertex
	currentExecutionVertex dag.Vertex

	mu sync.RWMutex
}

type ComputerWorker struct{}

func (w ComputerWorker) Run(ctx context.Context, data *retrypool.RequestResponse[ComputeRequest, ComputeResponse]) error {
	fmt.Println("Running worker")

	switch data.Request.entityType {
	case entity.TypeWorkflow:
		values := []reflect.Value{reflect.ValueOf(data.Request.context)}

		for _, v := range data.Request.inputs {
			values = append(values, reflect.ValueOf(v))
		}
		returnedValues := reflect.ValueOf(data.Request.context.Handler().Handler).Call(values)

		var res []interface{}
		var errRes error
		if len(returnedValues) > 0 {
			res = make([]interface{}, len(returnedValues)-1)
			for i := 0; i < len(returnedValues)-1; i++ {
				res[i] = returnedValues[i].Interface()
			}
			if !returnedValues[len(returnedValues)-1].IsNil() {
				errRes = returnedValues[len(returnedValues)-1].Interface().(error)
			}
		}

		if errRes != nil {
			data.CompleteWithError(errRes)
		} else {
			data.Complete(ComputeResponse{
				ouputs: res,
			})
		}

	}

	fmt.Println("\t Worker done with ", data.Request.workflow.ID)

	return nil
}

func New(
	ctx context.Context,
	db repository.Repository,
	registry *registry.Registry,
) *Orchestrator {
	o := &Orchestrator{
		ctx:      ctx,
		db:       db,
		registry: registry,
		state:    Searching,
	}

	o.commands = commands.New(o.ctx, o.db, o.registry)
	o.queries = queries.New(o.ctx, o.db, o.registry)

	o.pool = retrypool.New(ctx, []retrypool.Worker[*retrypool.RequestResponse[ComputeRequest, ComputeResponse]]{})

	go o.monitoringStatus()

	return o
}

func (o *Orchestrator) monitoringStatus() {
	ticker := time.NewTicker(time.Second / 2)

	defer func() {
		if r := recover(); r != nil {
			buf := make([]byte, 4096)
			n := runtime.Stack(buf, false)
			stackTrace := string(buf[:n])

			// Create a concise error message
			panic(fmt.Errorf("Panic in triggerTickers: %v \n %v", r, stackTrace))
		}
	}()

	for {
		select {
		case <-o.ctx.Done():
			logs.Debug(o.ctx, "Monitoring status context canceled")
			ticker.Stop()
			return
		case <-ticker.C:

			tasks := []CurrentTask[ComputeRequest, ComputeResponse]{}
			o.pool.RangeTasks(func(data *retrypool.TaskWrapper[*retrypool.RequestResponse[ComputeRequest, ComputeResponse]], workerID int, status retrypool.TaskStatus) bool {
				tasks = append(tasks, CurrentTask[ComputeRequest, ComputeResponse]{
					WorkerID: workerID,
					Status:   status,
					Data:     data.Data().Request,
				})
				return true
			})

			logTasks := ""
			for _, v := range tasks {
				logTasks += fmt.Sprintf("\n\t\tWorkerID: %d, Status: %v, Data: %v", v.WorkerID, v.Status, v.Data)
			}

			fmt.Println("Worker logs \n", "\tAvailable Workers ", o.pool.AvailableWorkers(), "\n\tTotal Workers ", o.pool.GetWorkerCount(), "\n\tTasks ", logTasks)
		}
	}
}

func (o *Orchestrator) Wait() {
	o.mu.RLock()
	o.wg.Wait()
	o.mu.RUnlock()
}

func (o *Orchestrator) Stop() {
	o.mu.RLock()
	o.mu.RUnlock()
	fmt.Println("Stopping orchestrator")
	o.wg.Wait()
	o.pool.Shutdown()
}

// Temporary function
func (o *Orchestrator) Workflow(workflowFunc interface{}, opts types.WorkflowOptions, params ...any) (types.WorkflowID, error) {
	return o.commands.CommandWorkflow(workflowFunc, opts, params...)
}

func (o *Orchestrator) ChangeState(state State) {
	o.state = state
}

type CurrentTask[Request any, Response any] struct {
	WorkerID int
	Status   retrypool.TaskStatus
	Data     Request
}

func (o *Orchestrator) Tick(ctx context.Context) error {
	var tx *ent.Tx
	var err error

	o.mu.Lock()
	defer o.mu.Unlock()

	switch o.state {

	// Initial state searching for a new workflow to run
	case Searching:
		if tx, err = o.db.Tx(); err != nil {
			return err
		}

		if o.workflow, err = o.db.Workflows().PullPending(tx); err != nil {
			tx.Rollback()
			return nil
		}

		if err = tx.Commit(); err != nil {
			return err
		}

		o.wg.Add(1) // block for compute

		o.graph = dag.AcyclicGraph{}

		o.currentWorkflowID = o.workflow.ID
		o.currentExecutionID = o.workflow.Execution.ID

		// prepare the initial vertices
		o.workflowEntityVertex = o.graph.Add(VertexEntity{
			EntityID: o.workflow.ID,
			Type:     entity.TypeWorkflow,
			RunID:    o.workflow.RunID,
			StepID:   o.workflow.StepID,
			Queue:    o.workflow.QueueName,
			// Inputs:   o.workflow.Data.Input,
		})

		o.workflowExecutionVertex = o.graph.Add(VertexExecution{
			ExecutionID: o.currentExecutionID,
			Type:        entity.TypeWorkflow,
		})

		o.graph.Connect(dag.BasicEdge(o.workflowEntityVertex, o.workflowExecutionVertex))

		o.currentEntityType = entity.TypeWorkflow
		o.currentEntityVertex = o.workflowEntityVertex
		o.currentExecutionVertex = o.workflowExecutionVertex

		fmt.Println("\t Searching -> SearchingToRun")

		fmt.Println(
			string(
				o.graph.Dot(&dag.DotOpts{
					DrawCycles: true,
					Verbose:    true,
				}),
			),
		)

		o.ChangeState(BeforeRun)

	case BeforeRun:
		o.pool.RedistributeTasks()

		if o.pool.AvailableWorkers() == 0 {
			fmt.Println("\t scale up")
			o.pool.AddWorker(ComputerWorker{})
		}

		o.ChangeState(Running)

	case Running:

		entityVertex := o.currentEntityVertex.(VertexEntity)
		executionVertex := o.currentExecutionVertex.(VertexExecution)

		var handlerInfo types.HandlerInfo
		switch o.currentEntityType {
		case entity.TypeWorkflow:
			workflow, err := o.registry.GetWorkflow(types.HandlerIdentity(o.workflow.HandlerName))
			if err != nil {
				o.ChangeState(RunningToErr)
				return err
			}
			handlerInfo = types.HandlerInfo(workflow)
		}

		params, err := io.ConvertInputsFromSerialization(handlerInfo, o.workflow.Data.Input)
		if err != nil {
			return err
		}

		fmt.Println("create task")
		task := retrypool.
			NewRequestResponse[ComputeRequest, ComputeResponse](ComputeRequest{
			inputs:     params,
			workflow:   o.workflow,
			entityType: o.currentEntityType,
			context: tempoliteContext.NewWorkflowContext(
				o.ctx,
				entityVertex.EntityID,
				executionVertex.ExecutionID,
				entityVertex.RunID,
				entityVertex.Queue,
				entityVertex.StepID,
				handlerInfo,
				struct {
					*commands.Commands
					*queries.Queries
				}{
					o.commands,
					o.queries,
				},
			),
		})

		executionVertex.Task = task // assign the task
		o.graph.Replace(o.currentExecutionVertex, executionVertex)
		o.currentExecutionVertex = executionVertex

		beingProcess := retrypool.NewProcessedNotification()
		beingQueued := retrypool.NewQueuedNotification()

		if err := o.pool.Submit(
			task,
			retrypool.WithBeingProcessed[*retrypool.RequestResponse[ComputeRequest, ComputeResponse]](beingProcess),
			retrypool.WithQueued[*retrypool.RequestResponse[ComputeRequest, ComputeResponse]](beingQueued),
		); err != nil {
			o.ChangeState(RunningToErr)
			return err
		}

		o.pool.RedistributeTasks()

		fmt.Println("waiting for beingQueued", task)
		<-beingQueued

		if tx, err = o.db.Tx(); err != nil {
			o.ChangeState(RunningToErr)
			return err
		}
		if err = o.db.Workflows().UpdateExecutionRunning(tx, o.currentExecutionID); err != nil {
			tx.Rollback()
			o.ChangeState(RunningToErr)
			return err
		}
		if err = tx.Commit(); err != nil {
			o.ChangeState(RunningToErr)
			return err
		}

		fmt.Println("waiting for beingProcess", task)
		<-beingProcess

		if tx, err = o.db.Tx(); err != nil {
			o.ChangeState(RunningToErr)
			return err
		}
		if err = o.db.Workflows().UpdateExecutionRunning(tx, o.currentExecutionID); err != nil {
			tx.Rollback()
			o.ChangeState(RunningToErr)
			return err
		}
		if err = tx.Commit(); err != nil {
			o.ChangeState(RunningToErr)
			return err
		}

		o.ChangeState(SwitchBranch)

	case SwitchBranch:

		entityVertex := o.currentEntityVertex.(VertexEntity)
		executionVertex := o.currentExecutionVertex.(VertexExecution)

		switch entityVertex.Type {
		// for workflow, we cannot aford to wait
		case entity.TypeWorkflow:
			go func(repo repository.Repository, executionID int, task *retrypool.RequestResponse[ComputeRequest, ComputeResponse]) {
				response, err := task.Wait(o.ctx)
				if err != nil {
					if tx, err = repo.Tx(); err != nil {
						o.ChangeState(RunningToErr)
						return
					}
					if err = repo.Workflows().UpdateExecutionFailed(tx, executionID); err != nil {
						tx.Rollback()
						o.ChangeState(RunningToErr)
						return
					}
					if err = tx.Commit(); err != nil {
						o.ChangeState(RunningToErr)
						return
					}
					o.ChangeState(RunningToErr)
					return
				}

				if tx, err = repo.Tx(); err != nil {
					o.ChangeState(RunningToErr)
					return
				}
				if err = repo.Workflows().UpdateExecutionSuccess(tx, executionID); err != nil {
					tx.Rollback()
					o.ChangeState(RunningToErr)
					return
				}

				fmt.Println("\t Response", response.ouputs)
				ouputs, err := io.ConvertOutputsForSerialization(response.ouputs)
				if err != nil {
					tx.Rollback()
					o.ChangeState(RunningToErr)
					return
				}

				if err = repo.Workflows().UpdateExecutionDataOuput(tx, executionID, ouputs); err != nil {
					tx.Rollback()
					o.ChangeState(RunningToErr)
					return
				}

				if err = tx.Commit(); err != nil {
					o.ChangeState(RunningToErr)
					return
				}
			}(o.db, executionVertex.ExecutionID, executionVertex.Task)

			for {
				if tx, err = o.db.Tx(); err != nil {
					o.ChangeState(RunningToErr)
					return err
				}

				if o.workflow, err = o.db.Workflows().Get(tx, entityVertex.EntityID); err != nil {
					tx.Rollback()
					o.ChangeState(RunningToErr)
					return nil
				}

				if err = tx.Commit(); err != nil {
					o.ChangeState(RunningToErr)
					return err
				}

				if o.workflow.Execution.Status == repository.Status(entity.StatusCompleted) {
					o.ChangeState(Completed)
					break
				}

				if o.workflow.Execution.Status == repository.Status(entity.StatusFailed) {
					o.ChangeState(RunningToErr)
					break
				}

				var children []*repository.HierarchyInfo

				if tx, err = o.db.Tx(); err != nil {
					o.ChangeState(RunningToErr)
					return err
				}

				if children, err = o.db.Hierarchies().GetChildren(tx, entityVertex.EntityID); err != nil {
					tx.Rollback()
					o.ChangeState(RunningToErr)
					return nil
				}

				if err = tx.Commit(); err != nil {
					o.ChangeState(RunningToErr)
					return err
				}

				if len(children) > 0 {
					fmt.Println("Workflow got children ", children)

					if tx, err = o.db.Tx(); err != nil {
						o.ChangeState(RunningToErr)
						return err
					}

					for _, row := range children {
						switch entity.Type(row.ChildType) {
						case entity.TypeWorkflow:
							var childWorkflow *repository.WorkflowInfo
							if childWorkflow, err = o.db.Workflows().Get(tx, row.ChildEntityID); err != nil {
								tx.Rollback()
								o.ChangeState(RunningToErr)
								return nil
							}
							entityVertex := o.graph.Add(VertexEntity{
								RunID:    row.RunID,
								EntityID: row.ChildEntityID,
								StepID:   row.ChildStepID,
								Queue:    childWorkflow.QueueName,
								Type:     entity.TypeWorkflow,
							})
							executionVertex := o.graph.Add(VertexExecution{
								ExecutionID: row.ChildExecutionID,
								Type:        entity.TypeWorkflow,
							})
							o.graph.Connect(dag.BasicEdge(entityVertex, executionVertex))
							// since the execution created it...
							o.graph.Connect(dag.BasicEdge(o.currentExecutionVertex, entityVertex))
						}
					}

					if err = tx.Commit(); err != nil {
						o.ChangeState(RunningToErr)
						return err
					}

					fmt.Println(
						string(
							o.graph.Dot(&dag.DotOpts{
								DrawCycles: true,
								Verbose:    true,
							}),
						),
					)

					o.ChangeState(NextBranch)
					break
				}
			}
		}

	case NextBranch:

		// that' mean that our current vertex execution have a sub-node

		switch o.currentEntityType {
		case entity.TypeWorkflow:
			edges := o.graph.EdgesFrom(o.currentExecutionVertex)
			nextEntity := edges[0].Target().(VertexEntity)
			fmt.Println("\t next entity", nextEntity)
			o.currentEntityVertex = nextEntity
			edges = o.graph.EdgesFrom(nextEntity)
			nextExecution := edges[0].Target().(VertexExecution)
			fmt.Println("\t next execution")
			o.currentExecutionVertex = nextExecution
			o.currentEntityType = nextEntity.Type
		}

		o.ChangeState(BeforeRun)

	case Wait:

		value := o.currentEntityVertex.(VertexEntity)

		if tx, err = o.db.Tx(); err != nil {
			o.ChangeState(RunningToErr)
			return err
		}

		children := []*repository.HierarchyInfo{}
		if children, err = o.db.Hierarchies().GetChildren(tx, value.EntityID); err != nil {
			tx.Rollback()
			o.ChangeState(RunningToErr)
			return nil
		}

		if err = tx.Commit(); err != nil {
			o.ChangeState(RunningToErr)
			return err
		}

		fmt.Println("Children", children)
		o.ChangeState(Completed)

	case Completed:
		// meaning it is really completed and finished, we need to wrap it up
		o.currentExecutionID = 0
		o.currentWorkflowID = 0
		o.wg.Done()
		o.ChangeState(CompletedToEnd)

		var handlerInfo types.HandlerInfo
		switch o.currentEntityType {
		case entity.TypeWorkflow:
			workflow, err := o.registry.GetWorkflow(types.HandlerIdentity(o.workflow.HandlerName))
			if err != nil {
				o.ChangeState(RunningToErr)
				return err
			}
			handlerInfo = types.HandlerInfo(workflow)
		}

		if tx, err = o.db.Tx(); err != nil {
			o.ChangeState(RunningToErr)
			return err
		}

		if o.workflow, err = o.db.Workflows().Get(tx, o.workflow.ID); err != nil {
			tx.Rollback()
			o.ChangeState(RunningToErr)
			return nil
		}

		if err = tx.Commit(); err != nil {
			o.ChangeState(RunningToErr)
			return err
		}

		fmt.Println("workflow", o.workflow.ID)
		fmt.Println("outputs", o.workflow.Execution.Output)

		outputs, err := io.ConvertOutputsFromSerialization(handlerInfo, o.workflow.Execution.Output)
		if err != nil {
			return err
		}

		fmt.Println("outputs", outputs)

	case End:
		// noooooothing

	}

	return nil
}
