package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/qmuntal/stateless"
)

const (
	// FSM states
	StateIdle      = "Idle"
	StateExecuting = "Executing"
	StateCompleted = "Completed"
	StateFailed    = "Failed"
	StateRetried   = "Retried"

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
	if result != nil && f.result != nil {
		reflect.ValueOf(result).Elem().Set(reflect.ValueOf(f.result))
		log.Printf("Future.Get returning result: %v", f.result)
	}
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

type RetryPolicy struct {
	MaxAttempts     int
	InitialInterval time.Duration
}

type WorkflowOptions struct {
	RetryPolicy *RetryPolicy
}

type ActivityOptions struct {
	RetryPolicy *RetryPolicy
}

type WorkflowContext struct {
	orchestrator *Orchestrator
	ctx          context.Context
	workflowID   string
}

func (ctx *WorkflowContext) Workflow(stepID string, workflowFunc interface{}, options WorkflowOptions, args ...interface{}) *Future {
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

	// Create a new WorkflowEntity
	entity := &WorkflowEntity{
		ID:     cacheKey,
		Status: StateIdle,
	}

	// Add the entity to the database
	ctx.orchestrator.db.AddWorkflowEntity(entity)

	subWorkflowInstance := &WorkflowInstance{
		stepID:       cacheKey,
		handler:      handler,
		input:        args,
		future:       future,
		ctx:          ctx.ctx,
		orchestrator: ctx.orchestrator,
		workflowID:   cacheKey,
		options:      options,
		entity:       entity,
	}

	ctx.orchestrator.addWorkflowInstance(subWorkflowInstance)
	subWorkflowInstance.Start()

	return future
}

func (ctx *WorkflowContext) Activity(stepID string, activityFunc interface{}, options ActivityOptions, args ...interface{}) *Future {
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

	// Create a new ActivityEntity
	entity := &ActivityEntity{
		ID:     cacheKey,
		Status: StateIdle,
	}

	// Add the entity to the database
	ctx.orchestrator.db.AddActivityEntity(entity)

	activityInstance := &ActivityInstance{
		stepID:       cacheKey,
		handler:      handler,
		input:        args,
		future:       future,
		ctx:          ctx.ctx,
		orchestrator: ctx.orchestrator,
		workflowID:   ctx.workflowID,
		options:      options,
		entity:       entity,
	}

	go activityInstance.executeWithRetry()

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

	// SideEffect has a default retry policy with MaxAttempts=1
	retryPolicy := &RetryPolicy{
		MaxAttempts:     1,
		InitialInterval: 0,
	}

	// Create a new SideEffectEntity
	entity := &SideEffectEntity{
		ID:     cacheKey,
		Status: StateIdle,
	}

	// Add the entity to the database
	ctx.orchestrator.db.AddSideEffectEntity(entity)

	go func() {
		var attempt int
		for attempt = 1; attempt <= retryPolicy.MaxAttempts; attempt++ {
			execution := &SideEffectExecution{
				Attempt: attempt,
				Status:  StateExecuting,
			}
			entity.Executions = append(entity.Executions, execution)

			defer func() {
				if r := recover(); r != nil {
					err := fmt.Errorf("panic in side effect: %v", r)
					log.Printf("Panic in side effect: %v", err)
					execution.Status = StateFailed
					execution.Error = err
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
				execution.Status = StateFailed
				execution.Error = err
				future.setError(err)
				ctx.orchestrator.stopWithError(err)
				return
			}

			result := results[0].Interface()
			log.Printf("Side effect returned result: %v", result)
			future.setResult(result)
			// Cache the result
			ctx.orchestrator.setCache(cacheKey, result)
			execution.Status = StateCompleted
			entity.Status = StateCompleted
			entity.Result = result
			return
		}
		// Max attempts reached
		entity.Status = StateFailed
		future.setError(fmt.Errorf("side effect failed after %d attempts", retryPolicy.MaxAttempts))
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

type WorkflowEntity struct {
	ID         string
	Status     string // e.g., "Completed", "Failed", etc.
	Result     interface{}
	Executions []*WorkflowExecution
}

type WorkflowExecution struct {
	Attempt int
	Status  string // "Executing", "Completed", "Failed", "Retried"
	Error   error
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
	options      WorkflowOptions
	entity       *WorkflowEntity
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
	wi.executeWithRetry()
	return nil
}

func (wi *WorkflowInstance) executeWithRetry() {
	var attempt int
	var maxAttempts int
	var initialInterval time.Duration

	if wi.options.RetryPolicy != nil {
		maxAttempts = wi.options.RetryPolicy.MaxAttempts
		initialInterval = wi.options.RetryPolicy.InitialInterval
	} else {
		// Default retry policy
		maxAttempts = 1
		initialInterval = 0
	}

	for attempt = 1; attempt <= maxAttempts; attempt++ {
		execution := &WorkflowExecution{
			Attempt: attempt,
			Status:  StateExecuting,
		}
		wi.entity.Executions = append(wi.entity.Executions, execution)

		err := wi.runWorkflow(execution)
		if err == nil {
			// Success
			execution.Status = StateCompleted
			wi.entity.Status = StateCompleted
			wi.entity.Result = wi.result
			wi.fsm.Fire(TriggerComplete)
			return
		} else {
			execution.Status = StateFailed
			execution.Error = err
			wi.err = err
			if attempt < maxAttempts {
				execution.Status = StateRetried
				log.Printf("Retrying workflow %s, attempt %d/%d after %v", wi.stepID, attempt+1, maxAttempts, initialInterval)
				time.Sleep(initialInterval)
			} else {
				// Max attempts reached
				wi.entity.Status = StateFailed
				wi.fsm.Fire(TriggerFail)
				return
			}
		}
	}
}

func (wi *WorkflowInstance) runWorkflow(execution *WorkflowExecution) error {
	log.Printf("WorkflowInstance %s runWorkflow attempt %d", wi.stepID, execution.Attempt)

	// Check cache
	if result, ok := wi.orchestrator.getCache(wi.stepID); ok {
		log.Printf("Cache hit for stepID: %s", wi.stepID)
		wi.result = result
		return nil
	}

	// Register workflow on-the-fly
	handler, ok := wi.orchestrator.registry.workflows[wi.handler.HandlerName]
	if !ok {
		err := fmt.Errorf("error getting handler for workflow: %s", wi.handler.HandlerName)
		return err
	}

	f := handler.Handler

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
		}
	}()

	select {
	case <-wi.ctx.Done():
		log.Printf("Context cancelled in workflow")
		wi.err = wi.ctx.Err()
		return wi.err
	default:
	}

	results := reflect.ValueOf(f).Call(argsValues)

	numOut := len(results)
	if numOut == 0 {
		err := fmt.Errorf("function %s should return at least an error", handler.HandlerName)
		log.Printf("Error: %v", err)
		wi.err = err
		return err
	}

	errInterface := results[numOut-1].Interface()

	if errInterface != nil {
		log.Printf("Workflow returned error: %v", errInterface)
		wi.err = errInterface.(error)
		return wi.err
	} else {
		if numOut > 1 {
			result := results[0].Interface()
			log.Printf("Workflow returned result: %v", result)
			wi.result = result
		}
		// Cache the result
		wi.orchestrator.setCache(wi.stepID, wi.result)
		return nil
	}
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

type ActivityEntity struct {
	ID         string
	Status     string // e.g., "Completed", "Failed", etc.
	Result     interface{}
	Executions []*ActivityExecution
}

type ActivityExecution struct {
	Attempt int
	Status  string // "Executing", "Completed", "Failed", "Retried"
	Error   error
}

type ActivityInstance struct {
	stepID       string
	handler      HandlerInfo
	input        []interface{}
	result       interface{}
	err          error
	ctx          context.Context
	orchestrator *Orchestrator
	workflowID   string
	options      ActivityOptions
	entity       *ActivityEntity
	future       *Future
}

func (ai *ActivityInstance) executeWithRetry() {
	var attempt int
	var maxAttempts int
	var initialInterval time.Duration

	if ai.options.RetryPolicy != nil {
		maxAttempts = ai.options.RetryPolicy.MaxAttempts
		initialInterval = ai.options.RetryPolicy.InitialInterval
	} else {
		// Default retry policy
		maxAttempts = 1
		initialInterval = 0
	}

	for attempt = 1; attempt <= maxAttempts; attempt++ {
		execution := &ActivityExecution{
			Attempt: attempt,
			Status:  StateExecuting,
		}
		ai.entity.Executions = append(ai.entity.Executions, execution)

		err := ai.runActivity(execution)
		if err == nil {
			// Success
			execution.Status = StateCompleted
			ai.entity.Status = StateCompleted
			ai.entity.Result = ai.result
			ai.future.setResult(ai.result)
			return
		} else {
			execution.Status = StateFailed
			execution.Error = err
			ai.err = err
			if attempt < maxAttempts {
				execution.Status = StateRetried
				log.Printf("Retrying activity %s, attempt %d/%d after %v", ai.stepID, attempt+1, maxAttempts, initialInterval)
				time.Sleep(initialInterval)
			} else {
				// Max attempts reached
				ai.entity.Status = StateFailed
				ai.future.setError(err)
				return
			}
		}
	}
}

func (ai *ActivityInstance) runActivity(execution *ActivityExecution) error {
	log.Printf("ActivityInstance %s runActivity attempt %d", ai.stepID, execution.Attempt)

	// Check cache
	if result, ok := ai.orchestrator.getCache(ai.stepID); ok {
		log.Printf("Cache hit for stepID: %s", ai.stepID)
		ai.result = result
		return nil
	}

	handler := ai.handler
	f := handler.Handler

	defer func() {
		if r := recover(); r != nil {
			err := fmt.Errorf("panic in activity: %v", r)
			log.Printf("Panic in activity: %v", err)
			ai.err = err
		}
	}()

	argsValues := []reflect.Value{reflect.ValueOf(ActivityContext{ai.ctx})}
	for _, arg := range ai.input {
		argsValues = append(argsValues, reflect.ValueOf(arg))
	}

	results := reflect.ValueOf(f).Call(argsValues)
	numOut := len(results)
	if numOut == 0 {
		err := fmt.Errorf("activity %s should return at least an error", handler.HandlerName)
		log.Printf("Error: %v", err)
		ai.err = err
		return err
	}

	errInterface := results[numOut-1].Interface()
	if errInterface != nil {
		log.Printf("Activity returned error: %v", errInterface)
		ai.err = errInterface.(error)
		return ai.err
	} else {
		var result interface{}
		if numOut > 1 {
			result = results[0].Interface()
			log.Printf("Activity returned result: %v", result)
		}
		ai.result = result
		// Cache the result
		ai.orchestrator.setCache(ai.stepID, result)
		return nil
	}
}

type SideEffectEntity struct {
	ID         string
	Status     string // e.g., "Completed", "Failed", etc.
	Result     interface{}
	Executions []*SideEffectExecution
}

type SideEffectExecution struct {
	Attempt int
	Status  string // "Executing", "Completed", "Failed", "Retried"
	Error   error
}

type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
}

type MapCache struct {
	cache   map[string]interface{}
	cacheMu sync.Mutex
}

func NewMapCache() *MapCache {
	return &MapCache{
		cache: make(map[string]interface{}),
	}
}

func (c *MapCache) Get(key string) (interface{}, bool) {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()
	value, ok := c.cache[key]
	return value, ok
}

func (c *MapCache) Set(key string, value interface{}) {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()
	c.cache[key] = value
}

type Database interface {
	AddWorkflowEntity(entity *WorkflowEntity)
	GetWorkflowEntity(id string) *WorkflowEntity
	AddActivityEntity(entity *ActivityEntity)
	GetActivityEntity(id string) *ActivityEntity
	AddSideEffectEntity(entity *SideEffectEntity)
	GetSideEffectEntity(id string) *SideEffectEntity
}

type DefaultDatabase struct {
	workflows   map[string]*WorkflowEntity
	activities  map[string]*ActivityEntity
	sideEffects map[string]*SideEffectEntity
	mu          sync.Mutex
}

func NewDefaultDatabase() *DefaultDatabase {
	return &DefaultDatabase{
		workflows:   make(map[string]*WorkflowEntity),
		activities:  make(map[string]*ActivityEntity),
		sideEffects: make(map[string]*SideEffectEntity),
	}
}

func (db *DefaultDatabase) AddWorkflowEntity(entity *WorkflowEntity) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.workflows[entity.ID] = entity
}

func (db *DefaultDatabase) GetWorkflowEntity(id string) *WorkflowEntity {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.workflows[id]
}

func (db *DefaultDatabase) AddActivityEntity(entity *ActivityEntity) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.activities[entity.ID] = entity
}

func (db *DefaultDatabase) GetActivityEntity(id string) *ActivityEntity {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.activities[id]
}

func (db *DefaultDatabase) AddSideEffectEntity(entity *SideEffectEntity) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.sideEffects[entity.ID] = entity
}

func (db *DefaultDatabase) GetSideEffectEntity(id string) *SideEffectEntity {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.sideEffects[id]
}

type Orchestrator struct {
	db          Database
	rootWf      *WorkflowInstance
	ctx         context.Context
	instances   []*WorkflowInstance
	instancesMu sync.Mutex
	registry    *Registry
	cache       Cache
	err         error
}

func NewOrchestrator(db Database, cache Cache, ctx context.Context) *Orchestrator {
	log.Printf("NewOrchestrator called")
	o := &Orchestrator{
		db:       db,
		ctx:      ctx,
		registry: newRegistry(),
		cache:    cache,
	}
	return o
}

func (o *Orchestrator) Workflow(workflowFunc interface{}, options WorkflowOptions, args ...interface{}) error {
	// Register the workflow if not already registered
	handler, err := o.registerWorkflow(workflowFunc)
	if err != nil {
		return err
	}

	// Create a new WorkflowEntity
	entity := &WorkflowEntity{
		ID:     "root", // For simplicity, use "root" as ID
		Status: StateIdle,
	}

	// Add the entity to the database
	o.db.AddWorkflowEntity(entity)

	// Create a new WorkflowInstance
	instance := &WorkflowInstance{
		stepID:       "root",
		handler:      handler,
		input:        args,
		ctx:          o.ctx,
		orchestrator: o,
		workflowID:   "root",
		options:      options,
		entity:       entity,
	}

	// Store the root workflow instance
	o.rootWf = instance

	// Start the instance
	instance.Start()

	return nil
}

func (o *Orchestrator) RegisterWorkflow(workflowFunc interface{}) error {
	_, err := o.registerWorkflow(workflowFunc)
	return err
}

func (o *Orchestrator) getCache(key string) (interface{}, bool) {
	return o.cache.Get(key)
}

func (o *Orchestrator) setCache(key string, value interface{}) {
	o.cache.Set(key, value)
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

	if o.rootWf == nil {
		log.Printf("No root workflow to execute")
		return
	}

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

func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

// Example Activity
func SomeActivity(ctx ActivityContext, data int) (int, error) {
	log.Printf("SomeActivity called with data: %d", data)
	select {
	case <-time.After(0):
		// Simulate processing
	case <-ctx.ctx.Done():
		log.Printf("SomeActivity context cancelled")
		return -1, ctx.ctx.Err()
	}
	result := data * 3
	log.Printf("SomeActivity returning result: %d", result)
	return result, nil
}

func SubSubSubWorkflow(ctx *WorkflowContext, data int) (int, error) {
	log.Printf("SubSubSubWorkflow called with data: %d", data)
	var result int
	if err := ctx.Activity("activity-step", SomeActivity, ActivityOptions{
		RetryPolicy: &RetryPolicy{MaxAttempts: 3, InitialInterval: time.Second},
	}, data).Get(&result); err != nil {
		log.Printf("SubSubSubWorkflow encountered error from activity: %v", err)
		return -1, err
	}
	log.Printf("SubSubSubWorkflow returning result: %d", result)
	return result, nil
}

func SubSubWorkflow(ctx *WorkflowContext, data int) (int, error) {
	log.Printf("SubSubWorkflow called with data: %d", data)
	var result int
	if err := ctx.Activity("activity-step", SomeActivity, ActivityOptions{
		RetryPolicy: &RetryPolicy{MaxAttempts: 3, InitialInterval: time.Second},
	}, data).Get(&result); err != nil {
		log.Printf("SubSubWorkflow encountered error from activity: %v", err)
		return -1, err
	}
	if err := ctx.Workflow("subsubsubworkflow-step", SubSubSubWorkflow, WorkflowOptions{
		RetryPolicy: &RetryPolicy{MaxAttempts: 2, InitialInterval: time.Second},
	}, result).Get(&result); err != nil {
		log.Printf("SubSubWorkflow encountered error from sub-sub-workflow: %v", err)
		return -1, err
	}
	log.Printf("SubSubWorkflow returning result: %d", result)
	return result, nil
}

var subWorkflowFailed atomic.Bool

func SubWorkflow(ctx *WorkflowContext, data int) (int, error) {
	log.Printf("SubWorkflow called with data: %d", data)
	var result int
	if err := ctx.Activity("activity-step", SomeActivity, ActivityOptions{
		RetryPolicy: &RetryPolicy{MaxAttempts: 3, InitialInterval: time.Second},
	}, data).Get(&result); err != nil {
		log.Printf("SubWorkflow encountered error from activity: %v", err)
		return -1, err
	}

	if subWorkflowFailed.Load() {
		subWorkflowFailed.Store(false)
		return -1, fmt.Errorf("subworkflow failed on purpose")
	}

	if err := ctx.Workflow("subsubworkflow-step", SubSubWorkflow, WorkflowOptions{
		RetryPolicy: &RetryPolicy{MaxAttempts: 2, InitialInterval: time.Second},
	}, result).Get(&result); err != nil {
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

	var shouldDouble bool
	if err := ctx.SideEffect("side-effect-step", func() bool {
		log.Printf("Side effect called")
		return rand.Float32() < 0.5
	}).Get(&shouldDouble); err != nil {
		log.Printf("Workflow encountered error from side effect: %v", err)
		return -1, err
	}

	if err := ctx.Workflow("subworkflow-step", SubWorkflow, WorkflowOptions{
		RetryPolicy: &RetryPolicy{MaxAttempts: 2, InitialInterval: time.Second},
	}, data).Get(&value); err != nil {
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

	database := NewDefaultDatabase()
	cache := NewMapCache()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	orchestrator := NewOrchestrator(database, cache, ctx)
	orchestrator.RegisterWorkflow(Workflow)

	subWorkflowFailed.Store(true)

	err := orchestrator.Workflow(Workflow, WorkflowOptions{
		RetryPolicy: &RetryPolicy{MaxAttempts: 2, InitialInterval: time.Second},
	}, 40)
	if err != nil {
		log.Fatalf("Error starting workflow: %v", err)
	}

	orchestrator.Wait()

	log.Printf("main finished")
}
