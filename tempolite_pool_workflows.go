package tempolite

import (
	"context"
	"fmt"
	"log"
	"reflect"

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
	}

	return retrypool.New(
		tp.ctx,
		[]retrypool.Worker[*workflowTask]{
			workflowWorker{
				tp: tp,
			},
		},
		opts...)
}

func (tp *Tempolite) workflowOnSuccess(controller retrypool.WorkerController[*workflowTask], workerID int, worker retrypool.Worker[*workflowTask], task *retrypool.TaskWrapper[*workflowTask]) {

	log.Printf("workflowOnSuccess: %d", workerID)

	if _, err := tp.client.WorkflowExecution.Update().Where(workflowexecution.IDEQ(task.Data().ctx.executionID)).SetStatus(workflowexecution.StatusCompleted).Save(tp.ctx); err != nil {
		log.Printf("workflowOnSuccess: WorkflowExecution.Update failed: %v", err)
	}
}

func (tp *Tempolite) workflowOnFailure(controller retrypool.WorkerController[*workflowTask], workerID int, worker retrypool.Worker[*workflowTask], task *retrypool.TaskWrapper[*workflowTask], err error) retrypool.DeadTaskAction {

	// Simple retry mechanism
	// We know in advance the config and the retry value, we can manage in-memory
	if task.Data().retryCount < task.Data().maxRetry {
		task.Data().retryCount++
		fmt.Println("retry it the task: ", err)
		// Just by creating a new workflow execution, we're incrementing the total count of executions which is the retry count in the database
		if _, err := tp.client.WorkflowExecution.Update().SetStatus(workflowexecution.StatusRetried).Save(tp.ctx); err != nil {
			log.Printf("workflowOnFailure: WorkflowExecution.Update failed: %v", err)
		}
		if err := task.Data().retry(); err != nil {
			log.Printf("workflowOnFailure: retry failed: %v", err)
			return retrypool.DeadTaskActionAddToDeadTasks
		}
		return retrypool.DeadTaskActionDoNothing
	}

	if _, err := tp.client.WorkflowExecution.Update().SetStatus(workflowexecution.StatusFailed).Save(tp.ctx); err != nil {
		log.Printf("workflowOnFailure: WorkflowExecution.Update failed: %v", err)
	}

	fmt.Println("remove it the task: ", err)

	return retrypool.DeadTaskActionDoNothing
}

type workflowWorker struct {
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

	if _, err := w.tp.client.WorkflowExecution.Update().Where(workflowexecution.IDEQ(data.ctx.executionID)).SetOutput(res).Save(w.tp.ctx); err != nil {
		log.Printf("workflowWorker: WorkflowExecution.Update failed: %v", err)
	}

	return errRes
}
