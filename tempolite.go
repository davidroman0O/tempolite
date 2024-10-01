package tempolite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/davidroman0O/retrypool"
)

// --- Models and Repositories ---

// TaskStatus represents the status of a task
type TaskStatus int

const (
	TaskStatusPending TaskStatus = iota
	TaskStatusInProgress
	TaskStatusCompleted
	TaskStatusFailed
	TaskStatusCancelled
	TaskStatusTerminated
)

// Task represents a task to be processed
type Task struct {
	ID                 int64      // Primary key
	ExecutionContextID string     // Shared context ID among related tasks
	HandlerName        string     // Name of the handler function
	Payload            []byte     // Serialized task parameters
	Status             TaskStatus // Task status
	RetryCount         int        // Number of retries attempted
	ScheduledAt        time.Time  // When the task is scheduled to run
	CreatedAt          time.Time
	UpdatedAt          time.Time
	Options            []byte        // Serialized task options (if needed)
	Result             []byte        // Serialized result of the task
	ParentTaskID       sql.NullInt64 // Parent task ID, nullable
}

// SideEffect represents a recorded side effect to avoid redundant computations
type SideEffect struct {
	ID                 int64  // Primary key
	ExecutionContextID string // Associated execution context
	Key                string // Unique key for the side effect
	Result             []byte // Serialized result of the side effect
	CreatedAt          time.Time
}

// Signal represents a message sent to or from a task
type Signal struct {
	ID        int64  // Primary key
	TaskID    int64  // Associated task ID
	Name      string // Signal name
	Payload   []byte // Serialized signal data
	CreatedAt time.Time
	Direction string // "inbound" or "outbound"
}

// --- Repository Interfaces ---

type TaskRepository interface {
	CreateTask(ctx context.Context, task *Task) (int64, error)
	GetPendingTasks(ctx context.Context, limit int) ([]*Task, error)
	UpdateTask(ctx context.Context, task *Task) error
	GetTaskByID(ctx context.Context, id int64) (*Task, error)
	CancelTask(ctx context.Context, id int64, status TaskStatus) error
	ResetInProgressTasks(ctx context.Context) error
}

type SideEffectRepository interface {
	GetSideEffect(ctx context.Context, executionContextID, key string) (*SideEffect, error)
	SaveSideEffect(ctx context.Context, sideEffect *SideEffect) error
}

type SignalRepository interface {
	SaveSignal(ctx context.Context, signal *Signal) error
	GetSignals(ctx context.Context, taskID int64, name string, direction string) ([]*Signal, error)
	DeleteSignals(ctx context.Context, taskID int64, name string, direction string) error
}

// --- SQLite Implementations ---

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
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		execution_context_id TEXT NOT NULL,
		handler_name TEXT NOT NULL,
		payload BLOB NOT NULL,
		status INTEGER NOT NULL,
		retry_count INTEGER NOT NULL,
		scheduled_at INTEGER NOT NULL,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL,
		options BLOB,
		result BLOB,
		parent_task_id INTEGER,
		FOREIGN KEY(parent_task_id) REFERENCES tasks(id)
	);
	`
	_, err := r.db.Exec(query)
	return err
}

func (r *SQLiteTaskRepository) CreateTask(ctx context.Context, task *Task) (int64, error) {
	query := `
	INSERT INTO tasks (execution_context_id, handler_name, payload, status, retry_count, scheduled_at, created_at, updated_at, options, result, parent_task_id)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`
	var parentTaskID interface{}
	if task.ParentTaskID.Valid {
		parentTaskID = task.ParentTaskID.Int64
	} else {
		parentTaskID = nil
	}
	result, err := r.db.ExecContext(ctx, query,
		task.ExecutionContextID,
		task.HandlerName,
		task.Payload,
		int(task.Status),
		task.RetryCount,
		task.ScheduledAt.Unix(),
		task.CreatedAt.Unix(),
		task.UpdatedAt.Unix(),
		task.Options,
		task.Result,
		parentTaskID,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *SQLiteTaskRepository) GetPendingTasks(ctx context.Context, limit int) ([]*Task, error) {
	query := `
	SELECT id, execution_context_id, handler_name, payload, status, retry_count, scheduled_at, created_at, updated_at, options, result, parent_task_id
	FROM tasks
	WHERE status = ? AND scheduled_at <= ?
	ORDER BY scheduled_at
	LIMIT ?;
	`
	now := time.Now().Unix()
	rows, err := r.db.QueryContext(ctx, query, int(TaskStatusPending), now, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		task := &Task{}
		var scheduledAt, createdAt, updatedAt int64
		var status int
		var parentTaskID sql.NullInt64
		err := rows.Scan(
			&task.ID,
			&task.ExecutionContextID,
			&task.HandlerName,
			&task.Payload,
			&status,
			&task.RetryCount,
			&scheduledAt,
			&createdAt,
			&updatedAt,
			&task.Options,
			&task.Result,
			&parentTaskID,
		)
		if err != nil {
			return nil, err
		}
		task.Status = TaskStatus(status)
		task.ScheduledAt = time.Unix(scheduledAt, 0)
		task.CreatedAt = time.Unix(createdAt, 0)
		task.UpdatedAt = time.Unix(updatedAt, 0)
		task.ParentTaskID = parentTaskID
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *SQLiteTaskRepository) UpdateTask(ctx context.Context, task *Task) error {
	query := `
	UPDATE tasks
	SET execution_context_id = ?, handler_name = ?, payload = ?, status = ?, retry_count = ?,  scheduled_at = ?, created_at = ?, updated_at = ?, options = ?, result = ?, parent_task_id = ?
	WHERE id = ?;
	`
	var parentTaskID interface{}
	if task.ParentTaskID.Valid {
		parentTaskID = task.ParentTaskID.Int64
	} else {
		parentTaskID = nil
	}
	_, err := r.db.ExecContext(ctx, query,
		task.ExecutionContextID,
		task.HandlerName,
		task.Payload,
		int(task.Status),
		task.RetryCount,
		task.ScheduledAt.Unix(),
		task.CreatedAt.Unix(),
		task.UpdatedAt.Unix(),
		task.Options,
		task.Result,
		parentTaskID,
		task.ID,
	)
	return err
}

func (r *SQLiteTaskRepository) GetTaskByID(ctx context.Context, id int64) (*Task, error) {
	query := `
	SELECT id, execution_context_id, handler_name, payload, status, retry_count, scheduled_at, created_at, updated_at, options, result, parent_task_id
	FROM tasks
	WHERE id = ?;
	`
	row := r.db.QueryRowContext(ctx, query, id)
	task := &Task{}
	var scheduledAt, createdAt, updatedAt int64
	var status int
	var parentTaskID sql.NullInt64
	err := row.Scan(
		&task.ID,
		&task.ExecutionContextID,
		&task.HandlerName,
		&task.Payload,
		&status,
		&task.RetryCount,
		&scheduledAt,
		&createdAt,
		&updatedAt,
		&task.Options,
		&task.Result,
		&parentTaskID,
	)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			log.Printf("GetTaskByID: context canceled")
		}
		return nil, err
	}
	task.Status = TaskStatus(status)
	task.ScheduledAt = time.Unix(scheduledAt, 0)
	task.CreatedAt = time.Unix(createdAt, 0)
	task.UpdatedAt = time.Unix(updatedAt, 0)
	task.ParentTaskID = parentTaskID
	return task, nil
}

func (r *SQLiteTaskRepository) CancelTask(ctx context.Context, id int64, status TaskStatus) error {
	query := `
	UPDATE tasks
	SET status = ?
	WHERE id = ? AND status = ?;
	`
	result, err := r.db.ExecContext(ctx, query, int(status), id, int(TaskStatusPending))
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 && status == TaskStatusCancelled {
		return fmt.Errorf("task %d is not pending and cannot be cancelled", id)
	}
	return nil
}

func (r *SQLiteTaskRepository) ResetInProgressTasks(ctx context.Context) error {
	query := `
	UPDATE tasks
	SET status = ?
	WHERE status = ?;
	`
	_, err := r.db.ExecContext(ctx, query, int(TaskStatusPending), int(TaskStatusInProgress))
	return err
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
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		execution_context_id TEXT NOT NULL,
		key TEXT NOT NULL,
		result BLOB NOT NULL,
		created_at INTEGER NOT NULL,
		UNIQUE (execution_context_id, key)
	);
	`
	_, err := r.db.Exec(query)
	return err
}

func (r *SQLiteSideEffectRepository) GetSideEffect(ctx context.Context, executionContextID, key string) (*SideEffect, error) {
	query := `
	SELECT id, execution_context_id, key, result, created_at
	FROM side_effects
	WHERE execution_context_id = ? AND key = ?;
	`
	row := r.db.QueryRowContext(ctx, query, executionContextID, key)
	sideEffect := &SideEffect{}
	var createdAt int64
	err := row.Scan(
		&sideEffect.ID,
		&sideEffect.ExecutionContextID,
		&sideEffect.Key,
		&sideEffect.Result,
		&createdAt,
	)
	if err != nil {
		return nil, err
	}
	sideEffect.CreatedAt = time.Unix(createdAt, 0)
	return sideEffect, nil
}

func (r *SQLiteSideEffectRepository) SaveSideEffect(ctx context.Context, sideEffect *SideEffect) error {
	query := `
	INSERT OR REPLACE INTO side_effects (execution_context_id, key, result, created_at)
	VALUES (?, ?, ?, ?);
	`
	_, err := r.db.ExecContext(ctx, query,
		sideEffect.ExecutionContextID,
		sideEffect.Key,
		sideEffect.Result,
		sideEffect.CreatedAt.Unix(),
	)
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
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		task_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		payload BLOB NOT NULL,
		created_at INTEGER NOT NULL,
		direction TEXT NOT NULL
	);
	`
	_, err := r.db.Exec(query)
	return err
}

func (r *SQLiteSignalRepository) SaveSignal(ctx context.Context, signal *Signal) error {
	query := `
	INSERT INTO signals (task_id, name, payload, created_at, direction)
	VALUES (?, ?, ?, ?, ?);
	`
	_, err := r.db.ExecContext(ctx, query,
		signal.TaskID,
		signal.Name,
		signal.Payload,
		signal.CreatedAt.Unix(),
		signal.Direction,
	)
	return err
}

func (r *SQLiteSignalRepository) GetSignals(ctx context.Context, taskID int64, name string, direction string) ([]*Signal, error) {
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
		signal := &Signal{}
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
		signals = append(signals, signal)
	}
	return signals, nil
}

func (r *SQLiteSignalRepository) DeleteSignals(ctx context.Context, taskID int64, name string, direction string) error {
	query := `
	DELETE FROM signals
	WHERE task_id = ? AND name = ? AND direction = ?;
	`
	_, err := r.db.ExecContext(ctx, query, taskID, name, direction)
	return err
}

// --- Handler Registration and Retrieval ---

type HandlerFunc interface{}

type HandlerInfo struct {
	Handler    HandlerFunc
	ParamType  reflect.Type
	ReturnType reflect.Type
}

var (
	handlerRegistry      = make(map[string]HandlerInfo)
	handlerRegistryMutex sync.RWMutex
)

func RegisterHandler(handler HandlerFunc) {
	handlerType := reflect.TypeOf(handler)
	if handlerType.Kind() != reflect.Func {
		panic("Handler must be a function")
	}
	if handlerType.NumIn() != 2 || handlerType.NumOut() != 2 {
		panic("Handler must have two input parameters and two return values")
	}
	if handlerType.In(0) != reflect.TypeOf((*context.Context)(nil)).Elem() {
		panic("First parameter of handler must be context.Context")
	}
	if handlerType.Out(1) != reflect.TypeOf((*error)(nil)).Elem() {
		panic("Second return value of handler must be error")
	}

	name := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
	handlerInfo := HandlerInfo{
		Handler:    handler,
		ParamType:  handlerType.In(1),
		ReturnType: handlerType.Out(0),
	}

	handlerRegistryMutex.Lock()
	handlerRegistry[name] = handlerInfo
	handlerRegistryMutex.Unlock()
}

func GetHandler(name string) (HandlerInfo, bool) {
	handlerRegistryMutex.RLock()
	handlerInfo, exists := handlerRegistry[name]
	handlerRegistryMutex.RUnlock()
	return handlerInfo, exists
}

// --- Tempolite Implementation ---

type Tempolite struct {
	taskRepo       TaskRepository
	sideEffectRepo SideEffectRepository
	signalRepo     SignalRepository
	pool           *retrypool.Pool[*Task]
	workerWg       sync.WaitGroup
	ctx            context.Context
	cancel         context.CancelFunc
	taskOptions    sync.Map // map[int64][]retrypool.TaskOption[*Task]
	taskContexts   sync.Map // map[int64]context.CancelFunc

	signalChannels sync.Map // map[string]chan []byte
}

func New(ctx context.Context, taskRepo TaskRepository, sideEffectRepo SideEffectRepository, signalRepo SignalRepository) (*Tempolite, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithCancel(ctx)

	wb := &Tempolite{
		taskRepo:       taskRepo,
		sideEffectRepo: sideEffectRepo,
		signalRepo:     signalRepo,
		ctx:            ctx,
		cancel:         cancel,
	}

	// Initialize retrypool with custom options
	poolOptions := []retrypool.Option[*Task]{
		retrypool.WithOnTaskSuccess[*Task](wb.onTaskSuccess),
		retrypool.WithOnTaskFailure[*Task](wb.onTaskFailure),
		retrypool.WithOnRetry[*Task](wb.onRetry),                   // Update RetryCount on retry
		retrypool.WithAttempts[*Task](retrypool.UnlimitedAttempts), // Set maximum retry attempts
		retrypool.WithPanicHandler[*Task](wb.onTaskPanic),          // Add panic handler
	}

	wb.pool = retrypool.New(ctx, nil, poolOptions...)

	// Recover tasks that were in progress
	err := wb.taskRepo.ResetInProgressTasks(ctx)
	if err != nil {
		return nil, err
	}

	return wb, nil
}

func (wb *Tempolite) onRetry(attempt int, err error, task *retrypool.TaskWrapper[*Task]) {
	// Update the RetryCount in task.Data()
	taskData := task.Data()
	taskData.RetryCount = attempt
}

type enqueueOptions struct {
	poolOpts []retrypool.TaskOption[*Task]
	parentID *int64
}

type enqueueOption func(*enqueueOptions)

func EnqueueWithMaxDuration(duration time.Duration) enqueueOption {
	return func(opts *enqueueOptions) {
		opts.poolOpts = append(opts.poolOpts, retrypool.WithMaxDuration[*Task](duration))
	}
}

func EnqueueWithImmediateRetry() enqueueOption {
	return func(opts *enqueueOptions) {
		opts.poolOpts = append(opts.poolOpts, retrypool.WithImmediateRetry[*Task]())
	}
}

func EnqueueWithPanicOnTimeout() enqueueOption {
	return func(opts *enqueueOptions) {
		opts.poolOpts = append(opts.poolOpts, retrypool.WithPanicOnTimeout[*Task]())
	}
}

func EnqueueWithParent(parentID int64) enqueueOption {
	return func(opts *enqueueOptions) {
		opts.parentID = &parentID
	}
}

func (wb *Tempolite) EnqueueTask(ctx context.Context, executionContextID string, handler HandlerFunc, params interface{}, options ...enqueueOption) (int64, error) {
	// Get handler name
	handlerName := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()

	// Get handler info
	handlerInfo, exists := GetHandler(handlerName)
	if !exists {
		return 0, fmt.Errorf("no handler registered with name: %s", handlerName)
	}

	// Check if params match the expected type
	paramsType := reflect.TypeOf(params)
	if paramsType != handlerInfo.ParamType {
		return 0, fmt.Errorf("params type mismatch: expected %v, got %v", handlerInfo.ParamType, paramsType)
	}

	// Serialize parameters
	payload, err := json.Marshal(params)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal task parameters: %v", err)
	}

	opts := enqueueOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	// Serialize options if needed (placeholder in this example)
	var optionsData []byte

	task := &Task{
		ExecutionContextID: executionContextID,
		HandlerName:        handlerName,
		Payload:            payload,
		Status:             TaskStatusPending,
		RetryCount:         0,
		ScheduledAt:        time.Now(),
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
		Options:            optionsData,
	}

	// Set ParentTaskID if provided
	if opts.parentID != nil {
		task.ParentTaskID = sql.NullInt64{Int64: *opts.parentID, Valid: true}
	}

	id, err := wb.taskRepo.CreateTask(ctx, task)
	if err != nil {
		return 0, err
	}
	task.ID = id

	// Store task options for use during dispatch
	if len(options) > 0 {
		wb.taskOptions.Store(id, options)
	}

	return id, nil
}

func (wb *Tempolite) Start(workerCount int) {
	for i := 0; i < workerCount; i++ {
		worker := &TaskWorker{
			ID: i,
			wb: wb,
		}
		wb.pool.AddWorker(worker)
	}
	// Start dispatcher
	wb.workerWg.Add(1)
	go wb.dispatcher()
}

func (wb *Tempolite) Stop() {
	wb.cancel()
	wb.pool.Close()
	wb.workerWg.Wait()
}

func (wb *Tempolite) dispatcher() {
	defer wb.workerWg.Done()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	log.Println("Dispatcher started")

	for {
		select {
		case <-wb.ctx.Done():
			log.Println("Dispatcher stopping due to context cancellation")
			return
		case <-ticker.C:
			log.Println("Dispatcher tick: fetching pending tasks")
			// Fetch pending tasks
			tasks, err := wb.taskRepo.GetPendingTasks(wb.ctx, 100)
			if err != nil {
				log.Printf("Error fetching tasks: %v", err)
				continue
			}

			log.Printf("Fetched %d pending tasks", len(tasks))

			for _, task := range tasks {
				log.Printf("Processing task ID: %d", task.ID)
				// Update task status to in-progress
				task.Status = TaskStatusInProgress
				task.UpdatedAt = time.Now()
				err := wb.taskRepo.UpdateTask(wb.ctx, task)
				if err != nil {
					log.Printf("Error updating task status: %v", err)
					continue
				}

				// Retrieve task options
				var options []enqueueOption
				if opts, ok := wb.taskOptions.Load(task.ID); ok {
					options = opts.([]enqueueOption)
				}

				cfg := enqueueOptions{}
				for _, opt := range options {
					opt(&cfg)
				}

				// Dispatch task to retrypool with options
				log.Printf("Dispatching task ID: %d to retrypool", task.ID)
				wb.pool.Dispatch(task, cfg.poolOpts...)
			}
		}
	}
}

func (wb *Tempolite) SendSignal(ctx context.Context, taskID int64, name string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	signal := &Signal{
		TaskID:    taskID,
		Name:      name,
		Payload:   data,
		CreatedAt: time.Now(),
		Direction: "inbound",
	}

	key := fmt.Sprintf("%d:%s", taskID, name)

	// Check if there is a channel registered
	if chInterface, ok := wb.signalChannels.Load(key); ok {
		ch := chInterface.(chan []byte)
		select {
		case ch <- data:
			// Signal sent to the task
			return nil
		case <-ctx.Done():
			return ctx.Err()
		default:
			// If channel is blocked, store the signal in the database
			return wb.signalRepo.SaveSignal(ctx, signal)
		}
	} else {
		// No channel registered, store the signal
		return wb.signalRepo.SaveSignal(ctx, signal)
	}
}

func (wb *Tempolite) ReceiveSignal(ctx context.Context, taskID int64, name string) (<-chan []byte, error) {
	key := fmt.Sprintf("%d:%s", taskID, name)
	ch := make(chan []byte)

	_, loaded := wb.signalChannels.LoadOrStore(key, ch)
	if loaded {
		// A channel already exists for this signal name and task, this shouldn't happen
		// But for safety, we can return an error
		return nil, fmt.Errorf("ReceiveSignal already called for task %d and signal %s", taskID, name)
	}

	// Fetch any stored signals from the database and send them on the channel
	go func() {
		// Defer cleanup
		defer func() {
			wb.signalChannels.Delete(key)
			close(ch)
		}()

		// Fetch signals from the database
		signals, err := wb.signalRepo.GetSignals(ctx, taskID, name, "outbound")
		if err != nil {
			// Handle error, maybe log it
			return
		}
		// Send stored signals on the channel
		for _, signal := range signals {
			select {
			case ch <- signal.Payload:
			case <-ctx.Done():
				return
			}
		}
		// Delete signals after retrieving
		err = wb.signalRepo.DeleteSignals(ctx, taskID, name, "outbound")
		if err != nil {
			// Handle error, maybe log it
			return
		}

		// Now wait until context is done
		<-ctx.Done()
	}()

	return ch, nil
}

func (wb *Tempolite) CancelTask(taskID int64) error {
	// Cancel pending task
	err := wb.taskRepo.CancelTask(wb.ctx, taskID, TaskStatusCancelled)
	if err != nil {
		return err
	}
	return nil
}

func (wb *Tempolite) TerminateTask(taskID int64) error {
	// Terminate running task and set status to Terminated
	err := wb.taskRepo.CancelTask(wb.ctx, taskID, TaskStatusTerminated)
	if err != nil {
		return err
	}

	// Cancel running task
	if cancelFunc, ok := wb.taskContexts.Load(taskID); ok {
		cancelFunc.(context.CancelFunc)()
	}
	return nil
}

func (wb *Tempolite) WaitForTaskCompletion(ctx context.Context, taskID int64) ([]byte, error) {
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			task, err := wb.taskRepo.GetTaskByID(ctx, taskID)
			if err != nil {
				return nil, err
			}
			if task.Status == TaskStatusCompleted {
				return task.Result, nil
			} else if task.Status == TaskStatusFailed || task.Status == TaskStatusCancelled || task.Status == TaskStatusTerminated {
				return nil, fmt.Errorf("task %d failed, was cancelled, or terminated", taskID)
			}
			// Sleep for a bit before checking again
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func (wb *Tempolite) onTaskPanic(task *Task, v interface{}) {
	taskData := task
	log.Printf("Task %d panicked: %v", taskData.ID, v)

	// Update task status to failed
	taskData.Status = TaskStatusFailed
	taskData.UpdatedAt = time.Now()
	err := wb.taskRepo.UpdateTask(wb.ctx, taskData)
	if err != nil {
		log.Printf("Error updating task status after panic: %v", err)
	}

	// Remove task options
	wb.taskOptions.Delete(taskData.ID)
}

// --- TaskWorker Implementation ---

type TaskWorker struct {
	ID              int
	wb              *Tempolite
	sideEffectCache map[string][]byte // Cache for side effects within a task
}

func (tw *TaskWorker) Run(ctx context.Context, task *Task) error {
	log.Printf("TaskWorker %d: Starting task %d", tw.ID, task.ID)

	// Check if the task was cancelled or terminated before starting
	currentTask, err := tw.wb.taskRepo.GetTaskByID(context.Background(), task.ID)
	if err != nil {
		log.Printf("TaskWorker %d: Error getting task %d: %v", tw.ID, task.ID, err)
		return err
	}
	if currentTask.Status == TaskStatusCancelled {
		log.Printf("TaskWorker %d: Task %d was cancelled", tw.ID, task.ID)
		return fmt.Errorf("task %d was cancelled", task.ID)
	}
	if currentTask.Status == TaskStatusTerminated {
		log.Printf("TaskWorker %d: Task %d was terminated", tw.ID, task.ID)
		return fmt.Errorf("task %d was terminated", task.ID)
	}

	// Create a cancellable context for the task
	taskCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Store the cancel function to allow task termination
	tw.wb.taskContexts.Store(task.ID, cancel)
	defer tw.wb.taskContexts.Delete(task.ID)

	// Initialize side effect cache
	tw.sideEffectCache = make(map[string][]byte)

	log.Printf("TaskWorker %d: Getting handler for task %d", tw.ID, task.ID)
	handlerInfo, exists := GetHandler(task.HandlerName)
	if !exists {
		log.Printf("TaskWorker %d: No handler found for task %d", tw.ID, task.ID)
		return fmt.Errorf("no handler registered with name: %s", task.HandlerName)
	}

	// Create a new instance of the parameter type
	paramInstancePtr := reflect.New(handlerInfo.ParamType).Interface()

	// Unmarshal the payload into the parameter instance
	log.Printf("TaskWorker %d: Unmarshaling payload for task %d: %v", tw.ID, task.ID, task.Payload)
	err = json.Unmarshal(task.Payload, paramInstancePtr)
	if err != nil {
		log.Printf("TaskWorker %d: Error unmarshaling payload for task %d: %v", tw.ID, task.ID, err)
		return fmt.Errorf("failed to unmarshal task payload: %v", err)
	}

	// Get the actual parameter value (not a pointer)
	paramValue := reflect.ValueOf(paramInstancePtr).Elem()

	// Create the task context
	taskCtx = &TaskContext{
		Context:            taskCtx,
		executionContextID: task.ExecutionContextID,
		taskID:             task.ID,
		wb:                 tw.wb,
		sideEffectCache:    tw.sideEffectCache,
		dbCtx:              tw.wb.ctx,
		task:               task,
	}

	// Call the handler using reflection
	log.Printf("TaskWorker %d: Calling handler for task %d", tw.ID, task.ID)
	handlerValue := reflect.ValueOf(handlerInfo.Handler)
	results := handlerValue.Call([]reflect.Value{
		reflect.ValueOf(taskCtx),
		paramValue,
	})

	// Check for error
	if !results[1].IsNil() {
		err := results[1].Interface().(error)
		log.Printf("TaskWorker %d: Handler returned error for task %d: %v", tw.ID, task.ID, err)
		return err
	}

	// Marshal the result
	log.Printf("TaskWorker %d: Marshaling result for task %d", tw.ID, task.ID)
	result, err := json.Marshal(results[0].Interface())
	if err != nil {
		log.Printf("TaskWorker %d: Error marshaling result for task %d: %v", tw.ID, task.ID, err)
		return fmt.Errorf("failed to marshal task result: %v", err)
	}

	// Store the result
	task.Result = result

	log.Printf("TaskWorker %d: Completed task %d", tw.ID, task.ID)
	return nil
}

// --- TaskContext for Side Effects and Signals ---

type TaskContext struct {
	context.Context
	executionContextID string
	taskID             int64
	wb                 *Tempolite
	sideEffectCache    map[string][]byte
	dbCtx              context.Context
	task               *Task // Include task to access RetryCount
}

func (t *TaskContext) GetID() int64 {
	return t.taskID
}

func (tc *TaskContext) RecordSideEffect(key string, computeFunc func() (interface{}, error)) (interface{}, error) {
	// Check cache
	if data, ok := tc.sideEffectCache[key]; ok {
		var result interface{}
		err := json.Unmarshal(data, &result)
		return result, err
	}

	// Check repository using dbCtx
	sideEffect, err := tc.wb.sideEffectRepo.GetSideEffect(tc.dbCtx, tc.executionContextID, key)
	if err == nil {
		var result interface{}
		err := json.Unmarshal(sideEffect.Result, &result)
		if err == nil {
			tc.sideEffectCache[key] = sideEffect.Result
			return result, nil
		}
	} else if err != sql.ErrNoRows {
		log.Printf("Error getting side effect: %v", err)
	}

	// Compute side effect
	result, err := computeFunc()
	if err != nil {
		return nil, err
	}

	// Serialize and save result
	data, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	sideEffect = &SideEffect{
		ExecutionContextID: tc.executionContextID,
		Key:                key,
		Result:             data,
		CreatedAt:          time.Now(),
	}
	err = tc.wb.sideEffectRepo.SaveSideEffect(tc.dbCtx, sideEffect)
	if err != nil {
		return nil, err
	}

	tc.sideEffectCache[key] = data
	return result, nil
}

func (tc *TaskContext) ReceiveSignal(name string) (<-chan []byte, error) {
	key := fmt.Sprintf("%d:%s", tc.taskID, name)
	ch := make(chan []byte)

	_, loaded := tc.wb.signalChannels.LoadOrStore(key, ch)
	if loaded {
		// A channel already exists for this signal name and task, this shouldn't happen
		// But for safety, we can return an error
		return nil, fmt.Errorf("ReceiveSignal already called for task %d and signal %s", tc.taskID, name)
	}

	// Fetch any stored signals from the database and send them on the channel
	go func() {
		// Defer cleanup
		defer func() {
			tc.wb.signalChannels.Delete(key)
			close(ch)
		}()

		// Fetch signals from the database
		signals, err := tc.wb.signalRepo.GetSignals(tc.dbCtx, tc.taskID, name, "inbound")
		if err != nil {
			// Handle error, maybe log it
			return
		}
		// Send stored signals on the channel
		for _, signal := range signals {
			select {
			case ch <- signal.Payload:
			case <-tc.Done():
				return
			}
		}
		// Delete signals after retrieving
		err = tc.wb.signalRepo.DeleteSignals(tc.dbCtx, tc.taskID, name, "inbound")
		if err != nil {
			// Handle error, maybe log it
			return
		}

		// Now wait until context is done
		<-tc.Done()
	}()

	return ch, nil
}

func (tc *TaskContext) SendSignal(name string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	signal := &Signal{
		TaskID:    tc.taskID,
		Name:      name,
		Payload:   data,
		CreatedAt: time.Now(),
		Direction: "outbound",
	}

	key := fmt.Sprintf("%d:%s", tc.taskID, name)

	// Check if there is a channel registered
	if chInterface, ok := tc.wb.signalChannels.Load(key); ok {
		ch := chInterface.(chan []byte)
		select {
		case ch <- data:
			// Signal sent
			return nil
		case <-tc.Done():
			return tc.Err()
		default:
			// If channel is blocked, store the signal in the database
			return tc.wb.signalRepo.SaveSignal(tc.dbCtx, signal)
		}
	} else {
		// No channel registered, store the signal
		return tc.wb.signalRepo.SaveSignal(tc.dbCtx, signal)
	}
}

func (tc *TaskContext) EnqueueTask(handler HandlerFunc, params interface{}, options ...enqueueOption) (int64, error) {
	parentTaskID := tc.taskID
	opts := []enqueueOption{}
	opts = append(opts, EnqueueWithParent(parentTaskID))
	opts = append(opts, options...)
	return tc.wb.EnqueueTask(tc, tc.executionContextID, handler, params, opts...)
}

func (tc *TaskContext) EnqueueTaskAndWait(handler HandlerFunc, params interface{}, options ...enqueueOption) ([]byte, error) {
	taskID, err := tc.EnqueueTask(handler, params, options...)
	if err != nil {
		return nil, err
	}
	return tc.wb.WaitForTaskCompletion(tc, taskID)
}

func (tc *TaskContext) RetryCount() int {
	return tc.task.RetryCount
}

// --- Tempolite Callback Implementations ---

func (wb *Tempolite) onTaskSuccess(controller retrypool.WorkerController[*Task], workerID int, worker retrypool.Worker[*Task], taskWrapper *retrypool.TaskWrapper[*Task]) {
	task := taskWrapper.Data()

	// Update task status to completed
	task.Status = TaskStatusCompleted
	task.UpdatedAt = time.Now()
	// The result should already be in task.Result
	err := wb.taskRepo.UpdateTask(wb.ctx, task)
	if err != nil {
		log.Printf("Error updating task status: %v", err)
	}

	// Remove task options
	wb.taskOptions.Delete(task.ID)
}

func (wb *Tempolite) onTaskFailure(controller retrypool.WorkerController[*Task], workerID int, worker retrypool.Worker[*Task], taskWrapper *retrypool.TaskWrapper[*Task], err error) {
	task := taskWrapper.Data()
	// Update retry count (already updated in onRetry)
	// task.RetryCount = taskWrapper.Retries()

	// Check if max retries reached
	// if task.RetryCount >= task.MaxRetries {
	// 	task.Status = TaskStatusFailed
	// } else {
	task.Status = TaskStatusPending
	// }

	// Schedule next retry with exponential backoff
	delay := time.Duration(task.RetryCount) * time.Second
	task.ScheduledAt = time.Now().Add(delay)

	task.UpdatedAt = time.Now()
	updateErr := wb.taskRepo.UpdateTask(wb.ctx, task)
	if updateErr != nil {
		log.Printf("Error updating task status: %v", updateErr)
	}
}
