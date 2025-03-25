// Package persistence provides storage and retrieval of workflow execution state.
package persistence

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// SQLiteOptions contains configuration options for SQLite persistence
type SQLiteOptions struct {
	// Path to the SQLite database file, use ":memory:" for in-memory database
	Path string

	// Enable Write-Ahead Logging mode for better concurrent performance
	UseWAL bool

	// Journal mode (DELETE, TRUNCATE, PERSIST, MEMORY, WAL, OFF)
	// If UseWAL is true, this will be set to "WAL" regardless
	JournalMode string

	// Synchronous mode (0=OFF, 1=NORMAL, 2=FULL, 3=EXTRA)
	SynchronousMode int

	// Cache size in kilobytes (negative values are pages)
	CacheSize int

	// Enable foreign key constraints
	ForeignKeys bool

	// Busy timeout in milliseconds
	BusyTimeout int

	// Maximum number of open connections
	MaxOpenConns int

	// Maximum number of idle connections
	MaxIdleConns int

	// Connection max lifetime
	ConnMaxLifetime time.Duration
}

// DefaultSQLiteOptions returns the default SQLite options
func DefaultSQLiteOptions() *SQLiteOptions {
	return &SQLiteOptions{
		Path:            "",
		UseWAL:          true,
		JournalMode:     "WAL", // WAL is generally the best option for performance
		SynchronousMode: 1,     // NORMAL - good balance of safety and performance
		CacheSize:       2000,  // 2MB cache
		ForeignKeys:     true,
		BusyTimeout:     5000, // 5 seconds
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
	}
}

// SQLiteOption is a function that configures SQLiteOptions
type SQLiteOption func(*SQLiteOptions)

// WithInMemoryDatabase configures SQLite to use an in-memory database
func WithInMemoryDatabase() SQLiteOption {
	return func(opts *SQLiteOptions) {
		opts.Path = ":memory:"
	}
}

// WithDatabasePath sets the path to the SQLite database file
func WithDatabasePath(path string) SQLiteOption {
	return func(opts *SQLiteOptions) {
		opts.Path = path
	}
}

// WithWALMode enables or disables Write-Ahead Logging mode
func WithWALMode(enabled bool) SQLiteOption {
	return func(opts *SQLiteOptions) {
		opts.UseWAL = enabled
		if enabled {
			opts.JournalMode = "WAL"
		}
	}
}

// WithJournalMode sets the journal mode for SQLite
func WithJournalMode(mode string) SQLiteOption {
	return func(opts *SQLiteOptions) {
		opts.JournalMode = mode
	}
}

// WithSynchronousMode sets the synchronous mode for SQLite
func WithSynchronousMode(mode int) SQLiteOption {
	return func(opts *SQLiteOptions) {
		opts.SynchronousMode = mode
	}
}

// WithCacheSize sets the cache size in kilobytes
func WithCacheSize(kb int) SQLiteOption {
	return func(opts *SQLiteOptions) {
		opts.CacheSize = kb
	}
}

// WithForeignKeys enables or disables foreign key constraints
func WithForeignKeys(enabled bool) SQLiteOption {
	return func(opts *SQLiteOptions) {
		opts.ForeignKeys = enabled
	}
}

// WithBusyTimeout sets the busy timeout in milliseconds
func WithBusyTimeout(timeout int) SQLiteOption {
	return func(opts *SQLiteOptions) {
		opts.BusyTimeout = timeout
	}
}

// WithMaxOpenConnections sets the maximum number of open connections
func WithMaxOpenConnections(n int) SQLiteOption {
	return func(opts *SQLiteOptions) {
		opts.MaxOpenConns = n
	}
}

// WithMaxIdleConnections sets the maximum number of idle connections
func WithMaxIdleConnections(n int) SQLiteOption {
	return func(opts *SQLiteOptions) {
		opts.MaxIdleConns = n
	}
}

// WithConnectionMaxLifetime sets the maximum connection lifetime
func WithConnectionMaxLifetime(d time.Duration) SQLiteOption {
	return func(opts *SQLiteOptions) {
		opts.ConnMaxLifetime = d
	}
}

// SQLitePersistence is a SQLite implementation of the Persistence interface.
type SQLitePersistence struct {
	db *sql.DB
	mu sync.RWMutex
}

// NewSQLitePersistence creates a new SQLite-based persistence layer.
func NewSQLitePersistence(path string, options ...SQLiteOption) (*SQLitePersistence, error) {
	// Use options pattern for configuration
	opts := DefaultSQLiteOptions()

	// Path is the only required parameter in the original API, so we set it first
	if path != "" {
		opts.Path = path
	}

	// Apply all provided options
	for _, option := range options {
		option(opts)
	}

	// If no path is set, return an error
	if opts.Path == "" {
		return nil, fmt.Errorf("database path is required")
	}

	// Build the DSN with pragmas for performance
	dsn := opts.Path

	// Add connection options
	dsn += "?"

	// Add cache size pragma
	dsn += fmt.Sprintf("_cache_size=%d", opts.CacheSize)

	// Add foreign keys pragma
	if opts.ForeignKeys {
		dsn += "&_foreign_keys=1"
	} else {
		dsn += "&_foreign_keys=0"
	}

	// Add busy timeout
	dsn += fmt.Sprintf("&_busy_timeout=%d", opts.BusyTimeout)

	// Open the SQLite database
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	// Configure the connection pool
	db.SetMaxOpenConns(opts.MaxOpenConns)
	db.SetMaxIdleConns(opts.MaxIdleConns)
	db.SetConnMaxLifetime(opts.ConnMaxLifetime)

	// Check the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping SQLite database: %w", err)
	}

	// Set journal mode and synchronous mode
	if opts.UseWAL {
		if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to set journal mode to WAL: %w", err)
		}
	} else if opts.JournalMode != "" {
		if _, err := db.Exec(fmt.Sprintf("PRAGMA journal_mode=%s", opts.JournalMode)); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to set journal mode to %s: %w", opts.JournalMode, err)
		}
	}

	if _, err := db.Exec(fmt.Sprintf("PRAGMA synchronous=%d", opts.SynchronousMode)); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to set synchronous mode to %d: %w", opts.SynchronousMode, err)
	}

	// Initialize the schema
	if err := initializeSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return &SQLitePersistence{
		db: db,
	}, nil
}

// NewSQLitePersistenceWithOptions creates a new SQLite-based persistence layer with the provided options.
// This is an alternative to NewSQLitePersistence with a more explicit options approach.
func NewSQLitePersistenceWithOptions(options ...SQLiteOption) (*SQLitePersistence, error) {
	return NewSQLitePersistence("", options...)
}

// initializeSchema creates the necessary tables in the database.
func initializeSchema(db *sql.DB) error {
	// Start a transaction for schema creation
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Core tables

	// Create runs table
	if _, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS runs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			status TEXT NOT NULL CHECK (status IN ('Pending', 'Running', 'Paused', 'Cancelled', 'Completed', 'Failed')),
			created_at TEXT NOT NULL DEFAULT (DATETIME('now')),
			updated_at TEXT NOT NULL DEFAULT (DATETIME('now'))
		)
	`); err != nil {
		return fmt.Errorf("failed to create runs table: %w", err)
	}

	// Create queues table
	if _, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS queues (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			created_at TEXT NOT NULL DEFAULT (DATETIME('now')),
			updated_at TEXT NOT NULL DEFAULT (DATETIME('now'))
		)
	`); err != nil {
		return fmt.Errorf("failed to create queues table: %w", err)
	}

	// Create event logs table
	if _, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS event_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			run_id INTEGER NOT NULL,
			entity_id INTEGER,
			entity_type TEXT,
			event_type TEXT NOT NULL,
			details TEXT,
			created_at TEXT NOT NULL DEFAULT (DATETIME('now')),
			FOREIGN KEY (run_id) REFERENCES runs(id) ON DELETE CASCADE
		)
	`); err != nil {
		return fmt.Errorf("failed to create event_logs table: %w", err)
	}

	// Workflow-related tables

	// Create workflow entities table
	if _, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS workflow_entities (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			run_id INTEGER NOT NULL,
			handler_name TEXT NOT NULL,
			step_id TEXT NOT NULL UNIQUE,
			status TEXT NOT NULL CHECK (status IN ('Pending', 'Queued', 'Running', 'Paused', 'Cancelled', 'Completed', 'Failed')),
			retry_policy TEXT NOT NULL,
			retry_state TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT (DATETIME('now')),
			updated_at TEXT NOT NULL DEFAULT (DATETIME('now')),
			FOREIGN KEY (run_id) REFERENCES runs(id) ON DELETE CASCADE
		)
	`); err != nil {
		return fmt.Errorf("failed to create workflow_entities table: %w", err)
	}

	// Create workflow executions table
	if _, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS workflow_executions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			entity_id INTEGER NOT NULL,
			status TEXT NOT NULL CHECK (status IN ('Pending', 'Queued', 'Running', 'Retried', 'Paused', 'Cancelled', 'Completed', 'Failed')),
			started_at TEXT NOT NULL DEFAULT (DATETIME('now')),
			completed_at TEXT,
			error TEXT,
			stack_trace TEXT,
			created_at TEXT NOT NULL DEFAULT (DATETIME('now')),
			updated_at TEXT NOT NULL DEFAULT (DATETIME('now')),
			FOREIGN KEY (entity_id) REFERENCES workflow_entities(id) ON DELETE CASCADE
		)
	`); err != nil {
		return fmt.Errorf("failed to create workflow_executions table: %w", err)
	}

	// Create workflow data table
	if _, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS workflow_data (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			entity_id INTEGER NOT NULL,
			duration TEXT,
			paused BOOLEAN DEFAULT 0,
			resumable BOOLEAN DEFAULT 0,
			is_root BOOLEAN DEFAULT 0,
			inputs BLOB,
			continued_from INTEGER,
			continued_execution_from INTEGER,
			workflow_step_id TEXT,
			workflow_from INTEGER,
			workflow_execution_from INTEGER,
			created_at TEXT NOT NULL DEFAULT (DATETIME('now')),
			updated_at TEXT NOT NULL DEFAULT (DATETIME('now')),
			FOREIGN KEY (entity_id) REFERENCES workflow_entities(id) ON DELETE CASCADE
		)
	`); err != nil {
		return fmt.Errorf("failed to create workflow_data table: %w", err)
	}

	// Create workflow execution data table
	if _, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS workflow_execution_data (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			execution_id INTEGER NOT NULL,
			last_heartbeat TEXT,
			outputs BLOB,
			created_at TEXT NOT NULL DEFAULT (DATETIME('now')),
			updated_at TEXT NOT NULL DEFAULT (DATETIME('now')),
			FOREIGN KEY (execution_id) REFERENCES workflow_executions(id) ON DELETE CASCADE
		)
	`); err != nil {
		return fmt.Errorf("failed to create workflow_execution_data table: %w", err)
	}

	// Activity-related tables

	// Create activity entities table
	if _, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS activity_entities (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workflow_id INTEGER NOT NULL,
			handler_name TEXT NOT NULL,
			step_id TEXT NOT NULL,
			status TEXT NOT NULL,
			retry_policy TEXT NOT NULL,
			retry_state TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT (DATETIME('now')),
			updated_at TEXT NOT NULL DEFAULT (DATETIME('now')),
			FOREIGN KEY (workflow_id) REFERENCES workflow_entities(id) ON DELETE CASCADE
		)
	`); err != nil {
		return fmt.Errorf("failed to create activity_entities table: %w", err)
	}

	// Create activity data table
	if _, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS activity_data (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			entity_id INTEGER NOT NULL,
			inputs BLOB,
			output BLOB,
			created_at TEXT NOT NULL DEFAULT (DATETIME('now')),
			updated_at TEXT NOT NULL DEFAULT (DATETIME('now')),
			FOREIGN KEY (entity_id) REFERENCES activity_entities(id) ON DELETE CASCADE
		)
	`); err != nil {
		return fmt.Errorf("failed to create activity_data table: %w", err)
	}

	// Create activity executions table
	if _, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS activity_executions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			activity_entity_id INTEGER NOT NULL,
			started_at TEXT NOT NULL DEFAULT (DATETIME('now')),
			completed_at TEXT,
			status TEXT NOT NULL,
			error TEXT,
			stack_trace TEXT,
			created_at TEXT NOT NULL DEFAULT (DATETIME('now')),
			updated_at TEXT NOT NULL DEFAULT (DATETIME('now')),
			FOREIGN KEY (activity_entity_id) REFERENCES activity_entities(id) ON DELETE CASCADE
		)
	`); err != nil {
		return fmt.Errorf("failed to create activity_executions table: %w", err)
	}

	// Create activity execution data table
	if _, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS activity_execution_data (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			execution_id INTEGER NOT NULL,
			last_heartbeat TEXT,
			outputs BLOB,
			created_at TEXT NOT NULL DEFAULT (DATETIME('now')),
			updated_at TEXT NOT NULL DEFAULT (DATETIME('now')),
			FOREIGN KEY (execution_id) REFERENCES activity_executions(id) ON DELETE CASCADE
		)
	`); err != nil {
		return fmt.Errorf("failed to create activity_execution_data table: %w", err)
	}

	// Signal-related tables

	// Create signal entities table
	if _, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS signal_entities (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			workflow_id INTEGER NOT NULL,
			handler_name TEXT NOT NULL,
			step_id TEXT NOT NULL,
			status TEXT NOT NULL,
			retry_policy TEXT NOT NULL,
			retry_state TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT (DATETIME('now')),
			updated_at TEXT NOT NULL DEFAULT (DATETIME('now')),
			FOREIGN KEY (workflow_id) REFERENCES workflow_entities(id) ON DELETE CASCADE
		)
	`); err != nil {
		return fmt.Errorf("failed to create signal_entities table: %w", err)
	}

	// Create signal data table
	if _, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS signal_data (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			entity_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT (DATETIME('now')),
			updated_at TEXT NOT NULL DEFAULT (DATETIME('now')),
			FOREIGN KEY (entity_id) REFERENCES signal_entities(id) ON DELETE CASCADE
		)
	`); err != nil {
		return fmt.Errorf("failed to create signal_data table: %w", err)
	}

	// Create signal executions table
	if _, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS signal_executions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			entity_id INTEGER NOT NULL,
			started_at TEXT NOT NULL DEFAULT (DATETIME('now')),
			completed_at TEXT,
			status TEXT NOT NULL,
			error TEXT,
			stack_trace TEXT,
			created_at TEXT NOT NULL DEFAULT (DATETIME('now')),
			updated_at TEXT NOT NULL DEFAULT (DATETIME('now')),
			FOREIGN KEY (entity_id) REFERENCES signal_entities(id) ON DELETE CASCADE
		)
	`); err != nil {
		return fmt.Errorf("failed to create signal_executions table: %w", err)
	}

	// Create signal execution data table
	if _, err := tx.Exec(`
		CREATE TABLE IF NOT EXISTS signal_execution_data (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			execution_id INTEGER NOT NULL,
			value BLOB,
			kind INTEGER NOT NULL,
			created_at TEXT NOT NULL DEFAULT (DATETIME('now')),
			updated_at TEXT NOT NULL DEFAULT (DATETIME('now')),
			FOREIGN KEY (execution_id) REFERENCES signal_executions(id) ON DELETE CASCADE
		)
	`); err != nil {
		return fmt.Errorf("failed to create signal_execution_data table: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Close closes the database connection.
func (p *SQLitePersistence) Close() error {
	return p.db.Close()
}

// CreateRun creates a new run and returns its ID.
func (p *SQLitePersistence) CreateRun() (RunID, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Insert a new run with Pending status
	result, err := p.db.Exec(
		"INSERT INTO runs (status) VALUES (?)",
		string(RunStatusPending),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create run: %w", err)
	}

	// Get the ID of the new run
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get run ID: %w", err)
	}

	return RunID(id), nil
}

// GetRun retrieves a run by ID.
func (p *SQLitePersistence) GetRun(id RunID) (*Run, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var run Run
	var statusStr string
	var createdAt, updatedAt string

	// Query the run
	err := p.db.QueryRow(
		"SELECT id, status, created_at, updated_at FROM runs WHERE id = ?",
		int64(id),
	).Scan(&run.ID, &statusStr, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get run: %w", err)
	}

	// Parse the timestamps
	run.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	run.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	run.Status = RunStatus(statusStr)

	return &run, nil
}

// UpdateRunStatus updates the status of a run.
func (p *SQLitePersistence) UpdateRunStatus(id RunID, status RunStatus) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Update the run status
	_, err := p.db.Exec(
		"UPDATE runs SET status = ?, updated_at = DATETIME('now') WHERE id = ?",
		string(status), int64(id),
	)
	if err != nil {
		return fmt.Errorf("failed to update run status: %w", err)
	}

	return nil
}

// CreateWorkflowEntity creates a new workflow entity and returns its ID.
func (p *SQLitePersistence) CreateWorkflowEntity(runID RunID, handlerName string, stepID WorkflowStepID, retryPolicy, retryState string) (WorkflowEntityID, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Insert a new workflow entity
	result, err := p.db.Exec(
		`INSERT INTO workflow_entities 
		(run_id, handler_name, step_id, status, retry_policy, retry_state) 
		VALUES (?, ?, ?, ?, ?, ?)`,
		int64(runID), handlerName, string(stepID), string(StatusPending), retryPolicy, retryState,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create workflow entity: %w", err)
	}

	// Get the ID of the new workflow entity
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get workflow entity ID: %w", err)
	}

	return WorkflowEntityID(id), nil
}

// GetWorkflowEntity retrieves a workflow entity by ID.
func (p *SQLitePersistence) GetWorkflowEntity(id WorkflowEntityID) (*WorkflowEntity, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var entity WorkflowEntity
	var statusStr string
	var createdAt, updatedAt string

	// Query the workflow entity
	err := p.db.QueryRow(
		`SELECT id, run_id, handler_name, step_id, status, retry_policy, retry_state, created_at, updated_at 
		FROM workflow_entities WHERE id = ?`,
		int64(id),
	).Scan(
		&entity.ID, &entity.RunID, &entity.HandlerName, &entity.StepID, &statusStr,
		&entity.RetryPolicy, &entity.RetryState, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow entity: %w", err)
	}

	// Parse the timestamps
	entity.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	entity.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	entity.Status = EntityStatus(statusStr)

	return &entity, nil
}

// UpdateWorkflowStatus updates the status of a workflow entity.
func (p *SQLitePersistence) UpdateWorkflowStatus(id WorkflowEntityID, status EntityStatus) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Update the workflow entity status
	_, err := p.db.Exec(
		"UPDATE workflow_entities SET status = ?, updated_at = DATETIME('now') WHERE id = ?",
		string(status), int64(id),
	)
	if err != nil {
		return fmt.Errorf("failed to update workflow status: %w", err)
	}

	return nil
}

// CreateWorkflowExecution creates a new workflow execution and returns its ID.
func (p *SQLitePersistence) CreateWorkflowExecution(workflowID WorkflowEntityID) (WorkflowExecutionID, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Insert a new workflow execution
	result, err := p.db.Exec(
		"INSERT INTO workflow_executions (entity_id, status) VALUES (?, ?)",
		int64(workflowID), string(ExecutionStatusPending),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create workflow execution: %w", err)
	}

	// Get the ID of the new workflow execution
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get workflow execution ID: %w", err)
	}

	return WorkflowExecutionID(id), nil
}

// UpdateWorkflowExecution updates a workflow execution with results.
func (p *SQLitePersistence) UpdateWorkflowExecution(id WorkflowExecutionID, status ExecutionStatus, output []byte, err error) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Start a transaction
	tx, dbErr := p.db.Begin()
	if dbErr != nil {
		return fmt.Errorf("failed to begin transaction: %w", dbErr)
	}
	defer tx.Rollback()

	var errStr, stackTrace *string
	if err != nil {
		str := err.Error()
		errStr = &str
		// In a real implementation, we would also capture the stack trace
	}

	// Update the workflow execution
	_, dbErr = tx.Exec(
		`UPDATE workflow_executions 
		SET status = ?, completed_at = DATETIME('now'), error = ?, stack_trace = ?, updated_at = DATETIME('now') 
		WHERE id = ?`,
		string(status), errStr, stackTrace, int64(id),
	)
	if dbErr != nil {
		return fmt.Errorf("failed to update workflow execution: %w", dbErr)
	}

	// If there's output, update the workflow execution data
	if output != nil {
		// Check if execution data exists
		var count int
		if err := tx.QueryRow(
			"SELECT COUNT(*) FROM workflow_execution_data WHERE execution_id = ?",
			int64(id),
		).Scan(&count); err != nil {
			return fmt.Errorf("failed to check for workflow execution data: %w", err)
		}

		if count == 0 {
			// Insert new execution data
			_, err := tx.Exec(
				"INSERT INTO workflow_execution_data (execution_id, outputs) VALUES (?, ?)",
				int64(id), output,
			)
			if err != nil {
				return fmt.Errorf("failed to insert workflow execution data: %w", err)
			}
		} else {
			// Update existing execution data
			_, err := tx.Exec(
				"UPDATE workflow_execution_data SET outputs = ?, updated_at = DATETIME('now') WHERE execution_id = ?",
				output, int64(id),
			)
			if err != nil {
				return fmt.Errorf("failed to update workflow execution data: %w", err)
			}
		}
	}

	// Get the workflow entity ID for this execution
	var workflowID int64
	if err := tx.QueryRow(
		"SELECT entity_id FROM workflow_executions WHERE id = ?",
		int64(id),
	).Scan(&workflowID); err != nil {
		return fmt.Errorf("failed to get workflow entity ID: %w", err)
	}

	// Update the workflow entity status based on execution status
	var entityStatus string
	switch status {
	case ExecutionStatusCompleted:
		entityStatus = string(StatusCompleted)
	case ExecutionStatusFailed:
		entityStatus = string(StatusFailed)
	case ExecutionStatusPaused:
		entityStatus = string(StatusPaused)
	case ExecutionStatusCancelled:
		entityStatus = string(StatusCancelled)
	case ExecutionStatusRunning:
		entityStatus = string(StatusRunning)
	default:
		// Don't update entity status for other execution statuses
		return tx.Commit()
	}

	// Update the workflow entity status
	_, dbErr = tx.Exec(
		"UPDATE workflow_entities SET status = ?, updated_at = DATETIME('now') WHERE id = ?",
		entityStatus, workflowID,
	)
	if dbErr != nil {
		return fmt.Errorf("failed to update workflow entity status: %w", dbErr)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// CreateActivityEntity creates a new activity entity and returns its ID.
func (p *SQLitePersistence) CreateActivityEntity(workflowID WorkflowEntityID, handlerName string, stepID ActivityStepID, retryPolicy, retryState string) (ActivityEntityID, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Insert a new activity entity
	result, err := p.db.Exec(
		`INSERT INTO activity_entities 
		(workflow_id, handler_name, step_id, status, retry_policy, retry_state) 
		VALUES (?, ?, ?, ?, ?, ?)`,
		int64(workflowID), handlerName, string(stepID), string(StatusPending), retryPolicy, retryState,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create activity entity: %w", err)
	}

	// Get the ID of the new activity entity
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get activity entity ID: %w", err)
	}

	return ActivityEntityID(id), nil
}

// UpdateActivityStatus updates the status of an activity entity.
func (p *SQLitePersistence) UpdateActivityStatus(id ActivityEntityID, status EntityStatus) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Update the activity entity status
	_, err := p.db.Exec(
		"UPDATE activity_entities SET status = ?, updated_at = DATETIME('now') WHERE id = ?",
		string(status), int64(id),
	)
	if err != nil {
		return fmt.Errorf("failed to update activity status: %w", err)
	}

	return nil
}

// CreateSignalEntity creates a new signal entity and returns its ID.
func (p *SQLitePersistence) CreateSignalEntity(workflowID WorkflowEntityID, name string) (SignalEntityID, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Generate a step ID for the signal
	stepID := fmt.Sprintf("signal-%s-%d", name, time.Now().UnixNano())

	// Create default retry policy and state
	retryPolicy := RetryPolicy{MaxAttempts: 1, MaxInterval: 0}
	retryState := RetryState{Attempts: 0}

	// Convert to JSON
	retryPolicyJSON, err := json.Marshal(retryPolicy)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal retry policy: %w", err)
	}

	retryStateJSON, err := json.Marshal(retryState)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal retry state: %w", err)
	}

	// Insert a new signal entity
	result, err := p.db.Exec(
		`INSERT INTO signal_entities 
		(workflow_id, handler_name, step_id, status, retry_policy, retry_state) 
		VALUES (?, ?, ?, ?, ?, ?)`,
		int64(workflowID), "signal-handler", stepID, string(StatusPending),
		string(retryPolicyJSON), string(retryStateJSON),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create signal entity: %w", err)
	}

	// Get the ID of the new signal entity
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get signal entity ID: %w", err)
	}

	// Insert signal data
	_, err = p.db.Exec(
		"INSERT INTO signal_data (entity_id, name) VALUES (?, ?)",
		id, name,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create signal data: %w", err)
	}

	return SignalEntityID(id), nil
}

// SendSignal sends a signal to a workflow.
func (p *SQLitePersistence) SendSignal(signalID SignalEntityID, value []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Create a signal execution
	result, err := p.db.Exec(
		"INSERT INTO signal_executions (entity_id, status) VALUES (?, ?)",
		int64(signalID), string(ExecutionStatusPending),
	)
	if err != nil {
		return fmt.Errorf("failed to create signal execution: %w", err)
	}

	// Get the ID of the new signal execution
	execID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get signal execution ID: %w", err)
	}

	// Insert signal execution data
	_, err = p.db.Exec(
		"INSERT INTO signal_execution_data (execution_id, value, kind) VALUES (?, ?, ?)",
		execID, value, 0, // Kind 0 = normal signal
	)
	if err != nil {
		return fmt.Errorf("failed to create signal execution data: %w", err)
	}

	// Update signal entity status
	_, err = p.db.Exec(
		"UPDATE signal_entities SET status = ?, updated_at = DATETIME('now') WHERE id = ?",
		string(StatusQueued), int64(signalID),
	)
	if err != nil {
		return fmt.Errorf("failed to update signal status: %w", err)
	}

	return nil
}

// LogEvent logs an event for a run.
func (p *SQLitePersistence) LogEvent(eventType EventType, runID RunID, entityType EntityType, entityID int64, details map[string]interface{}) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Convert details to JSON
	detailsJSON, err := json.Marshal(details)
	if err != nil {
		return fmt.Errorf("failed to marshal event details: %w", err)
	}

	// Insert event log
	_, err = p.db.Exec(
		"INSERT INTO event_logs (run_id, entity_id, entity_type, event_type, details) VALUES (?, ?, ?, ?, ?)",
		int64(runID), entityID, string(entityType), string(eventType), string(detailsJSON),
	)
	if err != nil {
		return fmt.Errorf("failed to log event: %w", err)
	}

	return nil
}

// UpdateWorkflowRetryState updates the retry state of a workflow entity.
func (p *SQLitePersistence) UpdateWorkflowRetryState(id WorkflowEntityID, retryState string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Update the workflow entity retry state
	_, err := p.db.Exec(
		"UPDATE workflow_entities SET retry_state = ?, updated_at = DATETIME('now') WHERE id = ?",
		retryState, int64(id),
	)
	if err != nil {
		return fmt.Errorf("failed to update workflow retry state: %w", err)
	}

	return nil
}

// GetWorkflowExecutions retrieves all executions for a workflow entity.
func (p *SQLitePersistence) GetWorkflowExecutions(workflowID WorkflowEntityID) ([]*WorkflowExecution, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Query all workflow executions for this entity
	rows, err := p.db.Query(
		`SELECT id, entity_id, started_at, completed_at, status, error, stack_trace, created_at, updated_at 
		FROM workflow_executions WHERE entity_id = ? ORDER BY id ASC`,
		int64(workflowID),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query workflow executions: %w", err)
	}
	defer rows.Close()

	var executions []*WorkflowExecution
	for rows.Next() {
		var execution WorkflowExecution
		var startedAt, createdAt, updatedAt string
		var completedAt sql.NullString
		var statusStr string
		var error sql.NullString
		var stackTrace sql.NullString

		// Scan row into variables
		err := rows.Scan(
			&execution.ID, &execution.WorkflowEntityID, &startedAt, &completedAt, &statusStr,
			&error, &stackTrace, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workflow execution: %w", err)
		}

		// Parse timestamps
		execution.StartedAt, _ = time.Parse(time.RFC3339, startedAt)
		execution.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		execution.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

		if completedAt.Valid {
			parsed, _ := time.Parse(time.RFC3339, completedAt.String)
			execution.CompletedAt = &parsed
		}

		// Set status and error
		execution.Status = ExecutionStatus(statusStr)
		if error.Valid {
			execution.Error = error.String
		}
		if stackTrace.Valid {
			execution.StackTrace = stackTrace.String
		}

		executions = append(executions, &execution)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating workflow executions: %w", err)
	}

	return executions, nil
}

// GetWorkflowExecution retrieves a single workflow execution by ID.
func (p *SQLitePersistence) GetWorkflowExecution(id WorkflowExecutionID) (*WorkflowExecution, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var execution WorkflowExecution
	var startedAt, createdAt, updatedAt string
	var completedAt sql.NullString
	var statusStr string
	var error sql.NullString
	var stackTrace sql.NullString

	// Query the workflow execution
	err := p.db.QueryRow(
		`SELECT id, entity_id, started_at, completed_at, status, error, stack_trace, created_at, updated_at 
		FROM workflow_executions WHERE id = ?`,
		int64(id),
	).Scan(
		&execution.ID, &execution.WorkflowEntityID, &startedAt, &completedAt, &statusStr,
		&error, &stackTrace, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow execution: %w", err)
	}

	// Parse timestamps
	execution.StartedAt, _ = time.Parse(time.RFC3339, startedAt)
	execution.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	execution.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	if completedAt.Valid {
		parsed, _ := time.Parse(time.RFC3339, completedAt.String)
		execution.CompletedAt = &parsed
	}

	// Set status and error
	execution.Status = ExecutionStatus(statusStr)
	if error.Valid {
		execution.Error = error.String
	}
	if stackTrace.Valid {
		execution.StackTrace = stackTrace.String
	}

	return &execution, nil
}

// UpdateWorkflowExecutionStatus updates just the status of a workflow execution.
func (p *SQLitePersistence) UpdateWorkflowExecutionStatus(id WorkflowExecutionID, status ExecutionStatus) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Update the workflow execution status
	_, err := p.db.Exec(
		"UPDATE workflow_executions SET status = ?, updated_at = DATETIME('now') WHERE id = ?",
		string(status), int64(id),
	)
	if err != nil {
		return fmt.Errorf("failed to update workflow execution status: %w", err)
	}

	return nil
}

// PauseWorkflow pauses a workflow and its current execution.
func (p *SQLitePersistence) PauseWorkflow(workflowID WorkflowEntityID, executionID WorkflowExecutionID) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Start a transaction to update both the entity and execution
	tx, err := p.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update the workflow entity status
	_, err = tx.Exec(
		"UPDATE workflow_entities SET status = ?, updated_at = DATETIME('now') WHERE id = ?",
		string(StatusPaused), int64(workflowID),
	)
	if err != nil {
		return fmt.Errorf("failed to update workflow status: %w", err)
	}

	// Update the workflow execution status
	_, err = tx.Exec(
		"UPDATE workflow_executions SET status = ?, updated_at = DATETIME('now') WHERE id = ?",
		string(ExecutionStatusPaused), int64(executionID),
	)
	if err != nil {
		return fmt.Errorf("failed to update workflow execution status: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ResumeWorkflow resumes a paused workflow and its execution.
func (p *SQLitePersistence) ResumeWorkflow(workflowID WorkflowEntityID, executionID WorkflowExecutionID) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Start a transaction to update both the entity and execution
	tx, err := p.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update the workflow entity status
	_, err = tx.Exec(
		"UPDATE workflow_entities SET status = ?, updated_at = DATETIME('now') WHERE id = ?",
		string(StatusRunning), int64(workflowID),
	)
	if err != nil {
		return fmt.Errorf("failed to update workflow status: %w", err)
	}

	// Update the workflow execution status
	_, err = tx.Exec(
		"UPDATE workflow_executions SET status = ?, updated_at = DATETIME('now') WHERE id = ?",
		string(ExecutionStatusRunning), int64(executionID),
	)
	if err != nil {
		return fmt.Errorf("failed to update workflow execution status: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetActivityEntity retrieves an activity entity by ID.
func (p *SQLitePersistence) GetActivityEntity(id ActivityEntityID) (*ActivityEntity, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var entity ActivityEntity
	var statusStr string
	var createdAt, updatedAt string

	// Query the activity entity
	err := p.db.QueryRow(
		`SELECT id, workflow_id, handler_name, step_id, status, retry_policy, retry_state, created_at, updated_at 
		FROM activity_entities WHERE id = ?`,
		int64(id),
	).Scan(
		&entity.ID, &entity.WorkflowID, &entity.HandlerName, &entity.StepID, &statusStr,
		&entity.RetryPolicy, &entity.RetryState, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity entity: %w", err)
	}

	// Parse the timestamps
	entity.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	entity.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	entity.Status = EntityStatus(statusStr)

	return &entity, nil
}

// GetActivityData retrieves the activity data for an entity.
func (p *SQLitePersistence) GetActivityData(id ActivityEntityID) (*ActivityData, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var data ActivityData
	var createdAt, updatedAt string

	// Query the activity data
	err := p.db.QueryRow(
		`SELECT id, entity_id, inputs, output, created_at, updated_at 
		FROM activity_data WHERE entity_id = ?`,
		int64(id),
	).Scan(
		&data.ID, &data.EntityID, &data.Inputs, &data.Output, &createdAt, &updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			// Create empty activity data entry if it doesn't exist
			result, err := p.db.Exec(
				"INSERT INTO activity_data (entity_id) VALUES (?)",
				int64(id),
			)
			if err != nil {
				return nil, fmt.Errorf("failed to create activity data: %w", err)
			}

			dataID, err := result.LastInsertId()
			if err != nil {
				return nil, fmt.Errorf("failed to get activity data ID: %w", err)
			}

			return &ActivityData{
				ID:        ActivityDataID(dataID),
				EntityID:  id,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("failed to get activity data: %w", err)
	}

	// Parse the timestamps
	data.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	data.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	return &data, nil
}

// UpdateActivityRetryState updates the retry state of an activity entity.
func (p *SQLitePersistence) UpdateActivityRetryState(id ActivityEntityID, retryState string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Update the activity entity retry state
	_, err := p.db.Exec(
		"UPDATE activity_entities SET retry_state = ?, updated_at = DATETIME('now') WHERE id = ?",
		retryState, int64(id),
	)
	if err != nil {
		return fmt.Errorf("failed to update activity retry state: %w", err)
	}

	return nil
}

// CreateActivityExecution creates a new activity execution and returns its ID.
func (p *SQLitePersistence) CreateActivityExecution(activityID ActivityEntityID) (ActivityExecutionID, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Insert a new activity execution
	result, err := p.db.Exec(
		"INSERT INTO activity_executions (activity_entity_id, started_at, status) VALUES (?, DATETIME('now'), ?)",
		int64(activityID), string(ExecutionStatusRunning),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create activity execution: %w", err)
	}

	// Get the ID of the new activity execution
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get activity execution ID: %w", err)
	}

	return ActivityExecutionID(id), nil
}

// UpdateActivityExecution updates an activity execution with results.
func (p *SQLitePersistence) UpdateActivityExecution(id ActivityExecutionID, status ExecutionStatus, output []byte, err error) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	tx, dbErr := p.db.Begin()
	if dbErr != nil {
		return fmt.Errorf("failed to begin transaction: %w", dbErr)
	}
	defer tx.Rollback()

	var errStr, stackTrace *string
	if err != nil {
		str := err.Error()
		errStr = &str
		// In a real implementation, we would also capture the stack trace
	}

	// Update the activity execution
	_, dbErr = tx.Exec(
		`UPDATE activity_executions 
		SET status = ?, completed_at = DATETIME('now'), error = ?, stack_trace = ?, updated_at = DATETIME('now') 
		WHERE id = ?`,
		string(status), errStr, stackTrace, int64(id),
	)
	if dbErr != nil {
		return fmt.Errorf("failed to update activity execution: %w", dbErr)
	}

	// If there's output, update the activity execution data
	if output != nil {
		// Check if execution data exists
		var count int
		if err := tx.QueryRow(
			"SELECT COUNT(*) FROM activity_execution_data WHERE execution_id = ?",
			int64(id),
		).Scan(&count); err != nil {
			return fmt.Errorf("failed to check for activity execution data: %w", err)
		}

		if count == 0 {
			// Insert new execution data
			_, err := tx.Exec(
				"INSERT INTO activity_execution_data (execution_id, outputs) VALUES (?, ?)",
				int64(id), output,
			)
			if err != nil {
				return fmt.Errorf("failed to insert activity execution data: %w", err)
			}
		} else {
			// Update existing execution data
			_, err := tx.Exec(
				"UPDATE activity_execution_data SET outputs = ?, updated_at = DATETIME('now') WHERE execution_id = ?",
				output, int64(id),
			)
			if err != nil {
				return fmt.Errorf("failed to update activity execution data: %w", err)
			}
		}
	}

	// Get the activity entity ID for this execution
	var activityID int64
	if err := tx.QueryRow(
		"SELECT activity_entity_id FROM activity_executions WHERE id = ?",
		int64(id),
	).Scan(&activityID); err != nil {
		return fmt.Errorf("failed to get activity entity ID: %w", err)
	}

	// Only update the activity entity status for terminal execution statuses
	if status == ExecutionStatusCompleted || status == ExecutionStatusFailed ||
		status == ExecutionStatusCancelled || status == ExecutionStatusCompensated {
		var entityStatus string

		switch status {
		case ExecutionStatusCompleted:
			entityStatus = string(StatusCompleted)
		case ExecutionStatusFailed:
			entityStatus = string(StatusFailed)
		case ExecutionStatusCancelled:
			entityStatus = string(StatusCancelled)
		case ExecutionStatusCompensated:
			entityStatus = string(StatusCompensated)
		}

		// Update the activity entity status
		_, dbErr = tx.Exec(
			"UPDATE activity_entities SET status = ?, updated_at = DATETIME('now') WHERE id = ?",
			entityStatus, activityID,
		)
		if dbErr != nil {
			return fmt.Errorf("failed to update activity entity status: %w", dbErr)
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetActivityExecution retrieves an activity execution by ID.
func (p *SQLitePersistence) GetActivityExecution(id ActivityExecutionID) (*ActivityExecution, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var execution ActivityExecution
	var startedAt, createdAt, updatedAt string
	var completedAt sql.NullString
	var statusStr string
	var error sql.NullString
	var stackTrace sql.NullString

	// Query the activity execution
	err := p.db.QueryRow(
		`SELECT id, activity_entity_id, started_at, completed_at, status, error, stack_trace, created_at, updated_at 
		FROM activity_executions WHERE id = ?`,
		int64(id),
	).Scan(
		&execution.ID, &execution.ActivityEntityID, &startedAt, &completedAt, &statusStr,
		&error, &stackTrace, &createdAt, &updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get activity execution: %w", err)
	}

	// Parse timestamps
	execution.StartedAt, _ = time.Parse(time.RFC3339, startedAt)
	execution.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	execution.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	if completedAt.Valid {
		parsed, _ := time.Parse(time.RFC3339, completedAt.String)
		execution.CompletedAt = &parsed
	}

	// Set status and error
	execution.Status = ExecutionStatus(statusStr)
	if error.Valid {
		execution.Error = error.String
	}
	if stackTrace.Valid {
		execution.StackTrace = stackTrace.String
	}

	return &execution, nil
}

// RecordActivityHeartbeat records a heartbeat for an activity execution.
func (p *SQLitePersistence) RecordActivityHeartbeat(id ActivityExecutionID, timestamp time.Time) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if execution data exists
	var count int
	if err := p.db.QueryRow(
		"SELECT COUNT(*) FROM activity_execution_data WHERE execution_id = ?",
		int64(id),
	).Scan(&count); err != nil {
		return fmt.Errorf("failed to check for activity execution data: %w", err)
	}

	if count == 0 {
		// Insert new execution data with heartbeat
		_, err := p.db.Exec(
			"INSERT INTO activity_execution_data (execution_id, last_heartbeat) VALUES (?, ?)",
			int64(id), timestamp.Format(time.RFC3339),
		)
		if err != nil {
			return fmt.Errorf("failed to insert activity execution data with heartbeat: %w", err)
		}
	} else {
		// Update existing execution data with heartbeat
		_, err := p.db.Exec(
			"UPDATE activity_execution_data SET last_heartbeat = ?, updated_at = DATETIME('now') WHERE execution_id = ?",
			timestamp.Format(time.RFC3339), int64(id),
		)
		if err != nil {
			return fmt.Errorf("failed to update activity execution data with heartbeat: %w", err)
		}
	}

	return nil
}

// GetWorkflowData retrieves the workflow data by entity ID.
func (p *SQLitePersistence) GetWorkflowData(id WorkflowEntityID) (*WorkflowData, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var data WorkflowData
	var createdAt, updatedAt string
	var continuedFrom, continuedExecutionFrom, workflowFrom, workflowExecutionFrom sql.NullInt64
	var workflowStepID sql.NullString

	// Query the workflow data
	err := p.db.QueryRow(
		`SELECT id, entity_id, duration, paused, resumable, is_root, inputs, 
		continued_from, continued_execution_from, workflow_step_id, 
		workflow_from, workflow_execution_from, created_at, updated_at 
		FROM workflow_data WHERE entity_id = ?`,
		int64(id),
	).Scan(
		&data.ID, &data.EntityID, &data.Duration, &data.Paused, &data.Resumable, &data.IsRoot, &data.Inputs,
		&continuedFrom, &continuedExecutionFrom, &workflowStepID,
		&workflowFrom, &workflowExecutionFrom, &createdAt, &updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			// Create empty workflow data entry if it doesn't exist
			result, err := p.db.Exec(
				"INSERT INTO workflow_data (entity_id, is_root) VALUES (?, ?)",
				int64(id), true,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to create workflow data: %w", err)
			}

			dataID, err := result.LastInsertId()
			if err != nil {
				return nil, fmt.Errorf("failed to get workflow data ID: %w", err)
			}

			return &WorkflowData{
				ID:        WorkflowDataID(dataID),
				EntityID:  id,
				IsRoot:    true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		}
		return nil, fmt.Errorf("failed to get workflow data: %w", err)
	}

	// Parse timestamps
	data.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	data.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	// Handle nullable fields
	if continuedFrom.Valid {
		wfID := WorkflowEntityID(continuedFrom.Int64)
		data.ContinuedFrom = &wfID
	}

	if continuedExecutionFrom.Valid {
		execID := WorkflowExecutionID(continuedExecutionFrom.Int64)
		data.ContinuedExecutionFrom = &execID
	}

	if workflowStepID.Valid {
		data.WorkflowStepID = &workflowStepID.String
	}

	if workflowFrom.Valid {
		wfID := WorkflowEntityID(workflowFrom.Int64)
		data.WorkflowFrom = &wfID
	}

	if workflowExecutionFrom.Valid {
		execID := WorkflowExecutionID(workflowExecutionFrom.Int64)
		data.WorkflowExecutionFrom = &execID
	}

	return &data, nil
}

// GetWorkflowExecutionData retrieves the execution data for a workflow execution
func (p *SQLitePersistence) GetWorkflowExecutionData(executionID WorkflowExecutionID) (*WorkflowExecutionData, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var data WorkflowExecutionData
	var createdAt, updatedAt string
	var lastHeartbeat sql.NullString

	// Query the workflow execution data
	err := p.db.QueryRow(
		`SELECT id, execution_id, last_heartbeat, outputs, created_at, updated_at 
		FROM workflow_execution_data WHERE execution_id = ?`,
		int64(executionID),
	).Scan(
		&data.ID, &data.ExecutionID, &lastHeartbeat, &data.Outputs, &createdAt, &updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no execution data found for execution ID %d", executionID)
		}
		return nil, fmt.Errorf("failed to get workflow execution data: %w", err)
	}

	// Parse timestamps
	data.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	data.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	if lastHeartbeat.Valid {
		parsed, _ := time.Parse(time.RFC3339, lastHeartbeat.String)
		data.LastHeartbeat = &parsed
	}

	return &data, nil
}
