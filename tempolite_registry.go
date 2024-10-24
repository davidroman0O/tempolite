package tempolite

import (
	"fmt"
	"log"
	"reflect"
	"runtime"
)

type SideEffect HandlerInfo

type Workflow HandlerInfo

func As[T any]() HandlerIdentity {
	dataType := reflect.TypeOf((*T)(nil)).Elem()
	return HandlerIdentity(fmt.Sprintf("%s/%s", dataType.PkgPath(), dataType.Name()))
}

type Activity HandlerInfo

type ActivityRegister func() (*Activity, error)

func activityRegisterType[T any](dataType reflect.Type, instance T) ActivityRegister {
	return func() (*Activity, error) {
		handlerValue := reflect.ValueOf(instance)
		runMethod := handlerValue.MethodByName("Run")
		if !runMethod.IsValid() {
			return nil, fmt.Errorf("%s must have a Run method", dataType)
		}

		runMethodType := runMethod.Type()

		if runMethodType.NumIn() < 1 {
			return nil, fmt.Errorf("Run method of %s must have at least one input parameter (ActivityContext)", dataType)
		}

		if runMethodType.In(0) != reflect.TypeOf(ActivityContext{}) {
			return nil, fmt.Errorf("first parameter of Run method in %s must be ActivityContext", dataType)
		}

		// Collect all parameter types after the context parameter
		paramTypes := []reflect.Type{}
		paramsKinds := []reflect.Kind{}
		for i := 1; i < runMethodType.NumIn(); i++ {
			paramTypes = append(paramTypes, runMethodType.In(i))
			paramsKinds = append(paramsKinds, runMethodType.In(i).Kind())
		}

		// Collect all return types except the last one if it's an error
		numOut := runMethodType.NumOut()
		if numOut == 0 {
			return nil, fmt.Errorf("Run method in %s must return at least an error", dataType)
		}

		returnTypes := []reflect.Type{}
		returnKinds := []reflect.Kind{}
		for i := 0; i < numOut-1; i++ {
			returnTypes = append(returnTypes, runMethodType.Out(i))
			returnKinds = append(returnKinds, runMethodType.Out(i).Kind())
		}

		// Verify that the last return type is error
		if runMethodType.Out(numOut-1) != reflect.TypeOf((*error)(nil)).Elem() {
			return nil, fmt.Errorf("last return value of Run method in %s must be error", dataType)
		}

		act := &Activity{
			HandlerName:     dataType.Name(),
			HandlerLongName: HandlerIdentity(fmt.Sprintf("%s/%s", dataType.PkgPath(), dataType.Name())),
			Handler:         runMethod.Interface(), // Set Handler to the Run method
			ParamTypes:      paramTypes,
			ParamsKinds:     paramsKinds,
			ReturnTypes:     returnTypes,
			ReturnKinds:     returnKinds,
			NumIn:           runMethodType.NumIn() - 1, // Exclude context
			NumOut:          numOut - 1,                // Exclude error
		}

		log.Printf("Registering activity %s with name %s", dataType.Name(), act.HandlerLongName)

		return act, nil
	}
}

func (tp *Tempolite) registerWorkflow(workflowFunc interface{}) error {
	handlerType := reflect.TypeOf(workflowFunc)

	if handlerType.Kind() != reflect.Func {
		return fmt.Errorf("workflow must be a function")
	}

	if handlerType.NumIn() < 1 {
		return fmt.Errorf("workflow function must have at least one input parameter (WorkflowContext)")
	}

	if handlerType.In(0) != reflect.TypeOf(WorkflowContext{}) {
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
	handlerIdentity := HandlerIdentity(funcName)

	workflow := &Workflow{
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

	log.Printf("Registering workflow %s with name %s", funcName, handlerIdentity)

	tp.workflows.Store(workflow.HandlerLongName, *workflow)
	return nil
}

func (tp *Tempolite) registerActivity(activityFunc interface{}) error {
	handlerType := reflect.TypeOf(activityFunc)

	if handlerType.Kind() != reflect.Func {
		return fmt.Errorf("activity must be a function")
	}

	if handlerType.NumIn() < 1 {
		return fmt.Errorf("activity function must have at least one input parameter (ActivityContext)")
	}

	if handlerType.In(0) != reflect.TypeOf(ActivityContext{}) {
		return fmt.Errorf("first parameter of activity function must be ActivityContext")
	}

	// Collect parameter types after the context parameter
	paramTypes := []reflect.Type{}
	paramsKinds := []reflect.Kind{}
	for i := 1; i < handlerType.NumIn(); i++ {
		paramTypes = append(paramTypes, handlerType.In(i))
		paramsKinds = append(paramsKinds, handlerType.In(i).Kind())
	}

	// Collect return types excluding error
	numOut := handlerType.NumOut()
	if numOut == 0 {
		return fmt.Errorf("activity function must return at least an error")
	}

	returnTypes := []reflect.Type{}
	returnKinds := []reflect.Kind{}
	for i := 0; i < numOut-1; i++ {
		returnTypes = append(returnTypes, handlerType.Out(i))
		returnKinds = append(returnKinds, handlerType.Out(i).Kind())
	}

	// Verify that the last return type is error
	if handlerType.Out(numOut-1) != reflect.TypeOf((*error)(nil)).Elem() {
		return fmt.Errorf("last return value of activity function must be error")
	}

	// Generate a unique identifier for the activity function
	funcName := runtime.FuncForPC(reflect.ValueOf(activityFunc).Pointer()).Name()
	handlerIdentity := HandlerIdentity(funcName)

	activity := &Activity{
		HandlerName:     funcName,
		HandlerLongName: handlerIdentity,
		Handler:         activityFunc,
		ParamTypes:      paramTypes,
		ParamsKinds:     paramsKinds,
		ReturnTypes:     returnTypes,
		ReturnKinds:     returnKinds,
		NumIn:           handlerType.NumIn() - 1, // Exclude context
		NumOut:          numOut - 1,              // Exclude error
	}

	log.Printf("Registering activity function %s with name %s", funcName, handlerIdentity)

	tp.activities.Store(activity.HandlerLongName, *activity)
	return nil
}

func (tp *Tempolite) IsActivityRegistered(longName HandlerIdentity) bool {
	_, ok := tp.activities.Load(longName)
	return ok
}
