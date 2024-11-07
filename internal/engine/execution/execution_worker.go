package execution

import (
	"context"
	"math/rand"
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

func (e *WorkerPool[Request, Response]) Submit(task *retrypool.RequestResponse[Request, Response]) error {
	return e.pool.Submit(task)
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
	id := e.pool.AddWorker(e.builder(e.ctx, e.queue))
	e.workers = append(e.workers, id)
	return id
}

func (e *WorkerPool[Request, Response]) RemoveWorker() error {
	// get random id in the e.workers
	id := e.workers[rand.Intn(len(e.workers))]
	return e.pool.RemoveWorker(id)
}
