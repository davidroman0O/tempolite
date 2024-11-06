package registry

import (
	"fmt"
	"reflect"
	"runtime"

	"github.com/davidroman0O/tempolite/internal/engine/context"
	"github.com/davidroman0O/tempolite/internal/types"
)

func (r *Registry) registerWorkflow(workflowFunc interface{}) error {
	handlerType := reflect.TypeOf(workflowFunc)

	if handlerType.Kind() != reflect.Func {
		return fmt.Errorf("workflow must be a function")
	}

	if handlerType.NumIn() < 1 {
		return fmt.Errorf("workflow function must have at least one input parameter (WorkflowContext)")
	}

	if handlerType.In(0) != reflect.TypeOf(context.WorkflowContext{}) {
		return fmt.Errorf("first parameter of workflow function must be WorkflowContext")
	}

	// Collect parameter types after the context parameter
	paramsKinds := []reflect.Kind{}
	paramTypes := []reflect.Type{}
	for i := 1; i < handlerType.NumIn(); i++ {
		paramTypes = append(paramTypes, handlerType.In(i))
		paramsKinds = append(paramsKinds, handlerType.In(i).Kind())
	}

	// Collect return types excluding error
	numOut := handlerType.NumOut()
	if numOut == 0 {
		return fmt.Errorf("workflow function must return at least an error")
	}

	returnKinds := []reflect.Kind{}
	returnTypes := []reflect.Type{}
	for i := 0; i < numOut-1; i++ {
		returnTypes = append(returnTypes, handlerType.Out(i))
		returnKinds = append(returnKinds, handlerType.Out(i).Kind())
	}

	// Verify that the last return type is error
	if handlerType.Out(numOut-1) != reflect.TypeOf((*error)(nil)).Elem() {
		return fmt.Errorf("last return value of workflow function must be error")
	}

	// Generate a unique identifier for the workflow function
	funcName := runtime.FuncForPC(reflect.ValueOf(workflowFunc).Pointer()).Name()
	handlerIdentity := types.HandlerIdentity(funcName)

	workflow := &types.Workflow{
		HandlerName:     funcName,
		HandlerLongName: handlerIdentity,
		Handler:         workflowFunc,
		ParamTypes:      paramTypes,
		ParamsKinds:     paramsKinds,
		ReturnTypes:     returnTypes,
		ReturnKinds:     returnKinds,
		NumIn:           handlerType.NumIn() - 1, // Exclude context
		NumOut:          numOut - 1,              // Exclude error
	}

	r.Lock()
	r.workflows[workflow.HandlerLongName] = *workflow
	r.Unlock()
	return nil
}
