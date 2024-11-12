package engine

import (
	"context"
	"fmt"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/davidroman0O/tempolite/internal/clock"
	"github.com/davidroman0O/tempolite/internal/engine/cq/commands"
	"github.com/davidroman0O/tempolite/internal/engine/cq/queries"
	"github.com/davidroman0O/tempolite/internal/engine/info"
	"github.com/davidroman0O/tempolite/internal/engine/queues"
	"github.com/davidroman0O/tempolite/internal/engine/registry"
	"github.com/davidroman0O/tempolite/internal/engine/schedulers"
	"github.com/davidroman0O/tempolite/internal/persistence/ent"
	"github.com/davidroman0O/tempolite/internal/persistence/repository"
	"github.com/davidroman0O/tempolite/internal/types"
	"github.com/davidroman0O/tempolite/pkg/logs"
	"golang.org/x/sync/errgroup"
)

type Engine struct {
	ctx          context.Context
	cancel       context.CancelFunc
	mu           sync.RWMutex
	registry     *registry.Registry
	workerQueues map[string]*queues.Queue
	db           repository.Repository
	// scheduler    *schedulers.Scheduler
	commands *commands.Commands
	queries  *queries.Queries
	clock    *clock.Clock

	schedulerWorkflowPending clock.TickerID
}

func New(
	ctx context.Context,
	builder registry.RegistryBuildFn,
	client *ent.Client,
) (*Engine, error) {
	var err error
	logs.Debug(ctx, "Creating engine")

	ctx, cancel := context.WithCancel(ctx)

	e := &Engine{
		ctx:          ctx,
		cancel:       cancel,
		workerQueues: make(map[string]*queues.Queue),
		clock: clock.NewClock(
			// ctx,
			time.Millisecond*10,
			func(err error) {
				logs.Error(ctx, err.Error())
			},
		),
		schedulerWorkflowPending: "scheduler_workflow_pending",
	}

	logs.Debug(ctx, "Creating registry")
	if e.registry, err = builder(); err != nil {
		logs.Error(ctx, "Error creating registry", "error", err)
		return nil, err
	}

	logs.Debug(ctx, "Creating repository")
	e.db = repository.NewRepository(
		ctx,
		client)

	if err := e.ReconciliateWorkflows(); err != nil {
		logs.Error(ctx, "Error reconciliating workflows", "error", err)
		return nil, err
	}

	e.clock.Start()

	e.commands = commands.New(e.ctx, e.db, e.registry)
	e.queries = queries.New(e.ctx, e.db, e.registry, e.clock)

	if err := e.AddQueue("default"); err != nil {
		logs.Error(ctx, "Error creating default queue", "error", err)
		return nil, err
	}

	go e.monitoringStatus()

	return e, nil
}

func (e *Engine) GetQueues() []string {
	e.mu.Lock()
	defer e.mu.Unlock()
	var queues []string
	for k := range e.workerQueues {
		queues = append(queues, k)
	}
	return queues
}

func (e *Engine) GetQueue(queue string) *queues.Queue {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.workerQueues[queue]
}

// When we restart, we need to check the state of the workflows
// - Entity Running / Execution Running => Reconciliate: Execution to Cancelled, no inc retry, new Execution Pending
func (e *Engine) ReconciliateWorkflows() error {

	tx, err := e.db.Tx()
	if err != nil {
		logs.Error(e.ctx, "ReconciliateWorkflows error getting tx", "error", err)
		return err
	}

	if err = e.db.Workflows().ReconciliateRunningRunning(tx); err != nil {
		logs.Error(e.ctx, "ReconciliateWorkflows error reconciliating running running", "error", err)
		return err
	}

	if err = tx.Commit(); err != nil {
		logs.Error(e.ctx, "ReconciliateWorkflows error committing tx", "error", err)
		return err
	}

	return nil
}

func (e *Engine) monitoringStatus() {
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
		case <-e.ctx.Done():
			logs.Debug(e.ctx, "Monitoring status context canceled")
			ticker.Stop()
			return
		case <-ticker.C:
			logs.Debug(e.ctx, "Monitoring status")
			e.mu.RLock()
			queues := e.workerQueues
			e.mu.RUnlock()

			megaLog := ""
			for _, q := range queues {
				megaLog += fmt.Sprintf("Current Queue %v\n", q.GetName())

				{
					tasks := q.GetCurrentWorkflowTasks()
					for _, v := range tasks {
						megaLog += fmt.Sprintf("\tWORKFLOW TASK workerID:%v taskStatus:%v workflowID:%v executionID:%v stepID:%v queueID:%v runID:%v  \n", v.WorkerID, v.Status, v.Data.WorkflowInfo.ID, v.Data.WorkflowInfo.Execution.ID, v.Data.WorkflowInfo.StepID, v.Data.WorkflowInfo.QueueID, v.Data.WorkflowInfo.RunID)
					}
					slots := q.AvailableWorkflowWorkers()
					megaLog += fmt.Sprintf("\tWORKFLOW SLOTS %v\n", slots)
				}

				// TODO: add them but the data is empty for now
				// {
				// 	tasks := q.GetCurrentActivityTasks()
				// 	for _, v := range tasks {
				// 		megaLog += fmt.Sprintf("\tACTIVITY TASK workerID:%v taskStatus:%v workflowID:%v stepID:%v queueID:%v runID:%v  \n", v.WorkerID, v.Status, v.Data.WorkflowInfo.ID, v.Data.WorkflowInfo.StepID, v.Data.WorkflowInfo.QueueID, v.Data.WorkflowInfo.RunID)
				// 	}
				// }

				// {
				// 	tasks := q.GetCurrentSideEffectTasks()
				// 	for _, v := range tasks {
				// 		megaLog += fmt.Sprintf("\tSIDE EFFECT TASK workerID:%v taskStatus:%v workflowID:%v stepID:%v queueID:%v runID:%v  \n", v.WorkerID, v.Status, v.Data.WorkflowInfo.ID, v.Data.WorkflowInfo.StepID, v.Data.WorkflowInfo.QueueID, v.Data.WorkflowInfo.RunID)
				// 	}
				// }

				// {
				// 	tasks := q.GetCurrentSagaTasks()
				// 	for _, v := range tasks {
				// 		megaLog += fmt.Sprintf("\tSAGA TASK workerID:%v taskStatus:%v workflowID:%v stepID:%v queueID:%v runID:%v  \n", v.WorkerID, v.Status, v.Data.WorkflowInfo.ID, v.Data.WorkflowInfo.StepID, v.Data.WorkflowInfo.QueueID, v.Data.WorkflowInfo.RunID)
				// 	}
				// }

			}

			e.mu.RLock()
			tickers := e.clock.GetMapTickers()
			e.mu.RUnlock()

			megaLog += fmt.Sprintf("Current Tickers %v\n", len(tickers))

			var tickerIDs []string
			for id := range tickers {
				tickerIDs = append(tickerIDs, fmt.Sprintf("%v", id))
			}
			sort.Strings(tickerIDs)
			for _, id := range tickerIDs {
				t := tickers[id]
				megaLog += fmt.Sprintf("\tTICKER ID: %v - Running: %v\n", id, t)
			}
			fmt.Println(megaLog)

		}
	}
}

func (e *Engine) AddQueue(queue string) error {
	var err error
	var newQueue *queues.Queue
	logs.Debug(e.ctx, "Adding queue", "queue", queue)

	// TODO: add options for initial workers
	if newQueue, err = queues.New(
		e.ctx,
		queue,
		e.registry,
		e.db,
		e.commands,
		e.queries,
	); err != nil {
		logs.Error(e.ctx, "Error creating queue", "queue", queue, "error", err)
		return err
	}

	logs.Debug(e.ctx, "Adding queue workflow pending to clock", "id", newQueue.GetTickerIDWorkflowPending())
	e.clock.AddTicker(
		newQueue.GetTickerIDWorkflowPending(),
		schedulers.NewSchedulerWorkflowPending(
			e.ctx,
			e.db,
			newQueue.GetName(),
			newQueue.AvailableWorkflowWorkers,
			newQueue.SubmitWorkflow,
		),
		clock.WithCleanup(func() {
			fmt.Println("Cleaning up ", newQueue.GetTickerIDWorkflowPending())
			// it will be removed on close
		}))

	e.mu.Lock()
	e.workerQueues[queue] = newQueue
	e.mu.Unlock()
	logs.Debug(e.ctx, "Queue added", "queue", queue)
	return nil
}

func (e *Engine) Shutdown() error {
	shutdown := errgroup.Group{}
	logs.Debug(e.ctx, "Shutting down engine")

	// logs.Debug(e.ctx, "Shutting down scheduler")
	// e.scheduler.Stop()
	// logs.Debug(e.ctx, "Scheduler shutdown complete")

	e.cancel()

	for n, q := range e.workerQueues {
		shutdown.Go(func() error {
			logs.Debug(e.ctx, "Shutting down queue", "queue", n)
			defer logs.Debug(e.ctx, "Queue shutdown complete", "queue", n)
			return q.Shutdown()
		})
	}
	shutdown.Wait()

	logs.Debug(e.ctx, "Shutting down clock")
	e.clock.Stop()
	logs.Debug(e.ctx, "Clock shutdown complete")

	defer logs.Debug(e.ctx, "Engine shutdown complete")

	logs.Debug(e.ctx, "Waiting for shutdown")
	return nil
}

func (e *Engine) Scale(queue string, targets map[string]int) error {
	logs.Debug(e.ctx, "Preparing to scale", "queue", queue, "targets", targets)
	e.mu.Lock()
	q, ok := e.workerQueues[queue]
	e.mu.Unlock()
	if !ok {
		if err := e.AddQueue(queue); err != nil {
			return err
		}
		e.mu.Lock()
		q = e.workerQueues[queue]
		e.mu.Unlock()
	}
	logs.Debug(e.ctx, "Scaling queue", "queue", queue, "targets", targets)
	return q.Scale(targets)
}

// Create a new run workflow
//
// # Pseudo code
//
// ```
// => Workflow
// -     - create workflow entity
// -     - create workflow execution
// -
// -     if fail
// -     if max attempt < attemps => retry
// -         - create workflow execution
// -
// - 		if fail
// - 		if max attempt < attemps => retry
// - 			- create workflow execution
// - 		else
// - 			- workflow failed
// - else
// - 	- workflow failed
// ```
func (e *Engine) Workflow(workflowFunc interface{}, options types.WorkflowOptions, params ...any) *info.WorkflowInfo {
	var id types.WorkflowID
	var err error
	logs.Debug(e.ctx, "Creating workflow")
	if id, err = e.commands.CommandWorkflow(workflowFunc, options, params...); err != nil {
		logs.Error(e.ctx, "Error creating workflow", "error", err)
		return e.queries.QueryNoWorkflow(err)
	}
	logs.Debug(e.ctx, "Workflow created", "id", id)
	return e.queries.QueryWorfklow(workflowFunc, id)
}

// IF you cancelled a Tempolite instance, you need to a new WorkflowInfo instance
func (e *Engine) GetWorkflow(id types.WorkflowID) *info.WorkflowInfo {
	logs.Debug(e.ctx, "Engine Getting workflow", "id", id)
	return e.queries.GetWorkflowInfo(id)
}
