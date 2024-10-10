package tempolite

import (
	"context"
	dbSQL "database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"runtime"
	"sync"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/davidroman0O/go-tempolite/ent"
	"github.com/davidroman0O/go-tempolite/ent/executioncontext"
	"github.com/davidroman0O/go-tempolite/ent/handlerexecution"
	"github.com/davidroman0O/go-tempolite/ent/handlertask"
	"github.com/davidroman0O/go-tempolite/ent/sideeffectresult"
	"github.com/google/uuid"
)

type Tempolite struct {
	ctx context.Context

	db              *dbSQL.DB
	client          *ent.Client
	handlers        sync.Map
	handlerTaskPool *HandlerTaskPool
	scheduler       *Scheduler
}

type TempoliteConfig struct {
	path        *string
	destructive bool
}

type TempoliteOption func(*TempoliteConfig)

func WithPath(path string) TempoliteOption {
	return func(c *TempoliteConfig) {
		c.path = &path
	}
}

func WithMemory() TempoliteOption {
	return func(c *TempoliteConfig) {
		c.path = nil
	}
}

func WithDestructive() TempoliteOption {
	return func(c *TempoliteConfig) {
		c.destructive = true
	}
}

func New(ctx context.Context, db *dbSQL.DB, opts ...TempoliteOption) (*Tempolite, error) {
	cfg := TempoliteConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}

	client := ent.NewClient(ent.Driver(sql.OpenDB(dialect.SQLite, db)))

	if cfg.destructive {
		if err := client.Schema.Create(ctx); err != nil {
			return nil, err
		}
	}

	t := &Tempolite{
		ctx:    ctx,
		db:     db,
		client: client,
	}

	t.handlerTaskPool = NewHandlerTaskPool(t, 5)
	t.scheduler = NewScheduler(t)

	return t, nil
}

func (tp *Tempolite) Close() error {
	var closeErr error
	tp.scheduler.Close()
	if err := tp.db.Close(); err != nil {
		closeErr = errors.New("error closing database")
	}
	if err := tp.client.Close(); err != nil {
		if closeErr != nil {
			closeErr = fmt.Errorf("%v; %v", closeErr, errors.New("error closing client"))
		} else {
			closeErr = errors.New("error closing client")
		}
	}
	return closeErr
}

func (tp *Tempolite) RegisterHandler(handler interface{}) error {
	handlerType := reflect.TypeOf(handler)
	log.Printf("Registering handler of type %v", handlerType)

	if handlerType.Kind() != reflect.Func {
		return errors.New("Handler must be a function")
	}

	if handlerType.NumIn() != 2 {
		return errors.New("Handler must have two input parameters")
	}

	if !handlerType.In(0).Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
		return errors.New("First parameter of handler must implement context.Context")
	}

	var returnType reflect.Type
	if handlerType.NumOut() == 2 {
		if !handlerType.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			return errors.New("Second return value of handler must be error")
		}
		returnType = handlerType.Out(0)
	} else if handlerType.NumOut() == 1 {
		if !handlerType.Out(0).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			return errors.New("Return value of handler must be error")
		}
	} else {
		return errors.New("Handler must have either one or two return values")
	}

	name := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
	log.Printf("Handler registered with name %s", name)

	tp.setHandler(name, handler, handlerType.In(1), returnType, handlerType.NumIn(), handlerType.NumOut())

	return nil
}

func (t *Tempolite) setHandler(name string, handler interface{}, paramType reflect.Type, returnType reflect.Type, numIn int, numOut int) {
	t.handlers.Store(name, HandlerInfo{
		Handler:    handler,
		ParamType:  paramType,
		ReturnType: returnType,
		NumIn:      numIn,
		NumOut:     numOut,
	})
}

func (t *Tempolite) getHandler(name string) (HandlerInfo, bool) {
	handler, ok := t.handlers.Load(name)
	if !ok {
		return HandlerInfo{}, false
	}
	return handler.(HandlerInfo), true
}

func (tp *Tempolite) Enqueue(ctx context.Context, handler interface{}, params interface{}, options ...EnqueueOption) (string, error) {
	return tp.enqueueTask(ctx, handler, params, "handler", options...)
}

func (tp *Tempolite) EnqueueSaga(ctx context.Context, saga SagaDefinition, params interface{}, options ...EnqueueOption) (string, error) {
	return tp.enqueueSaga(ctx, saga, options...)
}

func (tp *Tempolite) enqueueTask(ctx context.Context, handler interface{}, params interface{}, taskType string, options ...EnqueueOption) (string, error) {
	handlerName := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
	log.Printf("Enqueuing task with handler %s", handlerName)

	_, exists := tp.getHandler(handlerName)
	if !exists {
		return "", fmt.Errorf("no handler registered with name: %s", handlerName)
	}

	opts := enqueueOptions{
		maxRetries: 3, // Default value
	}
	for _, option := range options {
		option(&opts)
	}

	var payloadBytes []byte
	if params != nil {
		var err error
		payloadBytes, err = json.Marshal(params)
		if err != nil {
			return "", fmt.Errorf("failed to marshal task parameters: %v", err)
		}
	}

	tx, err := tp.client.Tx(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var execCtx *ent.ExecutionContext
	if opts.executionContextID == nil {
		execCtx, err = tx.ExecutionContext.
			Create().
			SetID(uuid.New().String()).
			SetCurrentRunID(uuid.New().String()).
			SetStatus(executioncontext.StatusRunning).
			SetStartTime(time.Now()).
			Save(ctx)
	} else {
		execCtx, err = tx.ExecutionContext.Get(ctx, *opts.executionContextID)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get or create execution context: %v", err)
	}

	handlerExec, err := tx.HandlerExecution.
		Create().
		SetID(uuid.New().String()).
		SetRunID(execCtx.CurrentRunID).
		SetHandlerName(handlerName).
		SetStatus(handlerexecution.StatusPending).
		SetStartTime(time.Now()).
		SetMaxRetries(opts.maxRetries).
		SetExecutionContext(execCtx).
		Save(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create handler execution: %v", err)
	}

	if opts.parentID != nil {
		parentExec, err := tx.HandlerExecution.Get(ctx, *opts.parentID)
		if err != nil {
			return "", fmt.Errorf("failed to get parent execution: %v", err)
		}
		_, err = tx.HandlerExecution.UpdateOne(handlerExec).
			SetParent(parentExec).
			Save(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to set parent execution: %v", err)
		}
	}

	handlerTask, err := tx.HandlerTask.
		Create().
		SetID(uuid.New().String()).
		SetTaskType(taskType).
		SetHandlerName(handlerName).
		SetPayload(payloadBytes).
		SetStatus(handlertask.StatusPending).
		SetCreatedAt(time.Now()).
		SetHandlerExecution(handlerExec).
		Save(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create handler task: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %v", err)
	}

	return handlerTask.ID, nil
}

func (tp *Tempolite) enqueueSaga(ctx context.Context, saga SagaDefinition, options ...EnqueueOption) (string, error) {
	opts := enqueueOptions{
		maxRetries: 3, // Default value
	}
	for _, option := range options {
		option(&opts)
	}

	tx, err := tp.client.Tx(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var execCtx *ent.ExecutionContext
	if opts.executionContextID == nil {
		execCtx, err = tx.ExecutionContext.
			Create().
			SetID(uuid.New().String()).
			SetCurrentRunID(uuid.New().String()).
			SetStatus(executioncontext.StatusRunning).
			SetStartTime(time.Now()).
			Save(ctx)
	} else {
		execCtx, err = tx.ExecutionContext.Get(ctx, *opts.executionContextID)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get or create execution context: %v", err)
	}

	sagaExec, err := tx.SagaExecution.
		Create().
		SetID(uuid.New().String()).
		SetExecutionContextID(execCtx.ID).
		SetStatus("running").
		SetStartTime(time.Now()).
		Save(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create saga execution: %v", err)
	}

	// Store the saga steps in the database and in-memory registry
	for i, step := range saga.Steps {
		stepID := uuid.New().String()
		_, err := tx.SagaStepExecution.
			Create().
			SetID(stepID).
			SetSagaExecutionID(sagaExec.ID).
			SetStepNumber(i).
			SetStatus("pending").
			Save(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to create saga step execution: %v", err)
		}

		transactionHandlerName := fmt.Sprintf("%s_transaction_%d", sagaExec.ID, i)
		compensationHandlerName := fmt.Sprintf("%s_compensation_%d", sagaExec.ID, i)
		tp.setHandler(transactionHandlerName, step.Transaction, nil, nil, 0, 0)
		tp.setHandler(compensationHandlerName, step.Compensation, nil, nil, 0, 0)
	}

	// Enqueue the first transaction step
	firstStepHandlerName := fmt.Sprintf("%s_transaction_%d", sagaExec.ID, 0)
	handlerExec, err := tx.HandlerExecution.
		Create().
		SetID(uuid.New().String()).
		SetRunID(execCtx.CurrentRunID).
		SetHandlerName(firstStepHandlerName).
		SetStatus(handlerexecution.StatusPending).
		SetStartTime(time.Now()).
		SetMaxRetries(opts.maxRetries).
		SetExecutionContext(execCtx).
		Save(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create handler execution: %v", err)
	}

	handlerTask, err := tx.HandlerTask.
		Create().
		SetID(uuid.New().String()).
		SetTaskType("transaction").
		SetHandlerName(firstStepHandlerName).
		SetStatus(handlertask.StatusPending).
		SetCreatedAt(time.Now()).
		SetHandlerExecution(handlerExec).
		Save(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create handler task: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %v", err)
	}

	return handlerTask.ID, nil
}

func (tp *Tempolite) WaitFor(ctx context.Context, id string) (interface{}, error) {
	ticker := time.NewTicker(time.Second / 16)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			task, err := tp.client.HandlerTask.Get(ctx, id)
			if err != nil {
				if ent.IsNotFound(err) {
					continue
				}
				return nil, err
			}

			switch task.Status {
			case handlertask.StatusCompleted:
				handlerExec, err := task.QueryHandlerExecution().Only(ctx)
				if err != nil {
					return nil, err
				}
				handlerInfo, exists := tp.getHandler(handlerExec.HandlerName)
				if !exists {
					return nil, fmt.Errorf("no handler registered with name %s", handlerExec.HandlerName)
				}
				return handlerInfo.ToInterface(task.Result)
			}
		}
	}
}

type HandlerContext struct {
	context.Context
	tp                 *Tempolite
	handlerExecutionID string
	executionContextID string
}

func (c *HandlerContext) Enqueue(handler interface{}, params interface{}, options ...EnqueueOption) (string, error) {
	opts := []EnqueueOption{
		WithParentID(c.handlerExecutionID),
		WithExecutionContextID(c.executionContextID),
	}
	opts = append(opts, options...)
	return c.tp.Enqueue(c, handler, params, opts...)
}

func (c *HandlerContext) EnqueueSaga(saga SagaDefinition, options ...EnqueueOption) (string, error) {
	opts := []EnqueueOption{
		WithParentID(c.handlerExecutionID),
		WithExecutionContextID(c.executionContextID),
	}
	opts = append(opts, options...)
	return c.tp.EnqueueSaga(c, saga, opts...)
}

func (c *HandlerContext) SideEffect(name string, sideEffect SideEffect, options ...EnqueueOption) (interface{}, error) {
	// Check if side effect result already exists
	result, err := c.tp.getSideEffectResult(c.executionContextID, name)
	if err == nil {
		return result, nil
	}
	// Enqueue side effect task
	taskID, err := c.tp.enqueueSideEffect(c, sideEffect, name, options...)
	if err != nil {
		return nil, err
	}
	// Wait for side effect to complete
	return c.tp.WaitFor(c, taskID)
}

func (c *HandlerContext) WaitFor(id string) (interface{}, error) {
	return c.tp.WaitFor(c, id)
}

type SideEffectContext struct {
	context.Context
	tp                 *Tempolite
	handlerExecutionID string
	executionContextID string
}

type TransactionContext struct {
	context.Context
	tp                 *Tempolite
	handlerExecutionID string
	executionContextID string
}

type CompensationContext struct {
	context.Context
	tp                 *Tempolite
	handlerExecutionID string
	executionContextID string
}

type HandlerInfo struct {
	Handler    interface{}
	ParamType  reflect.Type
	ReturnType reflect.Type
	NumIn      int
	NumOut     int
}

func (hi HandlerInfo) GetFn() reflect.Value {
	return reflect.ValueOf(hi.Handler)
}

func (hi HandlerInfo) ToInterface(data []byte) (interface{}, error) {
	if hi.ReturnType == nil {
		return nil, nil
	}
	param := reflect.New(hi.ReturnType).Interface()
	err := json.Unmarshal(data, &param)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %v", err)
	}
	return param, nil
}

type EnqueueOption func(*enqueueOptions)

type enqueueOptions struct {
	executionContextID *string
	parentID           *string
	maxRetries         int
}

func WithExecutionContextID(id string) EnqueueOption {
	return func(o *enqueueOptions) {
		o.executionContextID = &id
	}
}

func WithParentID(id string) EnqueueOption {
	return func(o *enqueueOptions) {
		o.parentID = &id
	}
}

func WithMaxRetries(maxRetries int) EnqueueOption {
	return func(o *enqueueOptions) {
		o.maxRetries = maxRetries
	}
}

func (tp *Tempolite) Wait(condition func(TempoliteInfo) bool, interval time.Duration) error {
	log.Printf("Starting wait loop with interval %v", interval)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-tp.ctx.Done():
			log.Printf("Context done during wait loop")
			return tp.ctx.Err()
		case <-ticker.C:
			info := tp.getInfo()
			if condition(info) {
				log.Printf("Wait condition satisfied")
				return nil
			}
		}
	}
}

func (tp *Tempolite) getInfo() TempoliteInfo {
	log.Printf("Getting pool stats")
	return TempoliteInfo{
		Tasks:           tp.handlerTaskPool.pool.QueueSize(),
		DeadTasks:       tp.handlerTaskPool.pool.DeadTaskCount(),
		ProcessingTasks: tp.handlerTaskPool.pool.ProcessingCount(),
	}
}

type TempoliteInfo struct {
	Tasks           int
	DeadTasks       int
	ProcessingTasks int
}

func (tpi TempoliteInfo) IsCompleted() bool {
	return tpi.Tasks == 0 && tpi.ProcessingTasks == 0 && tpi.DeadTasks == 0
}

// Interfaces

type Handler interface {
	Run(ctx HandlerContext, data interface{}) (interface{}, error)
}

type SideEffect interface {
	Run(ctx SideEffectContext) (interface{}, error)
}

type SagaStep interface {
	Transaction(ctx TransactionContext) (interface{}, error)
	Compensation(ctx CompensationContext) (interface{}, error)
}

type SagaDefinition struct {
	Steps []SagaStep
}

// SideEffect related methods

func (tp *Tempolite) getSideEffectResult(executionContextID string, name string) (interface{}, error) {
	// Query SideEffectResult
	res, err := tp.client.SideEffectResult.
		Query().
		Where(
			sideeffectresult.ExecutionContextID(executionContextID),
			sideeffectresult.Name(name),
		).
		Only(tp.ctx)
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = json.Unmarshal(res.Result, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (tp *Tempolite) enqueueSideEffect(ctx context.Context, sideEffect SideEffect, name string, options ...EnqueueOption) (string, error) {
	// Similar to Enqueue, but task_type is "side_effect"
	handlerName := name // Use the name as handlerName for side effect
	// Need to store the side effect function in the handler registry
	// So we can execute it later
	tp.setHandler(handlerName, sideEffect, nil, nil, 0, 0) // paramType and returnType are nil for now

	// Enqueue the side effect task
	return tp.enqueueTask(ctx, sideEffect, nil, "side_effect", options...)
}
