package tempolite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/davidroman0O/go-tempolite/dag"
	"github.com/davidroman0O/retrypool"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

// Enums
type TaskStatus int

const (
	TaskStatusPending TaskStatus = iota
	TaskStatusInProgress
	TaskStatusCompleted
	TaskStatusFailed
	TaskStatusCancelled
	TaskStatusTerminated
)

type SagaStatus int

const (
	SagaStatusPending SagaStatus = iota
	SagaStatusInProgress
	SagaStatusPaused
	SagaStatusCompleted
	SagaStatusFailed
	SagaStatusCancelled
	SagaStatusTerminating
	SagaStatusTerminated
	SagaStatusCriticallyFailed
)

type ExecutionNodeType int

const (
	ExecutionNodeTypeHandler ExecutionNodeType = iota
	ExecutionNodeTypeSagaHandler
	ExecutionNodeTypeSagaStep
	ExecutionNodeTypeSideEffect
	ExecutionNodeTypeCompensation
)

type ExecutionStatus int

const (
	ExecutionStatusPending ExecutionStatus = iota
	ExecutionStatusInProgress
	ExecutionStatusCompleted
	ExecutionStatusFailed
	ExecutionStatusCancelled
	ExecutionStatusCriticallyFailed
)

// Structs
type Task struct {
	ID                 string
	ExecutionContextID string
	HandlerName        string
	Payload            []byte
	Status             TaskStatus
	RetryCount         int
	ScheduledAt        time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
	CompletedAt        *time.Time
	Result             []byte
	ParentTaskID       *string
	SagaID             *string
}

type SagaInfo struct {
	ID              string
	Status          SagaStatus
	CurrentStep     int
	CreatedAt       time.Time
	LastUpdatedAt   time.Time
	CompletedAt     *time.Time
	HandlerName     string
	CancelRequested bool
}

type ExecutionNode struct {
	ID           string
	ParentID     *string
	Type         ExecutionNodeType
	Status       ExecutionStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CompletedAt  *time.Time
	HandlerName  string
	Payload      []byte
	Result       []byte
	ErrorMessage *string
}

type Compensation struct {
	ID        string
	SagaID    string
	StepIndex int
	Payload   []byte
	Status    ExecutionStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Signal struct {
	ID        string
	TaskID    string
	Name      string
	Payload   []byte
	CreatedAt time.Time
	Direction string
}

type WrappedResult struct {
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Data     interface{}            `json:"data"`
}

// Interfaces
type HandlerFunc interface{}
type HandlerSagaFunc interface{}

type TaskRepository interface {
	CreateTask(ctx context.Context, task *Task) error
	GetTask(ctx context.Context, id string) (*Task, error)
	UpdateTask(ctx context.Context, task *Task) error
	GetPendingTasks(ctx context.Context, limit int) ([]*Task, error)
	GetRunningTasksForSaga(ctx context.Context, sagaID string) ([]*Task, error)
}

type SideEffectRepository interface {
	GetSideEffect(ctx context.Context, executionContextID, key string) ([]byte, error)
	SaveSideEffect(ctx context.Context, executionContextID, key string, result []byte) error
}

type SignalRepository interface {
	SaveSignal(ctx context.Context, signal *Signal) error
	GetSignals(ctx context.Context, taskID string, name string, direction string) ([]*Signal, error)
	DeleteSignals(ctx context.Context, taskID string, name string, direction string) error
}

type SagaRepository interface {
	CreateSaga(ctx context.Context, saga *SagaInfo) error
	GetSaga(ctx context.Context, id string) (*SagaInfo, error)
	UpdateSaga(ctx context.Context, saga *SagaInfo) error
}

type ExecutionTreeRepository interface {
	CreateNode(ctx context.Context, node *ExecutionNode) error
	GetNode(ctx context.Context, id string) (*ExecutionNode, error)
	UpdateNode(ctx context.Context, node *ExecutionNode) error
	GetChildNodes(ctx context.Context, parentID string) ([]*ExecutionNode, error)
}

type CompensationRepository interface {
	CreateCompensation(ctx context.Context, compensation *Compensation) error
	GetCompensation(ctx context.Context, id string) (*Compensation, error)
	UpdateCompensation(ctx context.Context, compensation *Compensation) error
	GetCompensationsForSaga(ctx context.Context, sagaID string) ([]*Compensation, error)
}

type SagaStepRepository interface {
	CreateSagaStep(ctx context.Context, sagaID string, stepIndex int, payload []byte) error
	GetSagaStep(ctx context.Context, sagaID string, stepIndex int) ([]byte, error)
}

type HandlerContext interface {
	context.Context
	GetID() string
	EnqueueTask(handler HandlerFunc, params interface{}, options ...EnqueueOption) (string, error)
	EnqueueTaskAndWait(handler HandlerFunc, params interface{}, options ...EnqueueOption) (interface{}, error)
	SideEffect(key string, effect SideEffect) (interface{}, error)
	SendSignal(name string, payload interface{}) error
	ReceiveSignal(name string) (<-chan []byte, error)
	WaitForCompletion(taskID string) (interface{}, error)
}

type HandlerSagaContext interface {
	HandlerContext
	Step(step SagaStep) error
	EnqueueSaga(handler HandlerSagaFunc, params interface{}, options ...EnqueueOption) (string, error)
}

type TransactionContext interface {
	context.Context
	GetID() string
	SideEffect(key string, effect SideEffect) (interface{}, error)
	SendSignal(name string, payload interface{}) error
	ReceiveSignal(name string) (<-chan []byte, error)
}

type CompensationContext interface {
	context.Context
	GetID() string
	SideEffect(key string, effect SideEffect) (interface{}, error)
	SendSignal(name string, payload interface{}) error
	ReceiveSignal(name string) (<-chan []byte, error)
}

type SideEffectContext interface {
	context.Context
	GetID() string
	EnqueueTask(handler HandlerFunc, params interface{}, options ...EnqueueOption) (string, error)
	EnqueueTaskAndWait(handler HandlerFunc, params interface{}, options ...EnqueueOption) (interface{}, error)
	SideEffect(key string, effect SideEffect) (interface{}, error)
	SendSignal(name string, payload interface{}) error
	ReceiveSignal(name string) (<-chan []byte, error)
	WaitForCompletion(taskID string) (interface{}, error)
}

type SagaStep interface {
	Transaction(ctx TransactionContext) error
	Compensation(ctx CompensationContext) error
}

type SideEffect interface {
	Run(ctx SideEffectContext) (interface{}, error)
}

type sideEffectTask struct {
	sideEffect         SideEffect
	executionContextID string
	key                string
}

func (s *sideEffectTask) Run(ctx SideEffectContext) (interface{}, error) {
	effect, err := s.sideEffect.Run(ctx)
	log.Printf("Side effect task completed for key %s value %v", s.key, effect)
	return effect, err
}

// Tempolite struct and methods

type Tempolite struct {
	taskRepo          TaskRepository
	sideEffectRepo    SideEffectRepository
	signalRepo        SignalRepository
	sagaRepo          SagaRepository
	executionTreeRepo ExecutionTreeRepository
	compensationRepo  CompensationRepository
	sagaStepRepo      SagaStepRepository
	handlerPool       *retrypool.Pool[*Task]
	sagaHandlerPool   *retrypool.Pool[*Task]
	compensationPool  *retrypool.Pool[*Compensation]
	sideEffectPool    *retrypool.Pool[*sideEffectTask]
	db                *sql.DB
	handlers          map[string]handlerInfo
	handlersMutex     sync.RWMutex
	ctx               context.Context
	cancel            context.CancelFunc
	workersWg         sync.WaitGroup
	executionTrees    map[string]*dag.AcyclicGraph
	executionTreesMu  sync.RWMutex
}

type handlerInfo struct {
	Handler    interface{}
	ParamType  reflect.Type
	ReturnType reflect.Type
	IsSaga     bool
}

type TempoliteOption func(*Tempolite)

func WithHandlerWorkers(count int) TempoliteOption {
	return func(tp *Tempolite) {
		log.Printf("Initializing %d handler workers", count)
		workers := make([]retrypool.Worker[*Task], count)
		for i := 0; i < count; i++ {
			workers[i] = &TaskWorker{ID: i, tp: tp}
			log.Printf("Created handler worker %d", i)
		}
		tp.handlerPool = retrypool.New(tp.ctx, workers, tp.getHandlerPoolOptions()...)
		log.Printf("Handler pool initialized")
	}
}

func WithSagaWorkers(count int) TempoliteOption {
	return func(tp *Tempolite) {
		log.Printf("Initializing %d saga workers", count)
		workers := make([]retrypool.Worker[*Task], count)
		for i := 0; i < count; i++ {
			workers[i] = &SagaTaskWorker{ID: i, tp: tp}
			log.Printf("Created saga worker %d", i)
		}
		tp.sagaHandlerPool = retrypool.New(tp.ctx, workers, tp.getSagaHandlerPoolOptions()...)
		log.Printf("Saga pool initialized")
	}
}

func WithCompensationWorkers(count int) TempoliteOption {
	return func(tp *Tempolite) {
		log.Printf("Initializing %d compensation workers", count)
		workers := make([]retrypool.Worker[*Compensation], count)
		for i := 0; i < count; i++ {
			workers[i] = &CompensationWorker{ID: i, tp: tp}
			log.Printf("Created compensation worker %d", i)
		}
		tp.compensationPool = retrypool.New(tp.ctx, workers, tp.getCompensationPoolOptions()...)
		log.Printf("Compensation pool initialized")
	}
}

func WithSideEffectWorkers(count int) TempoliteOption {
	return func(tp *Tempolite) {
		log.Printf("Initializing %d side effect workers", count)
		workers := make([]retrypool.Worker[*sideEffectTask], count)
		for i := 0; i < count; i++ {
			workers[i] = &SideEffectWorker{ID: i, tp: tp}
			log.Printf("Created side effect worker %d", i)
		}
		tp.sideEffectPool = retrypool.New(tp.ctx, workers, tp.getSideEffectPoolOptions()...)
		log.Printf("Side effect pool initialized")
	}
}

func (tp *Tempolite) getHandlerPoolOptions() []retrypool.Option[*Task] {
	log.Printf("Getting handler pool options")
	return []retrypool.Option[*Task]{
		retrypool.WithOnTaskSuccess[*Task](tp.onHandlerSuccess),
		retrypool.WithOnTaskFailure[*Task](tp.onHandlerFailure),
		retrypool.WithOnRetry[*Task](tp.onHandlerRetry),
		retrypool.WithAttempts[*Task](1),
		retrypool.WithPanicHandler[*Task](tp.onHandlerPanic),
	}
}

func (tp *Tempolite) getSagaHandlerPoolOptions() []retrypool.Option[*Task] {
	log.Printf("Getting saga handler pool options")
	return []retrypool.Option[*Task]{
		retrypool.WithOnTaskSuccess[*Task](tp.onSagaHandlerSuccess),
		retrypool.WithOnTaskFailure[*Task](tp.onSagaHandlerFailure),
		retrypool.WithOnRetry[*Task](tp.onSagaHandlerRetry),
		retrypool.WithAttempts[*Task](1),
		retrypool.WithPanicHandler[*Task](tp.onSagaHandlerPanic),
	}
}

func (tp *Tempolite) getCompensationPoolOptions() []retrypool.Option[*Compensation] {
	log.Printf("Getting compensation pool options")
	return []retrypool.Option[*Compensation]{
		retrypool.WithOnTaskSuccess[*Compensation](tp.onCompensationSuccess),
		retrypool.WithOnTaskFailure[*Compensation](tp.onCompensationFailure),
		retrypool.WithOnRetry[*Compensation](tp.onCompensationRetry),
		retrypool.WithAttempts[*Compensation](1),
		retrypool.WithPanicHandler[*Compensation](tp.onCompensationPanic),
	}
}

func (tp *Tempolite) getSideEffectPoolOptions() []retrypool.Option[*sideEffectTask] {
	log.Printf("Getting side effect pool options")
	return []retrypool.Option[*sideEffectTask]{
		retrypool.WithOnTaskSuccess[*sideEffectTask](tp.onSideEffectSuccess),
		retrypool.WithOnTaskFailure[*sideEffectTask](tp.onSideEffectFailure),
		retrypool.WithOnRetry[*sideEffectTask](tp.onSideEffectRetry),
		retrypool.WithAttempts[*sideEffectTask](1),
		retrypool.WithPanicHandler[*sideEffectTask](tp.onSideEffectPanic),
	}
}

func New(ctx context.Context, db *sql.DB, options ...TempoliteOption) (*Tempolite, error) {
	log.Printf("Creating new Tempolite instance")
	ctx, cancel := context.WithCancel(ctx)

	tp := &Tempolite{
		db:             db,
		handlers:       make(map[string]handlerInfo),
		ctx:            ctx,
		cancel:         cancel,
		executionTrees: make(map[string]*dag.AcyclicGraph),
	}

	var err error

	log.Printf("Initializing SQLite repositories")
	tp.taskRepo, err = NewSQLiteTaskRepository(db)
	if err != nil {
		return nil, fmt.Errorf("error creating task repository: %w", err)
	}

	tp.sideEffectRepo, err = NewSQLiteSideEffectRepository(db)
	if err != nil {
		return nil, fmt.Errorf("error creating side effect repository: %w", err)
	}

	tp.signalRepo, err = NewSQLiteSignalRepository(db)
	if err != nil {
		return nil, fmt.Errorf("error creating signal repository: %w", err)
	}

	tp.sagaRepo, err = NewSQLiteSagaRepository(db)
	if err != nil {
		return nil, fmt.Errorf("error creating saga repository: %w", err)
	}

	tp.executionTreeRepo, err = NewSQLiteExecutionTreeRepository(db)
	if err != nil {
		return nil, fmt.Errorf("error creating execution tree repository: %w", err)
	}

	tp.compensationRepo, err = NewSQLiteCompensationRepository(db)
	if err != nil {
		return nil, fmt.Errorf("error creating compensation repository: %w", err)
	}

	tp.sagaStepRepo, err = NewSQLiteSagaStepRepository(db)
	if err != nil {
		return nil, fmt.Errorf("error creating saga step repository: %w", err)
	}

	log.Printf("Applying options")
	// Apply options
	for _, option := range options {
		option(tp)
	}

	// Initialize pools if not set by options
	if tp.handlerPool == nil {
		log.Printf("No handler pool set, creating default handler pool with 1 worker")
		WithHandlerWorkers(1)(tp)
	}
	if tp.sagaHandlerPool == nil {
		log.Printf("No saga handler pool set, creating default saga handler pool with 1 worker")
		WithSagaWorkers(1)(tp)
	}
	if tp.compensationPool == nil {
		log.Printf("No compensation pool set, creating default compensation pool with 1 worker")
		WithCompensationWorkers(1)(tp)
	}
	if tp.sideEffectPool == nil {
		log.Printf("No side effect pool set, creating default side effect pool with 1 worker")
		WithSideEffectWorkers(1)(tp)
	}

	log.Printf("Tempolite instance created successfully")
	return tp, nil
}

func (tp *Tempolite) RegisterHandler(handler interface{}) {
	handlerType := reflect.TypeOf(handler)
	log.Printf("Registering handler of type %v", handlerType)

	if handlerType.Kind() != reflect.Func {
		panic("Handler must be a function")
	}

	if handlerType.NumIn() != 2 {
		panic("Handler must have two input parameters")
	}

	if handlerType.In(0).Kind() != reflect.Interface || !handlerType.In(0).Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
		panic("First parameter of handler must implement context.Context")
	}

	isSaga := handlerType.In(0).Implements(reflect.TypeOf((*HandlerSagaContext)(nil)).Elem())

	var returnType reflect.Type
	if handlerType.NumOut() == 2 {
		if !handlerType.Out(1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			panic("Second return value of handler must be error")
		}
		returnType = handlerType.Out(0)
	} else if handlerType.NumOut() == 1 {
		if !handlerType.Out(0).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			panic("Return value of handler must be error")
		}
	} else {
		panic("Handler must have either one or two return values")
	}

	name := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
	log.Printf("Handler registered with name %s", name)
	tp.handlersMutex.Lock()
	tp.handlers[name] = handlerInfo{
		Handler:    handler,
		ParamType:  handlerType.In(1),
		ReturnType: returnType,
		IsSaga:     isSaga,
	}
	tp.handlersMutex.Unlock()
}

func (tp *Tempolite) getHandler(name string) (handlerInfo, bool) {
	log.Printf("Fetching handler with name %s", name)
	tp.handlersMutex.RLock()
	defer tp.handlersMutex.RUnlock()
	handler, exists := tp.handlers[name]
	return handler, exists
}

// WaitForTaskCompletion waits for the task with the given ID to reach a terminal state (completed, failed, etc.)
func (tp *Tempolite) WaitForTaskCompletion(ctx context.Context, taskID string, pollInterval time.Duration) (interface{}, error) {
	log.Printf("Waiting for completion of task ID %s", taskID)
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Context done while waiting for task ID %s", taskID)
			return nil, ctx.Err()
		case <-ticker.C:
			task, err := tp.taskRepo.GetTask(ctx, taskID)
			if err != nil {
				log.Printf("Error getting task with ID %s: %v", taskID, err)
				return nil, err
			}

			switch task.Status {
			case TaskStatusCompleted:
				if len(task.Result) == 0 {
					// Task completed with no result, return success
					log.Printf("Task with ID %s completed successfully but has no result", taskID)
					return nil, nil
				}

				// Handle the case where task has a result
				var wrappedResult WrappedResult
				if err := json.Unmarshal(task.Result, &wrappedResult); err != nil {
					log.Printf("Failed to unmarshal wrapped task result: %v", err)
					return nil, fmt.Errorf("failed to unmarshal wrapped task result: %v", err)
				}

				log.Printf("Task with ID %s completed successfully with result", taskID)
				return wrappedResult.Data, nil

			case TaskStatusFailed:
				log.Printf("Task with ID %s failed", taskID)
				return nil, fmt.Errorf("task failed with ID %s", taskID)

			case TaskStatusCancelled:
				log.Printf("Task with ID %s was cancelled", taskID)
				return nil, fmt.Errorf("task was cancelled")

			case TaskStatusTerminated:
				log.Printf("Task with ID %s was terminated", taskID)
				return nil, fmt.Errorf("task was terminated")

				// You can add more cases if there are additional task states.
			}
		}
	}
}

func (tp *Tempolite) GetInfo(ctx context.Context, id string) (interface{}, error) {
	log.Printf("Getting info for id %s", id)
	// Try to get task info
	task, err := tp.taskRepo.GetTask(ctx, id)
	if err == nil {
		log.Printf("Found task with id %s", id)
		return task, nil
	}

	// Try to get saga info
	saga, err := tp.sagaRepo.GetSaga(ctx, id)
	if err == nil {
		log.Printf("Found saga with id %s", id)
		return saga, nil
	}

	// Try to get side effect info
	sideEffect, err := tp.sideEffectRepo.GetSideEffect(ctx, id, "")
	if err == nil {
		log.Printf("Found side effect with id %s", id)
		return sideEffect, nil
	}

	log.Printf("No info found for id %s", id)
	return nil, fmt.Errorf("no info found for id: %s", id)
}

func (tp *Tempolite) GetExecutionTree(ctx context.Context, rootID string) (*dag.AcyclicGraph, error) {
	log.Printf("Getting execution tree for root ID %s", rootID)
	tp.executionTreesMu.RLock()
	tree, exists := tp.executionTrees[rootID]
	tp.executionTreesMu.RUnlock()

	if exists {
		log.Printf("Execution tree found in memory for root ID %s", rootID)
		return tree, nil
	}

	log.Printf("Execution tree not found in memory, reconstructing from database")
	// If the tree doesn't exist in memory, reconstruct it from the database
	node, err := tp.executionTreeRepo.GetNode(ctx, rootID)
	if err != nil {
		return nil, err
	}

	tree = &dag.AcyclicGraph{}
	err = tp.reconstructExecutionTree(ctx, node, tree)
	if err != nil {
		return nil, err
	}

	tp.executionTreesMu.Lock()
	tp.executionTrees[rootID] = tree
	tp.executionTreesMu.Unlock()

	log.Printf("Execution tree reconstructed and stored in memory for root ID %s", rootID)
	return tree, nil
}

func (tp *Tempolite) reconstructExecutionTree(ctx context.Context, node *ExecutionNode, tree *dag.AcyclicGraph) error {
	log.Printf("Adding node %s to execution tree", node.ID)
	tree.Add(node)

	children, err := tp.executionTreeRepo.GetChildNodes(ctx, node.ID)
	if err != nil {
		return err
	}

	log.Printf("Found %d child nodes for node %s", len(children), node.ID)
	for _, child := range children {
		err = tp.reconstructExecutionTree(ctx, child, tree)
		if err != nil {
			return err
		}
		log.Printf("Connecting node %s to child node %s", node.ID, child.ID)
		tree.Connect(dag.BasicEdge(node, child))
	}

	return nil
}

func (tp *Tempolite) SendSignal(ctx context.Context, taskID string, name string, payload interface{}) error {
	log.Printf("Sending signal '%s' for task ID %s", name, taskID)
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal signal payload: %v", err)
	}

	signal := &Signal{
		ID:        uuid.New().String(),
		TaskID:    taskID,
		Name:      name,
		Payload:   data,
		CreatedAt: time.Now(),
		Direction: "inbound",
	}

	err = tp.signalRepo.SaveSignal(ctx, signal)
	if err != nil {
		log.Printf("Failed to save signal: %v", err)
		return err
	}

	log.Printf("Signal '%s' sent successfully for task ID %s", name, taskID)
	return nil
}

func (tp *Tempolite) ReceiveSignal(ctx context.Context, taskID string, name string) (<-chan []byte, error) {
	log.Printf("Receiving signal '%s' for task ID %s", name, taskID)
	ch := make(chan []byte)

	go func() {
		defer close(ch)

		for {
			select {
			case <-ctx.Done():
				log.Printf("Context done while receiving signal '%s' for task ID %s", name, taskID)
				return
			case <-time.After(time.Second):
				signals, err := tp.signalRepo.GetSignals(ctx, taskID, name, "inbound")
				if err != nil {
					log.Printf("Error fetching signals: %v", err)
					continue
				}

				for _, signal := range signals {
					select {
					case ch <- signal.Payload:
						log.Printf("Received signal '%s' for task ID %s", name, taskID)
						if err := tp.signalRepo.DeleteSignals(ctx, taskID, name, "inbound"); err != nil {
							log.Printf("Error deleting signal: %v", err)
						}
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return ch, nil
}

func (tp *Tempolite) Cancel(ctx context.Context, id string) error {
	log.Printf("Cancelling task or saga with ID %s", id)
	// Try to cancel task
	task, err := tp.taskRepo.GetTask(ctx, id)
	if err == nil {
		task.Status = TaskStatusCancelled
		err = tp.taskRepo.UpdateTask(ctx, task)
		if err != nil {
			log.Printf("Failed to cancel task with ID %s: %v", id, err)
			return err
		}
		log.Printf("Task with ID %s cancelled successfully", id)
		return nil
	}

	// Try to cancel saga
	saga, err := tp.sagaRepo.GetSaga(ctx, id)
	if err == nil {
		saga.Status = SagaStatusCancelled
		err = tp.sagaRepo.UpdateSaga(ctx, saga)
		if err != nil {
			log.Printf("Failed to cancel saga with ID %s: %v", id, err)
			return err
		}
		log.Printf("Saga with ID %s cancelled successfully", id)
		return nil
	}

	log.Printf("No task or saga found with ID %s", id)
	return fmt.Errorf("no task or saga found with id: %s", id)
}

func (tp *Tempolite) Terminate(ctx context.Context, id string) error {
	log.Printf("Terminating task or saga with ID %s", id)
	// Try to terminate task
	task, err := tp.taskRepo.GetTask(ctx, id)
	if err == nil {
		task.Status = TaskStatusTerminated
		err = tp.taskRepo.UpdateTask(ctx, task)
		if err != nil {
			log.Printf("Failed to terminate task with ID %s: %v", id, err)
			return err
		}
		log.Printf("Task with ID %s terminated successfully", id)
		return nil
	}

	// Try to terminate saga
	saga, err := tp.sagaRepo.GetSaga(ctx, id)
	if err == nil {
		saga.Status = SagaStatusTerminated
		err = tp.sagaRepo.UpdateSaga(ctx, saga)
		if err != nil {
			log.Printf("Failed to terminate saga with ID %s: %v", id, err)
			return err
		}
		log.Printf("Saga with ID %s terminated successfully", id)
		return nil
	}

	log.Printf("No task or saga found with ID %s", id)
	return fmt.Errorf("no task or saga found with id: %s", id)
}

type EnqueueOption func(*enqueueOptions)

type enqueueOptions struct {
	maxDuration    time.Duration
	timeLimit      time.Duration
	immediate      bool
	panicOnTimeout bool
}

func WithMaxDuration(duration time.Duration) EnqueueOption {
	return func(o *enqueueOptions) {
		log.Printf("Setting max duration for enqueue option: %v", duration)
		o.maxDuration = duration
	}
}

func WithTimeLimit(limit time.Duration) EnqueueOption {
	return func(o *enqueueOptions) {
		log.Printf("Setting time limit for enqueue option: %v", limit)
		o.timeLimit = limit
	}
}

func WithImmediateRetry() EnqueueOption {
	return func(o *enqueueOptions) {
		log.Printf("Enabling immediate retry for enqueue option")
		o.immediate = true
	}
}

func WithPanicOnTimeout() EnqueueOption {
	return func(o *enqueueOptions) {
		log.Printf("Enabling panic on timeout for enqueue option")
		o.panicOnTimeout = true
	}
}

func (tp *Tempolite) Enqueue(ctx context.Context, handler interface{}, params interface{}, options ...EnqueueOption) (string, error) {
	handlerName := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
	log.Printf("Enqueuing task with handler %s", handlerName)
	handlerInfo, exists := tp.getHandler(handlerName)
	if !exists {
		log.Printf("No handler registered with name %s", handlerName)
		return "", fmt.Errorf("no handler registered with name: %s", handlerName)
	}

	opts := enqueueOptions{}
	for _, option := range options {
		option(&opts)
	}

	payload, err := json.Marshal(params)
	if err != nil {
		log.Printf("Failed to marshal task parameters for handler %s: %v", handlerName, err)
		return "", fmt.Errorf("failed to marshal task parameters: %v", err)
	}

	executionContextID := uuid.New().String()
	task := &Task{
		ID:                 uuid.New().String(),
		ExecutionContextID: executionContextID,
		HandlerName:        handlerName,
		Payload:            payload,
		Status:             TaskStatusPending,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
		ScheduledAt:        time.Now(),
	}

	if handlerInfo.IsSaga {
		log.Printf("Task is part of a saga, creating saga info")
		saga := &SagaInfo{
			ID:            task.ID,
			Status:        SagaStatusPending,
			CreatedAt:     time.Now(),
			LastUpdatedAt: time.Now(),
			HandlerName:   handlerName,
		}
		if err := tp.sagaRepo.CreateSaga(ctx, saga); err != nil {
			log.Printf("Failed to create saga: %v", err)
			return "", fmt.Errorf("failed to create saga: %v", err)
		}
		task.SagaID = &saga.ID
	}

	log.Printf("Creating task in repository")
	if err := tp.taskRepo.CreateTask(ctx, task); err != nil {
		log.Printf("Failed to create task: %v", err)
		return "", fmt.Errorf("failed to create task: %v", err)
	}

	log.Printf("Creating execution node for task %s", task.ID)
	executionNode := &ExecutionNode{
		ID:          task.ID,
		Type:        ExecutionNodeTypeHandler,
		Status:      ExecutionStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		HandlerName: handlerName,
		Payload:     payload,
	}

	if err := tp.executionTreeRepo.CreateNode(ctx, executionNode); err != nil {
		log.Printf("Failed to create execution node: %v", err)
		return "", fmt.Errorf("failed to create execution node: %v", err)
	}

	log.Printf("Dispatching task to pool")
	poolOptions := []retrypool.TaskOption[*Task]{
		retrypool.WithMaxDuration[*Task](opts.maxDuration),
		retrypool.WithTimeLimit[*Task](opts.timeLimit),
	}
	if opts.immediate {
		poolOptions = append(poolOptions, retrypool.WithImmediateRetry[*Task]())
	}
	if opts.panicOnTimeout {
		poolOptions = append(poolOptions, retrypool.WithPanicOnTimeout[*Task]())
	}

	if handlerInfo.IsSaga {
		tp.sagaHandlerPool.Dispatch(task, poolOptions...)
	} else {
		tp.handlerPool.Dispatch(task, poolOptions...)
	}

	log.Printf("Task with ID %s enqueued successfully", task.ID)
	return task.ID, nil
}

func (tp *Tempolite) EnqueueSaga(ctx context.Context, handler HandlerSagaFunc, params interface{}, options ...EnqueueOption) (string, error) {
	log.Printf("Enqueuing saga task")
	return tp.Enqueue(ctx, handler, params, options...)
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

type TempoliteInfo struct {
	Tasks             int
	SagaTasks         int
	CompensationTasks int
	SideEffectTasks   int
}

func (tp *Tempolite) getInfo() TempoliteInfo {
	log.Printf("Getting pool stats")
	return TempoliteInfo{
		Tasks:             tp.handlerPool.QueueSize(),
		SagaTasks:         tp.sagaHandlerPool.QueueSize(),
		CompensationTasks: tp.compensationPool.QueueSize(),
		SideEffectTasks:   tp.sideEffectPool.QueueSize(),
	}
}

func (tp *Tempolite) GetPoolStats() map[string]int {
	log.Printf("Getting pool statistics")
	return map[string]int{
		"handler":      tp.handlerPool.QueueSize(),
		"saga":         tp.sagaHandlerPool.QueueSize(),
		"compensation": tp.compensationPool.QueueSize(),
		"sideeffect":   tp.sideEffectPool.QueueSize(),
	}
}

func (tp *Tempolite) Close() error {
	log.Printf("Closing Tempolite instance")
	tp.cancel()
	tp.handlerPool.Close()
	tp.sagaHandlerPool.Close()
	tp.compensationPool.Close()
	tp.sideEffectPool.Close()
	tp.workersWg.Wait()
	log.Printf("Tempolite instance closed successfully")
	return nil
}

// Worker implementations

type TaskWorker struct {
	ID int
	tp *Tempolite
}

func (w *TaskWorker) Run(ctx context.Context, task *Task) error {
	log.Printf("Running task with ID %s on worker %d", task.ID, w.ID)
	handlerInfo, exists := w.tp.getHandler(task.HandlerName)
	if !exists {
		return fmt.Errorf("no handler registered with name: %s", task.HandlerName)
	}

	handlerValue := reflect.ValueOf(handlerInfo.Handler)
	paramType := handlerValue.Type().In(1)
	param := reflect.New(paramType).Interface()

	err := json.Unmarshal(task.Payload, param)
	if err != nil {
		log.Printf("Failed to unmarshal task payload: %v", err)
		return fmt.Errorf("failed to unmarshal task payload: %v", err)
	}

	handlerCtx := &handlerContext{
		Context:            ctx,
		tp:                 w.tp,
		taskID:             task.ID,
		executionContextID: task.ExecutionContextID,
	}

	log.Printf("Calling handler for task ID %s", task.ID)
	results := handlerValue.Call([]reflect.Value{
		reflect.ValueOf(handlerCtx),
		reflect.ValueOf(param).Elem(),
	})

	// Handle error if present in the last return value.
	if len(results) > 0 && !results[len(results)-1].IsNil() {
		return results[len(results)-1].Interface().(error)
	}

	// Handle result if present.
	if len(results) > 1 && !results[0].IsNil() {
		wrappedResult := WrappedResult{
			Metadata: map[string]interface{}{}, // Add any relevant metadata here
			Data:     results[0].Interface(),
		}
		resultBytes, err := json.Marshal(wrappedResult)
		if err != nil {
			log.Printf("Failed to marshal wrapped task result: %v", err)
			return fmt.Errorf("failed to marshal wrapped task result: %v", err)
		}
		task.Result = resultBytes
		log.Printf("Task result marshaled successfully for task ID %s", task.ID)
	}

	log.Printf("Task with ID %s completed successfully on worker %d", task.ID, w.ID)
	return nil
}

type SagaTaskWorker struct {
	ID int
	tp *Tempolite
}

func (w *SagaTaskWorker) Run(ctx context.Context, task *Task) error {
	log.Printf("Running saga task with ID %s on worker %d", task.ID, w.ID)
	handlerInfo, exists := w.tp.getHandler(task.HandlerName)
	if !exists {
		return fmt.Errorf("no handler registered with name: %s", task.HandlerName)
	}

	handlerValue := reflect.ValueOf(handlerInfo.Handler)
	paramType := handlerValue.Type().In(1)
	param := reflect.New(paramType).Interface()

	err := json.Unmarshal(task.Payload, param)
	if err != nil {
		log.Printf("Failed to unmarshal task payload: %v", err)
		return fmt.Errorf("failed to unmarshal task payload: %v", err)
	}

	handlerCtx := &handlerSagaContext{
		handlerContext: handlerContext{
			Context:            ctx,
			tp:                 w.tp,
			taskID:             task.ID,
			executionContextID: task.ExecutionContextID,
		},
		sagaID: *task.SagaID,
	}

	log.Printf("Calling saga handler for task ID %s", task.ID)
	results := handlerValue.Call([]reflect.Value{
		reflect.ValueOf(handlerCtx),
		reflect.ValueOf(param).Elem(),
	})

	// Handle error if present in the last return value.
	if len(results) > 0 && !results[len(results)-1].IsNil() {
		return results[len(results)-1].Interface().(error)
	}

	// Handle result if present.
	if len(results) > 1 && !results[0].IsNil() {
		result, err := json.Marshal(results[0].Interface())
		if err != nil {
			log.Printf("Failed to marshal task result: %v", err)
			return fmt.Errorf("failed to marshal task result: %v", err)
		}
		task.Result = result
		log.Printf("Task result marshaled successfully for saga task ID %s", task.ID)
	}

	log.Printf("Saga task with ID %s completed successfully on worker %d", task.ID, w.ID)
	return nil
}

type CompensationWorker struct {
	ID int
	tp *Tempolite
}

func (w *CompensationWorker) Run(ctx context.Context, compensation *Compensation) error {
	log.Printf("Running compensation with ID %s on worker %d", compensation.ID, w.ID)
	saga, err := w.tp.sagaRepo.GetSaga(ctx, compensation.SagaID)
	if err != nil {
		log.Printf("Failed to get saga: %v", err)
		return fmt.Errorf("failed to get saga: %v", err)
	}

	handlerInfo, exists := w.tp.getHandler(saga.HandlerName)
	if !exists {
		return fmt.Errorf("no handler registered with name: %s", saga.HandlerName)
	}

	handlerValue := reflect.ValueOf(handlerInfo.Handler)
	paramType := handlerValue.Type().In(1)
	param := reflect.New(paramType).Interface()

	err = json.Unmarshal(compensation.Payload, param)
	if err != nil {
		log.Printf("Failed to unmarshal compensation payload: %v", err)
		return fmt.Errorf("failed to unmarshal compensation payload: %v", err)
	}

	compensationCtx := &compensationContext{
		Context: ctx,
		tp:      w.tp,
		sagaID:  compensation.SagaID,
		stepID:  compensation.ID,
	}

	log.Printf("Calling compensation handler for compensation ID %s", compensation.ID)
	results := handlerValue.Call([]reflect.Value{
		reflect.ValueOf(compensationCtx),
		reflect.ValueOf(param).Elem(),
	})

	if len(results) > 0 && !results[len(results)-1].IsNil() {
		return results[len(results)-1].Interface().(error)
	}

	log.Printf("Compensation with ID %s completed successfully on worker %d", compensation.ID, w.ID)
	return nil
}

type SideEffectWorker struct {
	ID int
	tp *Tempolite
}

func (w *SideEffectWorker) Run(ctx context.Context, task *sideEffectTask) error {
	log.Printf("Running side effect on worker %d %s %s", w.ID, task.key, task.executionContextID)
	sideEffectCtx := &sideEffectContext{
		Context: ctx,
		tp:      w.tp,
		id:      task.executionContextID,
	}

	result, err := task.sideEffect.Run(sideEffectCtx)
	if err != nil {
		log.Printf("Side effect run failed: %v", err)
		return err
	}

	log.Printf("Side effect completed successfully, saving result %v", result)

	wrappedResult := WrappedResult{
		Metadata: map[string]interface{}{}, // Add any relevant metadata if needed
		Data:     result,
	}

	resultBytes, err := json.Marshal(wrappedResult)
	if err != nil {
		log.Printf("Failed to marshal side effect result: %v", err)
		return fmt.Errorf("failed to marshal side effect result: %v", err)
	}

	err = w.tp.sideEffectRepo.SaveSideEffect(ctx, task.executionContextID, task.key, resultBytes)
	if err != nil {
		log.Printf("Failed to save side effect result: %v", err)
		return fmt.Errorf("failed to save side effect result: %v", err)
	}

	log.Printf("Side effect completed successfully on worker %d", w.ID)
	return nil
}

// Context implementations

type handlerContext struct {
	context.Context
	tp                 *Tempolite
	taskID             string
	executionContextID string
}

func (c *handlerContext) GetID() string {
	return c.taskID
}

func (c *handlerContext) EnqueueTask(handler HandlerFunc, params interface{}, options ...EnqueueOption) (string, error) {
	log.Printf("Enqueuing child task from handler context, parent task ID %s", c.taskID)
	taskID, err := c.tp.Enqueue(c, handler, params, options...)
	if err != nil {
		log.Printf("Failed to enqueue child task: %v", err)
		return "", err
	}

	log.Printf("Linking child task %s with parent task %s in execution tree", taskID, c.taskID)
	parentNode, err := c.tp.executionTreeRepo.GetNode(c, c.taskID)
	if err != nil {
		log.Printf("Failed to get parent node: %v", err)
		return "", fmt.Errorf("failed to get parent node: %v", err)
	}

	childNode, err := c.tp.executionTreeRepo.GetNode(c, taskID)
	if err != nil {
		log.Printf("Failed to get child node: %v", err)
		return "", fmt.Errorf("failed to get child node: %v", err)
	}

	childNode.ParentID = &parentNode.ID
	if err := c.tp.executionTreeRepo.UpdateNode(c, childNode); err != nil {
		log.Printf("Failed to update child node: %v", err)
		return "", fmt.Errorf("failed to update child node: %v", err)
	}

	log.Printf("Child task %s enqueued successfully from parent task %s", taskID, c.taskID)
	return taskID, nil
}

func (c *handlerContext) EnqueueTaskAndWait(handler HandlerFunc, params interface{}, options ...EnqueueOption) (interface{}, error) {
	log.Printf("Enqueuing and waiting for task from handler context, parent task ID %s", c.taskID)
	taskID, err := c.EnqueueTask(handler, params, options...)
	if err != nil {
		return nil, err
	}

	return c.WaitForCompletion(taskID)
}

func (c *handlerContext) SideEffect(key string, effect SideEffect) (interface{}, error) {
	log.Printf("Running side effect with key %s for task ID %s", key, c.executionContextID)

	// Check if the side effect already exists
	result, err := c.tp.sideEffectRepo.GetSideEffect(c, c.executionContextID, key)
	if err == nil && len(result) > 0 {
		var value interface{}
		if err := json.Unmarshal(result, &value); err != nil {
			log.Printf("Failed to unmarshal side effect result: %v", err)
			return nil, fmt.Errorf("failed to unmarshal side effect result: %v", err)
		}
		return value, nil
	}

	log.Printf("Dispatching side effect with key %s for task ID %s", key, c.taskID)
	// c.tp.sideEffectPool.Dispatch(effect)
	c.tp.sideEffectPool.Dispatch(&sideEffectTask{
		sideEffect:         effect,
		executionContextID: c.executionContextID,
		key:                key,
	})

	// Retry fetching the result with a timeout
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("timeout while waiting for side effect result with key %s", key)
		case <-ticker.C:
			result, err = c.tp.sideEffectRepo.GetSideEffect(c, c.executionContextID, key)
			if err != nil {
				log.Printf("Error fetching side effect result: %v", err)
				continue
			}
			if len(result) > 0 {
				wrappedResult := WrappedResult{}
				// var value interface{}
				if err := json.Unmarshal(result, &wrappedResult); err != nil {
					log.Printf("Failed to unmarshal side effect result: %v", err)
					return nil, fmt.Errorf("failed to unmarshal side effect result: %v", err)
				}
				log.Printf("Side effect with key %s for task ID %s completed successfully: %v", key, c.taskID, wrappedResult.Data)
				return wrappedResult.Data, nil
			}
		}
	}
}

func (c *handlerContext) SendSignal(name string, payload interface{}) error {
	log.Printf("Sending signal '%s' from handler context, task ID %s", name, c.taskID)
	return c.tp.SendSignal(c, c.taskID, name, payload)
}

func (c *handlerContext) ReceiveSignal(name string) (<-chan []byte, error) {
	log.Printf("Receiving signal '%s' from handler context, task ID %s", name, c.taskID)
	return c.tp.ReceiveSignal(c, c.taskID, name)
}

func (c *handlerContext) WaitForCompletion(taskID string) (interface{}, error) {
	log.Printf("Waiting for completion of task ID %s", taskID)
	for {
		select {
		case <-c.Done():
			log.Printf("Context done while waiting for task ID %s", taskID)
			return nil, c.Err()
		case <-time.After(time.Second):
			task, err := c.tp.taskRepo.GetTask(c, taskID)
			if err != nil {
				log.Printf("Failed to get task: %v", err)
				return nil, err
			}

			switch task.Status {
			case TaskStatusCompleted:
				var result interface{}
				log.Printf("HandlerContext Task ID %s completed successfully - %s %s - %v", taskID, task.ID, task.ExecutionContextID, task.Result)
				if err := json.Unmarshal(task.Result, &result); err != nil {
					log.Printf("Failed to unmarshal task result: %v", err)
					return nil, fmt.Errorf("failed to unmarshal task result: %v", err)
				}
				log.Printf("Task ID %s completed successfully", taskID)
				return result, nil
			case TaskStatusFailed:
				log.Printf("Task ID %s failed", taskID)
				return nil, fmt.Errorf("task failed")
			case TaskStatusCancelled:
				log.Printf("Task ID %s cancelled", taskID)
				return nil, fmt.Errorf("task cancelled")
			case TaskStatusTerminated:
				log.Printf("Task ID %s terminated", taskID)
				return nil, fmt.Errorf("task terminated")
			}
		}
	}
}

type handlerSagaContext struct {
	handlerContext
	sagaID string
}

func (c *handlerSagaContext) Step(step SagaStep) error {
	log.Printf("Executing saga step for saga ID %s", c.sagaID)
	saga, err := c.tp.sagaRepo.GetSaga(c, c.sagaID)
	if err != nil {
		log.Printf("Failed to get saga: %v", err)
		return fmt.Errorf("failed to get saga: %v", err)
	}

	transactionCtx := &transactionContext{
		Context: c.Context,
		tp:      c.tp,
		sagaID:  c.sagaID,
		stepID:  uuid.New().String(),
	}

	log.Printf("Executing transaction for saga step")
	if err := step.Transaction(transactionCtx); err != nil {
		log.Printf("Transaction failed, executing compensation")
		compensationCtx := &compensationContext{
			Context: c.Context,
			tp:      c.tp,
			sagaID:  c.sagaID,
			stepID:  transactionCtx.stepID,
		}

		if compErr := step.Compensation(compensationCtx); compErr != nil {
			log.Printf("Compensation failed for saga step: %v", compErr)
			return fmt.Errorf("transaction failed and compensation failed: %v, compensation error: %v", err, compErr)
		}

		log.Printf("Transaction failed, but compensation succeeded")
		return fmt.Errorf("transaction failed, compensation succeeded: %v", err)
	}

	log.Printf("Updating saga current step")
	saga.CurrentStep++
	saga.LastUpdatedAt = time.Now()
	if err := c.tp.sagaRepo.UpdateSaga(c, saga); err != nil {
		log.Printf("Failed to update saga: %v", err)
		return fmt.Errorf("failed to update saga: %v", err)
	}

	log.Printf("Saga step completed successfully for saga ID %s", c.sagaID)
	return nil
}

func (c *handlerSagaContext) EnqueueSaga(handler HandlerSagaFunc, params interface{}, options ...EnqueueOption) (string, error) {
	log.Printf("Enqueuing saga task from handler saga context, saga ID %s", c.sagaID)
	return c.tp.EnqueueSaga(c, handler, params, options...)
}

type transactionContext struct {
	context.Context
	tp     *Tempolite
	sagaID string
	stepID string
}

func (c *transactionContext) GetID() string {
	return c.stepID
}

func (c *transactionContext) SideEffect(key string, effect SideEffect) (interface{}, error) {
	log.Printf("Getting side effect with key %s for transaction context, saga ID %s", key, c.sagaID)
	return c.tp.sideEffectRepo.GetSideEffect(c, c.sagaID, key)
}

func (c *transactionContext) SendSignal(name string, payload interface{}) error {
	log.Printf("Sending signal '%s' from transaction context, step ID %s", name, c.stepID)
	return c.tp.SendSignal(c, c.stepID, name, payload)
}

func (c *transactionContext) ReceiveSignal(name string) (<-chan []byte, error) {
	log.Printf("Receiving signal '%s' from transaction context, step ID %s", name, c.stepID)
	return c.tp.ReceiveSignal(c, c.stepID, name)
}

type compensationContext struct {
	context.Context
	tp     *Tempolite
	sagaID string
	stepID string
}

func (c *compensationContext) GetID() string {
	return c.stepID
}

func (c *compensationContext) SideEffect(key string, effect SideEffect) (interface{}, error) {
	log.Printf("Getting side effect with key %s for compensation context, saga ID %s", key, c.sagaID)
	return c.tp.sideEffectRepo.GetSideEffect(c, c.sagaID, key)
}

func (c *compensationContext) SendSignal(name string, payload interface{}) error {
	log.Printf("Sending signal '%s' from compensation context, step ID %s", name, c.stepID)
	return c.tp.SendSignal(c, c.stepID, name, payload)
}

func (c *compensationContext) ReceiveSignal(name string) (<-chan []byte, error) {
	log.Printf("Receiving signal '%s' from compensation context, step ID %s", name, c.stepID)
	return c.tp.ReceiveSignal(c, c.stepID, name)
}

type sideEffectContext struct {
	context.Context
	tp *Tempolite
	id string
}

func (c *sideEffectContext) GetID() string {
	return c.id
}

func (c *sideEffectContext) EnqueueTask(handler HandlerFunc, params interface{}, options ...EnqueueOption) (string, error) {
	log.Printf("Enqueuing task from side effect context, side effect ID %s", c.id)
	return c.tp.Enqueue(c, handler, params, options...)
}

func (c *sideEffectContext) EnqueueTaskAndWait(handler HandlerFunc, params interface{}, options ...EnqueueOption) (interface{}, error) {
	log.Printf("Enqueuing and waiting for task from side effect context, side effect ID %s", c.id)
	taskID, err := c.EnqueueTask(handler, params, options...)
	if err != nil {
		return nil, err
	}
	return c.WaitForCompletion(taskID)
}

func (c *sideEffectContext) SideEffect(key string, effect SideEffect) (interface{}, error) {
	log.Printf("Getting side effect with key %s from side effect context, side effect ID %s", key, c.id)
	return c.tp.sideEffectRepo.GetSideEffect(c, c.id, key)
}

func (c *sideEffectContext) SendSignal(name string, payload interface{}) error {
	log.Printf("Sending signal '%s' from side effect context, side effect ID %s", name, c.id)
	return c.tp.SendSignal(c, c.id, name, payload)
}

func (c *sideEffectContext) ReceiveSignal(name string) (<-chan []byte, error) {
	log.Printf("Receiving signal '%s' from side effect context, side effect ID %s", name, c.id)
	return c.tp.ReceiveSignal(c, c.id, name)
}

func (c *sideEffectContext) WaitForCompletion(taskID string) (interface{}, error) {
	log.Printf("Waiting for completion of task ID %s from side effect context", taskID)
	for {
		select {
		case <-c.Done():
			log.Printf("Context done while waiting for task ID %s", taskID)
			return nil, c.Err()
		case <-time.After(time.Second):
			task, err := c.tp.taskRepo.GetTask(c, taskID)
			if err != nil {
				log.Printf("Failed to get task: %v", err)
				return nil, err
			}

			switch task.Status {
			case TaskStatusCompleted:
				var result interface{}
				log.Printf("SideEffectContext Task ID %s completed successfully - %s %s - %v", taskID, task.ID, task.ExecutionContextID, task.Result)
				if err := json.Unmarshal(task.Result, &result); err != nil {
					log.Printf("Failed to unmarshal task result: %v", err)
					return nil, fmt.Errorf("failed to unmarshal task result: %v", err)
				}
				log.Printf("Task ID %s completed successfully", taskID)
				return result, nil
			case TaskStatusFailed:
				log.Printf("Task ID %s failed", taskID)
				return nil, fmt.Errorf("task failed")
			case TaskStatusCancelled:
				log.Printf("Task ID %s cancelled", taskID)
				return nil, fmt.Errorf("task cancelled")
			case TaskStatusTerminated:
				log.Printf("Task ID %s terminated", taskID)
				return nil, fmt.Errorf("task terminated")
			}
		}
	}
}

// Utility functions

func IsUnrecoverable(err error) bool {
	type unrecoverable interface {
		Unrecoverable() bool
	}

	if u, ok := err.(unrecoverable); ok {
		return u.Unrecoverable()
	}

	return false
}

func Unrecoverable(err error) error {
	return unrecoverableError{err}
}

type unrecoverableError struct {
	error
}

func (e unrecoverableError) Unrecoverable() bool {
	return true
}

func (e unrecoverableError) Unwrap() error {
	return e.error
}

func nullableTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return t.Unix()
}

func nullableString(s *string) interface{} {
	if s == nil {
		return nil
	}
	return *s
}

// Callback implementations

func (tp *Tempolite) onHandlerSuccess(controller retrypool.WorkerController[*Task], workerID int, worker retrypool.Worker[*Task], task *retrypool.TaskWrapper[*Task]) {
	log.Printf("Handler task with ID %s succeeded", task.Data().ID)
	taskData := task.Data()
	taskData.Status = TaskStatusCompleted
	now := time.Now()
	taskData.CompletedAt = &now
	if err := tp.taskRepo.UpdateTask(tp.ctx, taskData); err != nil {
		log.Printf("Failed to update task status: %v", err)
	}

	node, err := tp.executionTreeRepo.GetNode(tp.ctx, taskData.ID)
	if err != nil {
		log.Printf("Failed to get execution node: %v", err)
		return
	}

	node.Status = ExecutionStatusCompleted
	node.CompletedAt = taskData.CompletedAt
	node.Result = taskData.Result
	if err := tp.executionTreeRepo.UpdateNode(tp.ctx, node); err != nil {
		log.Printf("Failed to update execution node: %v", err)
	}
}

func (tp *Tempolite) onHandlerFailure(controller retrypool.WorkerController[*Task], workerID int, worker retrypool.Worker[*Task], task *retrypool.TaskWrapper[*Task], err error) retrypool.DeadTaskAction {
	log.Printf("Handler task with ID %s failed: %v", task.Data().ID, err)
	taskData := task.Data()
	taskData.Status = TaskStatusFailed
	if err := tp.taskRepo.UpdateTask(tp.ctx, taskData); err != nil {
		log.Printf("Failed to update task status: %v", err)
	}

	node, nodeErr := tp.executionTreeRepo.GetNode(tp.ctx, taskData.ID)
	if nodeErr != nil {
		log.Printf("Failed to get execution node: %v", nodeErr)
	} else {
		node.Status = ExecutionStatusFailed
		errorMsg := err.Error()
		node.ErrorMessage = &errorMsg
		if updateErr := tp.executionTreeRepo.UpdateNode(tp.ctx, node); updateErr != nil {
			log.Printf("Failed to update execution node: %v", updateErr)
		}
	}

	if IsUnrecoverable(err) {
		return retrypool.DeadTaskActionAddToDeadTasks
	}

	return retrypool.DeadTaskActionForceRetry
}

func (tp *Tempolite) onHandlerRetry(attempt int, err error, task *retrypool.TaskWrapper[*Task]) {
	log.Printf("Retrying handler task with ID %s, attempt %d, error: %v", task.Data().ID, attempt, err)
	taskData := task.Data()
	taskData.RetryCount = attempt
	if err := tp.taskRepo.UpdateTask(tp.ctx, taskData); err != nil {
		log.Printf("Failed to update task retry count: %v", err)
	}
}

func (tp *Tempolite) onHandlerPanic(task *Task, v interface{}) {
	log.Printf("Handler panicked for task ID %s: %v", task.ID, v)
	task.Status = TaskStatusFailed
	errorMsg := fmt.Sprintf("Handler panicked: %v", v)
	if err := tp.taskRepo.UpdateTask(tp.ctx, task); err != nil {
		log.Printf("Failed to update task status after panic: %v", err)
	}

	node, nodeErr := tp.executionTreeRepo.GetNode(tp.ctx, task.ID)
	if nodeErr != nil {
		log.Printf("Failed to get execution node: %v", nodeErr)
	} else {
		node.Status = ExecutionStatusFailed
		node.ErrorMessage = &errorMsg
		if updateErr := tp.executionTreeRepo.UpdateNode(tp.ctx, node); updateErr != nil {
			log.Printf("Failed to update execution node: %v", updateErr)
		}
	}
}

func (tp *Tempolite) onSagaHandlerSuccess(controller retrypool.WorkerController[*Task], workerID int, worker retrypool.Worker[*Task], task *retrypool.TaskWrapper[*Task]) {
	log.Printf("Saga handler task with ID %s succeeded", task.Data().ID)
	taskData := task.Data()
	taskData.Status = TaskStatusCompleted
	now := time.Now()
	taskData.CompletedAt = &now
	if err := tp.taskRepo.UpdateTask(tp.ctx, taskData); err != nil {
		log.Printf("Failed to update saga task status: %v", err)
	}

	saga, err := tp.sagaRepo.GetSaga(tp.ctx, *taskData.SagaID)
	if err != nil {
		log.Printf("Failed to get saga: %v", err)
		return
	}

	saga.Status = SagaStatusCompleted
	saga.CompletedAt = taskData.CompletedAt
	if err := tp.sagaRepo.UpdateSaga(tp.ctx, saga); err != nil {
		log.Printf("Failed to update saga status: %v", err)
	}

	node, err := tp.executionTreeRepo.GetNode(tp.ctx, taskData.ID)
	if err != nil {
		log.Printf("Failed to get execution node: %v", err)
		return
	}

	node.Status = ExecutionStatusCompleted
	node.CompletedAt = taskData.CompletedAt
	node.Result = taskData.Result
	if err := tp.executionTreeRepo.UpdateNode(tp.ctx, node); err != nil {
		log.Printf("Failed to update execution node: %v", err)
	}
}

func (tp *Tempolite) onSagaHandlerFailure(controller retrypool.WorkerController[*Task], workerID int, worker retrypool.Worker[*Task], task *retrypool.TaskWrapper[*Task], err error) retrypool.DeadTaskAction {
	log.Printf("Saga handler task with ID %s failed: %v", task.Data().ID, err)
	taskData := task.Data()
	taskData.Status = TaskStatusFailed
	if err := tp.taskRepo.UpdateTask(tp.ctx, taskData); err != nil {
		log.Printf("Failed to update saga task status: %v", err)
	}

	saga, sagaErr := tp.sagaRepo.GetSaga(tp.ctx, *taskData.SagaID)
	if sagaErr != nil {
		log.Printf("Failed to get saga: %v", sagaErr)
	} else {
		saga.Status = SagaStatusFailed
		if updateErr := tp.sagaRepo.UpdateSaga(tp.ctx, saga); updateErr != nil {
			log.Printf("Failed to update saga status: %v", updateErr)
		}
	}

	node, nodeErr := tp.executionTreeRepo.GetNode(tp.ctx, taskData.ID)
	if nodeErr != nil {
		log.Printf("Failed to get execution node: %v", nodeErr)
	} else {
		node.Status = ExecutionStatusFailed
		errorMsg := err.Error()
		node.ErrorMessage = &errorMsg
		if updateErr := tp.executionTreeRepo.UpdateNode(tp.ctx, node); updateErr != nil {
			log.Printf("Failed to update execution node: %v", updateErr)
		}
	}

	if IsUnrecoverable(err) {
		return retrypool.DeadTaskActionAddToDeadTasks
	}

	compensations, compErr := tp.compensationRepo.GetCompensationsForSaga(tp.ctx, *taskData.SagaID)
	if compErr != nil {
		log.Printf("Failed to get compensations for saga: %v", compErr)
		return retrypool.DeadTaskActionForceRetry
	}

	for _, compensation := range compensations {
		log.Printf("Dispatching compensation task ID %s for saga ID %s", compensation.ID, *taskData.SagaID)
		tp.compensationPool.Dispatch(compensation)
	}

	return retrypool.DeadTaskActionForceRetry
}

func (tp *Tempolite) onSagaHandlerRetry(attempt int, err error, task *retrypool.TaskWrapper[*Task]) {
	log.Printf("Retrying saga handler task with ID %s, attempt %d, error: %v", task.Data().ID, attempt, err)
	taskData := task.Data()
	taskData.RetryCount = attempt
	if err := tp.taskRepo.UpdateTask(tp.ctx, taskData); err != nil {
		log.Printf("Failed to update saga task retry count: %v", err)
	}
}

func (tp *Tempolite) onSagaHandlerPanic(task *Task, v interface{}) {
	log.Printf("Saga handler panicked for task ID %s: %v", task.ID, v)
	task.Status = TaskStatusFailed
	errorMsg := fmt.Sprintf("Saga handler panicked: %v", v)
	if err := tp.taskRepo.UpdateTask(tp.ctx, task); err != nil {
		log.Printf("Failed to update saga task status after panic: %v", err)
	}

	node, nodeErr := tp.executionTreeRepo.GetNode(tp.ctx, task.ID)
	if nodeErr != nil {
		log.Printf("Failed to get execution node: %v", nodeErr)
	} else {
		node.Status = ExecutionStatusFailed
		node.ErrorMessage = &errorMsg
		if updateErr := tp.executionTreeRepo.UpdateNode(tp.ctx, node); updateErr != nil {
			log.Printf("Failed to update execution node: %v", updateErr)
		}
	}

	saga, sagaErr := tp.sagaRepo.GetSaga(tp.ctx, *task.SagaID)
	if sagaErr != nil {
		log.Printf("Failed to get saga: %v", sagaErr)
	} else {
		saga.Status = SagaStatusFailed
		if updateErr := tp.sagaRepo.UpdateSaga(tp.ctx, saga); updateErr != nil {
			log.Printf("Failed to update saga status: %v", updateErr)
		}
	}
}

func (tp *Tempolite) onCompensationSuccess(controller retrypool.WorkerController[*Compensation], workerID int, worker retrypool.Worker[*Compensation], task *retrypool.TaskWrapper[*Compensation]) {
	log.Printf("Compensation task with ID %s succeeded", task.Data().ID)
	compensationData := task.Data()
	compensationData.Status = ExecutionStatusCompleted
	if err := tp.compensationRepo.UpdateCompensation(tp.ctx, compensationData); err != nil {
		log.Printf("Failed to update compensation status: %v", err)
	}
}

func (tp *Tempolite) onCompensationFailure(controller retrypool.WorkerController[*Compensation], workerID int, worker retrypool.Worker[*Compensation], task *retrypool.TaskWrapper[*Compensation], err error) retrypool.DeadTaskAction {
	log.Printf("Compensation task with ID %s failed: %v", task.Data().ID, err)
	compensationData := task.Data()
	compensationData.Status = ExecutionStatusFailed
	if err := tp.compensationRepo.UpdateCompensation(tp.ctx, compensationData); err != nil {
		log.Printf("Failed to update compensation status: %v", err)
	}

	saga, sagaErr := tp.sagaRepo.GetSaga(tp.ctx, compensationData.SagaID)
	if sagaErr != nil {
		log.Printf("Failed to get saga: %v", sagaErr)
	} else {
		saga.Status = SagaStatusCriticallyFailed
		if updateErr := tp.sagaRepo.UpdateSaga(tp.ctx, saga); updateErr != nil {
			log.Printf("Failed to update saga status: %v", updateErr)
		}
	}

	if IsUnrecoverable(err) {
		return retrypool.DeadTaskActionAddToDeadTasks
	}

	return retrypool.DeadTaskActionForceRetry
}

func (tp *Tempolite) onCompensationRetry(attempt int, err error, task *retrypool.TaskWrapper[*Compensation]) {
	log.Printf("Retrying compensation task with ID %s, attempt %d, error: %v", task.Data().ID, attempt, err)
	compensationData := task.Data()
	if err := tp.compensationRepo.UpdateCompensation(tp.ctx, compensationData); err != nil {
		log.Printf("Failed to update compensation retry count: %v", err)
	}
}

func (tp *Tempolite) onCompensationPanic(compensation *Compensation, v interface{}) {
	log.Printf("Compensation panicked for ID %s: %v", compensation.ID, v)
	compensation.Status = ExecutionStatusFailed
	if err := tp.compensationRepo.UpdateCompensation(tp.ctx, compensation); err != nil {
		log.Printf("Failed to update compensation status after panic: %v", err)
	}

	saga, sagaErr := tp.sagaRepo.GetSaga(tp.ctx, compensation.SagaID)
	if sagaErr != nil {
		log.Printf("Failed to get saga: %v", sagaErr)
	} else {
		saga.Status = SagaStatusCriticallyFailed
		if updateErr := tp.sagaRepo.UpdateSaga(tp.ctx, saga); updateErr != nil {
			log.Printf("Failed to update saga status: %v", updateErr)
		}
	}
}

func (tp *Tempolite) onSideEffectSuccess(controller retrypool.WorkerController[*sideEffectTask], workerID int, worker retrypool.Worker[*sideEffectTask], task *retrypool.TaskWrapper[*sideEffectTask]) {
	log.Printf("Side effect completed successfully")
}

func (tp *Tempolite) onSideEffectFailure(controller retrypool.WorkerController[*sideEffectTask], workerID int, worker retrypool.Worker[*sideEffectTask], task *retrypool.TaskWrapper[*sideEffectTask], err error) retrypool.DeadTaskAction {
	log.Printf("Side effect failed: %v", err)

	if IsUnrecoverable(err) {
		return retrypool.DeadTaskActionAddToDeadTasks
	}

	return retrypool.DeadTaskActionForceRetry
}

func (tp *Tempolite) onSideEffectRetry(attempt int, err error, task *retrypool.TaskWrapper[*sideEffectTask]) {
	log.Printf("Retrying side effect, attempt %d: %v", attempt, err)
}

func (tp *Tempolite) onSideEffectPanic(sideEffect *sideEffectTask, v interface{}) {
	log.Printf("Side effect panicked: %v", v)
}

// Helper functions for execution tree management

func (tp *Tempolite) addNodeToExecutionTree(ctx context.Context, node *ExecutionNode) error {
	log.Printf("Adding node %s to execution tree", node.ID)
	tp.executionTreesMu.Lock()
	defer tp.executionTreesMu.Unlock()

	tree, exists := tp.executionTrees[node.ID]
	if !exists {
		tree = &dag.AcyclicGraph{}
		tp.executionTrees[node.ID] = tree
	}

	tree.Add(node)

	if node.ParentID != nil {
		parentNode, err := tp.executionTreeRepo.GetNode(ctx, *node.ParentID)
		if err != nil {
			log.Printf("Failed to get parent node: %v", err)
			return fmt.Errorf("failed to get parent node: %v", err)
		}
		tree.Connect(dag.BasicEdge(parentNode, node))
	}

	log.Printf("Node %s added to execution tree successfully", node.ID)
	return nil
}

func (tp *Tempolite) updateNodeInExecutionTree(ctx context.Context, node *ExecutionNode) error {
	log.Printf("Updating node %s in execution tree", node.ID)
	tp.executionTreesMu.Lock()
	defer tp.executionTreesMu.Unlock()

	tree, exists := tp.executionTrees[node.ID]
	if !exists {
		return fmt.Errorf("execution tree not found for node %s", node.ID)
	}

	// Remove the old node and add the updated one
	tree.Remove(node)
	tree.Add(node)

	log.Printf("Node %s updated in execution tree successfully", node.ID)
	return nil
}

/////////////////////////////////////////////////////
// SQLite Repository Implementations

type SQLiteTaskRepository struct {
	db *sql.DB
}

func NewSQLiteTaskRepository(db *sql.DB) (*SQLiteTaskRepository, error) {
	log.Printf("NewSQLiteTaskRepository: initializing with db: %v", db)
	repo := &SQLiteTaskRepository{db: db}
	if err := repo.initDB(); err != nil {
		log.Printf("NewSQLiteTaskRepository: failed to initialize DB: %v", err)
		return nil, err
	}
	log.Printf("NewSQLiteTaskRepository: successfully initialized")
	return repo, nil
}

func (r *SQLiteTaskRepository) initDB() error {
	log.Printf("initDB: creating tasks table")
	query := `
	CREATE TABLE IF NOT EXISTS tasks (
		id TEXT PRIMARY KEY,
		execution_context_id TEXT NOT NULL,
		handler_name TEXT NOT NULL,
		payload BLOB,
		status INTEGER NOT NULL,
		retry_count INTEGER NOT NULL,
		scheduled_at INTEGER NOT NULL,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		completed_at INTEGER,
		result BLOB,
		parent_task_id TEXT,
		saga_id TEXT,
		FOREIGN KEY(parent_task_id) REFERENCES tasks(id),
		FOREIGN KEY(saga_id) REFERENCES sagas(id)
	);
	CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
	CREATE INDEX IF NOT EXISTS idx_tasks_saga_id ON tasks(saga_id);
	`
	_, err := r.db.Exec(query)
	if err != nil {
		log.Printf("initDB: error creating tasks table: %v", err)
	} else {
		log.Printf("initDB: tasks table created or already exists")
	}
	return err
}

func (r *SQLiteTaskRepository) CreateTask(ctx context.Context, task *Task) error {
	log.Printf("CreateTask: inserting task: %v", task)
	query := `
	INSERT INTO tasks (id, execution_context_id, handler_name, payload, status, retry_count, scheduled_at, created_at, updated_at, completed_at, result, parent_task_id, saga_id)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`
	_, err := r.db.ExecContext(ctx, query,
		task.ID,
		task.ExecutionContextID,
		task.HandlerName,
		task.Payload,
		task.Status,
		task.RetryCount,
		task.ScheduledAt.Unix(),
		task.CreatedAt.Unix(),
		task.UpdatedAt.Unix(),
		nullableTime(task.CompletedAt),
		nullableBytes(task.Result), // Properly handle nil results
		nullableString(task.ParentTaskID),
		nullableString(task.SagaID),
	)
	if err != nil {
		log.Printf("CreateTask: error inserting task: %v", err)
	} else {
		log.Printf("CreateTask: task inserted successfully")
	}
	return err
}

func (r *SQLiteTaskRepository) GetTask(ctx context.Context, id string) (*Task, error) {
	log.Printf("GetTask: fetching task with id: %s", id)
	query := `
    SELECT id, execution_context_id, handler_name, payload, status, retry_count, scheduled_at, created_at, updated_at, completed_at, result, parent_task_id, saga_id
    FROM tasks
    WHERE id = ?;
    `
	var task Task
	var scheduledAt, createdAt, updatedAt int64
	var completedAt sql.NullInt64
	var result sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&task.ID,
		&task.ExecutionContextID,
		&task.HandlerName,
		&task.Payload,
		&task.Status,
		&task.RetryCount,
		&scheduledAt,
		&createdAt,
		&updatedAt,
		&completedAt,
		&result,
		&task.ParentTaskID,
		&task.SagaID,
	)
	if err != nil {
		log.Printf("GetTask: error fetching task: %v", err)
		return nil, err
	}
	task.ScheduledAt = time.Unix(scheduledAt, 0)
	task.CreatedAt = time.Unix(createdAt, 0)
	task.UpdatedAt = time.Unix(updatedAt, 0)
	if completedAt.Valid {
		completedTime := time.Unix(completedAt.Int64, 0)
		task.CompletedAt = &completedTime
	}
	if result.Valid {
		task.Result = []byte(result.String) // Store result as bytes if it's not null.
	}
	log.Printf("GetTask: fetched task: %v - result: %v", task, task.Result)
	return &task, nil
}

func (r *SQLiteTaskRepository) UpdateTask(ctx context.Context, task *Task) error {
	log.Printf("UpdateTask: updating task: %v", task)
	query := `
	UPDATE tasks
	SET execution_context_id = ?, handler_name = ?, payload = ?, status = ?, retry_count = ?, scheduled_at = ?, updated_at = ?, completed_at = ?, result = ?, parent_task_id = ?, saga_id = ?
	WHERE id = ?;
	`
	_, err := r.db.ExecContext(ctx, query,
		task.ExecutionContextID,
		task.HandlerName,
		task.Payload,
		task.Status,
		task.RetryCount,
		task.ScheduledAt.Unix(),
		task.UpdatedAt.Unix(),
		nullableTime(task.CompletedAt),
		nullableBytes(task.Result), // Properly handle nil results
		nullableString(task.ParentTaskID),
		nullableString(task.SagaID),
		task.ID,
	)
	if err != nil {
		log.Printf("UpdateTask: error updating task: %v", err)
	} else {
		log.Printf("UpdateTask: task updated successfully")
	}
	return err
}

func (r *SQLiteTaskRepository) GetPendingTasks(ctx context.Context, limit int) ([]*Task, error) {
	log.Printf("GetPendingTasks: fetching up to %d pending tasks", limit)
	query := `
	SELECT id, execution_context_id, handler_name, payload, status, retry_count, scheduled_at, created_at, updated_at, completed_at, result, parent_task_id, saga_id
	FROM tasks
	WHERE status = ? AND scheduled_at <= ?
	ORDER BY scheduled_at
	LIMIT ?;
	`
	rows, err := r.db.QueryContext(ctx, query, TaskStatusPending, time.Now().Unix(), limit)
	if err != nil {
		log.Printf("GetPendingTasks: error fetching tasks: %v", err)
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var task Task
		var scheduledAt, createdAt, updatedAt, completedAt int64
		err := rows.Scan(
			&task.ID,
			&task.ExecutionContextID,
			&task.HandlerName,
			&task.Payload,
			&task.Status,
			&task.RetryCount,
			&scheduledAt,
			&createdAt,
			&updatedAt,
			&completedAt,
			&task.Result,
			&task.ParentTaskID,
			&task.SagaID,
		)
		if err != nil {
			log.Printf("GetPendingTasks: error scanning row: %v", err)
			return nil, err
		}
		task.ScheduledAt = time.Unix(scheduledAt, 0)
		task.CreatedAt = time.Unix(createdAt, 0)
		task.UpdatedAt = time.Unix(updatedAt, 0)
		if completedAt != 0 {
			completedTime := time.Unix(completedAt, 0)
			task.CompletedAt = &completedTime
		}
		tasks = append(tasks, &task)
	}
	log.Printf("GetPendingTasks: fetched %d tasks", len(tasks))
	return tasks, nil
}

func (r *SQLiteTaskRepository) GetRunningTasksForSaga(ctx context.Context, sagaID string) ([]*Task, error) {
	log.Printf("GetRunningTasksForSaga: fetching running tasks for sagaID: %s", sagaID)
	query := `
	SELECT id, execution_context_id, handler_name, payload, status, retry_count, scheduled_at, created_at, updated_at, completed_at, result, parent_task_id, saga_id
	FROM tasks
	WHERE saga_id = ? AND status = ?;
	`
	rows, err := r.db.QueryContext(ctx, query, sagaID, TaskStatusInProgress)
	if err != nil {
		log.Printf("GetRunningTasksForSaga: error fetching tasks: %v", err)
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var task Task
		var scheduledAt, createdAt, updatedAt, completedAt int64
		err := rows.Scan(
			&task.ID,
			&task.ExecutionContextID,
			&task.HandlerName,
			&task.Payload,
			&task.Status,
			&task.RetryCount,
			&scheduledAt,
			&createdAt,
			&updatedAt,
			&completedAt,
			&task.Result,
			&task.ParentTaskID,
			&task.SagaID,
		)
		if err != nil {
			log.Printf("GetRunningTasksForSaga: error scanning row: %v", err)
			return nil, err
		}
		task.ScheduledAt = time.Unix(scheduledAt, 0)
		task.CreatedAt = time.Unix(createdAt, 0)
		task.UpdatedAt = time.Unix(updatedAt, 0)
		if completedAt != 0 {
			completedTime := time.Unix(completedAt, 0)
			task.CompletedAt = &completedTime
		}
		tasks = append(tasks, &task)
	}
	log.Printf("GetRunningTasksForSaga: fetched %d tasks", len(tasks))
	return tasks, nil
}

type SQLiteSideEffectRepository struct {
	db *sql.DB
}

func NewSQLiteSideEffectRepository(db *sql.DB) (*SQLiteSideEffectRepository, error) {
	log.Printf("NewSQLiteSideEffectRepository: initializing with db: %v", db)
	repo := &SQLiteSideEffectRepository{db: db}
	if err := repo.initDB(); err != nil {
		log.Printf("NewSQLiteSideEffectRepository: failed to initialize DB: %v", err)
		return nil, err
	}
	log.Printf("NewSQLiteSideEffectRepository: successfully initialized")
	return repo, nil
}

func (r *SQLiteSideEffectRepository) initDB() error {
	log.Printf("initDB: creating side_effects table")
	query := `
	CREATE TABLE IF NOT EXISTS side_effects (
		execution_context_id TEXT NOT NULL,
		key TEXT NOT NULL,
		result BLOB NOT NULL,
		created_at INTEGER NOT NULL,
		PRIMARY KEY (execution_context_id, key)
	);
	CREATE INDEX IF NOT EXISTS idx_side_effects_execution_context_id ON side_effects(execution_context_id);
	`
	_, err := r.db.Exec(query)
	if err != nil {
		log.Printf("initDB: error creating side_effects table: %v", err)
	} else {
		log.Printf("initDB: side_effects table created or already exists")
	}
	return err
}

func (r *SQLiteSideEffectRepository) GetSideEffect(ctx context.Context, executionContextID, key string) ([]byte, error) {
	log.Printf("GetSideEffect: fetching side effect for executionContextID: %s, key: %s", executionContextID, key)
	query := `
	SELECT result
	FROM side_effects
	WHERE execution_context_id = ? AND key = ?;
	`
	var result []byte
	err := r.db.QueryRowContext(ctx, query, executionContextID, key).Scan(&result)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("GetSideEffect: no rows found for executionContextID: %s, key: %s", executionContextID, key)
			return nil, nil
		}
		log.Printf("GetSideEffect: error fetching side effect: %v", err)
		return nil, err
	}
	log.Printf("GetSideEffect: fetched side effect for executionContextID: %s, key: %s result: [%v]", executionContextID, key, string(result))
	return result, nil
}

func (r *SQLiteSideEffectRepository) SaveSideEffect(ctx context.Context, executionContextID, key string, result []byte) error {
	log.Printf("SaveSideEffect: saving side effect for executionContextID: %s, key: %s", executionContextID, key)
	query := `
	INSERT OR REPLACE INTO side_effects (execution_context_id, key, result, created_at)
	VALUES (?, ?, ?, ?);
	`
	_, err := r.db.ExecContext(ctx, query, executionContextID, key, result, time.Now().Unix())
	if err != nil {
		log.Printf("SaveSideEffect: error saving side effect: %v", err)
	} else {
		log.Printf("SaveSideEffect: side effect saved successfully for executionContextID: %s, key: %s", executionContextID, key)
	}
	// id, err := res.RowsAffected()
	// if err != nil {
	// 	log.Printf("SaveSideEffect: error getting last insert ID: %v", err)
	// }
	// log.Printf("SaveSideEffect: last insert ID: %d", id)
	return err
}

type SQLiteSignalRepository struct {
	db *sql.DB
}

func NewSQLiteSignalRepository(db *sql.DB) (*SQLiteSignalRepository, error) {
	log.Printf("NewSQLiteSignalRepository: initializing with db: %v", db)
	repo := &SQLiteSignalRepository{db: db}
	if err := repo.initDB(); err != nil {
		log.Printf("NewSQLiteSignalRepository: failed to initialize DB: %v", err)
		return nil, err
	}
	log.Printf("NewSQLiteSignalRepository: successfully initialized")
	return repo, nil
}

func (r *SQLiteSignalRepository) initDB() error {
	log.Printf("initDB: creating signals table")
	query := `
	CREATE TABLE IF NOT EXISTS signals (
		id TEXT PRIMARY KEY,
		task_id TEXT NOT NULL,
		name TEXT NOT NULL,
		payload BLOB,
		created_at INTEGER NOT NULL,
		direction TEXT NOT NULL,
		FOREIGN KEY(task_id) REFERENCES tasks(id)
	);
	CREATE INDEX IF NOT EXISTS idx_signals_task_id ON signals(task_id);
	CREATE INDEX IF NOT EXISTS idx_signals_name ON signals(name);
	`
	_, err := r.db.Exec(query)
	if err != nil {
		log.Printf("initDB: error creating signals table: %v", err)
	} else {
		log.Printf("initDB: signals table created or already exists")
	}
	return err
}

func (r *SQLiteSignalRepository) SaveSignal(ctx context.Context, signal *Signal) error {
	log.Printf("SaveSignal: saving signal: %v", signal)
	query := `
	INSERT INTO signals (id, task_id, name, payload, created_at, direction)
	VALUES (?, ?, ?, ?, ?, ?);
	`
	_, err := r.db.ExecContext(ctx, query,
		signal.ID,
		signal.TaskID,
		signal.Name,
		signal.Payload,
		signal.CreatedAt.Unix(),
		signal.Direction,
	)
	if err != nil {
		log.Printf("SaveSignal: error saving signal: %v", err)
	} else {
		log.Printf("SaveSignal: signal saved successfully")
	}
	return err
}

func (r *SQLiteSignalRepository) GetSignals(ctx context.Context, taskID string, name string, direction string) ([]*Signal, error) {
	log.Printf("GetSignals: fetching signals for taskID: %s, name: %s, direction: %s", taskID, name, direction)
	query := `
	SELECT id, task_id, name, payload, created_at, direction
	FROM signals
	WHERE task_id = ? AND name = ? AND direction = ?;
	`
	rows, err := r.db.QueryContext(ctx, query, taskID, name, direction)
	if err != nil {
		log.Printf("GetSignals: error fetching signals: %v", err)
		return nil, err
	}
	defer rows.Close()

	var signals []*Signal
	for rows.Next() {
		var signal Signal
		var createdAt int64
		err := rows.Scan(
			&signal.ID,
			&signal.TaskID,
			&signal.Name,
			&signal.Payload,
			&createdAt,
			&signal.Direction,
		)
		if err != nil {
			log.Printf("GetSignals: error scanning row: %v", err)
			return nil, err
		}
		signal.CreatedAt = time.Unix(createdAt, 0)
		signals = append(signals, &signal)
	}
	log.Printf("GetSignals: fetched %d signals", len(signals))
	return signals, nil
}

func (r *SQLiteSignalRepository) DeleteSignals(ctx context.Context, taskID string, name string, direction string) error {
	log.Printf("DeleteSignals: deleting signals for taskID: %s, name: %s, direction: %s", taskID, name, direction)
	query := `
	DELETE FROM signals
	WHERE task_id = ? AND name = ? AND direction = ?;
	`
	_, err := r.db.ExecContext(ctx, query, taskID, name, direction)
	if err != nil {
		log.Printf("DeleteSignals: error deleting signals: %v", err)
	} else {
		log.Printf("DeleteSignals: signals deleted successfully")
	}
	return err
}

type SQLiteSagaRepository struct {
	db *sql.DB
}

func NewSQLiteSagaRepository(db *sql.DB) (*SQLiteSagaRepository, error) {
	log.Printf("NewSQLiteSagaRepository: initializing with db: %v", db)
	repo := &SQLiteSagaRepository{db: db}
	if err := repo.initDB(); err != nil {
		log.Printf("NewSQLiteSagaRepository: failed to initialize DB: %v", err)
		return nil, err
	}
	log.Printf("NewSQLiteSagaRepository: successfully initialized")
	return repo, nil
}

func (r *SQLiteSagaRepository) initDB() error {
	log.Printf("initDB: creating sagas table")
	query := `
	CREATE TABLE IF NOT EXISTS sagas (
		id TEXT PRIMARY KEY,
		status INTEGER NOT NULL,
		current_step INTEGER NOT NULL,
		created_at INTEGER NOT NULL,
		last_updated_at INTEGER NOT NULL,
		completed_at INTEGER,
		handler_name TEXT NOT NULL,
		cancel_requested BOOLEAN NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_sagas_status ON sagas(status);
	`
	_, err := r.db.Exec(query)
	if err != nil {
		log.Printf("initDB: error creating sagas table: %v", err)
	} else {
		log.Printf("initDB: sagas table created or already exists")
	}
	return err
}

func (r *SQLiteSagaRepository) CreateSaga(ctx context.Context, saga *SagaInfo) error {
	log.Printf("CreateSaga: creating saga: %v", saga)
	query := `
	INSERT INTO sagas (id, status, current_step, created_at, last_updated_at, completed_at, handler_name, cancel_requested)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?);
	`
	_, err := r.db.ExecContext(ctx, query,
		saga.ID,
		saga.Status,
		saga.CurrentStep,
		saga.CreatedAt.Unix(),
		saga.LastUpdatedAt.Unix(),
		nullableTime(saga.CompletedAt),
		saga.HandlerName,
		saga.CancelRequested,
	)
	if err != nil {
		log.Printf("CreateSaga: error creating saga: %v", err)
	} else {
		log.Printf("CreateSaga: saga created successfully")
	}
	return err
}

func (r *SQLiteSagaRepository) GetSaga(ctx context.Context, id string) (*SagaInfo, error) {
	log.Printf("GetSaga: fetching saga with id: %s", id)
	query := `
    SELECT id, status, current_step, created_at, last_updated_at, completed_at, handler_name, cancel_requested
    FROM sagas
    WHERE id = ?;
    `
	var saga SagaInfo
	var createdAt, lastUpdatedAt int64
	var completedAt sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&saga.ID,
		&saga.Status,
		&saga.CurrentStep,
		&createdAt,
		&lastUpdatedAt,
		&completedAt,
		&saga.HandlerName,
		&saga.CancelRequested,
	)
	if err != nil {
		log.Printf("GetSaga: error fetching saga: %v", err)
		return nil, err
	}
	saga.CreatedAt = time.Unix(createdAt, 0)
	saga.LastUpdatedAt = time.Unix(lastUpdatedAt, 0)
	if completedAt.Valid {
		completedTime := time.Unix(completedAt.Int64, 0)
		saga.CompletedAt = &completedTime
	}
	log.Printf("GetSaga: fetched saga: %v", saga)
	return &saga, nil
}

func (r *SQLiteSagaRepository) UpdateSaga(ctx context.Context, saga *SagaInfo) error {
	log.Printf("UpdateSaga: updating saga: %v", saga)
	query := `
	UPDATE sagas
	SET status = ?, current_step = ?, last_updated_at = ?, completed_at = ?, cancel_requested = ?
	WHERE id = ?;
	`
	_, err := r.db.ExecContext(ctx, query,
		saga.Status,
		saga.CurrentStep,
		saga.LastUpdatedAt.Unix(),
		nullableTime(saga.CompletedAt),
		saga.CancelRequested,
		saga.ID,
	)
	if err != nil {
		log.Printf("UpdateSaga: error updating saga: %v", err)
	} else {
		log.Printf("UpdateSaga: saga updated successfully")
	}
	return err
}

type SQLiteExecutionTreeRepository struct {
	db *sql.DB
}

func NewSQLiteExecutionTreeRepository(db *sql.DB) (*SQLiteExecutionTreeRepository, error) {
	log.Printf("NewSQLiteExecutionTreeRepository: initializing with db: %v", db)
	repo := &SQLiteExecutionTreeRepository{db: db}
	if err := repo.initDB(); err != nil {
		log.Printf("NewSQLiteExecutionTreeRepository: failed to initialize DB: %v", err)
		return nil, err
	}
	log.Printf("NewSQLiteExecutionTreeRepository: successfully initialized")
	return repo, nil
}

func (r *SQLiteExecutionTreeRepository) initDB() error {
	log.Printf("initDB: creating execution_nodes table")
	query := `
	CREATE TABLE IF NOT EXISTS execution_nodes (
		id TEXT PRIMARY KEY,
		parent_id TEXT,
		type INTEGER NOT NULL,
		status INTEGER NOT NULL,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		completed_at INTEGER,
		handler_name TEXT NOT NULL,
		payload BLOB,
		result BLOB,
		error_message TEXT,
		FOREIGN KEY(parent_id) REFERENCES execution_nodes(id)
	);
	CREATE INDEX IF NOT EXISTS idx_execution_nodes_parent_id ON execution_nodes(parent_id);
	CREATE INDEX IF NOT EXISTS idx_execution_nodes_status ON execution_nodes(status);
	`
	_, err := r.db.Exec(query)
	if err != nil {
		log.Printf("initDB: error creating execution_nodes table: %v", err)
	} else {
		log.Printf("initDB: execution_nodes table created or already exists")
	}
	return err
}

func (r *SQLiteExecutionTreeRepository) CreateNode(ctx context.Context, node *ExecutionNode) error {
	log.Printf("CreateNode: creating node: %v", node)
	query := `
	INSERT INTO execution_nodes (id, parent_id, type, status, created_at, updated_at, completed_at, handler_name, payload, result, error_message)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`
	_, err := r.db.ExecContext(ctx, query,
		node.ID,
		nullableString(node.ParentID),
		node.Type,
		node.Status,
		node.CreatedAt.Unix(),
		node.UpdatedAt.Unix(),
		nullableTime(node.CompletedAt),
		node.HandlerName,
		node.Payload,
		node.Result,
		nullableString(node.ErrorMessage),
	)
	if err != nil {
		log.Printf("CreateNode: error creating node: %v", err)
	} else {
		log.Printf("CreateNode: node created successfully")
	}
	return err
}

func (r *SQLiteExecutionTreeRepository) GetNode(ctx context.Context, id string) (*ExecutionNode, error) {
	log.Printf("GetNode: fetching node with id: %s", id)
	query := `
    SELECT id, parent_id, type, status, created_at, updated_at, completed_at, handler_name, payload, result, error_message
    FROM execution_nodes
    WHERE id = ?;
    `
	var node ExecutionNode
	var createdAt, updatedAt int64
	var completedAt sql.NullInt64
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&node.ID,
		&node.ParentID,
		&node.Type,
		&node.Status,
		&createdAt,
		&updatedAt,
		&completedAt,
		&node.HandlerName,
		&node.Payload,
		&node.Result,
		&node.ErrorMessage,
	)
	if err != nil {
		log.Printf("GetNode: error fetching node: %v", err)
		return nil, err
	}
	node.CreatedAt = time.Unix(createdAt, 0)
	node.UpdatedAt = time.Unix(updatedAt, 0)
	if completedAt.Valid {
		completedTime := time.Unix(completedAt.Int64, 0)
		node.CompletedAt = &completedTime
	}
	log.Printf("GetNode: fetched node: %v", node)
	return &node, nil
}

func (r *SQLiteExecutionTreeRepository) UpdateNode(ctx context.Context, node *ExecutionNode) error {
	log.Printf("UpdateNode: updating node: %v", node)
	query := `
	UPDATE execution_nodes
	SET parent_id = ?, type = ?, status = ?, updated_at = ?, completed_at = ?, handler_name = ?, payload = ?, result = ?, error_message = ?
	WHERE id = ?;
	`
	_, err := r.db.ExecContext(ctx, query,
		nullableString(node.ParentID),
		node.Type,
		node.Status,
		node.UpdatedAt.Unix(),
		nullableTime(node.CompletedAt),
		node.HandlerName,
		node.Payload,
		node.Result,
		nullableString(node.ErrorMessage),
		node.ID,
	)
	if err != nil {
		log.Printf("UpdateNode: error updating node: %v", err)
	} else {
		log.Printf("UpdateNode: node updated successfully")
	}
	return err
}

func (r *SQLiteExecutionTreeRepository) GetChildNodes(ctx context.Context, parentID string) ([]*ExecutionNode, error) {
	log.Printf("GetChildNodes: fetching child nodes for parentID: %s", parentID)
	query := `
	SELECT id, parent_id, type, status, created_at, updated_at, completed_at, handler_name, payload, result, error_message
	FROM execution_nodes
	WHERE parent_id = ?;
	`
	rows, err := r.db.QueryContext(ctx, query, parentID)
	if err != nil {
		log.Printf("GetChildNodes: error fetching child nodes: %v", err)
		return nil, err
	}
	defer rows.Close()

	var nodes []*ExecutionNode
	for rows.Next() {
		var node ExecutionNode
		var createdAt, updatedAt, completedAt int64
		err := rows.Scan(
			&node.ID,
			&node.ParentID,
			&node.Type,
			&node.Status,
			&createdAt,
			&updatedAt,
			&completedAt,
			&node.HandlerName,
			&node.Payload,
			&node.Result,
			&node.ErrorMessage,
		)
		if err != nil {
			log.Printf("GetChildNodes: error scanning row: %v", err)
			return nil, err
		}
		node.CreatedAt = time.Unix(createdAt, 0)
		node.UpdatedAt = time.Unix(updatedAt, 0)
		if completedAt != 0 {
			completedTime := time.Unix(completedAt, 0)
			node.CompletedAt = &completedTime
		}
		nodes = append(nodes, &node)
	}
	log.Printf("GetChildNodes: fetched %d child nodes", len(nodes))
	return nodes, nil
}

type SQLiteCompensationRepository struct {
	db *sql.DB
}

func NewSQLiteCompensationRepository(db *sql.DB) (*SQLiteCompensationRepository, error) {
	log.Printf("NewSQLiteCompensationRepository: initializing with db: %v", db)
	repo := &SQLiteCompensationRepository{db: db}
	if err := repo.initDB(); err != nil {
		log.Printf("NewSQLiteCompensationRepository: failed to initialize DB: %v", err)
		return nil, err
	}
	log.Printf("NewSQLiteCompensationRepository: successfully initialized")
	return repo, nil
}

func (r *SQLiteCompensationRepository) initDB() error {
	log.Printf("initDB: creating compensations table")
	query := `
	CREATE TABLE IF NOT EXISTS compensations (
		id TEXT PRIMARY KEY,
		saga_id TEXT NOT NULL,
		step_index INTEGER NOT NULL,
		payload BLOB,
		status INTEGER NOT NULL,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		FOREIGN KEY(saga_id) REFERENCES sagas(id)
	);
	CREATE INDEX IF NOT EXISTS idx_compensations_saga_id ON compensations(saga_id);
	`
	_, err := r.db.Exec(query)
	if err != nil {
		log.Printf("initDB: error creating compensations table: %v", err)
	} else {
		log.Printf("initDB: compensations table created or already exists")
	}
	return err
}

func (r *SQLiteCompensationRepository) CreateCompensation(ctx context.Context, compensation *Compensation) error {
	log.Printf("CreateCompensation: creating compensation: %v", compensation)
	query := `
	INSERT INTO compensations (id, saga_id, step_index, payload, status, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?);
	`
	_, err := r.db.ExecContext(ctx, query,
		compensation.ID,
		compensation.SagaID,
		compensation.StepIndex,
		compensation.Payload,
		compensation.Status,
		compensation.CreatedAt.Unix(),
		compensation.UpdatedAt.Unix(),
	)
	if err != nil {
		log.Printf("CreateCompensation: error creating compensation: %v", err)
	} else {
		log.Printf("CreateCompensation: compensation created successfully")
	}
	return err
}

func (r *SQLiteCompensationRepository) GetCompensation(ctx context.Context, id string) (*Compensation, error) {
	log.Printf("GetCompensation: fetching compensation with id: %s", id)
	query := `
	SELECT id, saga_id, step_index, payload, status, created_at, updated_at
	FROM compensations
	WHERE id = ?;
	`
	var compensation Compensation
	var createdAt, updatedAt int64
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&compensation.ID,
		&compensation.SagaID,
		&compensation.StepIndex,
		&compensation.Payload,
		&compensation.Status,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		log.Printf("GetCompensation: error fetching compensation: %v", err)
		return nil, err
	}
	compensation.CreatedAt = time.Unix(createdAt, 0)
	compensation.UpdatedAt = time.Unix(updatedAt, 0)
	log.Printf("GetCompensation: fetched compensation: %v", compensation)
	return &compensation, nil
}

func (r *SQLiteCompensationRepository) UpdateCompensation(ctx context.Context, compensation *Compensation) error {
	log.Printf("UpdateCompensation: updating compensation: %v", compensation)
	query := `
	UPDATE compensations
	SET saga_id = ?, step_index = ?, payload = ?, status = ?, updated_at = ?
	WHERE id = ?;
	`
	_, err := r.db.ExecContext(ctx, query,
		compensation.SagaID,
		compensation.StepIndex,
		compensation.Payload,
		compensation.Status,
		compensation.UpdatedAt.Unix(),
		compensation.ID,
	)
	if err != nil {
		log.Printf("UpdateCompensation: error updating compensation: %v", err)
	} else {
		log.Printf("UpdateCompensation: compensation updated successfully")
	}
	return err
}

func (r *SQLiteCompensationRepository) GetCompensationsForSaga(ctx context.Context, sagaID string) ([]*Compensation, error) {
	log.Printf("GetCompensationsForSaga: fetching compensations for sagaID: %s", sagaID)
	query := `
	SELECT id, saga_id, step_index, payload, status, created_at, updated_at
	FROM compensations
	WHERE saga_id = ?
	ORDER BY step_index DESC;
	`
	rows, err := r.db.QueryContext(ctx, query, sagaID)
	if err != nil {
		log.Printf("GetCompensationsForSaga: error fetching compensations: %v", err)
		return nil, err
	}
	defer rows.Close()

	var compensations []*Compensation
	for rows.Next() {
		var compensation Compensation
		var createdAt, updatedAt int64
		err := rows.Scan(
			&compensation.ID,
			&compensation.SagaID,
			&compensation.StepIndex,
			&compensation.Payload,
			&compensation.Status,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			log.Printf("GetCompensationsForSaga: error scanning row: %v", err)
			return nil, err
		}
		compensation.CreatedAt = time.Unix(createdAt, 0)
		compensation.UpdatedAt = time.Unix(updatedAt, 0)
		compensations = append(compensations, &compensation)
	}
	log.Printf("GetCompensationsForSaga: fetched %d compensations", len(compensations))
	return compensations, nil
}

type SQLiteSagaStepRepository struct {
	db *sql.DB
}

func NewSQLiteSagaStepRepository(db *sql.DB) (*SQLiteSagaStepRepository, error) {
	log.Printf("NewSQLiteSagaStepRepository: initializing with db: %v", db)
	repo := &SQLiteSagaStepRepository{db: db}
	if err := repo.initDB(); err != nil {
		log.Printf("NewSQLiteSagaStepRepository: failed to initialize DB: %v", err)
		return nil, err
	}
	log.Printf("NewSQLiteSagaStepRepository: successfully initialized")
	return repo, nil
}

func (r *SQLiteSagaStepRepository) initDB() error {
	log.Printf("initDB: creating saga_steps table")
	query := `
	CREATE TABLE IF NOT EXISTS saga_steps (
		saga_id TEXT NOT NULL,
		step_index INTEGER NOT NULL,
		payload BLOB,
		PRIMARY KEY (saga_id, step_index),
		FOREIGN KEY(saga_id) REFERENCES sagas(id)
	);
	`
	_, err := r.db.Exec(query)
	if err != nil {
		log.Printf("initDB: error creating saga_steps table: %v", err)
	} else {
		log.Printf("initDB: saga_steps table created or already exists")
	}
	return err
}

func (r *SQLiteSagaStepRepository) CreateSagaStep(ctx context.Context, sagaID string, stepIndex int, payload []byte) error {
	log.Printf("CreateSagaStep: creating saga step for sagaID: %s, stepIndex: %d", sagaID, stepIndex)
	query := `
	INSERT INTO saga_steps (saga_id, step_index, payload)
	VALUES (?, ?, ?);
	`
	_, err := r.db.ExecContext(ctx, query, sagaID, stepIndex, payload)
	if err != nil {
		log.Printf("CreateSagaStep: error creating saga step: %v", err)
	} else {
		log.Printf("CreateSagaStep: saga step created successfully for sagaID: %s, stepIndex: %d", sagaID, stepIndex)
	}
	return err
}

func (r *SQLiteSagaStepRepository) GetSagaStep(ctx context.Context, sagaID string, stepIndex int) ([]byte, error) {
	log.Printf("GetSagaStep: fetching saga step for sagaID: %s, stepIndex: %d", sagaID, stepIndex)
	query := `
	SELECT payload
	FROM saga_steps
	WHERE saga_id = ? AND step_index = ?;
	`
	var payload []byte
	err := r.db.QueryRowContext(ctx, query, sagaID, stepIndex).Scan(&payload)
	if err != nil {
		log.Printf("GetSagaStep: error fetching saga step: %v", err)
		return nil, err
	}
	log.Printf("GetSagaStep: fetched saga step for sagaID: %s, stepIndex: %d", sagaID, stepIndex)
	return payload, nil
}

func nullableBytes(data []byte) interface{} {
	if data == nil {
		return nil
	}
	return data
}
