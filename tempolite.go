package tempolite

import (
	"context"
	dbSQL "database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sync"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/davidroman0O/comfylite3"
	"github.com/davidroman0O/go-tempolite/ent"
	"github.com/davidroman0O/go-tempolite/ent/executioncontext"
	"github.com/davidroman0O/go-tempolite/ent/handlerexecution"
	"github.com/davidroman0O/go-tempolite/ent/handlertask"
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

func New(ctx context.Context, opts ...TempoliteOption) (*Tempolite, error) {
	cfg := TempoliteConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}

	optsComfy := []comfylite3.ComfyOption{
		comfylite3.WithBuffer(77777),
	}

	var firstTime bool
	if cfg.path != nil {
		info, err := os.Stat(*cfg.path)
		if err == nil && !info.IsDir() {
			firstTime = false
		} else {
			firstTime = true
		}
		if cfg.destructive {
			os.Remove(*cfg.path)
		}
		if err := os.MkdirAll(filepath.Dir(*cfg.path), os.ModePerm); err != nil {
			return nil, err
		}
		optsComfy = append(optsComfy, comfylite3.WithPath(*cfg.path))
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

	if firstTime || (cfg.destructive && cfg.path != nil) {
		if err = client.Schema.Create(ctx); err != nil {
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

	payloadBytes, err := json.Marshal(params)
	if err != nil {
		return "", fmt.Errorf("failed to marshal task parameters: %v", err)
	}

	tx, err := tp.client.Tx(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	//
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
				// case handlertask.StatusFailed:
				// 	return nil, fmt.Errorf("task failed")
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

func (c *HandlerContext) WaitFor(id string) (interface{}, error) {
	return c.tp.WaitFor(c, id)
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
	param := reflect.New(hi.ParamType).Interface()
	err := json.Unmarshal(data, &param)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal task payload: %v", err)
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
