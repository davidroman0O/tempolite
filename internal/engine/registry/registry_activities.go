package registry

import (
	"fmt"
	"reflect"
	"runtime"

	"github.com/davidroman0O/tempolite/internal/engine/context"
	"github.com/davidroman0O/tempolite/internal/types"
)

func (r *Registry) registerActivity(activityFunc interface{}) error {
	handlerType := reflect.TypeOf(activityFunc)

	if handlerType.Kind() != reflect.Func {
		return fmt.Errorf("activity must be a function")
	}

	if handlerType.NumIn() < 1 {
		return fmt.Errorf("activity function must have at least one input parameter (ActivityContext)")
	}

	if handlerType.In(0) != reflect.TypeOf(context.ActivityContext{}) {
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
	handlerIdentity := types.HandlerIdentity(funcName)

	activity := &types.Activity{
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

	// log.Printf("Registering activity function %s with name %s", funcName, handlerIdentity)
	r.Lock()
	r.activities[activity.HandlerLongName] = *activity
	r.Unlock()

	return nil
}

func (r *Registry) IsActivityRegistered(longName types.HandlerIdentity) bool {
	r.Lock()
	_, ok := r.activities[longName]
	r.Unlock()
	return ok
}
