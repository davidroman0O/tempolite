package tempolite

import (
	"context"
	"log"
	"reflect"

	"github.com/davidroman0O/go-tempolite/ent/workflowexecution"
	"github.com/davidroman0O/retrypool"
)

type workflowTask struct {
	handler     interface{}
	handlerName HandlerIdentity
	params      []interface{}
}

func (tp *Tempolite) createWorkflowPool() *retrypool.Pool[*workflowTask] {

	opts := []retrypool.Option[*workflowTask]{
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
	if _, err := tp.client.WorkflowExecution.Update().SetStatus(workflowexecution.StatusCompleted).Save(tp.ctx); err != nil {
		log.Printf("workflowOnSuccess: WorkflowExecution.Update failed: %v", err)
	}
}

func (tp *Tempolite) workflowOnFailure(controller retrypool.WorkerController[*workflowTask], workerID int, worker retrypool.Worker[*workflowTask], task *retrypool.TaskWrapper[*workflowTask], err error) retrypool.DeadTaskAction {
	if _, err := tp.client.WorkflowExecution.Update().SetStatus(workflowexecution.StatusFailed).Save(tp.ctx); err != nil {
		log.Printf("workflowOnSuccess: WorkflowExecution.Update failed: %v", err)
	}
	return retrypool.DeadTaskActionAddToDeadTasks
}

type workflowWorker struct {
	tp *Tempolite
}

func (w workflowWorker) Run(ctx context.Context, data *workflowTask) error {
	log.Printf("workflowWorker: %s, %v", data.handlerName, data.params)

	contextWorkflow := WorkflowContext{
		TempoliteContext: ctx,
		tp:               w.tp,
	}

	values := []reflect.Value{reflect.ValueOf(contextWorkflow)}
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

	return errRes
}
