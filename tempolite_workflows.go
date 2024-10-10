package tempolite

import (
	"context"
	"fmt"
	"log"
	"reflect"

	"github.com/davidroman0O/retrypool"
)

type workflowTask struct {
	handler     interface{}
	handlerName HandlerIdentity
	params      []interface{}
}

func (tp *Tempolite) createWorkflowPool(opts ...retrypool.Option[*workflowTask]) *retrypool.Pool[*workflowTask] {
	return retrypool.New(
		tp.ctx,
		[]retrypool.Worker[*workflowTask]{
			workflowWorker{
				tp: tp,
			},
		},
		opts...)
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

	fmt.Println(res, errRes)

	return nil
}
