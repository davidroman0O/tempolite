package tempolite

import (
	"context"
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
	"github.com/davidroman0O/go-tempolite/ent/executionrelationship"
	"github.com/davidroman0O/go-tempolite/ent/run"
	"github.com/davidroman0O/go-tempolite/ent/schema"
	"github.com/davidroman0O/go-tempolite/ent/workflow"
	"github.com/davidroman0O/retrypool"
	"github.com/google/uuid"

	dbSQL "database/sql"
)

type Tempolite struct {
	db     *dbSQL.DB
	client *ent.Client

	workflows    sync.Map
	activities   sync.Map
	sideEffects  sync.Map
	sagas        sync.Map
	sagaBuilders sync.Map

	ctx    context.Context
	cancel context.CancelFunc

	workflowPool *retrypool.Pool[*workflowTask]
	activityPool *retrypool.Pool[*activityTask]

	schedulerExecutionWorkflowDone chan struct{}
	schedulerExecutionActivityDone chan struct{}
}

type tempoliteConfig struct {
	path        *string
	destructive bool
}

type tempoliteOption func(*tempoliteConfig)

func WithPath(path string) tempoliteOption {
	return func(c *tempoliteConfig) {
		c.path = &path
	}
}

func WithMemory() tempoliteOption {
	return func(c *tempoliteConfig) {
		c.path = nil
	}
}

func WithDestructive() tempoliteOption {
	return func(c *tempoliteConfig) {
		c.destructive = true
	}
}

func New(ctx context.Context, opts ...tempoliteOption) (*Tempolite, error) {
	cfg := tempoliteConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}

	ctx, cancel := context.WithCancel(ctx)

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
			cancel()
			return nil, err
		}
		optsComfy = append(optsComfy, comfylite3.WithPath(*cfg.path))
	} else {
		optsComfy = append(optsComfy, comfylite3.WithMemory())
	}

	comfy, err := comfylite3.New(optsComfy...)
	if err != nil {
		cancel()
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
			cancel()
			return nil, err
		}
	}

	tp := &Tempolite{
		ctx:                            ctx,
		db:                             db,
		client:                         client,
		cancel:                         cancel,
		schedulerExecutionWorkflowDone: make(chan struct{}),
		schedulerExecutionActivityDone: make(chan struct{}),
	}

	tp.workflowPool = tp.createWorkflowPool()
	tp.activityPool = tp.createActivityPool()

	go tp.schedulerExeutionWorkflow()
	go tp.schedulerExecutionActivity()

	return tp, nil
}

func (tp *Tempolite) Close() {
	fmt.Println("CANCEL EVERYTHING")
	tp.cancel() // it will stops other systems
	tp.workflowPool.Close()
}

func (tp *Tempolite) Wait() error {
	<-time.After(1 * time.Second)
	return tp.workflowPool.WaitWithCallback(tp.ctx, func(queueSize, processingCount, deadTaskCount int) bool {
		log.Printf("Wait: queueSize: %d, processingCount: %d, deadTaskCount: %d", queueSize, processingCount, deadTaskCount)
		tp.workflowPool.RangeTasks(func(data *workflowTask, workerID int, status retrypool.TaskStatus) bool {
			log.Printf("RangeTask: workerID: %d, status: %v task: %v", workerID, status, data.handlerName)
			return true
		})
		return queueSize > 0 || processingCount > 0 || deadTaskCount > 0
	}, time.Second)
}

func (tp *Tempolite) convertBackResults(handlerInfo HandlerInfo, executionOutput []interface{}) ([]interface{}, error) {
	outputs := []interface{}{}
	// TODO: we can probably parallelize this
	for idx, rawOutput := range executionOutput {
		inputType := handlerInfo.ParamTypes[idx]
		inputKind := handlerInfo.ParamsKinds[idx]
		realInput, err := convertInput(rawOutput, inputType, inputKind)
		if err != nil {
			log.Printf("get: convertInput failed: %v", err)
			return nil, err
		}
		outputs = append(outputs, realInput)
	}
	return outputs, nil
}

// TempoliteContext contains the information from where it was called, so we know the XXXInfo to which it belongs
// Saga only accepts one type of input
func (tp *Tempolite) enqueueSaga(ctx TempoliteContext, input interface{}) (*SagaInfo, error) {
	// todo: implement
	return nil, nil
}

func (tp *Tempolite) enqueueWorkflow(ctx TempoliteContext, input ...interface{}) (*workflowWorker, error) {
	// todo: implement
	return nil, nil
}

func (tp *Tempolite) enqueueSideEffect(ctx TempoliteContext, input ...interface{}) (*SideEffectInfo, error) {
	// todo: implement
	return nil, nil
}

func (tp *Tempolite) getSaga(id string) (*SagaInfo, error) {
	// todo: implement
	return nil, nil
}

func (tp *Tempolite) getSideEffect(id string) (*SideEffectInfo, error) {
	// todo: implement
	return nil, nil
}

func (tp *Tempolite) ProduceSignal(id string) chan interface{} {
	// whatever happen here, we have to create a channel that will then send the data to the other channel used by the consumer ON the correct type!!!
	return make(chan interface{}, 1)
}

func (tp *Tempolite) enqueueSubActivty(ctx TempoliteContext, longName HandlerIdentity, params ...interface{}) (ActivityExecutionID, error) {

	switch ctx.EntityType() {
	case "workflow", "activity":
		// nothing
	default:
		return "", fmt.Errorf("context entity type %s not supported", ctx.EntityType())
	}

	var value any
	var ok bool
	var err error
	var tx *ent.Tx

	if value, ok = tp.activities.Load(longName); ok {
		var activityHandlerInfo Activity
		if activityHandlerInfo, ok = value.(Activity); !ok {
			// could be development bug
			return "", fmt.Errorf("activity %s is not handler info", longName)
		}

		if len(params) != activityHandlerInfo.NumIn {
			return "", fmt.Errorf("parameter count mismatch: expected %d, got %d", activityHandlerInfo.NumIn, len(params))
		}

		for idx, param := range params {
			if reflect.TypeOf(param) != activityHandlerInfo.ParamTypes[idx] {
				return "", fmt.Errorf("parameter type mismatch: expected %s, got %s", activityHandlerInfo.ParamTypes[idx], reflect.TypeOf(param))
			}
		}

		if tx, err = tp.client.Tx(tp.ctx); err != nil {
			return "", err
		}

		var activityEntity *ent.Activity
		if activityEntity, err = tx.
			Activity.
			Create().
			SetID(ctx.RunID()).
			SetIdentity(string(longName)).
			SetHandlerName(activityHandlerInfo.HandlerName).
			SetInput(params).
			SetRetryPolicy(schema.RetryPolicy{
				MaximumAttempts: 1,
			}).
			Save(tp.ctx); err != nil {
			if err = tx.Rollback(); err != nil {
				return "", err
			}
			return "", err
		}

		log.Printf("EnqueueSubActivity - created activity entity %s for %s with params: %v", activityEntity.ID, longName, params)

		// instance of the workflow definition which will be used to create a workflow task and match with the in-memory registry
		var activityExecution *ent.ActivityExecution
		if activityExecution, err = tx.ActivityExecution.
			Create().
			SetID(uuid.NewString()).
			SetRunID(ctx.RunID()).
			SetActivity(activityEntity).
			Save(tp.ctx); err != nil {
			if err = tx.Rollback(); err != nil {
				return "", err
			}
			return "", err
		}

		switch ctx.EntityType() {
		case "workflow":
			if _, err := tx.ExecutionRelationship.Create().SetParentType(executionrelationship.ParentTypeWorkflow).SetChildType(executionrelationship.ChildTypeActivity).SetChildID(activityEntity.ID).SetParentID(ctx.ExecutionID()).Save(tp.ctx); err != nil {
				if err = tx.Rollback(); err != nil {
					return "", err
				}
				return "", err
			}
		case "activity":
			if _, err := tx.ExecutionRelationship.Create().SetParentType(executionrelationship.ParentTypeActivity).SetChildType(executionrelationship.ChildTypeActivity).SetChildID(activityEntity.ID).SetParentID(ctx.ExecutionID()).Save(tp.ctx); err != nil {
				if err = tx.Rollback(); err != nil {
					return "", err
				}
				return "", err
			}
		}

		log.Printf("EnqueueSubActivity - created workflow execution %s for %s with params: %v", activityExecution.ID, longName, params)

		if err = tx.Commit(); err != nil {
			if err = tx.Rollback(); err != nil {
				return "", err
			}
			return "", err
		}

		// when we enqueue a sub-activity, that mean we're within the execution model, so we don't care about the activity entity
		// only the execution
		return ActivityExecutionID(activityExecution.ID), nil

	} else {
		return "", fmt.Errorf("activity %s not found", longName)
	}
}

func (tp *Tempolite) enqueueSubActivtyFunc(ctx TempoliteContext, activityFunc interface{}, params ...interface{}) (ActivityExecutionID, error) {
	funcName := runtime.FuncForPC(reflect.ValueOf(activityFunc).Pointer()).Name()
	handlerIdentity := HandlerIdentity(funcName)
	return tp.enqueueSubActivty(ctx, handlerIdentity, params...)
}

func (tp *Tempolite) EnqueueActivityFunc(activityFunc interface{}, params ...interface{}) (ActivityID, error) {
	funcName := runtime.FuncForPC(reflect.ValueOf(activityFunc).Pointer()).Name()
	handlerIdentity := HandlerIdentity(funcName)
	return tp.EnqueueActivity(handlerIdentity, params...)
}

func (tp *Tempolite) EnqueueActivity(longName HandlerIdentity, params ...interface{}) (ActivityID, error) {
	var value any
	var ok bool
	var err error
	var tx *ent.Tx

	if value, ok = tp.activities.Load(longName); ok {
		var activityHandlerInfo Activity
		if activityHandlerInfo, ok = value.(Activity); !ok {
			// could be development bug
			return "", fmt.Errorf("activity %s is not handler info", longName)
		}

		if len(params) != activityHandlerInfo.NumIn {
			return "", fmt.Errorf("parameter count mismatch: expected %d, got %d", activityHandlerInfo.NumIn, len(params))
		}

		for idx, param := range params {
			if reflect.TypeOf(param) != activityHandlerInfo.ParamTypes[idx] {
				return "", fmt.Errorf("parameter type mismatch: expected %s, got %s", activityHandlerInfo.ParamTypes[idx], reflect.TypeOf(param))
			}
		}

		if tx, err = tp.client.Tx(tp.ctx); err != nil {
			return "", err
		}

		var runEntity *ent.Run
		// root Run entity that anchored the activity entity despite retries
		if runEntity, err = tx.
			Run.
			Create().
			SetID(uuid.NewString()).    // immutable
			SetRunID(uuid.NewString()). // can change if that flow change due to a retry
			SetType(run.TypeActivity).
			Save(tp.ctx); err != nil {
			if err = tx.Rollback(); err != nil {
				return "", err
			}
			return "", err
		}

		log.Printf("EnqueueActivity - created run entity %s for %s with params: %v", runEntity.ID, longName, params)

		var activityEntity *ent.Activity
		if activityEntity, err = tx.
			Activity.
			Create().
			SetID(runEntity.RunID).
			SetIdentity(string(longName)).
			SetHandlerName(activityHandlerInfo.HandlerName).
			SetInput(params).
			SetRetryPolicy(schema.RetryPolicy{
				MaximumAttempts: 1,
			}).
			Save(tp.ctx); err != nil {
			if err = tx.Rollback(); err != nil {
				return "", err
			}
			return "", err
		}

		log.Printf("EnqueueActivity - created activity entity %s for %s with params: %v", activityEntity.ID, longName, params)

		// instance of the workflow definition which will be used to create a workflow task and match with the in-memory registry
		var activityExecution *ent.ActivityExecution
		if activityExecution, err = tx.ActivityExecution.
			Create().
			SetID(uuid.NewString()).
			SetRunID(runEntity.ID).
			SetActivity(activityEntity).
			Save(tp.ctx); err != nil {
			if err = tx.Rollback(); err != nil {
				return "", err
			}
			return "", err
		}

		if _, err = tx.Run.UpdateOneID(runEntity.ID).SetActivity(activityEntity).Save(tp.ctx); err != nil {
			if err = tx.Rollback(); err != nil {
				return "", err
			}
			return "", err
		}

		log.Printf("EnqueueActivity - created activity execution %s for %s with params: %v", activityExecution.ID, longName, params)

		if err = tx.Commit(); err != nil {
			if err = tx.Rollback(); err != nil {
				return "", err
			}
			return "", err
		}

		// we're outside of the execution model, so we care about the activity entity
		return ActivityID(activityEntity.ID), nil
	} else {
		return "", fmt.Errorf("activity %s not found", longName)
	}
}

func (tp *Tempolite) enqueueSubWorkflow(ctx TempoliteContext, workflowFunc interface{}, params ...interface{}) (WorkflowExecutionID, error) {
	switch ctx.EntityType() {
	case "workflow", "activity":
		// nothing
	default:
		return "", fmt.Errorf("context entity type %s not supported", ctx.EntityType())
	}

	funcName := runtime.FuncForPC(reflect.ValueOf(workflowFunc).Pointer()).Name()
	handlerIdentity := HandlerIdentity(funcName)
	var value any
	var ok bool
	var err error
	var tx *ent.Tx

	if value, ok = tp.workflows.Load(handlerIdentity); ok {
		var workflowHandlerInfo Workflow
		if workflowHandlerInfo, ok = value.(Workflow); !ok {
			// could be development bug
			return "", fmt.Errorf("workflow %s is not handler info", handlerIdentity)
		}

		if len(params) != workflowHandlerInfo.NumIn {
			return "", fmt.Errorf("parameter count mismatch: expected %d, got %d", workflowHandlerInfo.NumIn, len(params))
		}

		for idx, param := range params {
			if reflect.TypeOf(param) != workflowHandlerInfo.ParamTypes[idx] {
				return "", fmt.Errorf("parameter type mismatch: expected %s, got %s", workflowHandlerInfo.ParamTypes[idx], reflect.TypeOf(param))
			}
		}

		if tx, err = tp.client.Tx(tp.ctx); err != nil {
			return "", err
		}

		// fmt.Println("===CREATE SUB WORKFLOW", ctx.ExecutionID(), ctx.RunID(), ctx.EntityID())

		//	definition of a workflow, it exists but it is nothing without an execution that will be created as long as it retries
		var workflowEntity *ent.Workflow
		if workflowEntity, err = tx.Workflow.
			Create().
			SetID(uuid.NewString()).
			SetIdentity(string(handlerIdentity)).
			SetHandlerName(workflowHandlerInfo.HandlerName).
			SetInput(params).
			SetRetryPolicy(schema.RetryPolicy{
				MaximumAttempts: 1,
			}).
			Save(tp.ctx); err != nil {
			fmt.Println("ERROR CREATE WORKFLOW", err)
			if err = tx.Rollback(); err != nil {
				return "", err
			}
			return "", err
		}

		log.Printf("EnqueueWorkflow - created workflow entity %s for %s with params: %v", workflowEntity.ID, handlerIdentity, params)

		// instance of the workflow definition which will be used to create a workflow task and match with the in-memory registry
		var workflowExecution *ent.WorkflowExecution
		if workflowExecution, err = tx.WorkflowExecution.
			Create().
			SetID(uuid.NewString()).
			SetRunID(ctx.RunID()).
			SetWorkflow(workflowEntity).
			Save(tp.ctx); err != nil {
			fmt.Println("ERROR CREATE WORKFLOW EXECUTION", err)
			if err = tx.Rollback(); err != nil {
				return "", err
			}
			return "", err
		}

		switch ctx.EntityType() {
		case "workflow":
			if _, err := tx.ExecutionRelationship.Create().SetParentType(executionrelationship.ParentTypeWorkflow).SetChildType(executionrelationship.ChildTypeWorkflow).SetChildID(workflowExecution.ID).SetParentID(ctx.ExecutionID()).Save(tp.ctx); err != nil {
				if err = tx.Rollback(); err != nil {
					return "", err
				}
				return "", err
			}
		case "activity":
			if _, err := tx.ExecutionRelationship.Create().SetParentType(executionrelationship.ParentTypeActivity).SetChildType(executionrelationship.ChildTypeWorkflow).SetChildID(workflowExecution.ID).SetParentID(ctx.ExecutionID()).Save(tp.ctx); err != nil {
				if err = tx.Rollback(); err != nil {
					return "", err
				}
				return "", err
			}
		}

		log.Printf("EnqueueWorkflow - created workflow execution %s for %s with params: %v", workflowExecution.ID, handlerIdentity, params)

		if err = tx.Commit(); err != nil {
			if err = tx.Rollback(); err != nil {
				return "", err
			}
			return "", err
		}

		// when we enqueue a sub-workflow, that mean we're within the execution model, so we don't care about the workflow entity
		// only the execution
		return WorkflowExecutionID(workflowExecution.ID), nil

	} else {
		return "", fmt.Errorf("workflow %s not found", handlerIdentity)
	}
}

func (tp *Tempolite) EnqueueWorkflow(workflowFunc interface{}, params ...interface{}) (WorkflowID, error) {
	funcName := runtime.FuncForPC(reflect.ValueOf(workflowFunc).Pointer()).Name()
	handlerIdentity := HandlerIdentity(funcName)
	var value any
	var ok bool
	var err error
	var tx *ent.Tx

	if value, ok = tp.workflows.Load(handlerIdentity); ok {
		var workflowHandlerInfo Workflow
		if workflowHandlerInfo, ok = value.(Workflow); !ok {
			// could be development bug
			return "", fmt.Errorf("workflow %s is not handler info", handlerIdentity)
		}

		if len(params) != workflowHandlerInfo.NumIn {
			return "", fmt.Errorf("parameter count mismatch: expected %d, got %d", workflowHandlerInfo.NumIn, len(params))
		}

		for idx, param := range params {
			if reflect.TypeOf(param) != workflowHandlerInfo.ParamTypes[idx] {
				return "", fmt.Errorf("parameter type mismatch: expected %s, got %s", workflowHandlerInfo.ParamTypes[idx], reflect.TypeOf(param))
			}
		}

		if tx, err = tp.client.Tx(tp.ctx); err != nil {
			return "", err
		}

		var runEntity *ent.Run
		// root Run entity that anchored the workflow entity despite retries
		if runEntity, err = tx.
			Run.
			Create().
			SetID(uuid.NewString()).    // immutable
			SetRunID(uuid.NewString()). // can change if that flow change due to a retry
			SetType(run.TypeWorkflow).
			Save(tp.ctx); err != nil {
			if err = tx.Rollback(); err != nil {
				return "", err
			}
			return "", err
		}

		log.Printf("EnqueueWorkflow - created run entity %s for %s with params: %v", runEntity.ID, handlerIdentity, params)

		//	definition of a workflow, it exists but it is nothing without an execution that will be created as long as it retries
		var workflowEntity *ent.Workflow
		if workflowEntity, err = tx.Workflow.
			Create().
			SetID(runEntity.RunID).
			SetStatus(workflow.StatusPending).
			SetIdentity(string(handlerIdentity)).
			SetHandlerName(workflowHandlerInfo.HandlerName).
			SetInput(params).
			SetRetryPolicy(schema.RetryPolicy{
				MaximumAttempts: 1,
			}).
			Save(tp.ctx); err != nil {
			if err = tx.Rollback(); err != nil {
				return "", err
			}
			return "", err
		}

		log.Printf("EnqueueWorkflow - created workflow entity %s for %s with params: %v", workflowEntity.ID, handlerIdentity, params)

		// instance of the workflow definition which will be used to create a workflow task and match with the in-memory registry
		var workflowExecution *ent.WorkflowExecution
		if workflowExecution, err = tx.WorkflowExecution.
			Create().
			SetID(uuid.NewString()).
			SetRunID(runEntity.ID).
			SetWorkflow(workflowEntity).
			Save(tp.ctx); err != nil {
			if err = tx.Rollback(); err != nil {
				return "", err
			}
			return "", err
		}

		if _, err = tx.Run.UpdateOneID(runEntity.ID).SetWorkflow(workflowEntity).Save(tp.ctx); err != nil {
			if err = tx.Rollback(); err != nil {
				return "", err
			}
			return "", err
		}

		log.Printf("EnqueueWorkflow - created workflow execution %s for %s with params: %v", workflowExecution.ID, handlerIdentity, params)

		if err = tx.Commit(); err != nil {
			if err = tx.Rollback(); err != nil {
				return "", err
			}
			return "", err
		}

		//	we're outside of the execution model, so we care about the workflow entity
		return WorkflowID(workflowEntity.ID), nil
	} else {
		return "", fmt.Errorf("workflow %s not found", handlerIdentity)
	}
}

func (tp *Tempolite) RemoveWorkflow(id string) (string, error) {
	return "", nil
}

func (tp *Tempolite) PauseWorkflow(id string) (string, error) {
	return "", nil
}

func (tp *Tempolite) ResumeWorkflow(id string) (string, error) {
	return "", nil
}

func (tp *Tempolite) GetSideEffect(id string) (*SideEffectInfo, error) {
	return tp.getSideEffect(id)
}

func (tp *Tempolite) GetSaga(id string) (*SagaInfo, error) {
	return tp.getSaga(id)
}

func (tp *Tempolite) GetWorkflow(id WorkflowID) (*WorkflowInfo, error) {
	log.Printf("GetWorkflow - looking for workflow %s", id)
	info := WorkflowInfo{
		tp:         tp,
		WorkflowID: id,
	}
	return &info, nil
}

func (tp *Tempolite) getWorkflowExecution(ctx TempoliteContext, id WorkflowExecutionID) (*WorkflowExecutionInfo, error) {
	log.Printf("getWorkflowExecution - looking for workflow execution %s", id)
	info := WorkflowExecutionInfo{
		tp:          tp,
		ExecutionID: id,
	}
	return &info, nil
}

func (tp *Tempolite) GetActivity(id ActivityID) (*ActivityInfo, error) {
	log.Printf("GetActivity - looking for activity %s", id)
	info := ActivityInfo{
		tp:         tp,
		ActivityID: id,
	}
	return &info, nil
}

func (tp *Tempolite) getActivity(ctx TempoliteContext, id ActivityExecutionID) (*ActivityExecutionInfo, error) {
	log.Printf("getActivity - looking for activity execution %s", id)
	info := ActivityExecutionInfo{
		tp:          tp,
		ExecutionID: id,
	}
	return &info, nil
}