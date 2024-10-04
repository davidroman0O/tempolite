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

// SQLite Repository Implementations
// ... (SQLite repository implementations remain the same)

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
	sideEffectPool    *retrypool.Pool[SideEffect]
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
		workers := make([]retrypool.Worker[*Task], count)
		for i := 0; i < count; i++ {
			workers[i] = &TaskWorker{ID: i, tp: tp}
		}
		tp.handlerPool = retrypool.New(tp.ctx, workers, tp.getHandlerPoolOptions()...)
	}
}

func WithSagaWorkers(count int) TempoliteOption {
	return func(tp *Tempolite) {
		workers := make([]retrypool.Worker[*Task], count)
		for i := 0; i < count; i++ {
			workers[i] = &SagaTaskWorker{ID: i, tp: tp}
		}
		tp.sagaHandlerPool = retrypool.New(tp.ctx, workers, tp.getSagaHandlerPoolOptions()...)
	}
}

func WithCompensationWorkers(count int) TempoliteOption {
	return func(tp *Tempolite) {
		workers := make([]retrypool.Worker[*Compensation], count)
		for i := 0; i < count; i++ {
			workers[i] = &CompensationWorker{ID: i, tp: tp}
		}
		tp.compensationPool = retrypool.New(tp.ctx, workers, tp.getCompensationPoolOptions()...)
	}
}

func WithSideEffectWorkers(count int) TempoliteOption {
	return func(tp *Tempolite) {
		workers := make([]retrypool.Worker[SideEffect], count)
		for i := 0; i < count; i++ {
			workers[i] = &SideEffectWorker{ID: i, tp: tp}
		}
		tp.sideEffectPool = retrypool.New(tp.ctx, workers, tp.getSideEffectPoolOptions()...)
	}
}

func (tp *Tempolite) getHandlerPoolOptions() []retrypool.Option[*Task] {
	return []retrypool.Option[*Task]{
		retrypool.WithOnTaskSuccess[*Task](tp.onHandlerSuccess),
		retrypool.WithOnTaskFailure[*Task](tp.onHandlerFailure),
		retrypool.WithOnRetry[*Task](tp.onHandlerRetry),
		retrypool.WithAttempts[*Task](3),
		retrypool.WithPanicHandler[*Task](tp.onHandlerPanic),
	}
}

func (tp *Tempolite) getSagaHandlerPoolOptions() []retrypool.Option[*Task] {
	return []retrypool.Option[*Task]{
		retrypool.WithOnTaskSuccess[*Task](tp.onSagaHandlerSuccess),
		retrypool.WithOnTaskFailure[*Task](tp.onSagaHandlerFailure),
		retrypool.WithOnRetry[*Task](tp.onSagaHandlerRetry),
		retrypool.WithAttempts[*Task](3),
		retrypool.WithPanicHandler[*Task](tp.onSagaHandlerPanic),
	}
}

func (tp *Tempolite) getCompensationPoolOptions() []retrypool.Option[*Compensation] {
	return []retrypool.Option[*Compensation]{
		retrypool.WithOnTaskSuccess[*Compensation](tp.onCompensationSuccess),
		retrypool.WithOnTaskFailure[*Compensation](tp.onCompensationFailure),
		retrypool.WithOnRetry[*Compensation](tp.onCompensationRetry),
		retrypool.WithAttempts[*Compensation](3),
		retrypool.WithPanicHandler[*Compensation](tp.onCompensationPanic),
	}
}

func (tp *Tempolite) getSideEffectPoolOptions() []retrypool.Option[SideEffect] {
	return []retrypool.Option[SideEffect]{
		retrypool.WithOnTaskSuccess[SideEffect](tp.onSideEffectSuccess),
		retrypool.WithOnTaskFailure[SideEffect](tp.onSideEffectFailure),
		retrypool.WithOnRetry[SideEffect](tp.onSideEffectRetry),
		retrypool.WithAttempts[SideEffect](3),
		retrypool.WithPanicHandler[SideEffect](tp.onSideEffectPanic),
	}
}

func New(ctx context.Context, db *sql.DB, options ...TempoliteOption) (*Tempolite, error) {
	ctx, cancel := context.WithCancel(ctx)

	tp := &Tempolite{
		db:             db,
		handlers:       make(map[string]handlerInfo),
		ctx:            ctx,
		cancel:         cancel,
		executionTrees: make(map[string]*dag.AcyclicGraph),
	}

	var err error

	tp.taskRepo, err = NewSQLiteTaskRepository(db)
	if err != nil {
		return nil, err
	}

	tp.sideEffectRepo, err = NewSQLiteSideEffectRepository(db)
	if err != nil {
		return nil, err
	}

	tp.signalRepo, err = NewSQLiteSignalRepository(db)
	if err != nil {
		return nil, err
	}

	tp.sagaRepo, err = NewSQLiteSagaRepository(db)
	if err != nil {
		return nil, err
	}

	tp.executionTreeRepo, err = NewSQLiteExecutionTreeRepository(db)
	if err != nil {
		return nil, err
	}

	tp.compensationRepo, err = NewSQLiteCompensationRepository(db)
	if err != nil {
		return nil, err
	}

	tp.sagaStepRepo, err = NewSQLiteSagaStepRepository(db)
	if err != nil {
		return nil, err
	}

	// Apply options
	for _, option := range options {
		option(tp)
	}

	// Initialize pools if not set by options
	if tp.handlerPool == nil {
		WithHandlerWorkers(1)(tp)
	}
	if tp.sagaHandlerPool == nil {
		WithSagaWorkers(1)(tp)
	}
	if tp.compensationPool == nil {
		WithCompensationWorkers(1)(tp)
	}
	if tp.sideEffectPool == nil {
		WithSideEffectWorkers(1)(tp)
	}

	return tp, nil
}

func (tp *Tempolite) RegisterHandler(handler interface{}) {
	handlerType := reflect.TypeOf(handler)
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
	tp.handlersMutex.RLock()
	defer tp.handlersMutex.RUnlock()
	handler, exists := tp.handlers[name]
	return handler, exists
}

func (tp *Tempolite) GetInfo(ctx context.Context, id string) (interface{}, error) {
	// Try to get task info
	task, err := tp.taskRepo.GetTask(ctx, id)
	if err == nil {
		return task, nil
	}

	// Try to get saga info
	saga, err := tp.sagaRepo.GetSaga(ctx, id)
	if err == nil {
		return saga, nil
	}

	// Try to get side effect info
	sideEffect, err := tp.sideEffectRepo.GetSideEffect(ctx, id, "")
	if err == nil {
		return sideEffect, nil
	}

	return nil, fmt.Errorf("no info found for id: %s", id)
}

func (tp *Tempolite) GetExecutionTree(ctx context.Context, rootID string) (*dag.AcyclicGraph, error) {
	tp.executionTreesMu.RLock()
	tree, exists := tp.executionTrees[rootID]
	tp.executionTreesMu.RUnlock()

	if exists {
		return tree, nil
	}

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

	return tree, nil
}

func (tp *Tempolite) reconstructExecutionTree(ctx context.Context, node *ExecutionNode, tree *dag.AcyclicGraph) error {
	tree.Add(node)

	children, err := tp.executionTreeRepo.GetChildNodes(ctx, node.ID)
	if err != nil {
		return err
	}

	for _, child := range children {
		err = tp.reconstructExecutionTree(ctx, child, tree)
		if err != nil {
			return err
		}
		tree.Connect(dag.BasicEdge(node, child))
	}

	return nil
}

func (tp *Tempolite) SendSignal(ctx context.Context, taskID string, name string, payload interface{}) error {
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

	return tp.signalRepo.SaveSignal(ctx, signal)
}

func (tp *Tempolite) ReceiveSignal(ctx context.Context, taskID string, name string) (<-chan []byte, error) {
	ch := make(chan []byte)

	go func() {
		defer close(ch)

		for {
			select {
			case <-ctx.Done():
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
	// Try to cancel task
	task, err := tp.taskRepo.GetTask(ctx, id)
	if err == nil {
		task.Status = TaskStatusCancelled
		return tp.taskRepo.UpdateTask(ctx, task)
	}

	// Try to cancel saga
	saga, err := tp.sagaRepo.GetSaga(ctx, id)
	if err == nil {
		saga.Status = SagaStatusCancelled
		return tp.sagaRepo.UpdateSaga(ctx, saga)
	}

	return fmt.Errorf("no task or saga found with id: %s", id)
}

func (tp *Tempolite) Terminate(ctx context.Context, id string) error {
	// Try to terminate task
	task, err := tp.taskRepo.GetTask(ctx, id)
	if err == nil {
		task.Status = TaskStatusTerminated
		return tp.taskRepo.UpdateTask(ctx, task)
	}

	// Try to terminate saga
	saga, err := tp.sagaRepo.GetSaga(ctx, id)
	if err == nil {
		saga.Status = SagaStatusTerminated
		return tp.sagaRepo.UpdateSaga(ctx, saga)
	}

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
		o.maxDuration = duration
	}
}

func WithTimeLimit(limit time.Duration) EnqueueOption {
	return func(o *enqueueOptions) {
		o.timeLimit = limit
	}
}

func WithImmediateRetry() EnqueueOption {
	return func(o *enqueueOptions) {
		o.immediate = true
	}
}

func WithPanicOnTimeout() EnqueueOption {
	return func(o *enqueueOptions) {
		o.panicOnTimeout = true
	}
}

func (tp *Tempolite) Enqueue(ctx context.Context, handler interface{}, params interface{}, options ...EnqueueOption) (string, error) {
	handlerName := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
	handlerInfo, exists := tp.getHandler(handlerName)
	if !exists {
		return "", fmt.Errorf("no handler registered with name: %s", handlerName)
	}

	opts := enqueueOptions{}
	for _, option := range options {
		option(&opts)
	}

	payload, err := json.Marshal(params)
	if err != nil {
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
		saga := &SagaInfo{
			ID:            task.ID,
			Status:        SagaStatusPending,
			CreatedAt:     time.Now(),
			LastUpdatedAt: time.Now(),
			HandlerName:   handlerName,
		}
		if err := tp.sagaRepo.CreateSaga(ctx, saga); err != nil {
			return "", fmt.Errorf("failed to create saga: %v", err)
		}
		task.SagaID = &saga.ID
	}

	if err := tp.taskRepo.CreateTask(ctx, task); err != nil {
		return "", fmt.Errorf("failed to create task: %v", err)
	}

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
		return "", fmt.Errorf("failed to create execution node: %v", err)
	}

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

	return task.ID, nil
}

func (tp *Tempolite) EnqueueSaga(ctx context.Context, handler HandlerSagaFunc, params interface{}, options ...EnqueueOption) (string, error) {
	return tp.Enqueue(ctx, handler, params, options...)
}

func (tp *Tempolite) Wait(condition func(TempoliteInfo) bool, interval time.Duration) error {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-tp.ctx.Done():
			return tp.ctx.Err()
		case <-ticker.C:
			info := tp.getInfo()
			if condition(info) {
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
	return TempoliteInfo{
		Tasks:             tp.handlerPool.QueueSize(),
		SagaTasks:         tp.sagaHandlerPool.QueueSize(),
		CompensationTasks: tp.compensationPool.QueueSize(),
		SideEffectTasks:   tp.sideEffectPool.QueueSize(),
	}
}

func (tp *Tempolite) GetPoolStats() map[string]int {
	return map[string]int{
		"handler":      tp.handlerPool.QueueSize(),
		"saga":         tp.sagaHandlerPool.QueueSize(),
		"compensation": tp.compensationPool.QueueSize(),
		"sideeffect":   tp.sideEffectPool.QueueSize(),
	}
}

func (tp *Tempolite) Close() error {
	tp.cancel()
	tp.handlerPool.Close()
	tp.sagaHandlerPool.Close()
	tp.compensationPool.Close()
	tp.sideEffectPool.Close()
	tp.workersWg.Wait()
	return nil
}

// Worker implementations

type TaskWorker struct {
	ID int
	tp *Tempolite
}

func (w *TaskWorker) Run(ctx context.Context, task *Task) error {
	handlerInfo, exists := w.tp.getHandler(task.HandlerName)
	if !exists {
		return fmt.Errorf("no handler registered with name: %s", task.HandlerName)
	}

	handlerValue := reflect.ValueOf(handlerInfo.Handler)
	paramType := handlerValue.Type().In(1)
	param := reflect.New(paramType).Interface()

	err := json.Unmarshal(task.Payload, param)
	if err != nil {
		return fmt.Errorf("failed to unmarshal task payload: %v", err)
	}

	handlerCtx := &handlerContext{
		Context: ctx,
		tp:      w.tp,
		taskID:  task.ID,
	}

	results := handlerValue.Call([]reflect.Value{
		reflect.ValueOf(handlerCtx),
		reflect.ValueOf(param).Elem(),
	})

	if len(results) > 0 && !results[len(results)-1].IsNil() {
		return results[len(results)-1].Interface().(error)
	}

	if len(results) > 1 {
		result, err := json.Marshal(results[0].Interface())
		if err != nil {
			return fmt.Errorf("failed to marshal task result: %v", err)
		}
		task.Result = result
	}

	return nil
}

type SagaTaskWorker struct {
	ID int
	tp *Tempolite
}

func (w *SagaTaskWorker) Run(ctx context.Context, task *Task) error {
	handlerInfo, exists := w.tp.getHandler(task.HandlerName)
	if !exists {
		return fmt.Errorf("no handler registered with name: %s", task.HandlerName)
	}

	handlerValue := reflect.ValueOf(handlerInfo.Handler)
	paramType := handlerValue.Type().In(1)
	param := reflect.New(paramType).Interface()

	err := json.Unmarshal(task.Payload, param)
	if err != nil {
		return fmt.Errorf("failed to unmarshal task payload: %v", err)
	}

	handlerCtx := &handlerSagaContext{
		handlerContext: handlerContext{
			Context: ctx,
			tp:      w.tp,
			taskID:  task.ID,
		},
		sagaID: *task.SagaID,
	}

	results := handlerValue.Call([]reflect.Value{
		reflect.ValueOf(handlerCtx),
		reflect.ValueOf(param).Elem(),
	})

	if len(results) > 0 && !results[len(results)-1].IsNil() {
		return results[len(results)-1].Interface().(error)
	}

	if len(results) > 1 {
		result, err := json.Marshal(results[0].Interface())
		if err != nil {
			return fmt.Errorf("failed to marshal task result: %v", err)
		}
		task.Result = result
	}

	return nil
}

type CompensationWorker struct {
	ID int
	tp *Tempolite
}

func (w *CompensationWorker) Run(ctx context.Context, compensation *Compensation) error {
	saga, err := w.tp.sagaRepo.GetSaga(ctx, compensation.SagaID)
	if err != nil {
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
		return fmt.Errorf("failed to unmarshal compensation payload: %v", err)
	}

	compensationCtx := &compensationContext{
		Context: ctx,
		tp:      w.tp,
		sagaID:  compensation.SagaID,
		stepID:  compensation.ID,
	}

	results := handlerValue.Call([]reflect.Value{
		reflect.ValueOf(compensationCtx),
		reflect.ValueOf(param).Elem(),
	})

	if len(results) > 0 && !results[len(results)-1].IsNil() {
		return results[len(results)-1].Interface().(error)
	}

	return nil
}

type SideEffectWorker struct {
	ID int
	tp *Tempolite
}

func (w *SideEffectWorker) Run(ctx context.Context, sideEffect SideEffect) error {
	sideEffectCtx := &sideEffectContext{
		Context: ctx,
		tp:      w.tp,
	}

	result, err := sideEffect.Run(sideEffectCtx)
	if err != nil {
		return err
	}

	resultBytes, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal side effect result: %v", err)
	}

	err = w.tp.sideEffectRepo.SaveSideEffect(ctx, sideEffectCtx.GetID(), "result", resultBytes)
	if err != nil {
		return fmt.Errorf("failed to save side effect result: %v", err)
	}

	return nil
}

// Context implementations

type handlerContext struct {
	context.Context
	tp     *Tempolite
	taskID string
}

func (c *handlerContext) GetID() string {
	return c.taskID
}

func (c *handlerContext) EnqueueTask(handler HandlerFunc, params interface{}, options ...EnqueueOption) (string, error) {
	taskID, err := c.tp.Enqueue(c, handler, params, options...)
	if err != nil {
		return "", err
	}

	parentNode, err := c.tp.executionTreeRepo.GetNode(c, c.taskID)
	if err != nil {
		return "", fmt.Errorf("failed to get parent node: %v", err)
	}

	childNode, err := c.tp.executionTreeRepo.GetNode(c, taskID)
	if err != nil {
		return "", fmt.Errorf("failed to get child node: %v", err)
	}

	childNode.ParentID = &parentNode.ID
	if err := c.tp.executionTreeRepo.UpdateNode(c, childNode); err != nil {
		return "", fmt.Errorf("failed to update child node: %v", err)
	}

	return taskID, nil
}

func (c *handlerContext) EnqueueTaskAndWait(handler HandlerFunc, params interface{}, options ...EnqueueOption) (interface{}, error) {
	taskID, err := c.EnqueueTask(handler, params, options...)
	if err != nil {
		return nil, err
	}

	return c.WaitForCompletion(taskID)
}

func (c *handlerContext) SideEffect(key string, effect SideEffect) (interface{}, error) {
	result, err := c.tp.sideEffectRepo.GetSideEffect(c, c.taskID, key)
	if err == nil {
		var value interface{}
		if err := json.Unmarshal(result, &value); err != nil {
			return nil, fmt.Errorf("failed to unmarshal side effect result: %v", err)
		}
		return value, nil
	}

	c.tp.sideEffectPool.Dispatch(effect)

	result, err = c.tp.sideEffectRepo.GetSideEffect(c, c.taskID, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get side effect result: %v", err)
	}

	var value interface{}
	if err := json.Unmarshal(result, &value); err != nil {
		return nil, fmt.Errorf("failed to unmarshal side effect result: %v", err)
	}

	return value, nil
}

func (c *handlerContext) SendSignal(name string, payload interface{}) error {
	return c.tp.SendSignal(c, c.taskID, name, payload)
}

func (c *handlerContext) ReceiveSignal(name string) (<-chan []byte, error) {
	return c.tp.ReceiveSignal(c, c.taskID, name)
}

func (c *handlerContext) WaitForCompletion(taskID string) (interface{}, error) {
	for {
		select {
		case <-c.Done():
			return nil, c.Err()
		case <-time.After(time.Second):
			task, err := c.tp.taskRepo.GetTask(c, taskID)
			if err != nil {
				return nil, err
			}

			switch task.Status {
			case TaskStatusCompleted:
				var result interface{}
				if err := json.Unmarshal(task.Result, &result); err != nil {
					return nil, fmt.Errorf("failed to unmarshal task result: %v", err)
				}
				return result, nil
			case TaskStatusFailed:
				return nil, fmt.Errorf("task failed")
			case TaskStatusCancelled:
				return nil, fmt.Errorf("task cancelled")
			case TaskStatusTerminated:
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
	saga, err := c.tp.sagaRepo.GetSaga(c, c.sagaID)
	if err != nil {
		return fmt.Errorf("failed to get saga: %v", err)
	}

	transactionCtx := &transactionContext{
		Context: c.Context,
		tp:      c.tp,
		sagaID:  c.sagaID,
		stepID:  uuid.New().String(),
	}

	if err := step.Transaction(transactionCtx); err != nil {
		compensationCtx := &compensationContext{
			Context: c.Context,
			tp:      c.tp,
			sagaID:  c.sagaID,
			stepID:  transactionCtx.stepID,
		}

		if compErr := step.Compensation(compensationCtx); compErr != nil {
			return fmt.Errorf("transaction failed and compensation failed: %v, compensation error: %v", err, compErr)
		}

		return fmt.Errorf("transaction failed, compensation succeeded: %v", err)
	}

	saga.CurrentStep++
	saga.LastUpdatedAt = time.Now()
	if err := c.tp.sagaRepo.UpdateSaga(c, saga); err != nil {
		return fmt.Errorf("failed to update saga: %v", err)
	}

	return nil
}

func (c *handlerSagaContext) EnqueueSaga(handler HandlerSagaFunc, params interface{}, options ...EnqueueOption) (string, error) {
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
	return c.tp.sideEffectRepo.GetSideEffect(c, c.sagaID, key)
}

func (c *transactionContext) SendSignal(name string, payload interface{}) error {
	return c.tp.SendSignal(c, c.stepID, name, payload)
}

func (c *transactionContext) ReceiveSignal(name string) (<-chan []byte, error) {
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
	return c.tp.sideEffectRepo.GetSideEffect(c, c.sagaID, key)
}

func (c *compensationContext) SendSignal(name string, payload interface{}) error {
	return c.tp.SendSignal(c, c.stepID, name, payload)
}

func (c *compensationContext) ReceiveSignal(name string) (<-chan []byte, error) {
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
	return c.tp.Enqueue(c, handler, params, options...)
}

func (c *sideEffectContext) EnqueueTaskAndWait(handler HandlerFunc, params interface{}, options ...EnqueueOption) (interface{}, error) {
	taskID, err := c.EnqueueTask(handler, params, options...)
	if err != nil {
		return nil, err
	}
	return c.WaitForCompletion(taskID)
}

func (c *sideEffectContext) SideEffect(key string, effect SideEffect) (interface{}, error) {
	return c.tp.sideEffectRepo.GetSideEffect(c, c.id, key)
}

func (c *sideEffectContext) SendSignal(name string, payload interface{}) error {
	return c.tp.SendSignal(c, c.id, name, payload)
}

func (c *sideEffectContext) ReceiveSignal(name string) (<-chan []byte, error) {
	return c.tp.ReceiveSignal(c, c.id, name)
}

func (c *sideEffectContext) WaitForCompletion(taskID string) (interface{}, error) {
	for {
		select {
		case <-c.Done():
			return nil, c.Err()
		case <-time.After(time.Second):
			task, err := c.tp.taskRepo.GetTask(c, taskID)
			if err != nil {
				return nil, err
			}

			switch task.Status {
			case TaskStatusCompleted:
				var result interface{}
				if err := json.Unmarshal(task.Result, &result); err != nil {
					return nil, fmt.Errorf("failed to unmarshal task result: %v", err)
				}
				return result, nil
			case TaskStatusFailed:
				return nil, fmt.Errorf("task failed")
			case TaskStatusCancelled:
				return nil, fmt.Errorf("task cancelled")
			case TaskStatusTerminated:
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
	taskData := task.Data()
	taskData.RetryCount = attempt
	if err := tp.taskRepo.UpdateTask(tp.ctx, taskData); err != nil {
		log.Printf("Failed to update task retry count: %v", err)
	}
}

func (tp *Tempolite) onHandlerPanic(task *Task, v interface{}) {
	log.Printf("Handler panicked: %v", v)
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
		tp.compensationPool.Dispatch(compensation)
	}

	return retrypool.DeadTaskActionForceRetry
}

func (tp *Tempolite) onSagaHandlerRetry(attempt int, err error, task *retrypool.TaskWrapper[*Task]) {
	taskData := task.Data()
	taskData.RetryCount = attempt
	if err := tp.taskRepo.UpdateTask(tp.ctx, taskData); err != nil {
		log.Printf("Failed to update saga task retry count: %v", err)
	}
}

func (tp *Tempolite) onSagaHandlerPanic(task *Task, v interface{}) {
	log.Printf("Saga handler panicked: %v", v)
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
	compensationData := task.Data()
	compensationData.Status = ExecutionStatusCompleted
	if err := tp.compensationRepo.UpdateCompensation(tp.ctx, compensationData); err != nil {
		log.Printf("Failed to update compensation status: %v", err)
	}
}

func (tp *Tempolite) onCompensationFailure(controller retrypool.WorkerController[*Compensation], workerID int, worker retrypool.Worker[*Compensation], task *retrypool.TaskWrapper[*Compensation], err error) retrypool.DeadTaskAction {
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
	compensationData := task.Data()
	if err := tp.compensationRepo.UpdateCompensation(tp.ctx, compensationData); err != nil {
		log.Printf("Failed to update compensation retry count: %v", err)
	}
}

func (tp *Tempolite) onCompensationPanic(compensation *Compensation, v interface{}) {
	log.Printf("Compensation panicked: %v", v)
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

func (tp *Tempolite) onSideEffectSuccess(controller retrypool.WorkerController[SideEffect], workerID int, worker retrypool.Worker[SideEffect], task *retrypool.TaskWrapper[SideEffect]) {
	log.Printf("Side effect completed successfully")
}

func (tp *Tempolite) onSideEffectFailure(controller retrypool.WorkerController[SideEffect], workerID int, worker retrypool.Worker[SideEffect], task *retrypool.TaskWrapper[SideEffect], err error) retrypool.DeadTaskAction {
	log.Printf("Side effect failed: %v", err)

	if IsUnrecoverable(err) {
		return retrypool.DeadTaskActionAddToDeadTasks
	}

	return retrypool.DeadTaskActionForceRetry
}

func (tp *Tempolite) onSideEffectRetry(attempt int, err error, task *retrypool.TaskWrapper[SideEffect]) {
	log.Printf("Retrying side effect, attempt %d: %v", attempt, err)
}

func (tp *Tempolite) onSideEffectPanic(sideEffect SideEffect, v interface{}) {
	log.Printf("Side effect panicked: %v", v)
}

// Helper functions for execution tree management

func (tp *Tempolite) addNodeToExecutionTree(ctx context.Context, node *ExecutionNode) error {
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
			return fmt.Errorf("failed to get parent node: %v", err)
		}
		tree.Connect(dag.BasicEdge(parentNode, node))
	}

	return nil
}

func (tp *Tempolite) updateNodeInExecutionTree(ctx context.Context, node *ExecutionNode) error {
	tp.executionTreesMu.Lock()
	defer tp.executionTreesMu.Unlock()

	tree, exists := tp.executionTrees[node.ID]
	if !exists {
		return fmt.Errorf("execution tree not found for node %s", node.ID)
	}

	// Remove the old node and add the updated one
	tree.Remove(node)
	tree.Add(node)

	return nil
}

/////////////////////////////////////////////////////

// SQLite Repository Implementations

type SQLiteTaskRepository struct {
	db *sql.DB
}

func NewSQLiteTaskRepository(db *sql.DB) (*SQLiteTaskRepository, error) {
	repo := &SQLiteTaskRepository{db: db}
	if err := repo.initDB(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *SQLiteTaskRepository) initDB() error {
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
	return err
}

func (r *SQLiteTaskRepository) CreateTask(ctx context.Context, task *Task) error {
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
		task.Result,
		nullableString(task.ParentTaskID),
		nullableString(task.SagaID),
	)
	return err
}

func (r *SQLiteTaskRepository) GetTask(ctx context.Context, id string) (*Task, error) {
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
		task.Result = []byte(result.String)
	}
	return &task, nil
}

func (r *SQLiteTaskRepository) UpdateTask(ctx context.Context, task *Task) error {
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
		task.Result,
		nullableString(task.ParentTaskID),
		nullableString(task.SagaID),
		task.ID,
	)
	return err
}

func (r *SQLiteTaskRepository) GetPendingTasks(ctx context.Context, limit int) ([]*Task, error) {
	query := `
	SELECT id, execution_context_id, handler_name, payload, status, retry_count, scheduled_at, created_at, updated_at, completed_at, result, parent_task_id, saga_id
	FROM tasks
	WHERE status = ? AND scheduled_at <= ?
	ORDER BY scheduled_at
	LIMIT ?;
	`
	rows, err := r.db.QueryContext(ctx, query, TaskStatusPending, time.Now().Unix(), limit)
	if err != nil {
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
	return tasks, nil
}

func (r *SQLiteTaskRepository) GetRunningTasksForSaga(ctx context.Context, sagaID string) ([]*Task, error) {
	query := `
	SELECT id, execution_context_id, handler_name, payload, status, retry_count, scheduled_at, created_at, updated_at, completed_at, result, parent_task_id, saga_id
	FROM tasks
	WHERE saga_id = ? AND status = ?;
	`
	rows, err := r.db.QueryContext(ctx, query, sagaID, TaskStatusInProgress)
	if err != nil {
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
	return tasks, nil
}

type SQLiteSideEffectRepository struct {
	db *sql.DB
}

func NewSQLiteSideEffectRepository(db *sql.DB) (*SQLiteSideEffectRepository, error) {
	repo := &SQLiteSideEffectRepository{db: db}
	if err := repo.initDB(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *SQLiteSideEffectRepository) initDB() error {
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
	return err
}

func (r *SQLiteSideEffectRepository) GetSideEffect(ctx context.Context, executionContextID, key string) ([]byte, error) {
	query := `
	SELECT result
	FROM side_effects
	WHERE execution_context_id = ? AND key = ?;
	`
	var result []byte
	err := r.db.QueryRowContext(ctx, query, executionContextID, key).Scan(&result)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return result, nil
}

func (r *SQLiteSideEffectRepository) SaveSideEffect(ctx context.Context, executionContextID, key string, result []byte) error {
	query := `
	INSERT OR REPLACE INTO side_effects (execution_context_id, key, result, created_at)
	VALUES (?, ?, ?, ?);
	`
	_, err := r.db.ExecContext(ctx, query, executionContextID, key, result, time.Now().Unix())
	return err
}

type SQLiteSignalRepository struct {
	db *sql.DB
}

func NewSQLiteSignalRepository(db *sql.DB) (*SQLiteSignalRepository, error) {
	repo := &SQLiteSignalRepository{db: db}
	if err := repo.initDB(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *SQLiteSignalRepository) initDB() error {
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
	return err
}

func (r *SQLiteSignalRepository) SaveSignal(ctx context.Context, signal *Signal) error {
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
	return err
}

func (r *SQLiteSignalRepository) GetSignals(ctx context.Context, taskID string, name string, direction string) ([]*Signal, error) {
	query := `
	SELECT id, task_id, name, payload, created_at, direction
	FROM signals
	WHERE task_id = ? AND name = ? AND direction = ?;
	`
	rows, err := r.db.QueryContext(ctx, query, taskID, name, direction)
	if err != nil {
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
			return nil, err
		}
		signal.CreatedAt = time.Unix(createdAt, 0)
		signals = append(signals, &signal)
	}
	return signals, nil
}

func (r *SQLiteSignalRepository) DeleteSignals(ctx context.Context, taskID string, name string, direction string) error {
	query := `
	DELETE FROM signals
	WHERE task_id = ? AND name = ? AND direction = ?;
	`
	_, err := r.db.ExecContext(ctx, query, taskID, name, direction)
	return err
}

type SQLiteSagaRepository struct {
	db *sql.DB
}

func NewSQLiteSagaRepository(db *sql.DB) (*SQLiteSagaRepository, error) {
	repo := &SQLiteSagaRepository{db: db}
	if err := repo.initDB(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *SQLiteSagaRepository) initDB() error {
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
	return err
}

func (r *SQLiteSagaRepository) CreateSaga(ctx context.Context, saga *SagaInfo) error {
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
	return err
}

func (r *SQLiteSagaRepository) GetSaga(ctx context.Context, id string) (*SagaInfo, error) {
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
		return nil, err
	}
	saga.CreatedAt = time.Unix(createdAt, 0)
	saga.LastUpdatedAt = time.Unix(lastUpdatedAt, 0)
	if completedAt.Valid {
		completedTime := time.Unix(completedAt.Int64, 0)
		saga.CompletedAt = &completedTime
	}
	return &saga, nil
}

func (r *SQLiteSagaRepository) UpdateSaga(ctx context.Context, saga *SagaInfo) error {
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
	return err
}

type SQLiteExecutionTreeRepository struct {
	db *sql.DB
}

func NewSQLiteExecutionTreeRepository(db *sql.DB) (*SQLiteExecutionTreeRepository, error) {
	repo := &SQLiteExecutionTreeRepository{db: db}
	if err := repo.initDB(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *SQLiteExecutionTreeRepository) initDB() error {
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
	return err
}

func (r *SQLiteExecutionTreeRepository) CreateNode(ctx context.Context, node *ExecutionNode) error {
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
	return err
}

func (r *SQLiteExecutionTreeRepository) GetNode(ctx context.Context, id string) (*ExecutionNode, error) {
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
		return nil, err
	}
	node.CreatedAt = time.Unix(createdAt, 0)
	node.UpdatedAt = time.Unix(updatedAt, 0)
	if completedAt.Valid {
		completedTime := time.Unix(completedAt.Int64, 0)
		node.CompletedAt = &completedTime
	}
	return &node, nil
}

func (r *SQLiteExecutionTreeRepository) UpdateNode(ctx context.Context, node *ExecutionNode) error {
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
	return err
}

func (r *SQLiteExecutionTreeRepository) GetChildNodes(ctx context.Context, parentID string) ([]*ExecutionNode, error) {
	query := `
	SELECT id, parent_id, type, status, created_at, updated_at, completed_at, handler_name, payload, result, error_message
	FROM execution_nodes
	WHERE parent_id = ?;
	`
	rows, err := r.db.QueryContext(ctx, query, parentID)
	if err != nil {
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
	return nodes, nil
}

type SQLiteCompensationRepository struct {
	db *sql.DB
}

func NewSQLiteCompensationRepository(db *sql.DB) (*SQLiteCompensationRepository, error) {
	repo := &SQLiteCompensationRepository{db: db}
	if err := repo.initDB(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *SQLiteCompensationRepository) initDB() error {
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
	return err
}

func (r *SQLiteCompensationRepository) CreateCompensation(ctx context.Context, compensation *Compensation) error {
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
	return err
}

func (r *SQLiteCompensationRepository) GetCompensation(ctx context.Context, id string) (*Compensation, error) {
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
		return nil, err
	}
	compensation.CreatedAt = time.Unix(createdAt, 0)
	compensation.UpdatedAt = time.Unix(updatedAt, 0)
	return &compensation, nil
}

func (r *SQLiteCompensationRepository) UpdateCompensation(ctx context.Context, compensation *Compensation) error {
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
	return err
}

func (r *SQLiteCompensationRepository) GetCompensationsForSaga(ctx context.Context, sagaID string) ([]*Compensation, error) {
	query := `
	SELECT id, saga_id, step_index, payload, status, created_at, updated_at
	FROM compensations
	WHERE saga_id = ?
	ORDER BY step_index DESC;
	`
	rows, err := r.db.QueryContext(ctx, query, sagaID)
	if err != nil {
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
			return nil, err
		}
		compensation.CreatedAt = time.Unix(createdAt, 0)
		compensation.UpdatedAt = time.Unix(updatedAt, 0)
		compensations = append(compensations, &compensation)
	}
	return compensations, nil
}

type SQLiteSagaStepRepository struct {
	db *sql.DB
}

func NewSQLiteSagaStepRepository(db *sql.DB) (*SQLiteSagaStepRepository, error) {
	repo := &SQLiteSagaStepRepository{db: db}
	if err := repo.initDB(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *SQLiteSagaStepRepository) initDB() error {
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
	return err
}

func (r *SQLiteSagaStepRepository) CreateSagaStep(ctx context.Context, sagaID string, stepIndex int, payload []byte) error {
	query := `
	INSERT INTO saga_steps (saga_id, step_index, payload)
	VALUES (?, ?, ?);
	`
	_, err := r.db.ExecContext(ctx, query, sagaID, stepIndex, payload)
	return err
}

func (r *SQLiteSagaStepRepository) GetSagaStep(ctx context.Context, sagaID string, stepIndex int) ([]byte, error) {
	query := `
	SELECT payload
	FROM saga_steps
	WHERE saga_id = ? AND step_index = ?;
	`
	var payload []byte
	err := r.db.QueryRowContext(ctx, query, sagaID, stepIndex).Scan(&payload)
	if err != nil {
		return nil, err
	}
	return payload, nil
}
