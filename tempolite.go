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
	"github.com/davidroman0O/go-tempolite/ent/activityexecution"
	"github.com/davidroman0O/go-tempolite/ent/executionrelationship"
	"github.com/davidroman0O/go-tempolite/ent/featureflagversion"
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

	workflowPool   *retrypool.Pool[*workflowTask]
	activityPool   *retrypool.Pool[*activityTask]
	sideEffectPool *retrypool.Pool[*sideEffectTask]

	schedulerExecutionWorkflowDone  chan struct{}
	schedulerExecutionActivityDone  chan struct{}
	schedulerSideEffectActivityDone chan struct{}

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

	go tp.schedulerExecutionWorkflow()
	go tp.schedulerExecutionActivity()

	return tp, nil
}

func (tp *Tempolite) getOrCreateVersion(workflowType, workflowID, changeID string, minSupported, maxSupported int) (int, error) {
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

func (tp *Tempolite) Close() {
	tp.cancel() // it will stops other systems
	tp.workflowPool.Close()
}

func (tp *Tempolite) Wait() error {
	<-time.After(1 * time.Second)

	activityDone := make(chan error)
	workflowDone := make(chan error)

	doneSignals := []chan error{activityDone, workflowDone}

	go func() {
		defer close(activityDone)
		defer log.Println("activityDone")
		activityDone <- tp.activityPool.WaitWithCallback(tp.ctx, func(queueSize, processingCount, deadTaskCount int) bool {
			log.Printf("Wait: queueSize: %d, processingCount: %d, deadTaskCount: %d", queueSize, processingCount, deadTaskCount)
			tp.activityPool.RangeTasks(func(data *activityTask, workerID int, status retrypool.TaskStatus) bool {
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
			tp.workflowPool.RangeTasks(func(data *workflowTask, workerID int, status retrypool.TaskStatus) bool {
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

func (tp *Tempolite) convertInputs(handlerInfo HandlerInfo, executionInputs []interface{}) ([]interface{}, error) {
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

func (tp *Tempolite) convertOuputs(handlerInfo HandlerInfo, executionOutput []interface{}) ([]interface{}, error) {
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

func (tp *Tempolite) enqueueSubActivityExecution(ctx WorkflowContext, longName HandlerIdentity, params ...interface{}) (ActivityExecutionID, error) {

	switch ctx.EntityType() {
	case "workflow":
		// nothing
	default:
		return "", fmt.Errorf("context entity type %s not supported", ctx.EntityType())
	}

	var value any
	var ok bool
	// var err error
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

		// Generate a deterministic activity ID using runID
		activityID := ctx.NextActivityID()

		// Check if activity execution already exists and is completed
		existingExec, err := tp.client.ActivityExecution.Query().
			Where(
				activityexecution.IDEQ(activityID),
				activityexecution.StatusEQ(activityexecution.StatusCompleted),
			).Only(tp.ctx)
		if err == nil {
			// Activity execution already exists and is completed
			return ActivityExecutionID(existingExec.ID), nil
		} else if !ent.IsNotFound(err) {
			// Some error occurred
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
			SetID(activityID).
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
			SetID(activityID). // Use the deterministic activity ID
			SetRunID(ctx.RunID()).
			SetActivity(activityEntity).
			Save(tp.ctx); err != nil {
			if err = tx.Rollback(); err != nil {
				return "", err
			}
			return "", err
		}

		// Add execution relationship
		if _, err := tx.ExecutionRelationship.Create().SetParentType(executionrelationship.ParentTypeWorkflow).SetChildType(executionrelationship.ChildTypeActivity).SetChildID(activityEntity.ID).SetParentID(ctx.ExecutionID()).Save(tp.ctx); err != nil {
			if err = tx.Rollback(); err != nil {
				return "", err
			}
			return "", err
		}

		if err = tx.Commit(); err != nil {
			if err = tx.Rollback(); err != nil {
				return "", err
			}
			return "", err
		}

		return ActivityExecutionID(activityExecution.ID), nil

	} else {
		return "", fmt.Errorf("activity %s not found", longName)
	}
}

func (tp *Tempolite) enqueueSubActivtyFunc(ctx WorkflowContext, activityFunc interface{}, params ...interface{}) (ActivityExecutionID, error) {
	funcName := runtime.FuncForPC(reflect.ValueOf(activityFunc).Pointer()).Name()
	handlerIdentity := HandlerIdentity(funcName)
	return tp.enqueueSubActivityExecution(ctx, handlerIdentity, params...)
}

func (tp *Tempolite) enqueueSubWorkflow(ctx TempoliteContext, workflowFunc interface{}, params ...interface{}) (WorkflowExecutionID, error) {
	// only a workflow can have a sub-workflow for determinstic reasons
	switch ctx.EntityType() {
	case "workflow":
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
			// case "activity":
			// 	if _, err := tx.ExecutionRelationship.Create().SetParentType(executionrelationship.ParentTypeActivity).SetChildType(executionrelationship.ChildTypeWorkflow).SetChildID(workflowExecution.ID).SetParentID(ctx.ExecutionID()).Save(tp.ctx); err != nil {
			// 		if err = tx.Rollback(); err != nil {
			// 			return "", err
			// 		}
			// 		return "", err
			// 	}
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

func (tp *Tempolite) Workflow(workflowFunc interface{}, params ...interface{}) (WorkflowID, error) {
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

func (tp *Tempolite) GetWorkflow(id WorkflowID) *WorkflowInfo {
	log.Printf("GetWorkflow - looking for workflow %s", id)
	info := WorkflowInfo{
		tp:         tp,
		WorkflowID: id,
	}
	return &info
}

func (tp *Tempolite) getWorkflowExecution(ctx TempoliteContext, id WorkflowExecutionID, err error) *WorkflowExecutionInfo {
	log.Printf("getWorkflowExecution - looking for workflow execution %s", id)
	info := WorkflowExecutionInfo{
		tp:          tp,
		ExecutionID: id,
		err:         err,
	}
	return &info
}

func (tp *Tempolite) GetActivity(id ActivityID) (*ActivityInfo, error) {
	log.Printf("GetActivity - looking for activity %s", id)
	info := ActivityInfo{
		tp:         tp,
		ActivityID: id,
	}
	return &info, nil
}

func (tp *Tempolite) getActivityExecution(ctx TempoliteContext, id ActivityExecutionID, err error) *ActivityExecutionInfo {
	log.Printf("getActivity - looking for activity execution %s", id)
	info := ActivityExecutionInfo{
		tp:          tp,
		ExecutionID: id,
		err:         err,
	}
	return &info
}

func (tp *Tempolite) GetSideEffect(id string) (*SideEffectInfo, error) {
	return tp.getSideEffect(id)
}

func (tp *Tempolite) getSideEffect(id string) (*SideEffectInfo, error) {
	// todo: implement
	return nil, nil
}

// Side Effects can't fail
func (tp *Tempolite) enqueueSideEffect(ctx TempoliteContext, longName HandlerIdentity) (SideEffectExecutionID, error) {
	return "", nil
}

func (tp *Tempolite) enqueueSideEffectFunc(ctx TempoliteContext, workflowFunc interface{}, input ...interface{}) (*SideEffectInfo, error) {
	// todo: implement
	return nil, nil
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func (tp *Tempolite) GetSaga(id string) (*SagaInfo, error) {
	return tp.getSaga(id)
}

func (tp *Tempolite) CancelWorkflow(id WorkflowID) (string, error) {
	return "", nil
}

func (tp *Tempolite) RemoveWorkflow(id WorkflowID) (string, error) {
	return "", nil
}

func (tp *Tempolite) PauseWorkflow(id WorkflowID) (string, error) {
	return "", nil
}

func (tp *Tempolite) ResumeWorkflow(id WorkflowID) (string, error) {
	return "", nil
}

// TempoliteContext contains the information from where it was called, so we know the XXXInfo to which it belongs
// Saga only accepts one type of input
func (tp *Tempolite) enqueueSaga(ctx TempoliteContext, input interface{}) (*SagaInfo, error) {
	// todo: implement
	return nil, nil
}

func (tp *Tempolite) getSaga(id string) (*SagaInfo, error) {
	// todo: implement
	return nil, nil
}

func (tp *Tempolite) ProduceSignal(id string) chan interface{} {
	// whatever happen here, we have to create a channel that will then send the data to the other channel used by the consumer ON the correct type!!!
	return make(chan interface{}, 1)
}
