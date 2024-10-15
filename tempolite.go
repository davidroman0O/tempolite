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
	"github.com/davidroman0O/go-tempolite/ent/activity"
	"github.com/davidroman0O/go-tempolite/ent/executionrelationship"
	"github.com/davidroman0O/go-tempolite/ent/featureflagversion"
	"github.com/davidroman0O/go-tempolite/ent/run"
	"github.com/davidroman0O/go-tempolite/ent/schema"
	"github.com/davidroman0O/go-tempolite/ent/sideeffect"
	"github.com/davidroman0O/go-tempolite/ent/sideeffectexecution"
	"github.com/davidroman0O/go-tempolite/ent/workflow"
	"github.com/davidroman0O/retrypool"
	"github.com/google/uuid"

	dbSQL "database/sql"
)

/// TODO: change the registry functions
/// - register structs with directly a pointer of the struct, we should guarantee will destroy that pointer
/// - I don't want to see that AsXXX functions anymore
///
/// TODO: change enqueue functions, i don't want to see As[T] functions anymore, you either put the function or an instance of the struct
///
/// In documentation warn that struct activities need a particular care since the developer might introduce even more non-deteministic code, it should be used for struct that hold a client/api but not a value what might change the output given the same inputs since it activities won't be replayed if sucessful.

// Trade-off of Tempolite vs Temporal
// Supporting the same similar concepts but now the exact similar features while having less time and resources implies that knowing how Workflows/Activities/Sideffect are behaving in a deterministic and non-deterministic environment, it is crucial
// to assign some kind of identifying value to each of the calls so the whole system can works with minimum overhead and just be reliable and predictable. For now it should works for sqlite, but one day... i want the same with no db.
type Identifier interface {
	~string | ~int
}

// TODO: to be renamed as tempoliteEngine since it will be used to rotate/roll new databases when its database size is too big
type Tempolite[T Identifier] struct {
	db     *dbSQL.DB
	client *ent.Client

	/// TODO: I should make a "register" instance that we can clone to create another instance of Tempolite with the same configuration when the size of the database is too big so we can automatically which new workflows/activities/sideeffects/saga to the new instance

	// Workflows are pre-registered
	// Deterministic function that should be replayed exactly the same way if successful without triggering code since it will have only the results of the previous execution
	workflows                      sync.Map
	workflowPool                   *retrypool.Pool[*workflowTask[T]]
	schedulerExecutionWorkflowDone chan struct{}

	// Activities are pre-registered
	// Un-deterministic function
	activities                     sync.Map
	activityPool                   *retrypool.Pool[*activityTask[T]]
	schedulerExecutionActivityDone chan struct{}

	// SideEffects are dynamically cached
	// You have to use side effects to guide the flow of a workflow
	// Help to manage un-deterministic code within a workflow
	sideEffects             sync.Map
	sideEffectPool          *retrypool.Pool[*sideEffectTask[T]]
	schedulerSideEffectDone chan struct{}

	// Saga are dynamically cached
	// Simplify the management of the Sagas, if we were to use Activities we would be subject to trickery to manage the flow
	sagas                     sync.Map
	transationctionPool       *retrypool.Pool[*transactionTask[T]]
	compensationPool          *retrypool.Pool[*compensationTask[T]]
	schedulerTransactionDone  chan struct{}
	schedulerCompensationDone chan struct{}

	ctx    context.Context
	cancel context.CancelFunc

	// Versioning system cache
	// TODO: we should make an analysis at start to know which versions we could cache
	versionCache sync.Map
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

func New[T Identifier](ctx context.Context, opts ...tempoliteOption) (*Tempolite[T], error) {
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
		firstTime = true
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

	tp := &Tempolite[T]{
		ctx:                            ctx,
		db:                             db,
		client:                         client,
		cancel:                         cancel,
		schedulerExecutionWorkflowDone: make(chan struct{}),
		schedulerExecutionActivityDone: make(chan struct{}),
		schedulerSideEffectDone:        make(chan struct{}),
	}

	tp.workflowPool = tp.createWorkflowPool()
	tp.activityPool = tp.createActivityPool()
	tp.sideEffectPool = tp.createSideEffectPool()

	go tp.schedulerExecutionSideEffect()
	go tp.schedulerExecutionWorkflow()
	go tp.schedulerExecutionActivity()

	return tp, nil
}

func (tp *Tempolite[T]) getOrCreateVersion(workflowType, workflowID, changeID string, minSupported, maxSupported int) (int, error) {
	key := fmt.Sprintf("%s-%s", workflowType, changeID)

	// Check cache first
	if cachedVersion, ok := tp.versionCache.Load(key); ok {
		version := cachedVersion.(int)
		log.Printf("Found cached version %d for key: %s", version, key)
		// Update version if necessary
		if version < maxSupported {
			version = maxSupported
			tp.versionCache.Store(key, version)
			log.Printf("Updated cached version to %d for key: %s", version, key)
		}
		return version, nil
	}

	tx, err := tp.client.Tx(tp.ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	v, err := tx.FeatureFlagVersion.Query().
		Where(featureflagversion.WorkflowTypeEQ(workflowType)).
		Where(featureflagversion.ChangeIDEQ(changeID)).
		Only(tp.ctx)

	if err != nil {
		if !ent.IsNotFound(err) {
			return 0, err
		}
		// Version not found, create a new one
		log.Printf("Creating new version for key: %s with value: %d", key, minSupported)
		v, err = tx.FeatureFlagVersion.Create().
			SetWorkflowType(workflowType).
			SetWorkflowID(workflowID).
			SetChangeID(changeID).
			SetVersion(minSupported).
			Save(tp.ctx)
		if err != nil {
			return 0, err
		}
	} else {
		log.Printf("Found existing version %d for key: %s", v.Version, key)
		// Update the version if maxSupported is greater
		if v.Version < maxSupported {
			v.Version = maxSupported
			v, err = tx.FeatureFlagVersion.UpdateOne(v).
				SetVersion(v.Version).
				SetWorkflowID(workflowID).
				Save(tp.ctx)
			if err != nil {
				return 0, err
			}
			log.Printf("Updated version to %d for key: %s", v.Version, key)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	// Update cache
	tp.versionCache.Store(key, v.Version)
	log.Printf("Stored version %d in cache for key: %s", v.Version, key)

	return v.Version, nil
}

func (tp *Tempolite[T]) Close() {
	tp.cancel() // it will stops other systems
	tp.workflowPool.Close()
	tp.activityPool.Close()
	tp.sideEffectPool.Close()
}

func (tp *Tempolite[T]) Wait() error {
	<-time.After(1 * time.Second)

	activityDone := make(chan error)
	workflowDone := make(chan error)

	doneSignals := []chan error{activityDone, workflowDone}

	go func() {
		defer close(activityDone)
		defer log.Println("activityDone")
		activityDone <- tp.activityPool.WaitWithCallback(tp.ctx, func(queueSize, processingCount, deadTaskCount int) bool {
			log.Printf("Wait: queueSize: %d, processingCount: %d, deadTaskCount: %d", queueSize, processingCount, deadTaskCount)
			tp.activityPool.RangeTasks(func(data *activityTask[T], workerID int, status retrypool.TaskStatus) bool {
				log.Printf("RangeTask: workerID: %d, status: %v task: %v", workerID, status, data.handlerName)
				return true
			})
			return queueSize > 0 || processingCount > 0 || deadTaskCount > 0
		}, time.Second)
	}()

	go func() {
		defer close(workflowDone)
		defer log.Println("workflowDone")
		workflowDone <- tp.workflowPool.WaitWithCallback(tp.ctx, func(queueSize, processingCount, deadTaskCount int) bool {
			log.Printf("Wait: queueSize: %d, processingCount: %d, deadTaskCount: %d", queueSize, processingCount, deadTaskCount)
			tp.workflowPool.RangeTasks(func(data *workflowTask[T], workerID int, status retrypool.TaskStatus) bool {
				log.Printf("RangeTask: workerID: %d, status: %v task: %v", workerID, status, data.handlerName)
				return true
			})
			return queueSize > 0 || processingCount > 0 || deadTaskCount > 0
		}, time.Second)
	}()

	for _, doneSignal := range doneSignals {
		if err := <-doneSignal; err != nil {
			return err
		}
	}

	return nil
}

func (tp *Tempolite[T]) convertInputs(handlerInfo HandlerInfo, executionInputs []interface{}) ([]interface{}, error) {
	outputs := []interface{}{}
	// TODO: we can probably parallelize this
	for idx, rawInputs := range executionInputs {
		inputType := handlerInfo.ParamTypes[idx]
		inputKind := handlerInfo.ParamsKinds[idx]
		// fmt.Println("inputType: ", inputType, "inputKind: ", inputKind, rawInputs)
		realInput, err := convertIO(rawInputs, inputType, inputKind)
		if err != nil {
			log.Printf("get: convertIO failed: %v", err)
			return nil, err
		}
		outputs = append(outputs, realInput)
	}
	return outputs, nil
}

func (tp *Tempolite[T]) convertOuputs(handlerInfo HandlerInfo, executionOutput []interface{}) ([]interface{}, error) {
	outputs := []interface{}{}
	// TODO: we can probably parallelize this
	for idx, rawOutput := range executionOutput {
		ouputType := handlerInfo.ReturnTypes[idx]
		outputKind := handlerInfo.ReturnKinds[idx]
		// fmt.Println("ouputType: ", ouputType, "outputKind: ", outputKind, rawOutput)
		realOutput, err := convertIO(rawOutput, ouputType, outputKind)
		if err != nil {
			log.Printf("get: convertIO failed: %v", err)
			return nil, err
		}
		outputs = append(outputs, realOutput)
	}
	return outputs, nil
}

func verifyHandlerAndParams(handlerInfo HandlerInfo, params []interface{}) error {
	if len(params) != handlerInfo.NumIn {
		return fmt.Errorf("parameter count mismatch (you probably put the wrong handler): expected %d, got %d", handlerInfo.NumIn, len(params))
	}

	for idx, param := range params {
		if reflect.TypeOf(param) != handlerInfo.ParamTypes[idx] {
			return fmt.Errorf("parameter type mismatch (you probably put the wrong handler) at index %d: expected %s, got %s", idx, handlerInfo.ParamTypes[idx], reflect.TypeOf(param))
		}
	}

	return nil
}

func (tp *Tempolite[T]) enqueueSubActivityExecution(ctx WorkflowContext[T], stepID T, longName HandlerIdentity, params ...interface{}) (ActivityID, error) {
	switch ctx.EntityType() {
	case "workflow":
		// nothing
	default:
		return "", fmt.Errorf("context entity type %s not supported", ctx.EntityType())
	}

	var err error

	// Check for existing activity with the same stepID within the current workflow
	exists, err := tp.client.ExecutionRelationship.Query().
		Where(
			executionrelationship.And(
				executionrelationship.RunID(ctx.RunID()),
				executionrelationship.ParentEntityID(ctx.EntityID()),
				executionrelationship.ChildStepID(fmt.Sprint(stepID)),
				executionrelationship.ParentStepID(ctx.StepID()),
			),
			executionrelationship.ChildTypeEQ(executionrelationship.ChildTypeActivity),
		).
		First(tp.ctx)
	if err == nil {
		// todo: is there a way to just get the status?
		act, err := tp.client.Activity.Get(tp.ctx, exists.ChildEntityID)
		if err != nil {
			return "", err
		}
		if act.Status == activity.StatusCompleted {
			return ActivityID(exists.ChildEntityID), nil
		}
	} else if !ent.IsNotFound(err) {
		return "", fmt.Errorf("error checking for existing stepID: %w", err)
	}

	var value any
	var ok bool
	var tx *ent.Tx

	if value, ok = tp.activities.Load(longName); ok {
		var activityHandlerInfo Activity
		if activityHandlerInfo, ok = value.(Activity); !ok {
			// could be development bug
			return "", fmt.Errorf("activity %s is not handler info", longName)
		}

		if err := verifyHandlerAndParams(HandlerInfo(activityHandlerInfo), params); err != nil {
			return "", err
		}

		// Proceed to create a new activity and activity execution
		if tx, err = tp.client.Tx(tp.ctx); err != nil {
			return "", err
		}

		var activityEntity *ent.Activity
		if activityEntity, err = tx.
			Activity.
			Create().
			SetID(uuid.NewString()).
			SetStepID(fmt.Sprint(stepID)).
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

		// Create activity execution with the deterministic ID
		var activityExecution *ent.ActivityExecution
		if activityExecution, err = tx.ActivityExecution.
			Create().
			SetID(activityEntity.ID). // Use the deterministic activity ID
			SetRunID(ctx.RunID()).
			SetActivity(activityEntity).
			Save(tp.ctx); err != nil {
			if err = tx.Rollback(); err != nil {
				return "", err
			}
			return "", err
		}

		// Add execution relationship
		if _, err := tx.ExecutionRelationship.Create().
			SetRunID(ctx.RunID()).
			// entity
			SetParentEntityID(ctx.EntityID()).
			SetChildEntityID(activityEntity.ID).
			// execution
			SetParentID(ctx.ExecutionID()).
			SetChildID(activityEntity.ID).
			// steps
			SetParentStepID(ctx.StepID()).
			SetChildStepID(fmt.Sprint(stepID)).
			// types
			SetParentType(executionrelationship.ParentTypeWorkflow).
			SetChildType(executionrelationship.ChildTypeActivity).
			Save(tp.ctx); err != nil {
			if err = tx.Rollback(); err != nil {
				return "", err
			}
			return "", err
		}

		log.Printf("EnqueueSubActivity - created activity execution %s for %s with params: %v", activityExecution.ID, longName, params)

		if err = tx.Commit(); err != nil {
			if err = tx.Rollback(); err != nil {
				return "", err
			}
			return "", err
		}

		return ActivityID(activityEntity.ID), nil

	} else {
		return "", fmt.Errorf("activity %s not found", longName)
	}
}

func (tp *Tempolite[T]) enqueueSubActivtyFunc(ctx WorkflowContext[T], stepID T, activityFunc interface{}, params ...interface{}) (ActivityID, error) {
	funcName := runtime.FuncForPC(reflect.ValueOf(activityFunc).Pointer()).Name()
	handlerIdentity := HandlerIdentity(funcName)
	return tp.enqueueSubActivityExecution(ctx, stepID, handlerIdentity, params...)
}

func (tp *Tempolite[T]) enqueueSubWorkflow(ctx TempoliteContext, stepID T, workflowFunc interface{}, params ...interface{}) (WorkflowID, error) {
	switch ctx.EntityType() {
	case "workflow":
		// Proceed with sub-workflow creation
	default:
		return "", fmt.Errorf("context entity type %s not supported", ctx.EntityType())
	}

	// Check for existing sub-workflow with the same stepID within the current workflow
	exists, err := tp.client.ExecutionRelationship.Query().
		Where(
			executionrelationship.And(
				executionrelationship.RunID(ctx.RunID()),
				executionrelationship.ParentEntityID(ctx.EntityID()),
				executionrelationship.ChildStepID(fmt.Sprint(stepID)),
				executionrelationship.ParentStepID(ctx.StepID()),
			),
			executionrelationship.ChildTypeEQ(executionrelationship.ChildTypeWorkflow),
		).
		First(tp.ctx)
	if err == nil {
		// todo: is there a way to just get the status?
		act, err := tp.client.Workflow.Get(tp.ctx, exists.ChildEntityID)
		if err != nil {
			return "", err
		}
		if act.Status == workflow.StatusCompleted {
			return WorkflowID(exists.ChildEntityID), nil
		}
	} else {
		if !ent.IsNotFound(err) {
			return "", fmt.Errorf("error checking for existing stepID: %w", err)
		}
	}

	funcName := runtime.FuncForPC(reflect.ValueOf(workflowFunc).Pointer()).Name()
	handlerIdentity := HandlerIdentity(funcName)
	var value any
	var ok bool
	var tx *ent.Tx

	if value, ok = tp.workflows.Load(handlerIdentity); ok {
		var workflowHandlerInfo Workflow
		if workflowHandlerInfo, ok = value.(Workflow); !ok {
			// could be development bug
			return "", fmt.Errorf("workflow %s is not handler info", handlerIdentity)
		}

		if err := verifyHandlerAndParams(HandlerInfo(workflowHandlerInfo), params); err != nil {
			return "", err
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

		//	definition of a workflow, it exists but it is nothing without an execution that will be created as long as it retries
		var workflowEntity *ent.Workflow
		if workflowEntity, err = tx.Workflow.
			Create().
			SetID(uuid.NewString()).
			SetStepID(fmt.Sprint(stepID)).
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

		if _, err := tx.ExecutionRelationship.Create().
			// run id
			SetRunID(ctx.RunID()).
			// entity
			SetParentEntityID(ctx.EntityID()).
			SetChildEntityID(workflowEntity.ID).
			// execution
			SetParentID(ctx.ExecutionID()).
			SetChildID(workflowExecution.ID).
			// Types
			SetParentType(executionrelationship.ParentTypeWorkflow).
			SetChildType(executionrelationship.ChildTypeWorkflow).
			// steps
			SetParentStepID(ctx.StepID()).
			SetChildStepID(fmt.Sprint(stepID)).
			//
			Save(tp.ctx); err != nil {
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

		// when we enqueue a sub-workflow, that mean we're within the execution model, so we don't care about the workflow entity
		// only the execution
		return WorkflowID(workflowEntity.ID), nil

	} else {
		return "", fmt.Errorf("workflow %s not found", handlerIdentity)
	}
}

func (tp *Tempolite[T]) Workflow(stepID T, workflowFunc interface{}, params ...interface{}) *WorkflowInfo[T] {
	id, err := tp.executeWorkflow(stepID, workflowFunc, params...)
	return tp.getWorkflowRoot(id, err)
}

func (tp *Tempolite[T]) executeWorkflow(stepID T, workflowFunc interface{}, params ...interface{}) (WorkflowID, error) {
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

		if err := verifyHandlerAndParams(HandlerInfo(workflowHandlerInfo), params); err != nil {
			return "", err
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
			SetStepID(fmt.Sprint(stepID)).
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

func (tp *Tempolite[T]) GetWorkflow(id WorkflowID) *WorkflowInfo[T] {
	log.Printf("GetWorkflow - looking for workflow %s", id)
	info := WorkflowInfo[T]{
		tp:         tp,
		WorkflowID: id,
	}
	return &info
}

func (tp *Tempolite[T]) getWorkflowRoot(id WorkflowID, err error) *WorkflowInfo[T] {
	log.Printf("getWorkflow - looking for workflow %s", id)
	info := WorkflowInfo[T]{
		tp:         tp,
		WorkflowID: id,
		err:        err,
	}
	return &info
}

func (tp *Tempolite[T]) getWorkflow(ctx TempoliteContext, id WorkflowID, err error) *WorkflowInfo[T] {
	log.Printf("getWorkflow - looking for workflow %s", id)
	info := WorkflowInfo[T]{
		tp:         tp,
		WorkflowID: id,
		err:        err,
	}
	return &info
}

func (tp *Tempolite[T]) getWorkflowExecution(ctx TempoliteContext, id WorkflowExecutionID, err error) *WorkflowExecutionInfo[T] {
	log.Printf("getWorkflowExecution - looking for workflow execution %s", id)
	info := WorkflowExecutionInfo[T]{
		tp:          tp,
		ExecutionID: id,
		err:         err,
	}
	return &info
}

func (tp *Tempolite[T]) GetActivity(id ActivityID) (*ActivityInfo[T], error) {
	log.Printf("GetActivity - looking for activity %s", id)
	info := ActivityInfo[T]{
		tp:         tp,
		ActivityID: id,
	}
	return &info, nil
}

func (tp *Tempolite[T]) getActivity(ctx TempoliteContext, id ActivityID, err error) *ActivityInfo[T] {
	log.Printf("getActivity - looking for activity %s", id)
	info := ActivityInfo[T]{
		tp:         tp,
		ActivityID: id,
		err:        err,
	}
	return &info
}

func (tp *Tempolite[T]) getActivityExecution(ctx TempoliteContext, id ActivityExecutionID, err error) *ActivityExecutionInfo[T] {
	log.Printf("getActivity - looking for activity execution %s", id)
	info := ActivityExecutionInfo[T]{
		tp:          tp,
		ExecutionID: id,
		err:         err,
	}
	return &info
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (tp *Tempolite[T]) GetSideEffect(id SideEffectID) *SideEffectInfo[T] {
	return &SideEffectInfo[T]{
		tp:       tp,
		EntityID: id,
	}
}

func (tp *Tempolite[T]) getSideEffect(ctx TempoliteContext, id SideEffectID, err error) *SideEffectInfo[T] {
	log.Printf("getSideEffect - looking for side effect %s", id)
	info := SideEffectInfo[T]{
		tp:       tp,
		EntityID: id,
		err:      err,
	}
	return &info
}

// That's how we should use it
//
// ```go
// var value int
//
//	err := ctx.SideEffect("eventual switch", func(ctx SideEffectContext[testIdentifier]) int {
//		return 420
//	}).Get(&value)
//
// ```
func (tp *Tempolite[T]) enqueueSubSideEffect(ctx TempoliteContext, stepID T, sideEffectHandler interface{}) (SideEffectID, error) {
	switch ctx.EntityType() {
	case "workflow":
		// Proceed with side effect creation
	default:
		return "", fmt.Errorf("context entity type %s not supported", ctx.EntityType())
	}

	var err error
	var tx *ent.Tx

	// Generate a unique identifier for the side effect function
	funcName := runtime.FuncForPC(reflect.ValueOf(sideEffectHandler).Pointer()).Name()
	handlerIdentity := HandlerIdentity(funcName)

	// Check for existing side effect with the same stepID within the current workflow
	exists, err := tp.client.ExecutionRelationship.Query().
		Where(
			executionrelationship.And(
				executionrelationship.RunID(ctx.RunID()),
				executionrelationship.ParentEntityID(ctx.EntityID()),
				executionrelationship.ChildStepID(fmt.Sprint(stepID)),
				executionrelationship.ParentStepID(ctx.StepID()),
			),
			executionrelationship.ChildTypeEQ(executionrelationship.ChildTypeSideEffect),
		).
		First(tp.ctx)
	if err == nil {
		sideEffectEntity, err := tp.client.SideEffect.Get(tp.ctx, exists.ChildEntityID)
		if err != nil {
			return "", err
		}
		if sideEffectEntity.Status == sideeffect.StatusCompleted {
			return SideEffectID(exists.ChildEntityID), nil
		}
	} else if !ent.IsNotFound(err) {
		return "", fmt.Errorf("error checking for existing stepID: %w", err)
	}

	if tx, err = tp.client.Tx(tp.ctx); err != nil {
		return "", err
	}

	sideEffectEntity, err := tx.SideEffect.
		Create().
		SetID(uuid.NewString()).
		SetStepID(fmt.Sprint(stepID)).
		SetIdentity(string(handlerIdentity)).
		SetHandlerName(funcName).
		SetStatus(sideeffect.StatusPending).
		Save(tp.ctx)
	if err != nil {
		tx.Rollback()
		return "", err
	}

	sideEffectExecution, err := tx.SideEffectExecution.
		Create().
		SetID(uuid.NewString()).
		SetRunID(ctx.RunID()).
		SetSideEffect(sideEffectEntity).
		SetStatus(sideeffectexecution.StatusPending).
		Save(tp.ctx)
	if err != nil {
		tx.Rollback()
		return "", fmt.Errorf("failed to create side effect execution: %w", err)
	}

	_, err = tx.ExecutionRelationship.Create().
		SetRunID(ctx.RunID()).
		SetParentEntityID(ctx.EntityID()).
		SetChildEntityID(sideEffectEntity.ID).
		SetParentID(ctx.ExecutionID()).
		SetChildID(sideEffectExecution.ID).
		SetParentStepID(ctx.StepID()).
		SetChildStepID(fmt.Sprint(stepID)).
		SetParentType(executionrelationship.ParentTypeWorkflow).
		SetChildType(executionrelationship.ChildTypeSideEffect).
		Save(tp.ctx)
	if err != nil {
		tx.Rollback()
		return "", err
	}

	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return "", err
	}

	// Analyze the side effect handler
	handlerType := reflect.TypeOf(sideEffectHandler)
	if handlerType.Kind() != reflect.Func {
		return "", fmt.Errorf("side effect must be a function")
	}

	if handlerType.NumIn() != 1 || handlerType.In(0) != reflect.TypeOf(SideEffectContext[T]{}) {
		return "", fmt.Errorf("side effect function must have exactly one input parameter of type SideEffectContext[T]")
	}

	// Collect all return types
	numOut := handlerType.NumOut()
	if numOut == 0 {
		return "", fmt.Errorf("side effect function must return at least one value")
	}

	returnTypes := make([]reflect.Type, numOut)
	returnKinds := make([]reflect.Kind, numOut)
	for i := 0; i < numOut; i++ {
		returnTypes[i] = handlerType.Out(i)
		returnKinds[i] = handlerType.Out(i).Kind()
	}

	// Cache the side effect info
	tp.sideEffects.Store(sideEffectEntity.ID, SideEffect{
		HandlerName:     funcName,
		HandlerLongName: handlerIdentity,
		Handler:         sideEffectHandler,
		ReturnTypes:     returnTypes,
		ReturnKinds:     returnKinds,
		NumOut:          numOut,
	})

	log.Printf("Enqueued side effect %s with ID %s", funcName, sideEffectEntity.ID)
	return SideEffectID(sideEffectEntity.ID), nil
}

func (tp *Tempolite[T]) enqueueSideEffectFunc(ctx TempoliteContext, stepID T, sideEffect interface{}) (SideEffectID, error) {
	funcName := runtime.FuncForPC(reflect.ValueOf(sideEffect).Pointer()).Name()
	handlerIdentity := HandlerIdentity(funcName)
	return tp.enqueueSubSideEffect(ctx, stepID, handlerIdentity)
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (tp *Tempolite[T]) CancelWorkflow(id WorkflowID) (string, error) {
	return "", nil
}

func (tp *Tempolite[T]) RemoveWorkflow(id WorkflowID) error {
	return nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
/// STAND BY
///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (tp *Tempolite[T]) PauseWorkflow(id WorkflowID) (string, error) {
	return "", nil
}

func (tp *Tempolite[T]) ResumeWorkflow(id WorkflowID) (string, error) {
	return "", nil
}

func (tp *Tempolite[T]) saga(ctx TempoliteContext, stepID T, sagaID SagaID, saga *SagaDefinition[T]) *SagaInfo[T] {
	id, err := tp.enqueueSaga(ctx, stepID, sagaID, saga)
	return tp.getSaga(id, err)
}

// TempoliteContext contains the information from where it was called, so we know the XXXInfo to which it belongs
// Saga only accepts one type of input
func (tp *Tempolite[T]) enqueueSaga(ctx TempoliteContext, stepID T, sagaID SagaID, saga *SagaDefinition[T]) (SagaID, error) {
	// todo: implement
	return sagaID, nil
}

func (tp *Tempolite[T]) getSaga(id SagaID, err error) *SagaInfo[T] {
	// todo: implement
	return &SagaInfo[T]{
		tp:     tp,
		SagaID: id,
		err:    err,
	}
}

func (tp *Tempolite[T]) ProduceSignal(id string) chan interface{} {
	// whatever happen here, we have to create a channel that will then send the data to the other channel used by the consumer ON the correct type!!!
	return make(chan interface{}, 1)
}
