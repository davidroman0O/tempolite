package data

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"github.com/davidroman0O/comfylite3"
	"github.com/davidroman0O/tempolite/ent"
	"github.com/davidroman0O/tempolite/ent/queue"
	"github.com/davidroman0O/tempolite/ent/schema"
	"github.com/davidroman0O/tempolite/ent/version"
	"github.com/davidroman0O/tempolite/ent/workflowdata"
	"github.com/davidroman0O/tempolite/ent/workflowentity"
	"github.com/davidroman0O/tempolite/logger"
	"github.com/sasha-s/go-deadlock"
)

type ActivityContext interface {
	Done() <-chan struct{}
	Err() error
}

type ActivityOptions struct {
	RetryPolicy *schema.RetryPolicy // special retry policy for this activity
}

// Default options means it a sub-workflow directly executed on the same worker/orchestrator
type WorkflowOptions struct {
	Queue       string              // Execute on a specific queue (Cross-Queue on different worker but keep same RunID and got own execution tree)
	Detach      bool                // detach from parent (fire-and-forget on another Run and got own execution tree) -- legacy is DeferExecution
	RetryPolicy *schema.RetryPolicy // special retry policy for this workflow
	// VersionOverrides map[string]int      // override feature flag versions
}

type TransactionContext interface{}
type CompensationContext interface{}

type WorkflowContext interface {
	// Get feature change by version range
	GetVersion(changeID schema.VersionChange, minSupported, maxSupported int) (int, error)

	// Continue workflow under the same workflow entity and same run ID within the same execution tree
	ContinueAsNew(ctx WorkflowContext, args ...interface{}) error

	Activity(stepID schema.ActivityStepID, activity interface{}, args ...interface{}) error
	ActivityWithOptions(stepID schema.ActivityStepID, activity interface{}, options ActivityOptions, args ...interface{}) error

	SideEffect(stepID schema.SideEffectStepID, sideEffect interface{}, args ...interface{}) error

	Saga(stepID schema.SagaStepID, builder *SagaDefinitionBuilder) error
	CompensateSaga(stepID schema.SagaStepID) error

	Workflow(stepID schema.WorkflowStepID, workflow interface{}, args ...interface{}) error
	WorkflowWithOptions(stepID schema.WorkflowStepID, workflow interface{}, options WorkflowOptions, args ...interface{}) error

	Signal(stepID schema.SignalStepID, signal interface{}, args ...interface{}) error
}

func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

// Handler types
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

type Data struct {
	context.Context
	mu         deadlock.RWMutex
	config     *DataConfig
	comfy      *comfylite3.ComfyDB
	client     *ent.Client
	workflows  map[string]HandlerInfo
	activities map[string]HandlerInfo
}

type DataConfig struct {
	memory   bool
	filePath string

	logger logger.Logger
}

type DataOption func(*DataConfig) error

func WithMemory() DataOption {
	return func(c *DataConfig) error {
		c.memory = true
		return nil
	}
}

func WithFilePath(filePath string) DataOption {
	return func(c *DataConfig) error {
		c.filePath = filePath
		return nil
	}
}

func WithLogger(logger logger.Logger) DataOption {
	return func(c *DataConfig) error {
		c.logger = logger
		return nil
	}
}

func New(ctx context.Context, opt ...DataOption) (*Data, error) {

	config := &DataConfig{}
	for _, o := range opt {
		if err := o(config); err != nil {
			return nil, err
		}
	}

	if config.logger == nil {
		config.logger = logger.NewDefaultLogger(slog.LevelDebug, logger.TextFormat)
	}
	var err error
	var comfy *comfylite3.ComfyDB

	var comfyOptions []comfylite3.ComfyOption = []comfylite3.ComfyOption{}
	if config.memory {
		comfyOptions = append(comfyOptions, comfylite3.WithMemory())
	} else {
		comfyOptions = append(comfyOptions, comfylite3.WithPath(config.filePath))
		// Ensure folder exists
		if err := os.MkdirAll(filepath.Dir(config.filePath), 0755); err != nil {
			return nil, err
		}
	}

	// Create a new ComfyDB instance
	if comfy, err = comfylite3.New(
		comfyOptions...,
	); err != nil {
		return nil, err
	}

	// Use the OpenDB function to create a sql.DB instance with SQLite options
	db := comfylite3.OpenDB(
		comfy,
		comfylite3.WithOption("_fk=1"),
		comfylite3.WithOption("cache=shared"),
		comfylite3.WithOption("mode=rwc"),
		comfylite3.WithForeignKeys(),
	)

	// Create a new ent client
	client := ent.NewClient(ent.Driver(sql.OpenDB(dialect.SQLite, db)))

	// Run the auto migration tool
	if err = client.Schema.Create(ctx); err != nil {
		return nil, err
	}

	// if not exists, create default queue "default"
	// check first
	if _, err = client.Queue.Query().Where(queue.Name(schema.DefaultQueue)).Only(ctx); err != nil {
		if _, err = client.Queue.Create().SetName(schema.DefaultQueue).Save(ctx); err != nil {
			return nil, err
		}
	}

	return &Data{
		Context:    ctx,
		config:     config,
		client:     client,
		comfy:      comfy,
		workflows:  map[string]HandlerInfo{},
		activities: map[string]HandlerInfo{},
	}, nil
}

func (d *Data) RegisterWorkflow(workflowFunc interface{}) (HandlerInfo, error) {
	funcName := GetFunctionName(workflowFunc)
	d.mu.Lock()
	defer d.mu.Unlock()

	// Check if already registered
	if handler, ok := d.workflows[funcName]; ok {
		return handler, nil
	}

	handlerType := reflect.TypeOf(workflowFunc)
	if handlerType.Kind() != reflect.Func {
		err := fmt.Errorf("workflow must be a function")
		return HandlerInfo{}, err
	}

	if handlerType.NumIn() < 1 {
		err := fmt.Errorf("workflow must have at least one parameter")
		return HandlerInfo{}, err
	}

	if handlerType.NumOut() < 1 {
		err := fmt.Errorf("workflow must have at least one return value")
		return HandlerInfo{}, err
	}

	handler := HandlerInfo{
		HandlerName:     funcName,
		HandlerLongName: HandlerIdentity(funcName),
		Handler:         workflowFunc,
		ParamsKinds:     make([]reflect.Kind, handlerType.NumIn()),
		ParamTypes:      make([]reflect.Type, handlerType.NumIn()),
		ReturnTypes:     make([]reflect.Type, handlerType.NumOut()),
		ReturnKinds:     make([]reflect.Kind, handlerType.NumOut()),
		NumIn:           handlerType.NumIn(),
		NumOut:          handlerType.NumOut(),
	}

	for i := 0; i < handlerType.NumIn(); i++ {
		handler.ParamsKinds[i] = handlerType.In(i).Kind()
		handler.ParamTypes[i] = handlerType.In(i)
	}

	for i := 0; i < handlerType.NumOut(); i++ {
		handler.ReturnKinds[i] = handlerType.Out(i).Kind()
		handler.ReturnTypes[i] = handlerType.Out(i)
	}

	d.workflows[funcName] = handler

	return handler, nil
}

func (r *Data) RegisterActivity(activityFunc interface{}) (HandlerInfo, error) {
	funcName := GetFunctionName(activityFunc)
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if already registered
	if handler, ok := r.activities[funcName]; ok {
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

	handler := HandlerInfo{
		HandlerName:     funcName,
		HandlerLongName: HandlerIdentity(funcName),
		Handler:         activityFunc,
		ParamsKinds:     make([]reflect.Kind, handlerType.NumIn()),
		ParamTypes:      make([]reflect.Type, handlerType.NumIn()),
		ReturnTypes:     make([]reflect.Type, handlerType.NumOut()),
		ReturnKinds:     make([]reflect.Kind, handlerType.NumOut()),
		NumIn:           handlerType.NumIn(),
		NumOut:          handlerType.NumOut(),
	}

	for i := 0; i < handlerType.NumIn(); i++ {
		handler.ParamsKinds[i] = handlerType.In(i).Kind()
		handler.ParamTypes[i] = handlerType.In(i)
	}

	for i := 0; i < handlerType.NumOut(); i++ {
		handler.ReturnKinds[i] = handlerType.Out(i).Kind()
		handler.ReturnTypes[i] = handlerType.Out(i)
	}

	r.activities[funcName] = handler

	return handler, nil
}

func (d *Data) Close() {
	d.client.Close()
	d.comfy.Close()
}

type CreateWorkflowOptions struct {
	stepID                    *schema.WorkflowStepID
	parentStepID              *schema.WorkflowStepID
	parentRunID               *schema.RunID
	parentWorkflowID          *schema.WorkflowEntityID
	parentWorkflowExecutionID *schema.WorkflowExecutionID
	continued                 bool
}

type CreateWorkflowOption func(*CreateWorkflowOptions)

func NewWorkflowWithStepID(stepID schema.WorkflowStepID) CreateWorkflowOption {
	return func(o *CreateWorkflowOptions) {
		o.stepID = &stepID
	}
}

func NewWorkflowWithParentStepID(parentStepID schema.WorkflowStepID) CreateWorkflowOption {
	return func(o *CreateWorkflowOptions) {
		o.parentStepID = &parentStepID
	}
}

func NewWorkflowWithParentRunID(parentRunID schema.RunID) CreateWorkflowOption {
	return func(o *CreateWorkflowOptions) {
		o.parentRunID = &parentRunID
	}
}

func NewWorkflowWithParentWorkflowID(parentWorkflowID schema.WorkflowEntityID) CreateWorkflowOption {
	return func(o *CreateWorkflowOptions) {
		o.parentWorkflowID = &parentWorkflowID
	}
}

func NewWorkflowWithParentWorkflowExecutionID(parentWorkflowExecutionID schema.WorkflowExecutionID) CreateWorkflowOption {
	return func(o *CreateWorkflowOptions) {
		o.parentWorkflowExecutionID = &parentWorkflowExecutionID
	}
}

func NewWorkflowWithContinued() CreateWorkflowOption {
	return func(o *CreateWorkflowOptions) {
		o.continued = true
	}
}

func NewWorkflowOptions(opts ...CreateWorkflowOption) CreateWorkflowOptions {
	options := CreateWorkflowOptions{}
	for _, o := range opts {
		o(&options)
	}
	return options
}

// Principal utility function to create a new workflow:
// - New workflow entity
// - Sub-workflow entity
// - Continue as new workflow entity
// - Detached workflow entity
func (d *Data) NewWorkflow(
	workflowFunc interface{},
	workflowOptions WorkflowOptions,
	options CreateWorkflowOptions,
	args ...interface{}) (*ent.WorkflowEntity, error) {

	var handler HandlerInfo
	var err error
	if handler, err = d.RegisterWorkflow(workflowFunc); err != nil {
		err := errors.Join(ErrPreparation, ErrRegistryRegisteration, fmt.Errorf("failed to register workflow: %w", err))
		d.config.logger.Error(context.Background(), err.Error(), "workflow_func", GetFunctionName(workflowFunc))
		return nil, err
	}

	var isRoot bool = true
	var inputBytes [][]byte

	// Arguments of the new workflow need to be converted to bytes for serialization and storage
	if inputBytes, err = convertInputsForSerialization(args); err != nil {
		err := errors.Join(ErrPreparation, ErrSerialization, fmt.Errorf("failed to serialize inputs: %w", err))
		d.config.logger.Error(context.Background(), err.Error(), "workflow_func", GetFunctionName(workflowFunc))
		return nil, err
	}

	// Check if this is a root workflow or a sub-workflow
	if options.parentRunID != nil ||
		options.parentWorkflowID != nil ||
		options.parentWorkflowExecutionID != nil ||
		options.continued {
		isRoot = false
	}

	var run *ent.Run
	var runID schema.RunID
	if options.parentRunID == nil { // if no parentRunID, means it's a real new workflow so create a new run
		d.config.logger.Debug(context.Background(), "preparing workflow got no parentRunID", "workflow_func", GetFunctionName(workflowFunc))
		if run, err = d.client.Run.Create().Save(d.Context); err != nil {
			err := errors.Join(ErrPreparation, fmt.Errorf("failed to create run: %w", err))
			d.config.logger.Error(context.Background(), err.Error(), "workflow_func", GetFunctionName(workflowFunc))
			return nil, err
		}
		runID = run.ID
	} else {
		d.config.logger.Debug(context.Background(), "preparing workflow got parentRunID", "workflow_func", GetFunctionName(workflowFunc), "parentRunID", options.parentRunID)
		runID = *options.parentRunID
	}

	var retryPolicy *schema.RetryPolicy
	if workflowOptions.RetryPolicy != nil {
		d.config.logger.Debug(context.Background(), "preparing workflow got retry policy", "workflow_func", GetFunctionName(workflowFunc), "retry_policy.max_attempts", workflowOptions.RetryPolicy.MaxAttempts, "retry_policy.max_interval", workflowOptions.RetryPolicy.MaxInterval)
		retryPolicy = workflowOptions.RetryPolicy
	} else {
		d.config.logger.Debug(context.Background(), "preparing workflow got no retry policy - applying default", "workflow_func", GetFunctionName(workflowFunc))
		retryPolicy = DefaultRetryPolicyInternal()
	}

	var queueName string = schema.DefaultQueue
	if workflowOptions.Queue != "" {
		d.config.logger.Debug(context.Background(), "preparing workflow got queue name", "workflow_func", GetFunctionName(workflowFunc), "queue_name", workflowOptions.Queue)
		queueName = workflowOptions.Queue
	}

	workflowDataBuilder := d.client.WorkflowData.
		Create().
		SetInputs(inputBytes).
		SetIsRoot(isRoot)

	if options.continued && options.parentWorkflowID != nil && options.parentWorkflowExecutionID != nil {
		var parentVersions []*ent.Version
		workflowDataBuilder.SetContinuedFrom(*options.parentWorkflowID)
		workflowDataBuilder.SetContinuedExecutionFrom(*options.parentWorkflowExecutionID)
		// If this is a continued workflow, we need to get the parent versions
		if parentVersions, err = d.GetVersionsByWorkflowEntityID(*options.parentWorkflowID); err != nil {
			err := errors.Join(ErrPreparation, fmt.Errorf("failed to get parent versions: %w", err))
			d.config.logger.Error(context.Background(), err.Error(), "workflow_func", GetFunctionName(workflowFunc))
			return nil, err
		}
		// We need to preemptively create versions of feature flags for the continued workflow
		for _, v := range parentVersions {
			//	create new versions for the continued workflow to conserve the same feature flags
			if _, err := d.client.Version.Create().
				SetEntityID(v.EntityID).
				SetChangeID(v.ChangeID).
				SetVersion(v.Version).
				SetData(v.Data).
				Save(d.Context); err != nil {
				err := errors.Join(ErrPreparation, fmt.Errorf("failed to create version: %w", err))
				d.config.logger.Error(context.Background(), err.Error(), "workflow_func", GetFunctionName(workflowFunc))
				return nil, err
			}
		}
	}

	var queueEntity *ent.Queue
	if queueEntity, err = d.client.Queue.Query().Where(queue.Name(queueName)).Only(d.Context); err != nil {
		err := errors.Join(ErrPreparation, fmt.Errorf("failed to get queue: %w", err))
		d.config.logger.Error(context.Background(), err.Error(), "workflow_func", GetFunctionName(workflowFunc))
		return nil, err
	}

	var stepID string = "root"
	if options.stepID != nil {
		stepID = string(*options.stepID)
	}

	// TODO: how do we link Versions with WorkflowEntity?

	var workflowEntity *ent.WorkflowEntity
	if workflowEntity, err = d.client.WorkflowEntity.Create().SetRunID(runID).SetRetryPolicy(*retryPolicy).SetRetryState(schema.RetryState{Attempts: 0}).SetStepID(schema.WorkflowStepID(stepID)).SetHandlerName(handler.HandlerName).SetQueue(queueEntity).Save(d.Context); err != nil {
		err := errors.Join(ErrPreparation, fmt.Errorf("failed to create workflow entity: %w", err))
		d.config.logger.Error(context.Background(), err.Error(), "workflow_func", GetFunctionName(workflowFunc))
		return nil, err
	}

	_, err = workflowDataBuilder.SetEntityID(workflowEntity.ID).SetWorkflow(workflowEntity).Save(d.Context)
	if err != nil {
		err := errors.Join(ErrPreparation, fmt.Errorf("failed to create workflow data: %w", err))
		d.config.logger.Error(context.Background(), err.Error(), "workflow_func", GetFunctionName(workflowFunc))
		return nil, errors.Join(ErrPreparation, fmt.Errorf("failed to create workflow data: %w", err))
	}

	return workflowEntity, nil
}

func DefaultRetryPolicyInternal() *schema.RetryPolicy {
	return &schema.RetryPolicy{
		MaxAttempts: 0, // 0 means no retries
		MaxInterval: (100 * time.Millisecond).Nanoseconds(),
	}
}

func (d *Data) GetVersionsByWorkflowEntityID(workflowEntityID schema.WorkflowEntityID) ([]*ent.Version, error) {
	versions, err := d.client.Version.Query().Where(version.EntityID(workflowEntityID)).All(d.Context)
	if err != nil {
		return nil, err
	}
	return versions, nil
}

type getterWorkflowEntityByID struct {
	withActivityChildren bool
	withExecutions       bool
	withRun              bool
	withVersions         bool
	withWorkflowData     bool
	withQueue            bool
}

type GetterWorkflowEntityByIDOption func(*getterWorkflowEntityByID)

func GetterWorkflowEntityWithActivityChildren() GetterWorkflowEntityByIDOption {
	return func(o *getterWorkflowEntityByID) {
		o.withActivityChildren = true
	}
}

func GetterWorkflowEntityWithExecutions() GetterWorkflowEntityByIDOption {
	return func(o *getterWorkflowEntityByID) {
		o.withExecutions = true
	}
}

func GetterWorkflowEntityWithRun() GetterWorkflowEntityByIDOption {
	return func(o *getterWorkflowEntityByID) {
		o.withRun = true
	}
}

func GetterWorkflowEntityWithVersions() GetterWorkflowEntityByIDOption {
	return func(o *getterWorkflowEntityByID) {
		o.withVersions = true
	}
}

func GetterWorkflowEntityWithWorkflowData() GetterWorkflowEntityByIDOption {
	return func(o *getterWorkflowEntityByID) {
		o.withWorkflowData = true
	}
}

func GetterWorkflowEntityWithQueue() GetterWorkflowEntityByIDOption {
	return func(o *getterWorkflowEntityByID) {
		o.withQueue = true
	}
}

func (d *Data) GetWorkflowEntityByID(workflowEntityID schema.WorkflowEntityID, opts ...GetterWorkflowEntityByIDOption) (*ent.WorkflowEntity, error) {
	options := getterWorkflowEntityByID{}
	for _, o := range opts {
		o(&options)
	}
	var workflowEntity *ent.WorkflowEntity
	var err error
	builder := d.client.WorkflowEntity.Query().Where(workflowentity.ID(workflowEntityID))
	if options.withActivityChildren {
		builder.WithActivityChildren(func(aeq *ent.ActivityEntityQuery) {})
	}
	if options.withExecutions {
		builder.WithExecutions(func(weq *ent.WorkflowExecutionQuery) {})
	}
	if options.withRun {
		builder.WithRun(func(rq *ent.RunQuery) {})
	}
	if options.withVersions {
		builder.WithVersions(func(vq *ent.VersionQuery) {})
	}
	if options.withWorkflowData {
		builder.WithWorkflowData(func(q *ent.WorkflowDataQuery) {})
	}
	if options.withQueue {
		builder.WithQueue(func(q *ent.QueueQuery) {})
	}
	if workflowEntity, err = builder.
		First(d.Context); err != nil {
		return nil, err
	}
	return workflowEntity, nil
}

func (d *Data) GetWorkflowData(id schema.WorkflowDataID) (*ent.WorkflowData, error) {
	workflowData, err := d.client.WorkflowData.Query().Where(workflowdata.ID(id)).First(d.Context)
	if err != nil {
		return nil, err
	}
	return workflowData, nil
}
