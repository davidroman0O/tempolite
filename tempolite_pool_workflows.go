package tempolite

import (
	"context"
	"fmt"
	"log"
	"reflect"

	"github.com/davidroman0O/go-tempolite/ent/workflow"
	"github.com/davidroman0O/go-tempolite/ent/workflowexecution"
	"github.com/davidroman0O/retrypool"
)

type workflowTask[T Identifier] struct {
	ctx         WorkflowContext[T]
	handler     interface{}
	handlerName HandlerIdentity
	params      []interface{}
	retryCount  int
	maxRetry    int
	retry       func() error
	isPaused    bool
}

func (tp *Tempolite[T]) createWorkflowPool() *retrypool.Pool[*workflowTask[T]] {

	opts := []retrypool.Option[*workflowTask[T]]{
		retrypool.WithAttempts[*workflowTask[T]](1),
		retrypool.WithOnTaskSuccess(tp.workflowOnSuccess),
		retrypool.WithOnTaskFailure(tp.workflowOnFailure),
		retrypool.WithPanicHandler(tp.workflowOnPanic),
		retrypool.WithOnRetry(tp.workflowOnRetry),
		retrypool.WithPanicWorker[*workflowTask[T]](tp.workflowWorkerPanic),
	}

	workers := []retrypool.Worker[*workflowTask[T]]{}

	for i := 0; i < 5; i++ {
		workers = append(workers, workflowWorker[T]{id: i, tp: tp})
	}

	return retrypool.New(
		tp.ctx,
		workers,
		opts...)
}

func (tp *Tempolite[T]) workflowWorkerPanic(workerID int, recovery any, err error, stackTrace string) {
	log.Printf("workflowWorkerPanic - workerID %d: %v", workerID, err)
	log.Println(stackTrace)
}

func (tp *Tempolite[T]) workflowOnPanic(task *workflowTask[T], v interface{}, stackTrace string) {
	log.Printf("Task panicked: %v", v)
	log.Println(stackTrace)
}

func (tp *Tempolite[T]) workflowOnRetry(attempt int, err error, task *retrypool.TaskWrapper[*workflowTask[T]]) {
	log.Printf("onHandlerTaskRetry: %d, %v", attempt, err)
}

func (tp *Tempolite[T]) workflowOnSuccess(controller retrypool.WorkerController[*workflowTask[T]], workerID int, worker retrypool.Worker[*workflowTask[T]], task *retrypool.TaskWrapper[*workflowTask[T]]) {
	if task.Data().isPaused {
		log.Printf("Workflow %s paused", task.Data().ctx.workflowID)
		// Update workflow status to paused in the database
		if _, err := tp.client.Workflow.UpdateOneID(task.Data().ctx.workflowID).
			SetIsPaused(true).
			SetIsReady(false).
			Save(tp.ctx); err != nil {
			log.Printf("Failed to update workflow pause status: %v", err)
		}
		return
	}

	log.Printf("workflowOnSuccess workflow %v - %v: %d - isPaused: %v", task.Data().ctx.workflowID, task.Data().ctx.executionID, workerID, task.Data().isPaused)

	if _, err := tp.client.WorkflowExecution.UpdateOneID(task.Data().ctx.executionID).SetStatus(workflowexecution.StatusCompleted).
		Save(tp.ctx); err != nil {
		log.Printf("workflowOnSuccess: WorkflowExecution.Update failed: %v", err)
	}

	if _, err := tp.client.Workflow.UpdateOneID(task.Data().ctx.workflowID).SetStatus(workflow.StatusCompleted).Save(tp.ctx); err != nil {
		log.Printf("workflowOnSuccess: Workflow.UpdateOneID failed: %v", err)
	}
}

func (tp *Tempolite[T]) workflowOnFailure(controller retrypool.WorkerController[*workflowTask[T]], workerID int, worker retrypool.Worker[*workflowTask[T]], task *retrypool.TaskWrapper[*workflowTask[T]], err error) retrypool.DeadTaskAction {

	// printf with err + retryCount + maxRetry
	log.Printf("workflowOnFailure:  err: %v,  data: %d, maxrety: %d", err, task.Data().retryCount, task.Data().maxRetry)

	total, err := tp.client.WorkflowExecution.Query().Where(workflowexecution.HasWorkflowWith(workflow.IDEQ(task.Data().ctx.workflowID))).Count(tp.ctx)
	if err != nil {
		log.Printf("workflowOnFailure: WorkflowExecution.Query failed: %v", err)
	}

	total = total - 1 // removing myself

	fmt.Printf("workflowOnFailure: total: %d\n", total)

	// Simple retry mechanism
	// We know in advance the config and the retry value, we can manage in-memory
	if task.Data().maxRetry > 0 && total < task.Data().maxRetry {

		fmt.Println("retry it the task: ", task.Data().retryCount, total, err)
		// Just by creating a new workflow execution, we're incrementing the total count of executions which is the retry count in the database
		if _, err := tp.client.WorkflowExecution.UpdateOneID(task.Data().ctx.executionID).SetStatus(workflowexecution.StatusRetried).Save(tp.ctx); err != nil {
			log.Printf("workflowOnFailure: WorkflowExecution.Update failed: %v", err)
		}

		if _, err := tp.client.Workflow.UpdateOneID(task.Data().ctx.workflowID).SetStatus(workflow.StatusRetried).Save(tp.ctx); err != nil {
			log.Printf("workflowOnSuccess: Workflow.UpdateOneID failed: %v", err)
		}

		if err := task.Data().retry(); err != nil {
			log.Printf("workflowOnFailure: retry failed: %v", err)
			return retrypool.DeadTaskActionAddToDeadTasks
		}

		return retrypool.DeadTaskActionForceRetry
	}

	log.Printf("workflowOnFailure workflow %v - %v: %d", task.Data().ctx.workflowID, task.Data().ctx.executionID, workerID)

	// we won't re-try the workflow, probably was an error somewhere
	if err != nil {
		if _, err := tp.client.WorkflowExecution.UpdateOneID(task.Data().ctx.executionID).SetStatus(workflowexecution.StatusFailed).SetError(err.Error()).Save(tp.ctx); err != nil {
			log.Printf("workflowOnFailure: WorkflowExecution.Update failed: %v", err)
		}
	} else {
		// not retrying and the there is no error
		if _, err := tp.client.WorkflowExecution.UpdateOneID(task.Data().ctx.executionID).SetStatus(workflowexecution.StatusFailed).SetError("unknown error").Save(tp.ctx); err != nil {
			log.Printf("workflowOnFailure: WorkflowExecution.Update failed: %v", err)
		}
	}

	if _, err := tp.client.Workflow.UpdateOneID(task.Data().ctx.workflowID).SetStatus(workflow.StatusFailed).Save(tp.ctx); err != nil {
		log.Printf("workflowOnSuccess: Workflow.UpdateOneID failed: %v", err)
	}

	return retrypool.DeadTaskActionDoNothing
}

type workflowWorker[T Identifier] struct {
	id int
	tp *Tempolite[T]
}

func (w workflowWorker[T]) Run(ctx context.Context, data *workflowTask[T]) error {
	log.Printf("workflowWorker: %s, %v", data.handlerName, data.params)

	values := []reflect.Value{reflect.ValueOf(data.ctx)}

	for _, v := range data.params {
		values = append(values, reflect.ValueOf(v))
	}

	returnedValues := reflect.ValueOf(data.handler).Call(values)

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

	if errRes == errWorkflowPaused {
		data.isPaused = true
		fmt.Println("pause detected", data.isPaused)
		return nil
	}

	fmt.Println("output to save", res, errRes)
	if _, err := w.tp.client.WorkflowExecution.UpdateOneID(data.ctx.executionID).SetOutput(res).Save(w.tp.ctx); err != nil {
		log.Printf("workflowWorker: WorkflowExecution.Update failed: %v", err)
		return err
	}

	return errRes
}
