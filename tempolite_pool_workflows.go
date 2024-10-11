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

type workflowTask struct {
	ctx         WorkflowContext
	handler     interface{}
	handlerName HandlerIdentity
	params      []interface{}
	retryCount  int
	maxRetry    int
	retry       func() error
}

func (tp *Tempolite) createWorkflowPool() *retrypool.Pool[*workflowTask] {

	opts := []retrypool.Option[*workflowTask]{
		retrypool.WithAttempts[*workflowTask](1),
		retrypool.WithOnTaskSuccess(tp.workflowOnSuccess),
		retrypool.WithOnTaskFailure(tp.workflowOnFailure),
		retrypool.WithPanicHandler(tp.workflowOnPanic),
		retrypool.WithOnRetry(tp.workflowOnRetry),
	}

	workers := []retrypool.Worker[*workflowTask]{}

	for i := 0; i < 5; i++ {
		workers = append(workers, workflowWorker{id: i, tp: tp})
	}

	return retrypool.New(
		tp.ctx,
		workers,
		opts...)
}

func (tp *Tempolite) workflowOnPanic(task *workflowTask, v interface{}, stackTrace string) {
	log.Printf("Task panicked: %v", v)
	log.Println(stackTrace)
}

func (tp *Tempolite) workflowOnRetry(attempt int, err error, task *retrypool.TaskWrapper[*workflowTask]) {
	log.Printf("onHandlerTaskRetry: %d, %v", attempt, err)
}

func (tp *Tempolite) workflowOnSuccess(controller retrypool.WorkerController[*workflowTask], workerID int, worker retrypool.Worker[*workflowTask], task *retrypool.TaskWrapper[*workflowTask]) {

	log.Printf("workflowOnSuccess workflow %v - %v: %d", task.Data().ctx.workflowID, task.Data().ctx.executionID, workerID)

	if _, err := tp.client.WorkflowExecution.UpdateOneID(task.Data().ctx.executionID).SetStatus(workflowexecution.StatusCompleted).
		Save(tp.ctx); err != nil {
		log.Printf("workflowOnSuccess: WorkflowExecution.Update failed: %v", err)
	}

	if _, err := tp.client.Workflow.UpdateOneID(task.Data().ctx.workflowID).SetStatus(workflow.StatusCompleted).Save(tp.ctx); err != nil {
		log.Printf("workflowOnSuccess: Workflow.UpdateOneID failed: %v", err)
	}
}

func (tp *Tempolite) workflowOnFailure(controller retrypool.WorkerController[*workflowTask], workerID int, worker retrypool.Worker[*workflowTask], task *retrypool.TaskWrapper[*workflowTask], err error) retrypool.DeadTaskAction {

	// printf with err + retryCount + maxRetry
	log.Printf("workflowOnFailure: %v, %d, %d", err, task.Data().retryCount, task.Data().maxRetry)

	// Simple retry mechanism
	// We know in advance the config and the retry value, we can manage in-memory
	if task.Data().maxRetry > 0 && task.Data().retryCount < task.Data().maxRetry {
		task.Data().retryCount++
		fmt.Println("retry it the task: ", err)
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

		return retrypool.DeadTaskActionDoNothing
	}

	log.Printf("workflowOnFailure workflow %v - %v: %d", task.Data().ctx.workflowID, task.Data().ctx.executionID, workerID)

	if _, err := tp.client.WorkflowExecution.UpdateOneID(task.Data().ctx.executionID).SetStatus(workflowexecution.StatusFailed).SetError(err.Error()).Save(tp.ctx); err != nil {
		log.Printf("workflowOnFailure: WorkflowExecution.Update failed: %v", err)
	}
	if _, err := tp.client.Workflow.UpdateOneID(task.Data().ctx.workflowID).SetStatus(workflow.StatusFailed).Save(tp.ctx); err != nil {
		log.Printf("workflowOnSuccess: Workflow.UpdateOneID failed: %v", err)
	}

	return retrypool.DeadTaskActionDoNothing
}

type workflowWorker struct {
	id int
	tp *Tempolite
}

func (w workflowWorker) Run(ctx context.Context, data *workflowTask) error {
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

	fmt.Println("output to save", res, errRes)
	if _, err := w.tp.client.WorkflowExecution.UpdateOneID(data.ctx.executionID).SetOutput(res).Save(w.tp.ctx); err != nil {
		log.Printf("workflowWorker: WorkflowExecution.Update failed: %v", err)
		return err
	}

	return errRes
}
