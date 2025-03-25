package persistence

import (
	"os"
	"path/filepath"
	"testing"
)

// TestSQLiteWorkflowRetry tests the workflow retry functionality with SQLite persistence
func TestSQLiteWorkflowRetry(t *testing.T) {
	// Create a temporary directory for the test database
	tempDir, err := os.MkdirTemp("", "tempolite-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a database file path
	dbPath := filepath.Join(tempDir, "test.db")

	// Create a new SQLite persistence instance
	db, err := NewSQLitePersistence(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite persistence: %v", err)
	}
	defer db.Close()

	// Create a test run
	runID, err := db.CreateRun()
	if err != nil {
		t.Fatalf("Failed to create run: %v", err)
	}

	// Create a workflow entity with retry policy
	retryPolicy := `{"max_attempts": 3, "max_interval": 1000}`
	retryState := `{"attempts": 0, "timeout": 0}`
	workflowID, err := db.CreateWorkflowEntity(runID, "TestWorkflow", "workflow-1", retryPolicy, retryState)
	if err != nil {
		t.Fatalf("Failed to create workflow entity: %v", err)
	}

	// Create a workflow execution
	executionID, err := db.CreateWorkflowExecution(workflowID)
	if err != nil {
		t.Fatalf("Failed to create workflow execution: %v", err)
	}

	// Simulate a failure - update execution with a failure
	err = db.UpdateWorkflowExecution(executionID, ExecutionStatusFailed, nil,
		NewError("test workflow failure"))
	if err != nil {
		t.Fatalf("Failed to update workflow execution: %v", err)
	}

	// Check the workflow entity status
	entity, err := db.GetWorkflowEntity(workflowID)
	if err != nil {
		t.Fatalf("Failed to get workflow entity: %v", err)
	}

	// Ensure the workflow is marked for retry
	if entity.Status != StatusFailed {
		t.Errorf("Expected workflow status to be Failed, got %s", entity.Status)
	}

	// Update retry state
	newRetryState := `{"attempts": 1, "timeout": 0}`
	err = db.UpdateWorkflowRetryState(workflowID, newRetryState)
	if err != nil {
		t.Fatalf("Failed to update retry state: %v", err)
	}

	// Create a second execution for retry
	execution2ID, err := db.CreateWorkflowExecution(workflowID)
	if err != nil {
		t.Fatalf("Failed to create second workflow execution: %v", err)
	}

	// Simulate successful completion on retry
	err = db.UpdateWorkflowExecution(execution2ID, ExecutionStatusCompleted, []byte(`{"result": "success"}`), nil)
	if err != nil {
		t.Fatalf("Failed to update second workflow execution: %v", err)
	}

	// Check the workflow entity status again
	entity, err = db.GetWorkflowEntity(workflowID)
	if err != nil {
		t.Fatalf("Failed to get workflow entity: %v", err)
	}

	// Ensure the workflow is now completed
	if entity.Status != StatusCompleted {
		t.Errorf("Expected workflow status to be Completed, got %s", entity.Status)
	}

	// Verify execution history
	executions, err := db.GetWorkflowExecutions(workflowID)
	if err != nil {
		t.Fatalf("Failed to get workflow executions: %v", err)
	}

	if len(executions) != 2 {
		t.Errorf("Expected 2 executions, got %d", len(executions))
	}

	if executions[0].Status != ExecutionStatusFailed {
		t.Errorf("Expected first execution status to be Failed, got %s", executions[0].Status)
	}

	if executions[1].Status != ExecutionStatusCompleted {
		t.Errorf("Expected second execution status to be Completed, got %s", executions[1].Status)
	}
}

// Helper function for the test
func NewError(msg string) error {
	return ErrTestFailure
}

// Define error for testing
var ErrTestFailure = &Error{message: "test workflow failure"}

// Simple Error implementation for testing
type Error struct {
	message string
}

func (e *Error) Error() string {
	return e.message
}

// TestSQLiteWorkflowPause tests pausing and resuming a workflow
func TestSQLiteWorkflowPause(t *testing.T) {
	// Create a temporary directory for the test database
	tempDir, err := os.MkdirTemp("", "tempolite-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a database file path
	dbPath := filepath.Join(tempDir, "test.db")

	// Create a new SQLite persistence instance
	db, err := NewSQLitePersistence(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite persistence: %v", err)
	}
	defer db.Close()

	// Create a test run
	runID, err := db.CreateRun()
	if err != nil {
		t.Fatalf("Failed to create run: %v", err)
	}

	// Create a workflow entity
	retryPolicy := `{"max_attempts": 3, "max_interval": 1000}`
	retryState := `{"attempts": 0, "timeout": 0}`
	workflowID, err := db.CreateWorkflowEntity(runID, "TestPauseWorkflow", "workflow-pause-1", retryPolicy, retryState)
	if err != nil {
		t.Fatalf("Failed to create workflow entity: %v", err)
	}

	// Create a workflow execution
	executionID, err := db.CreateWorkflowExecution(workflowID)
	if err != nil {
		t.Fatalf("Failed to create workflow execution: %v", err)
	}

	// Simulate workflow running
	err = db.UpdateWorkflowStatus(workflowID, StatusRunning)
	if err != nil {
		t.Fatalf("Failed to update workflow status: %v", err)
	}

	// Ensure execution is running
	err = db.UpdateWorkflowExecutionStatus(executionID, ExecutionStatusRunning)
	if err != nil {
		t.Fatalf("Failed to update execution status: %v", err)
	}

	// Simulate workflow pause
	err = db.PauseWorkflow(workflowID, executionID)
	if err != nil {
		t.Fatalf("Failed to pause workflow: %v", err)
	}

	// Check the workflow entity status
	entity, err := db.GetWorkflowEntity(workflowID)
	if err != nil {
		t.Fatalf("Failed to get workflow entity: %v", err)
	}

	// Ensure the workflow is paused
	if entity.Status != StatusPaused {
		t.Errorf("Expected workflow status to be Paused, got %s", entity.Status)
	}

	// Get the execution
	execution, err := db.GetWorkflowExecution(executionID)
	if err != nil {
		t.Fatalf("Failed to get workflow execution: %v", err)
	}

	// Ensure the execution is paused
	if execution.Status != ExecutionStatusPaused {
		t.Errorf("Expected execution status to be Paused, got %s", execution.Status)
	}

	// Resume the workflow
	err = db.ResumeWorkflow(workflowID, executionID)
	if err != nil {
		t.Fatalf("Failed to resume workflow: %v", err)
	}

	// Check the workflow entity status again
	entity, err = db.GetWorkflowEntity(workflowID)
	if err != nil {
		t.Fatalf("Failed to get workflow entity: %v", err)
	}

	// Ensure the workflow is now running
	if entity.Status != StatusRunning {
		t.Errorf("Expected workflow status to be Running, got %s", entity.Status)
	}

	// Get the execution again
	execution, err = db.GetWorkflowExecution(executionID)
	if err != nil {
		t.Fatalf("Failed to get workflow execution: %v", err)
	}

	// Ensure the execution is running
	if execution.Status != ExecutionStatusRunning {
		t.Errorf("Expected execution status to be Running, got %s", execution.Status)
	}
}

func TestGetWorkflowExecutions(t *testing.T) {
	// Create a temporary directory for the test database
	tempDir, err := os.MkdirTemp("", "tempolite-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a database file path
	dbPath := filepath.Join(tempDir, "test.db")

	// Create a new SQLite persistence instance
	db, err := NewSQLitePersistence(dbPath)
	if err != nil {
		t.Fatalf("Failed to create SQLite persistence: %v", err)
	}
	defer db.Close()

	// Create a test run
	runID, err := db.CreateRun()
	if err != nil {
		t.Fatalf("Failed to create run: %v", err)
	}

	// Create a workflow entity
	retryPolicy := `{"max_attempts": 3, "max_interval": 1000}`
	retryState := `{"attempts": 0, "timeout": 0}`
	workflowID, err := db.CreateWorkflowEntity(runID, "TestExecutionWorkflow", "workflow-exec-1", retryPolicy, retryState)
	if err != nil {
		t.Fatalf("Failed to create workflow entity: %v", err)
	}

	// Create multiple workflow executions to test retrieval
	// First execution - Failed
	execution1ID, err := db.CreateWorkflowExecution(workflowID)
	if err != nil {
		t.Fatalf("Failed to create first workflow execution: %v", err)
	}
	err = db.UpdateWorkflowExecution(execution1ID, ExecutionStatusFailed, nil, NewError("first execution failure"))
	if err != nil {
		t.Fatalf("Failed to update first workflow execution: %v", err)
	}

	// Second execution - Completed
	execution2ID, err := db.CreateWorkflowExecution(workflowID)
	if err != nil {
		t.Fatalf("Failed to create second workflow execution: %v", err)
	}
	err = db.UpdateWorkflowExecution(execution2ID, ExecutionStatusCompleted, []byte(`{"result": "success"}`), nil)
	if err != nil {
		t.Fatalf("Failed to update second workflow execution: %v", err)
	}

	// Third execution - Paused
	execution3ID, err := db.CreateWorkflowExecution(workflowID)
	if err != nil {
		t.Fatalf("Failed to create third workflow execution: %v", err)
	}
	err = db.UpdateWorkflowExecutionStatus(execution3ID, ExecutionStatusPaused)
	if err != nil {
		t.Fatalf("Failed to update third workflow execution: %v", err)
	}

	// Get all workflow executions
	executions, err := db.GetWorkflowExecutions(workflowID)
	if err != nil {
		t.Fatalf("Failed to get workflow executions: %v", err)
	}

	// Verify we got the expected number of executions
	if len(executions) != 3 {
		t.Errorf("Expected 3 executions, got %d", len(executions))
	}

	// Verify the executions are returned in the correct order (by ID ascending)
	if len(executions) >= 1 && executions[0].ID != execution1ID {
		t.Errorf("Expected first execution ID to be %d, got %d", execution1ID, executions[0].ID)
	}

	if len(executions) >= 2 && executions[1].ID != execution2ID {
		t.Errorf("Expected second execution ID to be %d, got %d", execution2ID, executions[1].ID)
	}

	if len(executions) >= 3 && executions[2].ID != execution3ID {
		t.Errorf("Expected third execution ID to be %d, got %d", execution3ID, executions[2].ID)
	}

	// Verify the status of each execution
	if len(executions) >= 1 && executions[0].Status != ExecutionStatusFailed {
		t.Errorf("Expected first execution status to be Failed, got %s", executions[0].Status)
	}

	if len(executions) >= 2 && executions[1].Status != ExecutionStatusCompleted {
		t.Errorf("Expected second execution status to be Completed, got %s", executions[1].Status)
	}

	if len(executions) >= 3 && executions[2].Status != ExecutionStatusPaused {
		t.Errorf("Expected third execution status to be Paused, got %s", executions[2].Status)
	}

	// Verify the first execution has an error
	if len(executions) >= 1 && executions[0].Error == "" {
		t.Errorf("Expected first execution to have an error message")
	}

	// Test getting a single execution
	singleExecution, err := db.GetWorkflowExecution(execution2ID)
	if err != nil {
		t.Fatalf("Failed to get single workflow execution: %v", err)
	}

	if singleExecution.ID != execution2ID {
		t.Errorf("Expected to get execution with ID %d, got %d", execution2ID, singleExecution.ID)
	}

	if singleExecution.Status != ExecutionStatusCompleted {
		t.Errorf("Expected execution status to be Completed, got %s", singleExecution.Status)
	}
}

// TestSQLiteInMemoryDatabase tests creating and using an in-memory SQLite database
func TestSQLiteInMemoryDatabase(t *testing.T) {
	// Create a new in-memory SQLite persistence instance
	db, err := NewSQLitePersistenceWithOptions(
		WithInMemoryDatabase(),
		WithWALMode(true),
		WithSynchronousMode(0), // OFF for faster tests
	)
	if err != nil {
		t.Fatalf("Failed to create in-memory SQLite persistence: %v", err)
	}
	defer db.Close()

	// Create a test run
	runID, err := db.CreateRun()
	if err != nil {
		t.Fatalf("Failed to create run: %v", err)
	}

	// Verify the run was created
	run, err := db.GetRun(runID)
	if err != nil {
		t.Fatalf("Failed to get run: %v", err)
	}

	if run.ID != runID {
		t.Errorf("Expected run ID %d, got %d", runID, run.ID)
	}

	// Create a workflow entity to ensure schema works
	retryPolicy := `{"max_attempts": 3, "max_interval": 1000}`
	retryState := `{"attempts": 0, "timeout": 0}`
	workflowID, err := db.CreateWorkflowEntity(runID, "TestWorkflow", "test-step", retryPolicy, retryState)
	if err != nil {
		t.Fatalf("Failed to create workflow entity: %v", err)
	}

	// Verify the workflow was created
	entity, err := db.GetWorkflowEntity(workflowID)
	if err != nil {
		t.Fatalf("Failed to get workflow entity: %v", err)
	}

	if entity.ID != workflowID {
		t.Errorf("Expected workflow ID %d, got %d", workflowID, entity.ID)
	}

	if entity.HandlerName != "TestWorkflow" {
		t.Errorf("Expected handler name 'TestWorkflow', got '%s'", entity.HandlerName)
	}
}

// TestSQLiteOptions tests various SQLite configuration options
func TestSQLiteOptions(t *testing.T) {
	// Create a temporary directory for the test database
	tempDir, err := os.MkdirTemp("", "tempolite-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test different combinations of options
	testCases := []struct {
		name    string
		options []SQLiteOption
	}{
		{
			name: "Default with path",
			options: []SQLiteOption{
				WithDatabasePath(filepath.Join(tempDir, "default.db")),
			},
		},
		{
			name: "WAL mode explicit",
			options: []SQLiteOption{
				WithDatabasePath(filepath.Join(tempDir, "wal.db")),
				WithWALMode(true),
			},
		},
		{
			name: "Delete journal mode",
			options: []SQLiteOption{
				WithDatabasePath(filepath.Join(tempDir, "delete.db")),
				WithWALMode(false),
				WithJournalMode("DELETE"),
			},
		},
		{
			name: "Full sync mode",
			options: []SQLiteOption{
				WithDatabasePath(filepath.Join(tempDir, "fullsync.db")),
				WithSynchronousMode(2), // FULL
			},
		},
		{
			name: "Large cache",
			options: []SQLiteOption{
				WithDatabasePath(filepath.Join(tempDir, "largecache.db")),
				WithCacheSize(10000), // 10MB
			},
		},
		{
			name: "No foreign keys",
			options: []SQLiteOption{
				WithDatabasePath(filepath.Join(tempDir, "nofk.db")),
				WithForeignKeys(false),
			},
		},
		{
			name: "Short timeout",
			options: []SQLiteOption{
				WithDatabasePath(filepath.Join(tempDir, "shorttimeout.db")),
				WithBusyTimeout(1000), // 1 second
			},
		},
		{
			name: "Many connections",
			options: []SQLiteOption{
				WithDatabasePath(filepath.Join(tempDir, "manyconn.db")),
				WithMaxOpenConnections(20),
				WithMaxIdleConnections(10),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create persistence with the given options
			db, err := NewSQLitePersistenceWithOptions(tc.options...)
			if err != nil {
				t.Fatalf("Failed to create SQLite persistence with %s options: %v", tc.name, err)
			}

			// Verify basic functionality
			runID, err := db.CreateRun()
			if err != nil {
				t.Fatalf("Failed to create run with %s options: %v", tc.name, err)
			}

			// Get the run to verify it works
			_, err = db.GetRun(runID)
			if err != nil {
				t.Fatalf("Failed to get run with %s options: %v", tc.name, err)
			}

			// Clean up
			db.Close()
		})
	}
}
