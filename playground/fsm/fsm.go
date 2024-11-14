package main

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/qmuntal/stateless"
)

const (
	// FSM states
	StateIdle      = "Idle"
	StateExecuting = "Executing"
	StateCompleted = "Completed"
	StateFailed    = "Failed"

	// FSM triggers
	TriggerStart              = "Start"
	TriggerExecuteSubWorkflow = "ExecuteSubWorkflow"
	TriggerComplete           = "Complete"
	TriggerFail               = "Fail"
)

type Future struct {
	result interface{}
	err    error
	done   chan struct{}
}

func NewFuture() *Future {
	return &Future{
		done: make(chan struct{}),
	}
}

func (f *Future) Get(result interface{}) error {
	log.Printf("Future.Get called")
	<-f.done
	if f.err != nil {
		log.Printf("Future.Get returning error: %v", f.err)
		return f.err
	}
	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(f.result))
	log.Printf("Future.Get returning result: %v", f.result)
	return nil
}

func (f *Future) setResult(result interface{}) {
	log.Printf("Future.setResult called with result: %v", result)
	f.result = result
	close(f.done)
}

func (f *Future) setError(err error) {
	log.Printf("Future.setError called with error: %v", err)
	f.err = err
	close(f.done)
}

type WorkflowContext struct {
	orchestrator *Orchestrator
	ctx          context.Context
}

func (ctx *WorkflowContext) Workflow(workflowFunc interface{}, args ...interface{}) *Future {
	log.Printf("WorkflowContext.Workflow called with workflowFunc: %v, args: %v", getFunctionName(workflowFunc), args)
	future := NewFuture()

	handler, ok := ctx.orchestrator.registry.GetHandler(workflowFunc)
	if !ok {
		err := fmt.Errorf("function not found")
		log.Printf("Error: %v", err)
		future.setError(err)
		return future
	}

	subWorkflowInstance := &WorkflowInstance{
		stepID:       getFunctionName(workflowFunc),
		handler:      handler,
		input:        args,
		future:       future,
		ctx:          ctx.ctx,
		orchestrator: ctx.orchestrator,
	}

	ctx.orchestrator.addWorkflowInstance(subWorkflowInstance)
	subWorkflowInstance.Start()

	return future
}

type RegistryBuilder struct {
	workflows  []interface{}
	activities []interface{}
}

func NewRegistryBuilder() *RegistryBuilder {
	return &RegistryBuilder{
		workflows:  []interface{}{},
		activities: []interface{}{},
	}
}

func (r *RegistryBuilder) RegisterWorkflow(workflowFunc interface{}) *RegistryBuilder {
	log.Printf("RegistryBuilder.Register called with workflowFunc: %v", getFunctionName(workflowFunc))
	r.workflows = append(r.workflows, workflowFunc)
	return r
}

func (r *RegistryBuilder) RegisterActivity(activityFunc interface{}) *RegistryBuilder {
	log.Printf("RegistryBuilder.Register called with activityFunc: %v", getFunctionName(activityFunc))
	r.activities = append(r.activities, activityFunc)
	return r
}

func (r *RegistryBuilder) Build() (*Registry, error) {
	log.Printf("RegistryBuilder.Build called")
	registry := newRegistry()

	for _, v := range r.workflows {
		log.Printf("Registering workflow function: %v", getFunctionName(v))
		if err := registry.registerWorkflow(v); err != nil {
			return nil, err
		}
	}

	for _, v := range r.activities {
		log.Printf("Registering activities function: %v", getFunctionName(v))
		if err := registry.registerActivity(v); err != nil {
			return nil, err
		}
	}

	return registry, nil
}

type HandlerIdentity string

func (h HandlerIdentity) String() string {
	return string(h)
}

type HandlerInfo struct {
	HandlerName     string
	HandlerLongName HandlerIdentity
	Handler         interface{}
	ParamsKinds     []reflect.Kind
	ParamTypes      []reflect.Type
	ReturnTypes     []reflect.Type
	ReturnKinds     []reflect.Kind
	NumIn           int
	NumOut          int
}

type Registry struct {
	workflows  map[string]HandlerInfo
	activities map[HandlerIdentity]HandlerInfo
	mu         sync.Mutex
}

func newRegistry() *Registry {
	return &Registry{
		workflows:  make(map[string]HandlerInfo),
		activities: make(map[HandlerIdentity]HandlerInfo),
	}
}

func (r *Registry) registerActivity(activityFunc interface{}) error {
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

	activity := &HandlerInfo{
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
	r.mu.Lock()
	r.activities[activity.HandlerLongName] = *activity
	r.mu.Unlock()

	return nil
}

func (r *Registry) registerWorkflow(workflowFunc interface{}) error {
	log.Printf("Registry.registerWorkflow called with workflowFunc: %v", getFunctionName(workflowFunc))
	handlerType := reflect.TypeOf(workflowFunc)

	if handlerType.Kind() != reflect.Func {
		err := fmt.Errorf("workflow must be a function")
		log.Printf("Error: %v", err)
		return err
	}

	if handlerType.NumIn() < 1 {
		err := fmt.Errorf("workflow function must have at least one input parameter (WorkflowContext)")
		log.Printf("Error: %v", err)
		return err
	}

	expectedContextType := reflect.TypeOf(&WorkflowContext{})
	if handlerType.In(0) != expectedContextType {
		err := fmt.Errorf("first parameter of workflow function must be *WorkflowContext")
		log.Printf("Error: %v", err)
		return err
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
		log.Printf("Error: %v", err)
		return err
	}

	returnKinds := []reflect.Kind{}
	returnTypes := []reflect.Type{}
	for i := 0; i < numOut-1; i++ {
		returnTypes = append(returnTypes, handlerType.Out(i))
		returnKinds = append(returnKinds, handlerType.Out(i).Kind())
	}

	if handlerType.Out(numOut-1) != reflect.TypeOf((*error)(nil)).Elem() {
		err := fmt.Errorf("last return value of workflow function must be error")
		log.Printf("Error: %v", err)
		return err
	}

	funcName := getFunctionName(workflowFunc)
	handlerIdentity := HandlerIdentity(funcName)

	workflow := &HandlerInfo{
		HandlerName:     funcName,
		HandlerLongName: handlerIdentity,
		Handler:         workflowFunc,
		ParamTypes:      paramTypes,
		ParamsKinds:     paramsKinds,
		ReturnTypes:     returnTypes,
		ReturnKinds:     returnKinds,
		NumIn:           handlerType.NumIn() - 1,
		NumOut:          numOut - 1,
	}

	r.mu.Lock()
	r.workflows[funcName] = *workflow
	r.mu.Unlock()
	log.Printf("Workflow function registered: %s", funcName)

	return nil
}

func (r *Registry) GetHandler(workflowFunc interface{}) (HandlerInfo, bool) {
	funcName := getFunctionName(workflowFunc)
	log.Printf("Registry.GetHandler called with funcName: %s", funcName)
	r.mu.Lock()
	defer r.mu.Unlock()
	v, ok := r.workflows[funcName]
	if !ok {
		log.Printf("Handler not found for funcName: %s", funcName)
	} else {
		log.Printf("Handler found for funcName: %s", funcName)
	}
	return v, ok
}

type Database struct {
	workflows []*WorkflowInstance
	mu        sync.Mutex
	registry  *Registry
}

type WorkflowInstance struct {
	stepID       string
	handler      HandlerInfo
	input        []interface{}
	result       interface{}
	err          error
	fsm          *stateless.StateMachine
	future       *Future
	ctx          context.Context
	orchestrator *Orchestrator
}

func (db *Database) Workflow(workflowFunc interface{}, args ...interface{}) error {
	log.Printf("Database.Workflow called with workflowFunc: %v, args: %v", getFunctionName(workflowFunc), args)
	db.mu.Lock()
	defer db.mu.Unlock()

	handler, ok := db.registry.GetHandler(workflowFunc)
	if !ok {
		err := fmt.Errorf("function not found")
		log.Printf("Error: %v", err)
		return err
	}

	instance := &WorkflowInstance{
		stepID:  "root",
		handler: handler,
		input:   args,
	}
	db.workflows = append(db.workflows, instance)
	log.Printf("Workflow instance added to database: %+v", instance)
	return nil
}

func (db *Database) PullNextWorkflow() *WorkflowInstance {
	log.Printf("Database.PullNextWorkflow called")
	db.mu.Lock()
	defer db.mu.Unlock()
	if len(db.workflows) == 0 {
		log.Printf("No workflows in database")
		return nil
	}
	instance := db.workflows[0]
	db.workflows = db.workflows[1:]
	log.Printf("Pulled workflow instance from database: %+v", instance)
	return instance
}

type Orchestrator struct {
	db          *Database
	registry    *Registry
	rootWf      *WorkflowInstance
	ctx         context.Context
	instances   []*WorkflowInstance
	instancesMu sync.Mutex
}

func NewOrchestrator(db *Database, registry *Registry, ctx context.Context) *Orchestrator {
	log.Printf("NewOrchestrator called")
	o := &Orchestrator{
		db:       db,
		registry: registry,
		ctx:      ctx,
	}

	return o
}

func (o *Orchestrator) addWorkflowInstance(wi *WorkflowInstance) {
	o.instancesMu.Lock()
	o.instances = append(o.instances, wi)
	o.instancesMu.Unlock()
}

func (o *Orchestrator) Wait() {
	log.Printf("Orchestrator.Wait called")
	o.rootWf = o.db.PullNextWorkflow()
	if o.rootWf == nil {
		log.Printf("No root workflow to execute")
		return
	}
	o.rootWf.ctx = o.ctx
	o.rootWf.orchestrator = o
	o.rootWf.Start()

	// Wait for root workflow to complete
	for {
		state := o.rootWf.fsm.MustState()
		log.Printf("Root Workflow FSM state: %s", state)
		if state == StateCompleted || state == StateFailed {
			break
		}
		select {
		case <-o.ctx.Done():
			log.Printf("Context cancelled: %v", o.ctx.Err())
			o.rootWf.fsm.Fire(TriggerFail)
			return
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}

	// Check if rootWf has any error
	if o.rootWf.err != nil {
		fmt.Printf("Root workflow failed with error: %v\n", o.rootWf.err)
	} else {
		fmt.Printf("Root workflow completed successfully with result: %v\n", o.rootWf.result)
	}
}

func (wi *WorkflowInstance) Start() {
	// Initialize the FSM
	wi.fsm = stateless.NewStateMachine(StateIdle)
	wi.fsm.Configure(StateIdle).
		Permit(TriggerStart, StateExecuting)

	wi.fsm.Configure(StateExecuting).
		OnEntry(wi.executeWorkflow).
		PermitReentry(TriggerExecuteSubWorkflow).
		Permit(TriggerComplete, StateCompleted).
		Permit(TriggerFail, StateFailed)

	wi.fsm.Configure(StateCompleted).
		OnEntry(wi.onCompleted)

	wi.fsm.Configure(StateFailed).
		OnEntry(wi.onFailed)

	// Start the FSM
	wi.fsm.Fire(TriggerStart)
}

func (wi *WorkflowInstance) executeWorkflow(_ context.Context, _ ...interface{}) error {
	log.Printf("WorkflowInstance %s executeWorkflow called", wi.stepID)

	handler := wi.handler
	f := handler.Handler

	ctxWorkflow := &WorkflowContext{
		orchestrator: wi.orchestrator,
		ctx:          wi.ctx,
	}

	argsValues := []reflect.Value{reflect.ValueOf(ctxWorkflow)}
	for _, arg := range wi.input {
		argsValues = append(argsValues, reflect.ValueOf(arg))
	}

	log.Printf("Executing workflow: %s with args: %v", handler.HandlerName, wi.input)

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic in workflow: %v", r)
			wi.err = fmt.Errorf("panic: %v", r)
			wi.fsm.Fire(TriggerFail)
		}
	}()

	select {
	case <-wi.ctx.Done():
		log.Printf("Context cancelled in workflow")
		wi.err = wi.ctx.Err()
		wi.fsm.Fire(TriggerFail)
		return nil
	default:
	}

	results := reflect.ValueOf(f).Call(argsValues)

	numOut := len(results)
	if numOut == 0 {
		err := fmt.Errorf("function %s should return at least an error", handler.HandlerName)
		log.Printf("Error: %v", err)
		wi.err = err
		wi.fsm.Fire(TriggerFail)
		return nil
	}

	errInterface := results[numOut-1].Interface()

	if errInterface != nil {
		log.Printf("Workflow returned error: %v", errInterface)
		wi.err = errInterface.(error)
		wi.fsm.Fire(TriggerFail)
	} else {
		if numOut > 1 {
			result := results[0].Interface()
			log.Printf("Workflow returned result: %v", result)
			wi.result = result
		}
		wi.fsm.Fire(TriggerComplete)
	}

	return nil
}

func (wi *WorkflowInstance) onCompleted(_ context.Context, _ ...interface{}) error {
	log.Printf("WorkflowInstance %s onCompleted called", wi.stepID)
	if wi.future != nil {
		wi.future.setResult(wi.result)
	}
	return nil
}

func (wi *WorkflowInstance) onFailed(_ context.Context, _ ...interface{}) error {
	log.Printf("WorkflowInstance %s onFailed called", wi.stepID)
	if wi.future != nil {
		wi.future.setError(wi.err)
	}
	return nil
}

func SubSubSubWorkflow(ctx *WorkflowContext, data int) (int, error) {
	log.Printf("SubSubWorkflow called with data: %d", data)
	select {
	case <-time.After(0):
		// Simulate long processing
	case <-ctx.ctx.Done():
		log.Printf("SubSubWorkflow context cancelled")
		return -1, ctx.ctx.Err()
	}
	result := data * 4
	log.Printf("SubSubWorkflow returning result: %d", result)
	return result, nil
}

func SubSubWorkflow(ctx *WorkflowContext, data int) (int, error) {
	log.Printf("SubSubWorkflow called with data: %d", data)
	select {
	case <-time.After(0):
		// Simulate long processing
	case <-ctx.ctx.Done():
		log.Printf("SubSubWorkflow context cancelled")
		return -1, ctx.ctx.Err()
	}
	result := data * 4
	if err := ctx.Workflow(SubSubSubWorkflow, data+1).Get(&result); err != nil {
		log.Printf("SubSubSubWorkflow encountered error from sub-sub-workflow: %v", err)
		return -1, err
	}
	log.Printf("SubSubWorkflow returning result: %d", result)
	return result, nil
}

func SubWorkflow(ctx *WorkflowContext, data int) (int, error) {
	log.Printf("SubWorkflow called with data: %d", data)
	var result int
	if err := ctx.Workflow(SubSubWorkflow, data*2).Get(&result); err != nil {
		log.Printf("SubWorkflow encountered error from sub-sub-workflow: %v", err)
		return -1, err
	}
	log.Printf("SubWorkflow returning result: %d", result)
	return result, nil
}

func Workflow(ctx *WorkflowContext, data int) (int, error) {
	log.Printf("Workflow called with data: %d", data)
	var value int

	select {
	case <-ctx.ctx.Done():
		log.Printf("Workflow context cancelled")
		return -1, ctx.ctx.Err()
	default:
	}

	if err := ctx.Workflow(SubWorkflow, data).Get(&value); err != nil {
		log.Printf("Workflow encountered error from sub-workflow: %v", err)
		return -1, err
	}
	log.Printf("Workflow received value from sub-workflow: %d", value)

	result := value + data
	log.Printf("Workflow returning result: %d", result)
	return result, nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	log.Printf("main started")
	registry, err := NewRegistryBuilder().
		RegisterWorkflow(Workflow).
		RegisterWorkflow(SubWorkflow).
		RegisterWorkflow(SubSubWorkflow).
		RegisterWorkflow(SubSubSubWorkflow).
		Build()
	if err != nil {
		log.Fatalf("Error building registry: %v", err)
	}

	database := &Database{
		registry: registry,
	}

	err = database.Workflow(Workflow, 40)
	if err != nil {
		log.Fatalf("Error adding workflow to database: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	orchestrator := NewOrchestrator(database, registry, ctx)

	orchestrator.Wait()

	log.Printf("main finished")
}

func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
