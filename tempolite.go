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
	"github.com/davidroman0O/go-tempolite/ent/entry"
	"github.com/davidroman0O/go-tempolite/ent/handlertask"
	"github.com/davidroman0O/go-tempolite/ent/node"
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

	// You need as much workers as you have have sub-tasks
	// The more you abuse the sub-tasks of sub-tasks, the more workers you need
	t.handlerTaskPool = NewHandlerTaskPool(t, 5)

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
	var retryCount int = 0
	var maxRetry int = 3
	delay := time.Second / 16
	var err error

	var total int
	for total == 0 && retryCount < maxRetry {
		if total, err = tp.client.Entry.Query().Where(entry.TaskID(id)).Count(ctx); err != nil {
			return nil, err
		}
		if total == 0 {
			retryCount++
			time.Sleep(delay)
		}
	}

	if total == 0 {
		return nil, fmt.Errorf("no task with id %s", id)
	}

	var entryValue *ent.Entry
	if entryValue, err = tp.client.Entry.Query().Where(entry.TaskID(id)).First(ctx); err != nil {
		return nil, err
	}

	ticker := time.NewTicker(time.Second / 16)
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
				}
				continue
			case "saga":
			case "compensation":
			case "side_effect":
			}
			return nil, fmt.Errorf("not implemented")
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

	// Before attempting to create a new taskContext or handlerTask, we need to check if we're not creating a duplicate of a retry
	// So if I execited before and succeeded, but right after me my parent failed, then i will be re-executed since I'm not a side effect, therefore i need to re-identify myself eventually.
	// Just to prevent my creation again
	var existingNode *ent.Node

	if opts.nodeID != nil && opts.index != nil {
		log.Printf("Searching for existing node: ParentID=%s, Index=%d", *opts.nodeID, *opts.index)
		existingNode, err = tp.findExistingNode(ctx, tx, *opts.nodeID, *opts.index)
		if err != nil {
			log.Printf("Error finding existing node: %v", err)
			if err := tx.Rollback(); err != nil {
				return "", err
			}
			return "", err
		}
		if existingNode != nil {
			log.Printf("Existing node found: ID=%s, Index=%d", existingNode.ID, existingNode.Index)

			// Update the existing handler task
			handlerTask, err := existingNode.QueryHandlerTask().Only(ctx)
			if err != nil {
				return "", err
			}

			updatedHandlerTask, err := tx.HandlerTask.UpdateOne(handlerTask).
				SetStatus(handlertask.Status(types.TaskStatusToString(types.TaskStatusPending))).
				SetPayload(payloadBytes).
				Save(ctx)
			if err != nil {
				if err := tx.Rollback(); err != nil {
					return "", err
				}
				return "", err
			}

			if err := tx.Commit(); err != nil {
				return "", err
			}
			log.Printf("Existing task updated with ID %s", updatedHandlerTask.ID)
			return updatedHandlerTask.ID, nil
		} else {
			log.Printf("No existing node found for ParentID=%s, Index=%d", *opts.nodeID, *opts.index)
		}
	}

	fmt.Println("====")
	if opts.parentID != nil {
		fmt.Println(*opts.parentID)
	}
	if opts.nodeID != nil {
		fmt.Println(*opts.nodeID)
	}
	if opts.index != nil {
		fmt.Println(*opts.index)
	}
	fmt.Println("====")

	if existingNode != nil {
		log.Printf("Found existing node %s for index %d", existingNode.ID, *opts.index)
		handlerTask, err := existingNode.QueryHandlerTask().Only(ctx)
		if err != nil {
			return "", err
		}

		// Update the existing handler task
		updatedHandlerTask, err := tx.HandlerTask.UpdateOne(handlerTask).
			SetStatus(handlertask.Status(types.TaskStatusToString(types.TaskStatusPending))).
			SetPayload(payloadBytes).
			Save(ctx)
		if err != nil {
			if err := tx.Rollback(); err != nil {
				return "", err
			}
			return "", err
		}

		if err := tx.Commit(); err != nil {
			return "", err
		}
		log.Printf("Existing task updated with ID %s", updatedHandlerTask.ID)
		return updatedHandlerTask.ID, nil
	}

	var executionContext *ent.ExecutionContext

	// if no parent, then we create a new execution context
	if opts.executionContextID == nil {
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
	handlerTaskCreator := tx.
		HandlerTask.
		Create().
		SetID(uuid.NewString()).
		SetHandlerName(handlerName).
		SetStatus(handlertask.Status(types.TaskStatusToString(types.TaskStatusPending))).
		SetPayload(payloadBytes).
		SetNumIn(handkerInfo.NumIn).
		SetNumOut(handkerInfo.NumOut).
		SetTaskContext(taskCtx)

	if opts.executionContextID != nil {
		handlerTaskCreator.SetExecutionContextID(*opts.executionContextID)
	} else {
		handlerTaskCreator.SetExecutionContext(executionContext)
	}

	if handlerTask, err = handlerTaskCreator.Save(ctx); err != nil {
		if err := tx.Rollback(); err != nil {
			return "", err
		}
		return "", err
	}

	var nodeEntity *ent.Node
	nodeCreator := tx.Node.Create().SetID(uuid.NewString()).SetHandlerTask(handlerTask)

	// so we can track who is who
	if opts.index != nil {
		nodeCreator = nodeCreator.SetIndex(*opts.index)
	} else {
		nodeCreator = nodeCreator.SetIndex(0)
	}

	if nodeEntity, err = nodeCreator.Save(ctx); err != nil {
		if err := tx.Rollback(); err != nil {
			return "", err
		}
		return "", err
	}

	if opts.nodeID != nil {
		if err := tp.ensureRelationships(ctx, *opts.nodeID, nodeEntity.ID); err != nil {
			log.Printf("Failed to ensure relationships: %v", err)
			if err := tx.Rollback(); err != nil {
				return "", err
			}
			return "", err
		}
	}

	// Verify relationships
	if err := tp.verifyAndFixNodeRelationships(ctx, nodeEntity.ID); err != nil {
		log.Printf("Failed to verify and fix node relationships: %v", err)
	}

	if opts.nodeID != nil {
		log.Printf("Adding node %s as child of node %s", nodeEntity.ID, *opts.nodeID)
		parentNode, err := tx.Node.Query().Where(node.ID(*opts.nodeID)).Only(ctx)
		if err != nil {
			log.Printf("Parent node not found: %v", err)
			if err := tx.Rollback(); err != nil {
				return "", err
			}
			return "", fmt.Errorf("parent node not found: %v", err)
		}
		if _, err = tx.Node.UpdateOneID(nodeEntity.ID).AddChildren(parentNode).Save(ctx); err != nil {
			if err := tx.Rollback(); err != nil {
				return "", err
			}
			return "", err
		}
	}

	if _, err = tx.HandlerTask.Update().SetNode(nodeEntity).SetNodeID(nodeEntity.ID).Where(handlertask.ID(handlerTask.ID)).Save(ctx); err != nil {
		if err := tx.Rollback(); err != nil {
			return "", err
		}
		return "", err
	}

	var entry *ent.Entry
	entryCreator := tx.Entry.Create().SetTaskID(handlerTask.ID).SetHandlerTaskID(handlerTask.ID).SetType("handler")
	if opts.executionContextID != nil {
		entryCreator.SetExecutionContextID(*opts.executionContextID)
	} else {
		entryCreator.SetExecutionContext(executionContext)
	}
	if entry, err = entryCreator.Save(ctx); err != nil {
		if err := tx.Rollback(); err != nil {
			return "", err
		}
		return "", err
	}

	log.Printf("Task enqueued with ID %s", handlerTask.ID)
	if opts.executionContextID != nil {
		log.Printf("Execution Context created with ID %s", *opts.executionContextID)
	} else {
		log.Printf("Execution Context created with ID %s", executionContext.ID)
	}
	log.Printf("Task Context created with ID %s", taskCtx.ID)
	log.Printf("Node created with ID %s", nodeEntity.ID)
	log.Printf("Entry created with ID %d", entry.ID)

	if err := tx.Commit(); err != nil {
		if err := tx.Rollback(); err != nil {
			return "", err
		}
		return "", err
	}

	return handlerTask.ID, nil
}

func (tp *Tempolite) verifyAndFixNodeRelationships(ctx context.Context, nodeID string) error {
	n, err := tp.client.Node.Query().
		Where(node.ID(nodeID)).
		WithChildren().
		WithParent().
		Only(ctx)
	if err != nil {
		return fmt.Errorf("failed to query node: %v", err)
	}

	log.Printf("Verifying node %s relationships:", nodeID)
	log.Printf("  Parent: %v", n.Edges.Parent)
	log.Printf("  Children count: %d", len(n.Edges.Children))

	// Check if children are properly associated
	for _, child := range n.Edges.Children {
		childNode, err := tp.client.Node.Query().
			Where(node.ID(child.ID)).
			WithParent().
			Only(ctx)
		if err != nil {
			log.Printf("Error querying child node %s: %v", child.ID, err)
			continue
		}
		if childNode.Edges.Parent == nil || childNode.Edges.Parent.ID != nodeID {
			log.Printf("Fixing parent relationship for child %s", child.ID)
			_, err := tp.client.Node.UpdateOne(childNode).
				SetParent(n).
				Save(ctx)
			if err != nil {
				log.Printf("Error fixing parent relationship: %v", err)
			}
		}
	}

	return nil
}

func (tp *Tempolite) queryChildren(ctx context.Context, parentID string) ([]*ent.Node, error) {
	children, err := tp.client.Node.Query().
		Where(node.HasParentWith(node.ID(parentID))).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query children: %v", err)
	}
	return children, nil
}

func (tp *Tempolite) ensureRelationships(ctx context.Context, parentID, childID string) error {
	return tp.client.Node.UpdateOneID(childID).
		SetParentID(parentID).
		Exec(ctx)
}

func (tp *Tempolite) findExistingNode(ctx context.Context, tx *ent.Tx, parentID string, index int) (*ent.Node, error) {
	var query *ent.NodeQuery
	if tx != nil {
		query = tx.Node.Query()
	} else {
		query = tp.client.Node.Query()
	}

	existingNode, err := query.
		Where(
			node.HasParentWith(node.ID(parentID)),
			node.Index(index),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error querying existing node: %v", err)
	}

	return existingNode, nil
}

type HandlerContext struct {
	context.Context
	tp                 *Tempolite
	taskID             string
	nodeID             string
	executionContextID string
	enqueueCounter     int
}

func (c *HandlerContext) GetID() string {
	return c.taskID
}

// TODO FIX: when we enqueue a children and it fails AND retries, we have a second handler_tasks, we need to check the relationships so the final dag is correct
func (c *HandlerContext) Enqueue(handler interface{}, params interface{}, options ...EnqueueOption) (string, error) {
	opts := []EnqueueOption{
		WithParentID(c.taskID),
		WithNodeID(c.nodeID),
		WithExecutionContextID(c.executionContextID),
		WithIndex(c.enqueueCounter),
	}
	opts = append(opts, options...)
	id, err := c.tp.Enqueue(c, handler, params, opts...)
	if err != nil {
		if ent.IsConstraintError(err) {
			log.Printf("Constraint error encountered, attempting to find existing node")
			existingNode, findErr := c.tp.findExistingNode(c, nil, c.nodeID, c.enqueueCounter)
			if findErr == nil && existingNode != nil {
				log.Printf("Found existing node on constraint error: %s", existingNode.ID)
				return existingNode.ID, nil
			}
			log.Printf("Unable to find existing node: %v", findErr)
		}
		return "", err
	}
	c.enqueueCounter++
	return id, nil
}

func (c *HandlerContext) WaitFor(id string) (interface{}, error) {
	return c.tp.WaitFor(c, id)
}
