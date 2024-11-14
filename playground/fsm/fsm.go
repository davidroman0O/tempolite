package main

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
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
	TriggerStart    = "Start"
	TriggerComplete = "Complete"
	TriggerFail     = "Fail"
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
	workflowID   string
}

func (ctx *WorkflowContext) Workflow(stepID string, workflowFunc interface{}, args ...interface{}) *Future {
	log.Printf("WorkflowContext.Workflow called with stepID: %s, workflowFunc: %v, args: %v", stepID, getFunctionName(workflowFunc), args)
	future := NewFuture()

	// Check cache
	cacheKey := ctx.workflowID + "." + stepID
	if result, ok := ctx.orchestrator.getCache(cacheKey); ok {
		log.Printf("Cache hit for stepID: %s", cacheKey)
		future.setResult(result)
		return future
	}

	// Register workflow on-the-fly
	handler, err := ctx.orchestrator.registerWorkflow(workflowFunc)
	if err != nil {
		log.Printf("Error registering workflow: %v", err)
		future.setError(err)
		ctx.orchestrator.stopWithError(err)
		return future
	}

	fmt.Println("Workflow handler", handler.Handler)

	subWorkflowInstance := &WorkflowInstance{
		stepID:       cacheKey,
		handler:      handler,
		input:        args,
		future:       future,
		ctx:          ctx.ctx,
		orchestrator: ctx.orchestrator,
		workflowID:   cacheKey,
	}

	ctx.orchestrator.addWorkflowInstance(subWorkflowInstance)
	subWorkflowInstance.Start()

	return future
}

func (ctx *WorkflowContext) Activity(stepID string, activityFunc interface{}, args ...interface{}) *Future {
	log.Printf("WorkflowContext.Activity called with stepID: %s, activityFunc: %v, args: %v", stepID, getFunctionName(activityFunc), args)
	future := NewFuture()

	// Check cache
	cacheKey := ctx.workflowID + "." + stepID
	if result, ok := ctx.orchestrator.getCache(cacheKey); ok {
		log.Printf("Cache hit for stepID: %s", cacheKey)
		future.setResult(result)
		return future
	}

	// Register activity on-the-fly
	handler, err := ctx.orchestrator.registerActivity(activityFunc)
	if err != nil {
		log.Printf("Error registering activity: %v", err)
		future.setError(err)
		ctx.orchestrator.stopWithError(err)
		return future
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("panic in activity: %v", r)
				log.Printf("Panic in activity: %v", err)
				future.setError(err)
				ctx.orchestrator.stopWithError(err)
			}
		}()

		argsValues := []reflect.Value{reflect.ValueOf(ActivityContext{ctx.ctx})}
		for _, arg := range args {
			argsValues = append(argsValues, reflect.ValueOf(arg))
		}

		results := reflect.ValueOf(handler.Handler).Call(argsValues)
		numOut := len(results)
		if numOut == 0 {
			err := fmt.Errorf("activity %s should return at least an error", handler.HandlerName)
			log.Printf("Error: %v", err)
			future.setError(err)
			ctx.orchestrator.stopWithError(err)
			return
		}

		errInterface := results[numOut-1].Interface()
		if errInterface != nil {
			log.Printf("Activity returned error: %v", errInterface)
			future.setError(errInterface.(error))
			ctx.orchestrator.stopWithError(errInterface.(error))
		} else {
			var result interface{}
			if numOut > 1 {
				result = results[0].Interface()
				log.Printf("Activity returned result: %v", result)
			}
			future.setResult(result)
			// Cache the result
			ctx.orchestrator.setCache(cacheKey, result)
		}
	}()

	return future
}

func (ctx *WorkflowContext) SideEffect(stepID string, sideEffectFunc interface{}) *Future {
	log.Printf("WorkflowContext.SideEffect called with stepID: %s", stepID)
	future := NewFuture()

	// Check cache
	cacheKey := ctx.workflowID + "." + stepID
	if result, ok := ctx.orchestrator.getCache(cacheKey); ok {
		log.Printf("Cache hit for stepID: %s", cacheKey)
		future.setResult(result)
		return future
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("panic in side effect: %v", r)
				log.Printf("Panic in side effect: %v", err)
				future.setError(err)
				ctx.orchestrator.stopWithError(err)
			}
		}()

		argsValues := []reflect.Value{}
		results := reflect.ValueOf(sideEffectFunc).Call(argsValues)
		numOut := len(results)
		if numOut == 0 {
			err := fmt.Errorf("side effect should return at least a value")
			log.Printf("Error: %v", err)
			future.setError(err)
			ctx.orchestrator.stopWithError(err)
			return
		}

		result := results[0].Interface()
		log.Printf("Side effect returned result: %v", result)
		future.setResult(result)
		// Cache the result
		ctx.orchestrator.setCache(cacheKey, result)
	}()

	return future
}

type ActivityContext struct {
	ctx context.Context
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
	workflowID   string
}

type Orchestrator struct {
	db          *Database
	rootWf      *WorkflowInstance
	ctx         context.Context
	instances   []*WorkflowInstance
	instancesMu sync.Mutex
	registry    *Registry
	cache       map[string]interface{}
	cacheMu     sync.Mutex
	err         error
}

func NewOrchestrator(db *Database, ctx context.Context) *Orchestrator {
	log.Printf("NewOrchestrator called")
	o := &Orchestrator{
		db:       db,
		ctx:      ctx,
		registry: newRegistry(),
		cache:    make(map[string]interface{}),
	}
	return o
}

func (o *Orchestrator) RegistryWorkflow(workflowFunc interface{}) error {
	_, err := o.registerWorkflow(workflowFunc)
	return err
}

func (o *Orchestrator) getCache(key string) (interface{}, bool) {
	o.cacheMu.Lock()
	defer o.cacheMu.Unlock()
	result, ok := o.cache[key]
	return result, ok
}

func (o *Orchestrator) setCache(key string, value interface{}) {
	o.cacheMu.Lock()
	defer o.cacheMu.Unlock()
	o.cache[key] = value
}

func (o *Orchestrator) stopWithError(err error) {
	o.err = err
	if o.rootWf != nil && o.rootWf.fsm != nil {
		o.rootWf.fsm.Fire(TriggerFail)
	}
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
	o.rootWf.workflowID = "root"
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

	// Check cache
	if result, ok := wi.orchestrator.getCache(wi.stepID); ok {
		log.Printf("Cache hit for stepID: %s", wi.stepID)
		wi.result = result
		wi.fsm.Fire(TriggerComplete)
		return nil
	}

	// Register workflow on-the-fly
	handler, ok := wi.orchestrator.registry.workflows[wi.handler.HandlerName]
	if !ok {
		wi.err = fmt.Errorf("error getting handler for workflow: %s", wi.handler.HandlerName)
		wi.fsm.Fire(TriggerFail)
		return nil
	}

	f := handler.Handler
	fmt.Println("handler", handler.Handler)

	ctxWorkflow := &WorkflowContext{
		orchestrator: wi.orchestrator,
		ctx:          wi.ctx,
		workflowID:   wi.workflowID,
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
		// Cache the result
		wi.orchestrator.setCache(wi.stepID, wi.result)
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

type Registry struct {
	workflows  map[string]HandlerInfo
	activities map[string]HandlerInfo
	mu         sync.Mutex
}

func newRegistry() *Registry {
	return &Registry{
		workflows:  make(map[string]HandlerInfo),
		activities: make(map[string]HandlerInfo),
	}
}

func (o *Orchestrator) registerWorkflow(workflowFunc interface{}) (HandlerInfo, error) {
	funcName := getFunctionName(workflowFunc)
	o.registry.mu.Lock()
	defer o.registry.mu.Unlock()

	// Check if already registered
	if handler, ok := o.registry.workflows[funcName]; ok {
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

	expectedContextType := reflect.TypeOf(&WorkflowContext{})
	if handlerType.In(0) != expectedContextType {
		err := fmt.Errorf("first parameter of workflow function must be *WorkflowContext")
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

	o.registry.workflows[funcName] = handler
	return handler, nil
}

func (o *Orchestrator) registerActivity(activityFunc interface{}) (HandlerInfo, error) {
	funcName := getFunctionName(activityFunc)
	o.registry.mu.Lock()
	defer o.registry.mu.Unlock()

	// Check if already registered
	if handler, ok := o.registry.activities[funcName]; ok {
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

	o.registry.activities[funcName] = handler
	return handler, nil
}

type Database struct {
	workflows []*WorkflowInstance
	mu        sync.Mutex
}

func (db *Database) Workflow(workflowFunc interface{}, args ...interface{}) error {
	log.Printf("Database.Workflow called with workflowFunc: %v, args: %v", getFunctionName(workflowFunc), args)
	instance := &WorkflowInstance{
		stepID: "root",
		handler: HandlerInfo{
			HandlerName:     getFunctionName(workflowFunc),
			HandlerLongName: HandlerIdentity(getFunctionName(workflowFunc)),
		}, // Will be set during execution
		input: args,
	}
	db.mu.Lock()
	db.workflows = append(db.workflows, instance)
	db.mu.Unlock()
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

func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

// Example Activity
func SomeActivity(ctx ActivityContext, data int) (int, error) {
	log.Printf("SomeActivity called with data: %d", data)
	select {
	case <-time.After(1 * time.Second):
		// Simulate processing
	case <-ctx.ctx.Done():
		log.Printf("SomeActivity context cancelled")
		return -1, ctx.ctx.Err()
	}
	result := data * 3
	log.Printf("SomeActivity returning result: %d", result)
	return result, nil
}

func SubWorkflow(ctx *WorkflowContext, data int) (int, error) {
	log.Printf("SubWorkflow called with data: %d", data)
	var result int
	if err := ctx.Activity("activity-step", SomeActivity, data).Get(&result); err != nil {
		log.Printf("SubWorkflow encountered error from activity: %v", err)
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

	var shouldDouble bool
	if err := ctx.SideEffect("side-effect-step", func() bool {
		log.Printf("Side effect called")
		return rand.Float32() < 0.5
	}).Get(&shouldDouble); err != nil {
		log.Printf("Workflow encountered error from side effect: %v", err)
		return -1, err
	}

	if err := ctx.Workflow("subworkflow-step", SubWorkflow, data).Get(&value); err != nil {
		log.Printf("Workflow encountered error from sub-workflow: %v", err)
		return -1, err
	}
	log.Printf("Workflow received value from sub-workflow: %d", value)

	result := value + data
	if shouldDouble {
		result *= 2
	}
	log.Printf("Workflow returning result: %d", result)
	return result, nil
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	log.Printf("main started")

	database := &Database{}

	err := database.Workflow(Workflow, 40)
	if err != nil {
		log.Fatalf("Error adding workflow to database: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	orchestrator := NewOrchestrator(database, ctx)
	orchestrator.RegistryWorkflow(Workflow)

	orchestrator.Wait()

	log.Printf("main finished")
}
