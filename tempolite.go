package tempolite

import (
	"context"
	"os"
	"path/filepath"

	dbSQL "database/sql"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"runtime"
	"sync"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/davidroman0O/comfylite3"
	"github.com/davidroman0O/go-tempolite/ent"
	"github.com/davidroman0O/go-tempolite/ent/executioncontext"
	"github.com/davidroman0O/go-tempolite/ent/executionunit"
	"github.com/davidroman0O/go-tempolite/ent/task"
	"github.com/google/uuid"
)

type Tempolite struct {
	ctx    context.Context
	db     *dbSQL.DB
	client *ent.Client

	handlers           sync.Map
	handlerTaskPool    *TaskPool
	sideEffectTaskPool *TaskPool
	sagaTaskPool       *TaskPool

	scheduler *Scheduler
}

type HandlerInfo struct {
	Handler    interface{}
	ParamType  reflect.Type
	ReturnType reflect.Type
	NumIn      int
	NumOut     int
	Type       string // "handler", "side_effect", "saga_transaction", or "saga_compensation"
}

func (hi HandlerInfo) toInterface(data []byte) (interface{}, error) {
	if hi.ParamType == nil {
		return nil, fmt.Errorf("parameter type is nil")
	}
	param := reflect.New(hi.ParamType).Interface()
	if err := json.Unmarshal(data, param); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task payload: %w", err)
	}
	return reflect.ValueOf(param).Elem().Interface(), nil
}

type TempoliteConfig struct {
	Path        string
	Destructive bool
}

type TempoliteOption func(*TempoliteConfig)

func WithPath(path string) TempoliteOption {
	return func(c *TempoliteConfig) {
		c.Path = path
	}
}

func WithDestructive() TempoliteOption {
	return func(c *TempoliteConfig) {
		c.Destructive = true
	}
}

func New(ctx context.Context, opts ...TempoliteOption) (*Tempolite, error) {
	cfg := &TempoliteConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	optsComfy := []comfylite3.ComfyOption{
		comfylite3.WithBuffer(77777),
	}

	var firstTime bool
	if cfg.Path != "" {
		info, err := os.Stat(cfg.Path)
		if err == nil && !info.IsDir() {
			firstTime = false
		} else {
			firstTime = true
		}
		if cfg.Destructive {
			os.Remove(cfg.Path)
		}
		if err := os.MkdirAll(filepath.Dir(cfg.Path), os.ModePerm); err != nil {
			return nil, err
		}
		optsComfy = append(optsComfy, comfylite3.WithPath(cfg.Path))
	} else {
		optsComfy = append(optsComfy, comfylite3.WithMemory())
	}

	comfy, err := comfylite3.New(optsComfy...)
	if err != nil {
		return nil, err
	}

	db := comfylite3.OpenDB(
		comfy,
		comfylite3.WithOption("_fk=1"),
		comfylite3.WithOption("cache=shared"),
		comfylite3.WithOption("mode=rwc"),
		comfylite3.WithForeignKeys(),
	)

	client := ent.NewClient(ent.Driver(sql.OpenDB(dialect.SQLite, db)))

	if firstTime || (cfg.Destructive && cfg.Path != "") {
		if err = client.Schema.Create(ctx); err != nil {
			return nil, err
		}
	}

	t := &Tempolite{
		ctx:    ctx,
		db:     db,
		client: client,
	}

	t.handlerTaskPool = NewTaskPool(t, 5, "handler")
	t.sideEffectTaskPool = NewTaskPool(t, 5, "side_effect")
	t.sagaTaskPool = NewTaskPool(t, 5, "saga")

	t.scheduler = NewScheduler(t)

	return t, nil
}

func (t *Tempolite) Close() error {
	t.scheduler.Close()
	t.handlerTaskPool.pool.Close()
	t.sideEffectTaskPool.pool.Close()
	t.sagaTaskPool.pool.Close()

	if err := t.client.Close(); err != nil {
		return fmt.Errorf("error closing client: %w", err)
	}

	if err := t.db.Close(); err != nil {
		return fmt.Errorf("error closing database: %w", err)
	}

	return nil
}

func (t *Tempolite) RegisterHandler(handler interface{}) error {
	return t.registerFunction(handler, "handler")
}

func (t *Tempolite) RegisterSideEffect(handler interface{}) error {
	return t.registerFunction(handler, "side_effect")
}

func (t *Tempolite) RegisterSagaTransaction(handler interface{}) error {
	return t.registerFunction(handler, "saga_transaction")
}

func (t *Tempolite) RegisterSagaCompensation(handler interface{}) error {
	return t.registerFunction(handler, "saga_compensation")
}

func (t *Tempolite) registerFunction(handler interface{}, handlerKind string) error {
	handlerValue := reflect.ValueOf(handler)
	handlerType := handlerValue.Type()

	if handlerType.Kind() != reflect.Func {
		return fmt.Errorf("%s must be a function", handlerType)
	}

	if handlerType.NumIn() != 2 {
		return fmt.Errorf("%s must have two input parameters", handlerType)
	}

	if !handlerType.In(0).Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
		return fmt.Errorf("first parameter of %s must implement context.Context", handlerType)
	}

	var returnType reflect.Type
	if handlerType.NumOut() == 2 {
		if !handlerType.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			return fmt.Errorf("second return value of %s must be error", handlerType)
		}
		returnType = handlerType.Out(0)
	} else if handlerType.NumOut() == 1 {
		if !handlerType.Out(0).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			return fmt.Errorf("return value of %s must be error", handlerType)
		}
	} else {
		return fmt.Errorf("%s must have either one or two return values", handlerType)
	}

	name := runtime.FuncForPC(handlerValue.Pointer()).Name()
	log.Printf("Registering %s with name %s", handlerType, name)

	t.handlers.Store(name, HandlerInfo{
		Handler:    handler,
		ParamType:  handlerType.In(1),
		ReturnType: returnType,
		NumIn:      handlerType.NumIn(),
		NumOut:     handlerType.NumOut(),
		Type:       handlerKind,
	})

	return nil
}

func (t *Tempolite) getHandler(name string) (HandlerInfo, bool) {
	handler, ok := t.handlers.Load(name)
	if !ok {
		return HandlerInfo{}, false
	}
	return handler.(HandlerInfo), true
}

func (t *Tempolite) Enqueue(ctx context.Context, handler interface{}, params interface{}, options ...EnqueueOption) (string, error) {
	handlerName := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
	log.Printf("Enqueuing task with handler %s", handlerName)

	handlerInfo, exists := t.getHandler(handlerName)
	if !exists {
		return "", fmt.Errorf("no handler registered with name: %s", handlerName)
	}

	opts := &enqueueOptions{
		maxRetries: 3, // Default value
	}
	for _, option := range options {
		option(opts)
	}

	payloadBytes, err := json.Marshal(params)
	if err != nil {
		return "", fmt.Errorf("failed to marshal task parameters: %w", err)
	}

	tx, err := t.client.Tx(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	var execCtx *ent.ExecutionContext
	if opts.executionContextID == "" {
		execCtx, err = tx.ExecutionContext.
			Create().
			SetID(uuid.New().String()).
			SetCurrentRunID(uuid.New().String()).
			SetStatus(executioncontext.StatusRunning).
			SetStartTime(time.Now()).
			Save(ctx)
	} else {
		execCtx, err = tx.ExecutionContext.Get(ctx, opts.executionContextID)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get or create execution context: %w", err)
	}

	execUnit, err := tx.ExecutionUnit.
		Create().
		SetID(uuid.New().String()).
		SetType(executionunit.Type(handlerInfo.Type)).
		SetStatus(executionunit.StatusPending).
		SetStartTime(time.Now()).
		SetMaxRetries(opts.maxRetries).
		SetExecutionContext(execCtx).
		Save(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create execution unit: %w", err)
	}

	if opts.parentID != "" {
		parentExecID := opts.parentID
		execUnit, err = tx.ExecutionUnit.UpdateOneID(execUnit.ID).
			SetParentID(parentExecID).
			Save(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to set parent execution unit: %w", err)
		}
	}

	taskEntity, err := tx.Task.
		Create().
		SetID(uuid.New().String()).
		SetType(task.Type(handlerInfo.Type)).
		SetHandlerName(handlerName).
		SetPayload(payloadBytes).
		SetStatus(task.StatusPending).
		SetCreatedAt(time.Now()).
		SetExecutionUnitID(execUnit.ID).
		Save(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create task: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Dispatch the task to the appropriate pool
	var dispatchErr error
	switch handlerInfo.Type {
	case "handler":
		dispatchErr = t.handlerTaskPool.Dispatch(taskEntity)
	case "side_effect":
		dispatchErr = t.sideEffectTaskPool.Dispatch(taskEntity)
	case "saga_transaction", "saga_compensation":
		dispatchErr = t.sagaTaskPool.Dispatch(taskEntity)
	default:
		return "", fmt.Errorf("unknown handler type: %s", handlerInfo.Type)
	}

	if dispatchErr != nil {
		return "", fmt.Errorf("failed to dispatch task: %w", dispatchErr)
	}

	return taskEntity.ID, nil
}

func (t *Tempolite) WaitFor(ctx context.Context, id string) (interface{}, error) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			// Use a new context for database operations
			dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			taskEntity, err := t.client.Task.Get(dbCtx, id)
			if err != nil {
				if ent.IsNotFound(err) {
					continue
				}
				return nil, fmt.Errorf("failed to get task: %w", err)
			}

			switch taskEntity.Status {
			case task.StatusCompleted:
				handlerInfo, exists := t.getHandler(taskEntity.HandlerName)
				if !exists {
					return nil, fmt.Errorf("no handler registered with name %s", taskEntity.HandlerName)
				}
				return handlerInfo.toInterface(taskEntity.Result)
			case task.StatusFailed:
				// Check if there's a retry task associated with the same ExecutionUnit
				execUnitID, err := taskEntity.QueryExecutionUnit().OnlyID(dbCtx)
				if err != nil {
					return nil, fmt.Errorf("failed to get execution unit ID: %w", err)
				}
				retryTask, err := t.client.Task.Query().
					Where(
						task.HasExecutionUnitWith(executionunit.IDEQ(execUnitID)),
						task.IDNEQ(taskEntity.ID),
						task.CreatedAtGT(taskEntity.CreatedAt),
					).
					Order(ent.Desc(task.FieldCreatedAt)).
					First(dbCtx)
				if err == nil {
					// Found a retry, wait for it
					return t.WaitFor(ctx, retryTask.ID)
				} else if !ent.IsNotFound(err) {
					return nil, fmt.Errorf("failed to query retry task: %w", err)
				}
				// No retry found, return the error
				return nil, fmt.Errorf("task failed: %s", string(taskEntity.Error))

			}
		}
	}
}

type HandlerContext struct {
	context.Context
	tp                 *Tempolite
	executionUnitID    string
	executionContextID string
}

func (c *HandlerContext) Enqueue(handler interface{}, params interface{}, options ...EnqueueOption) (string, error) {
	opts := []EnqueueOption{
		WithParentID(c.executionUnitID),
		WithExecutionContextID(c.executionContextID),
	}
	opts = append(opts, options...)
	return c.tp.Enqueue(c, handler, params, opts...)
}

func (c *HandlerContext) WaitFor(id string) (interface{}, error) {
	return c.tp.WaitFor(c, id)
}

type enqueueOptions struct {
	executionContextID string
	parentID           string
	maxRetries         int
}

type EnqueueOption func(*enqueueOptions)

func WithExecutionContextID(id string) EnqueueOption {
	return func(o *enqueueOptions) {
		o.executionContextID = id
	}
}

func WithParentID(id string) EnqueueOption {
	return func(o *enqueueOptions) {
		o.parentID = id
	}
}

func WithMaxRetries(maxRetries int) EnqueueOption {
	return func(o *enqueueOptions) {
		o.maxRetries = maxRetries
	}
}

func (t *Tempolite) Wait(condition func(TempoliteInfo) bool, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-t.ctx.Done():
			return t.ctx.Err()
		case <-ticker.C:
			info := t.getInfo()
			if condition(info) {
				return nil
			}
		}
	}
}

func (t *Tempolite) getInfo() TempoliteInfo {
	return TempoliteInfo{
		HandlerTasks:    t.handlerTaskPool.QueueSize(),
		SideEffectTasks: t.sideEffectTaskPool.QueueSize(),
		SagaTasks:       t.sagaTaskPool.QueueSize(),
		DeadTasks:       t.handlerTaskPool.DeadTaskCount() + t.sideEffectTaskPool.DeadTaskCount() + t.sagaTaskPool.DeadTaskCount(),
		ProcessingTasks: t.handlerTaskPool.ProcessingCount() + t.sideEffectTaskPool.ProcessingCount() + t.sagaTaskPool.ProcessingCount(),
	}
}

type TempoliteInfo struct {
	HandlerTasks    int
	SideEffectTasks int
	SagaTasks       int
	DeadTasks       int
	ProcessingTasks int
}

func (tpi TempoliteInfo) IsCompleted() bool {
	return tpi.HandlerTasks == 0 && tpi.SideEffectTasks == 0 && tpi.SagaTasks == 0 && tpi.ProcessingTasks == 0
}
