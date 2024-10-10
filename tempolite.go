package tempolite

import (
	"context"
	"encoding/json"
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
	"github.com/davidroman0O/go-tempolite/ent/run"
	"github.com/davidroman0O/go-tempolite/ent/schema"
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

	schedulerExecutionWorkflowDone chan struct{}
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

	tp := &Tempolite{
		ctx:    ctx,
		db:     db,
		client: client,
		cancel: cancel,
	}

	tp.workflowPool = tp.createWorkflowPool()

	go tp.schedulerExeutionWorkflow()

	return tp, nil
}

func (tp *Tempolite) Close() {
	tp.cancel() // it will stops other systems
	tp.workflowPool.Close()
}

func (tp *Tempolite) Wait() error {
	<-time.After(1 * time.Second)
	return tp.workflowPool.WaitWithCallback(tp.ctx, func(queueSize, processingCount, deadTaskCount int) bool {
		log.Printf("Wait: queueSize: %d, processingCount: %d, deadTaskCount: %d", queueSize, processingCount, deadTaskCount)
		return queueSize > 0 || processingCount > 0 || deadTaskCount > 0
	}, time.Second)
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

func (tp *Tempolite) enqueueActivty(ctx TempoliteContext, input ...interface{}) (*ActivityInfo, error) {
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

func (tp *Tempolite) getActivity(id string) (*ActivityInfo, error) {
	// todo: implement
	return nil, nil
}

func (tp *Tempolite) getWorkflow(id string) (*WorkflowInfo, error) {
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

func (tp *Tempolite) EnqueueActivityFunc(activityFunc interface{}, params ...interface{}) (string, error) {
	funcName := runtime.FuncForPC(reflect.ValueOf(activityFunc).Pointer()).Name()
	handlerIdentity := HandlerIdentity(funcName)
	return tp.EnqueueActivity(handlerIdentity, params...)
}

func (tp *Tempolite) EnqueueActivity(longName HandlerIdentity, params ...interface{}) (string, error) {
	// The whole implementation will be totally different, it's just for fun here
	// The real implementation will be with ActivityTask save on a db, then fetched from a scheduler to be sent to a pool.
	handlerInterface, ok := tp.activities.Load(longName)
	if !ok {
		return "", fmt.Errorf("activity %s not found", longName)
	}

	activity := handlerInterface.(Activity)

	if len(params) != activity.NumIn {
		return "", fmt.Errorf("parameter count mismatch: expected %d, got %d", activity.NumIn, len(params))
	}

	// Serialize parameters
	payloadBytes, err := json.Marshal(params)
	if err != nil {
		return "", fmt.Errorf("failed to marshal task parameters: %v", err)
	}

	// Enqueue the activity task (implementation depends on your task queue system)
	taskID := "task-id" // Generate or obtain a unique task ID
	log.Printf("Enqueued activity %s with payload: %s", longName, string(payloadBytes))

	// values := []reflect.Value{
	// 	reflect.ValueOf(ActivityContext{}),
	// }
	// for _, param := range params {
	// 	values = append(values, reflect.ValueOf(param))
	// }

	// reflect.ValueOf(activity.Handler).Call(values)

	return taskID, nil
}

func (tp *Tempolite) EnqueueWorkflow(workflowFunc interface{}, params ...interface{}) (string, error) {
	funcName := runtime.FuncForPC(reflect.ValueOf(workflowFunc).Pointer()).Name()
	handlerIdentity := HandlerIdentity(funcName)
	var value any
	var ok bool
	var err error

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

		var runEntity *ent.Run
		// root Run entity that anchored the workflow entity despite retries
		if runEntity, err = tp.
			client.
			Run.
			Create().
			SetID(uuid.NewString()).    // immutable
			SetRunID(uuid.NewString()). // can change if that flow change due to a retry
			SetType(run.TypeWorkflow).
			Save(tp.ctx); err != nil {
			return "", err
		}

		log.Printf("EnqueueWorkflow - created run entity %s for %s with params: %v", runEntity.ID, handlerIdentity, params)

		//	definition of a workflow, it exists but it is nothing without an execution that will be created as long as it retries
		var workflowEntity *ent.Workflow
		if workflowEntity, err = tp.client.Workflow.
			Create().
			SetID(runEntity.RunID).
			SetIdentity(string(handlerIdentity)).
			SetHandlerName(workflowHandlerInfo.HandlerName).
			SetInput(params).
			SetRetryPolicy(schema.RetryPolicy{
				MaximumAttempts: 3,
			}).
			Save(tp.ctx); err != nil {
			return "", err
		}

		log.Printf("EnqueueWorkflow - created workflow entity %s for %s with params: %v", workflowEntity.ID, handlerIdentity, params)

		// instance of the workflow definition which will be used to create a workflow task and match with the in-memory registry
		var workflowExecution *ent.WorkflowExecution
		if workflowExecution, err = tp.client.WorkflowExecution.
			Create().
			SetID(uuid.NewString()).
			SetRunID(runEntity.ID).
			SetWorkflow(workflowEntity).
			Save(tp.ctx); err != nil {
			return "", err
		}

		log.Printf("EnqueueWorkflow - created workflow execution %s for %s with params: %v", workflowExecution.ID, handlerIdentity, params)

		return runEntity.ID, nil
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

func (tp *Tempolite) GetWorkflow(id string) (*WorkflowInfo, error) {
	return tp.getWorkflow(id)
}

func (tp *Tempolite) GetActivity(id string) (*ActivityInfo, error) {
	return tp.getActivity(id)
}

func (tp *Tempolite) GetSideEffect(id string) (*SideEffectInfo, error) {
	return tp.getSideEffect(id)
}

func (tp *Tempolite) GetSaga(id string) (*SagaInfo, error) {
	return tp.getSaga(id)
}
