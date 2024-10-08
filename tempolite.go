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
	"github.com/davidroman0O/go-tempolite/dag"
	"github.com/davidroman0O/go-tempolite/ent"
	"github.com/davidroman0O/go-tempolite/ent/entry"
	"github.com/davidroman0O/go-tempolite/ent/handlertask"
	"github.com/davidroman0O/go-tempolite/types"
	"github.com/google/uuid"
)

type Tempolite struct {
	ctx      context.Context
	db       *dbSQL.DB
	client   *ent.Client
	handlers sync.Map

	handlerTaskPool *HandlerTaskPool

	scheduler *Scheduler
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
	var err error

	cfg := TempoliteConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}

	optsComfy := []comfylite3.ComfyOption{
		comfylite3.WithBuffer(77777), // TODO: make this configurable
	}

	var firstTime bool
	if cfg.path != nil {
		// check if the file exists before
		info, err := os.Stat(*cfg.path)
		if err == nil && !info.IsDir() {
			firstTime = false
		} else {
			firstTime = true
		}
		if cfg.destructive {
			os.Remove(*cfg.path)
		}
		// we make sure the path exists
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

	// first time or we asked for destruction so we create the schema
	if firstTime || (cfg.destructive && cfg.path != nil) {
		if err = client.Schema.Create(
			ctx,
		); err != nil {
			return nil, err
		}
	}

	t := &Tempolite{
		ctx:    ctx,
		db:     db,
		client: client,
	}

	t.handlerTaskPool = NewHandlerTaskPool(t, 1)

	t.scheduler = NewScheduler(t) // starts immediately

	return t, nil
}

func (tp *Tempolite) Close() error {
	var closeErr error
	tp.scheduler.Close() // first
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

type TempoliteInfo struct {
	Tasks                       int
	SagaTasks                   int
	CompensationTasks           int
	SideEffectTasks             int
	ProcessingTasks             int
	ProcessingSagaTasks         int
	ProcessingCompensationTasks int
	ProcessingSideEffectTasks   int
	DeadTasks                   int
	DeadSagaTasks               int
	DeadCompensationTasks       int
	DeadSideEffectTasks         int
}

func (tpi *TempoliteInfo) IsCompleted() bool {
	return tpi.Tasks == 0 && tpi.SagaTasks == 0 && tpi.CompensationTasks == 0 && tpi.SideEffectTasks == 0 &&
		tpi.ProcessingTasks == 0 && tpi.ProcessingSagaTasks == 0 && tpi.ProcessingCompensationTasks == 0 && tpi.ProcessingSideEffectTasks == 0
}

func (tp *Tempolite) getInfo() TempoliteInfo {
	log.Printf("Getting pool stats")
	return TempoliteInfo{
		Tasks: tp.handlerTaskPool.pool.QueueSize(),
		// SagaTasks:                   tp.sagaHandlerPool.QueueSize(),
		// CompensationTasks:           tp.compensationPool.QueueSize(),
		// SideEffectTasks:             tp.sideEffectPool.QueueSize(),
		DeadTasks: tp.handlerTaskPool.pool.DeadTaskCount(),
		// DeadSagaTasks:               tp.sagaHandlerPool.DeadTaskCount(),
		// DeadCompensationTasks:       tp.compensationPool.DeadTaskCount(),
		// DeadSideEffectTasks:         tp.sideEffectPool.DeadTaskCount(),
		ProcessingTasks: tp.handlerTaskPool.pool.ProcessingCount(),
		// ProcessingSagaTasks:         tp.sagaHandlerPool.ProcessingCount(),
		// ProcessingCompensationTasks: tp.compensationPool.ProcessingCount(),
		// ProcessingSideEffectTasks:   tp.sideEffectPool.ProcessingCount(),
	}
}

func (tp *Tempolite) WaitFor(ctx context.Context, id string) (interface{}, error) {
	var total int
	var err error
	if total, err = tp.client.Entry.Query().Where(entry.TaskID(id)).Count(ctx); err != nil {
		return nil, err
	}
	if total == 0 {
		return nil, fmt.Errorf("no task with id %s", id)
	}
	var entryValue *ent.Entry
	if entryValue, err = tp.client.Entry.Query().Where(entry.TaskID(id)).First(ctx); err != nil {
		return nil, err
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Context done while waiting for ID %s", id)
			return nil, ctx.Err()
		case <-ticker.C:
			switch entryValue.Type {
			case "handler":
				var handlerTaskValue *ent.HandlerTask
				if handlerTaskValue, err = tp.client.HandlerTask.Query().Where(handlertask.ID(id)).First(ctx); err != nil {
					return nil, err
				}
				switch handlerTaskValue.Status {
				case handlertask.StatusCompleted:
					handlerInfo, exists := tp.getHandler(handlerTaskValue.HandlerName)
					if !exists {
						return nil, fmt.Errorf("no handler registered with name %s", handlerTaskValue.HandlerName)
					}
					return handlerInfo.ToInterface(handlerTaskValue.Payload)
				case "saga":
				case "compensation":
				case "side_effect":
				default:
					return nil, fmt.Errorf("unknown task type %s", entryValue.Type)
				}
				return nil, fmt.Errorf("not implemented")
			}
		}
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

func (tp *Tempolite) RegisterHandler(handler interface{}) error {
	handlerType := reflect.TypeOf(handler)
	log.Printf("Registering handler of type %v", handlerType)

	if handlerType.Kind() != reflect.Func {
		return errors.New("Handler must be a function")
	}

	if handlerType.NumIn() != 2 {
		return errors.New("Handler must have two input parameters")
	}

	notInterface := handlerType.In(0).Kind() != reflect.Interface
	gotContext := handlerType.In(0).Implements(reflect.TypeOf((*context.Context)(nil)).Elem())

	if !notInterface || !gotContext {
		return errors.New("First parameter of handler must implement context.Context")
	}

	// notStruct := handlerType.In(1).Kind() != reflect.Struct
	// if notStruct {
	// 	return errors.New("Second parameter of handler must be a struct")
	// }

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
	var err error

	handlerName := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
	log.Printf("Enqueuing task with handler %s", handlerName)

	handkerInfo, exists := tp.getHandler(handlerName)
	if !exists {
		log.Printf("No handler registered with name %s", handlerName)
		return "", fmt.Errorf("no handler registered with name: %s", handlerName)
	}

	opts := enqueueOptions{}
	for _, option := range options {
		option(&opts)
	}

	payloadBytes, err := json.Marshal(params)
	if err != nil {
		log.Printf("Failed to marshal task parameters for handler %s: %v", handlerName, err)
		return "", fmt.Errorf("failed to marshal task parameters: %v", err)
	}

	var tx *ent.Tx

	log.Printf("Creating transaction for enqueuing task with handler %s", handlerName)

	if tx, err = tp.client.Tx(ctx); err != nil {
		return "", err
	}

	var executionContext *ent.ExecutionContext

	if executionContext, err = tx.
		ExecutionContext.
		Create().
		SetID(uuid.NewString()).
		Save(ctx); err != nil {
		if err := tx.Rollback(); err != nil {
			return "", err
		}
		return "", err
	}

	var taskCtx *ent.TaskContext
	if taskCtx, err = tx.
		TaskContext.
		Create().
		SetID(uuid.NewString()).
		SetRetryCount(0).
		SetMaxRetry(1).
		Save(ctx); err != nil {
		if err := tx.Rollback(); err != nil {
			return "", err
		}
		return "", err
	}

	var handlerTask *ent.HandlerTask
	if handlerTask, err = tx.
		HandlerTask.
		Create().
		SetID(uuid.NewString()).
		SetHandlerName(handlerName).
		SetStatus(handlertask.Status(types.TaskStatusToString(types.TaskStatusPending))).
		SetPayload(payloadBytes).
		SetExecutionContext(executionContext).
		SetTaskContext(taskCtx).
		SetNumIn(handkerInfo.NumIn).
		SetNumOut(handkerInfo.NumOut).
		Save(ctx); err != nil {
		if err := tx.Rollback(); err != nil {
			return "", err
		}
		return "", err
	}

	var node *ent.Node
	if node, err = tx.
		Node.
		Create().
		SetID(uuid.NewString()).
		SetHandlerTaskID(handlerTask.ID).
		Save(ctx); err != nil {
		if err := tx.Rollback(); err != nil {
			return "", err
		}
	}

	graph := dag.AcyclicGraph{}
	graph.Add(node)

	var dagBytes []byte
	if dagBytes, err = graph.MarshalJSON(); err != nil {
		if err := tx.Rollback(); err != nil {
			return "", err
		}
		return "", err
	}

	var execution *ent.Execution
	if execution, err = tx.
		Execution.
		Create().
		SetID(uuid.NewString()).
		SetExecutionContext(executionContext).
		SetDag(dagBytes).
		Save(ctx); err != nil {
		if err := tx.Rollback(); err != nil {
			return "", err
		}
		return "", err
	}

	var entry *ent.Entry
	if entry, err = tx.Entry.Create().SetTaskID(handlerTask.ID).SetHandlerTaskID(handlerTask.ID).SetType("handler").Save(ctx); err != nil {
		if err := tx.Rollback(); err != nil {
			return "", err
		}
		return "", err
	}

	log.Printf("Task enqueued with ID %s", handlerTask.ID)
	log.Printf("Execution created with ID %s", execution.ID)
	log.Printf("Execution Context created with ID %s", executionContext.ID)
	log.Printf("Task Context created with ID %s", taskCtx.ID)
	log.Printf("Node created with ID %s", node.ID)
	log.Printf("Entry created with ID %d", entry.ID)

	if err := tx.Commit(); err != nil {
		if err := tx.Rollback(); err != nil {
			return "", err
		}
		return "", err
	}

	return handlerTask.ID, nil
}

type HandlerContext struct {
	context.Context
	tp                 *Tempolite
	taskID             string
	executionContextID string
}

func (c *HandlerContext) GetID() string {
	return c.taskID
}

func (c *HandlerContext) Enqueue(handler interface{}, params interface{}, options ...EnqueueOption) (string, error) {
	return c.tp.Enqueue(c, handler, params)
}
