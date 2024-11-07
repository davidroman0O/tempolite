package execution

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/davidroman0O/retrypool"
)

type testRequest struct{}
type testResponse struct {
	msg      string
	workedBy int
}

type executorTests struct {
	ctx   context.Context
	queue string
}

func newTestsExecutor(ctx context.Context, queue string) executorTests {
	p := executorTests{
		ctx:   ctx,
		queue: queue,
	}
	return p
}

func (w executorTests) OnRetry(attempt int, err error, task *retrypool.TaskWrapper[*retrypool.RequestResponse[testRequest, *testResponse]]) {
	fmt.Println("OnRetry")
}

func (w executorTests) OnSuccess(controller retrypool.WorkerController[*retrypool.RequestResponse[testRequest, *testResponse]], workerID int, worker retrypool.Worker[*retrypool.RequestResponse[testRequest, *testResponse]], task *retrypool.TaskWrapper[*retrypool.RequestResponse[testRequest, *testResponse]]) {
	fmt.Println("OnSuccess")
}

func (w executorTests) OnFailure(controller retrypool.WorkerController[*retrypool.RequestResponse[testRequest, *testResponse]], workerID int, worker retrypool.Worker[*retrypool.RequestResponse[testRequest, *testResponse]], task *retrypool.TaskWrapper[*retrypool.RequestResponse[testRequest, *testResponse]], err error) retrypool.DeadTaskAction {
	fmt.Println("OnFailure")
	return retrypool.DeadTaskActionDoNothing
}

func (w executorTests) OnDeadTask(task *retrypool.DeadTask[*retrypool.RequestResponse[testRequest, *testResponse]]) {
	fmt.Println("OnDeadTask")
}

func (w executorTests) OnPanic(task *retrypool.RequestResponse[testRequest, *testResponse], v interface{}, stackTrace string) {
	fmt.Println("OnPanic")
}

func (w executorTests) OnExecutorPanic(worker int, recovery any, err error, stackTrace string) {
	fmt.Println("OnExecutorPanic")
}

type workerTests struct {
	ctx   context.Context
	queue string
	id    int
}

func newTestsWorker(ctx context.Context, queue string, id int) workerTests {
	w := workerTests{
		ctx:   ctx,
		queue: queue,
		id:    id,
	}
	return w
}

func (w workerTests) Run(ctx context.Context, data *retrypool.RequestResponse[testRequest, *testResponse]) error {

	fmt.Println("Run", w.id)
	data.Complete(&testResponse{
		msg:      "Hello",
		workedBy: w.id,
	})

	return nil
}

func TestExecutorWorkerPool(t *testing.T) {
	ctx := context.Background()
	queue := "test"
	count := 0

	wp := NewWorkerPool(
		ctx,
		queue,
		newTestsExecutor(ctx, queue),
		func(ctx context.Context, queue string) retrypool.Worker[*retrypool.RequestResponse[testRequest, *testResponse]] {
			count++
			fmt.Println("Worker", count)
			return newTestsWorker(ctx, queue, count)
		},
	)

	fmt.Println(wp.AddWorker())
	fmt.Println(wp.AddWorker())

	for i := 0; i < 10; i++ {
		task := retrypool.NewRequestResponse[testRequest, *testResponse](testRequest{})

		if err := wp.Submit(task); err != nil {
			t.Fatalf("Error: %v", err)
		}

		nCtx := context.Background()
		nCtx, cancel := context.WithTimeout(nCtx, 1*time.Second)
		defer cancel()

		fmt.Println(wp.pool.QueueSize(), wp.pool.ProcessingCount(), wp.pool.DeadTaskCount())

		resp, err := task.Wait(nCtx)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}

		fmt.Println(wp.pool.QueueSize(), wp.pool.ProcessingCount(), wp.pool.DeadTaskCount())

		if resp != nil {
			fmt.Println(resp)
		}
	}

	if err := wp.Wait(); err != nil {
		t.Fatalf("Error: %v", err)
	}

	if err := wp.Shutdown(); err != nil {
		if err != context.Canceled {
			t.Fatalf("Error: %v", err)
		}
	}
}
