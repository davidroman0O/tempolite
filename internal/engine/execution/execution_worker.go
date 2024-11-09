package execution

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/davidroman0O/retrypool"
)

// TODO: I still don't have a deadtask mechanism

type WorkerBuilder[Request, Response any] func(ctx context.Context, queue string) retrypool.Worker[*retrypool.RequestResponse[Request, Response]]

type Executor[Request, Response any] interface {
	OnRetry(attempt int, err error, task *retrypool.TaskWrapper[*retrypool.RequestResponse[Request, Response]])
	OnSuccess(controller retrypool.WorkerController[*retrypool.RequestResponse[Request, Response]], workerID int, worker retrypool.Worker[*retrypool.RequestResponse[Request, Response]], task *retrypool.TaskWrapper[*retrypool.RequestResponse[Request, Response]])
	OnFailure(controller retrypool.WorkerController[*retrypool.RequestResponse[Request, Response]], workerID int, worker retrypool.Worker[*retrypool.RequestResponse[Request, Response]], task *retrypool.TaskWrapper[*retrypool.RequestResponse[Request, Response]], err error) retrypool.DeadTaskAction
	OnDeadTask(task *retrypool.DeadTask[*retrypool.RequestResponse[Request, Response]])
	OnPanic(task *retrypool.RequestResponse[Request, Response], v interface{}, stackTrace string)
	OnExecutorPanic(worker int, recovery any, err error, stackTrace string)
}

type WorkerPool[Request, Response any] struct {
	ctx     context.Context
	pool    *retrypool.Pool[*retrypool.RequestResponse[Request, Response]]
	builder WorkerBuilder[Request, Response]
	queue   string
	workers []int
	mu      sync.Mutex
}

func NewWorkerPool[Request, Response any](
	ctx context.Context,
	queue string, // TODO: i don't think a queue will be just a string
	executor Executor[Request, Response],
	builder WorkerBuilder[Request, Response],
	// TODO: add options, we have rate limiter and all that stuff
) *WorkerPool[Request, Response] {
	e := &WorkerPool[Request, Response]{
		ctx:     ctx,
		builder: builder,
		queue:   queue,
		workers: []int{},
	}

	opts := []retrypool.Option[*retrypool.RequestResponse[Request, Response]]{}

	opts = append(opts, retrypool.WithAttempts[*retrypool.RequestResponse[Request, Response]](1)) // we will manage the retry ourselves
	opts = append(opts, retrypool.WithOnRetry(executor.OnRetry))
	opts = append(opts, retrypool.WithOnTaskSuccess(executor.OnSuccess))
	opts = append(opts, retrypool.WithOnTaskFailure(executor.OnFailure))
	opts = append(opts, retrypool.WithOnNewDeadTask(executor.OnDeadTask))
	opts = append(opts, retrypool.WithPanicWorker[*retrypool.RequestResponse[Request, Response]](executor.OnExecutorPanic))
	opts = append(opts, retrypool.WithPanicHandler[*retrypool.RequestResponse[Request, Response]](executor.OnPanic))
	opts = append(opts, retrypool.WithRoundRobinAssignment[*retrypool.RequestResponse[Request, Response]]())

	e.pool = retrypool.New[*retrypool.RequestResponse[Request, Response]](ctx, []retrypool.Worker[*retrypool.RequestResponse[Request, Response]]{}, opts...)

	return e
}

func (e *WorkerPool[Request, Response]) Submit(task *retrypool.RequestResponse[Request, Response], options ...retrypool.TaskOption[*retrypool.RequestResponse[Request, Response]]) error {
	return e.pool.Submit(task, options...)
}

func (e *WorkerPool[Request, Response]) Shutdown() error {
	return e.pool.Shutdown()
}

func (e *WorkerPool[Request, Response]) Wait() error {
	return e.pool.WaitWithCallback(e.ctx, func(queueSize, processingCount, deadTaskCount int) bool {
		return queueSize > 0 || processingCount > 0 || deadTaskCount > 0
	}, time.Second)
}

func (e *WorkerPool[Request, Response]) AddWorker() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	id := e.pool.AddWorker(e.builder(e.ctx, e.queue))
	e.workers = append(e.workers, id)
	return id
}

func (e *WorkerPool[Request, Response]) RemoveWorker() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(e.workers) == 0 {
		return fmt.Errorf("no workers to remove")
	}

	index := rand.Intn(len(e.workers))
	id := e.workers[index]

	// Remove the worker ID from the slice
	e.workers = append(e.workers[:index], e.workers[index+1:]...)

	err := e.pool.RemoveWorker(id)
	if err != nil {
		// If removal failed, add the worker back to the list
		e.workers = append(e.workers, id)
		return err
	}

	return nil
}

func (e *WorkerPool[Request, Response]) AvailableWorkers() int {
	e.mu.Lock()
	defer e.mu.Unlock()
	return len(e.workers)
}
