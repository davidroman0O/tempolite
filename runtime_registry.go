package tempolite

import (
	"fmt"
	"reflect"
	"runtime"

	"github.com/sasha-s/go-deadlock"
)

// Registry holds registered workflows, activities, and side effects.
type Registry struct {
	workflows   map[string]HandlerInfo
	activities  map[string]HandlerInfo
	sideEffects map[string]HandlerInfo
	mu          deadlock.Mutex
}

func NewRegistry() *Registry {
	return &Registry{
		workflows:   make(map[string]HandlerInfo),
		activities:  make(map[string]HandlerInfo),
		sideEffects: make(map[string]HandlerInfo),
	}
}

func (r *Registry) GetWorkflow(name string) (HandlerInfo, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	handler, ok := r.workflows[name]
	return handler, ok
}

func (r *Registry) GetActivity(name string) (HandlerInfo, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	handler, ok := r.activities[name]
	return handler, ok
}

func (r *Registry) GetWorkflowFunc(f interface{}) (HandlerInfo, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	name := getFunctionName(f)
	handler, ok := r.workflows[name]
	return handler, ok
}

func (r *Registry) GetActivityFunc(f interface{}) (HandlerInfo, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	name := getFunctionName(f)
	handler, ok := r.activities[name]
	return handler, ok
}

// TODO: add description schema so later features
func (r *Registry) RegisterWorkflow(workflowFunc interface{}) (HandlerInfo, error) {
	funcName := getFunctionName(workflowFunc)
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if already registered
	if handler, ok := r.workflows[funcName]; ok {
		return handler, nil
	}

	handlerType := reflect.TypeOf(workflowFunc)
	if handlerType.Kind() != reflect.Func {
		err := fmt.Errorf("workflow must be a function")
		return HandlerInfo{}, err
	}

	if handlerType.NumIn() < 1 {
		err := fmt.Errorf("workflow function must have at least one input parameter (WorkflowContext)")
		return HandlerInfo{}, err
	}

	expectedContextType := reflect.TypeOf(WorkflowContext{})
	if handlerType.In(0) != expectedContextType {
		err := fmt.Errorf("first parameter of workflow function must be WorkflowContext")
		return HandlerInfo{}, err
	}

	paramsKinds := []reflect.Kind{}
	paramTypes := []reflect.Type{}
	for i := 1; i < handlerType.NumIn(); i++ {
		paramTypes = append(paramTypes, handlerType.In(i))
		paramsKinds = append(paramsKinds, handlerType.In(i).Kind())
	}

	numOut := handlerType.NumOut()
	if numOut == 0 {
		err := fmt.Errorf("workflow function must return at least an error")
		return HandlerInfo{}, err
	}

	returnKinds := []reflect.Kind{}
	returnTypes := []reflect.Type{}
	for i := 0; i < numOut-1; i++ {
		returnTypes = append(returnTypes, handlerType.Out(i))
		returnKinds = append(returnKinds, handlerType.Out(i).Kind())
	}

	if handlerType.Out(numOut-1) != reflect.TypeOf((*error)(nil)).Elem() {
		err := fmt.Errorf("last return value of workflow function must be error")
		return HandlerInfo{}, err
	}

	handler := HandlerInfo{
		HandlerName:     funcName,
		HandlerLongName: HandlerIdentity(funcName),
		Handler:         workflowFunc,
		ParamTypes:      paramTypes,
		ParamsKinds:     paramsKinds,
		ReturnTypes:     returnTypes,
		ReturnKinds:     returnKinds,
		NumIn:           handlerType.NumIn() - 1,
		NumOut:          numOut - 1,
	}

	r.workflows[funcName] = handler
	return handler, nil
}

func (r *Registry) RegisterActivity(activityFunc interface{}) (HandlerInfo, error) {
	funcName := getFunctionName(activityFunc)
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if already registered
	if handler, ok := r.activities[funcName]; ok {
		return handler, nil
	}

	handlerType := reflect.TypeOf(activityFunc)
	if handlerType.Kind() != reflect.Func {
		err := fmt.Errorf("activity must be a function")
		return HandlerInfo{}, err
	}

	if handlerType.NumIn() < 1 {
		err := fmt.Errorf("activity function must have at least one input parameter (ActivityContext)")
		return HandlerInfo{}, err
	}

	expectedContextType := reflect.TypeOf(ActivityContext{})
	if handlerType.In(0) != expectedContextType {
		err := fmt.Errorf("first parameter of activity function must be ActivityContext")
		return HandlerInfo{}, err
	}

	paramsKinds := []reflect.Kind{}
	paramTypes := []reflect.Type{}
	for i := 1; i < handlerType.NumIn(); i++ {
		paramTypes = append(paramTypes, handlerType.In(i))
		paramsKinds = append(paramsKinds, handlerType.In(i).Kind())
	}

	numOut := handlerType.NumOut()
	if numOut == 0 {
		err := fmt.Errorf("activity function must return at least an error")
		return HandlerInfo{}, err
	}

	returnKinds := []reflect.Kind{}
	returnTypes := []reflect.Type{}
	for i := 0; i < numOut-1; i++ {
		returnTypes = append(returnTypes, handlerType.Out(i))
		returnKinds = append(returnKinds, handlerType.Out(i).Kind())
	}

	if handlerType.Out(numOut-1) != reflect.TypeOf((*error)(nil)).Elem() {
		err := fmt.Errorf("last return value of activity function must be error")
		return HandlerInfo{}, err
	}

	handler := HandlerInfo{
		HandlerName:     funcName,
		HandlerLongName: HandlerIdentity(funcName),
		Handler:         activityFunc,
		ParamTypes:      paramTypes,
		ParamsKinds:     paramsKinds,
		ReturnTypes:     returnTypes,
		ReturnKinds:     returnKinds,
		NumIn:           handlerType.NumIn() - 1,
		NumOut:          numOut - 1,
	}

	r.activities[funcName] = handler
	return handler, nil
}

func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

// TODO: trigger an error if the side effect got a `error` a return value
func (r *Registry) RegisterSideEffect(sideEffectFunc interface{}) (HandlerInfo, error) {
	funcName := getFunctionName(sideEffectFunc)
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if already registered
	if handler, ok := r.sideEffects[funcName]; ok {
		return handler, nil
	}

	handlerType := reflect.TypeOf(sideEffectFunc)
	if handlerType.Kind() != reflect.Func {
		return HandlerInfo{}, fmt.Errorf("side effect must be a function")
	}

	if handlerType.NumIn() != 0 {
		return HandlerInfo{}, fmt.Errorf("side effect function must take no parameters")
	}

	numOut := handlerType.NumOut()
	if numOut == 0 {
		return HandlerInfo{}, fmt.Errorf("side effect function must return at least one value")
	}

	returnTypes := make([]reflect.Type, numOut)
	returnKinds := make([]reflect.Kind, numOut)
	for i := 0; i < numOut; i++ {
		returnTypes[i] = handlerType.Out(i)
		returnKinds[i] = handlerType.Out(i).Kind()
	}

	handler := HandlerInfo{
		HandlerName:     funcName,
		HandlerLongName: HandlerIdentity(funcName),
		Handler:         sideEffectFunc,
		ReturnTypes:     returnTypes,
		ReturnKinds:     returnKinds,
		NumIn:           0,
		NumOut:          numOut,
	}

	r.sideEffects[funcName] = handler
	return handler, nil
}

func (r *Registry) GetSideEffect(name string) (HandlerInfo, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	handler, ok := r.sideEffects[name]
	return handler, ok
}

func (r *Registry) GetSideEffectFunc(f interface{}) (HandlerInfo, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := getFunctionName(f)
	handler, ok := r.sideEffects[name]
	return handler, ok
}
