# Tempolite Technical Bible

## Core Concepts and Architecture

Before diving into technical details, it's essential to understand the core concepts of Tempolite at both a conceptual and architectural level.

### Workflows

**What is a Workflow?**
A workflow is a deterministic function that orchestrates a sequence of operations to accomplish a business process. It represents the reliable, repeatable logic that coordinates various activities, handling errors and managing state.

**Architectural Role:**
- **Control Flow Engine**: Workflows define the "what" and "when" of business logic execution
- **State Management Hub**: They maintain and track the progress of business processes
- **Coordination Layer**: They orchestrate distributed activities to work together as a cohesive process

**Key Properties:**
- **Deterministic**: Given the same inputs and state, workflows must produce the same outputs and effects
- **Resilient**: Workflows can be paused, resumed, and retried from the last safe checkpoint
- **Persistent**: Workflow state is recorded to ensure it can be recovered after failures
- **Long-running**: Workflows can span seconds to years, waiting for external signals when needed

**When to Use:**
Use workflows to model business processes that have one or more of these characteristics:
- Multiple steps that need coordination
- Require reliability across system failures
- Long-running processes (minutes to years)
- Need to integrate with multiple systems
- Require human intervention (approvals, reviews)

**Implementation Pattern:**
```go
func OrderProcessWorkflow(ctx tempolite.WorkflowContext, orderID string) error {
    // Workflows should read almost like a story of your business process
    if err := ctx.Activity("validate-order").Run(ValidateOrder, orderID).Get(&isValid); err != nil {
        return err
    }
    
    if !isValid {
        return errors.New("invalid order")
    }
    
    if err := ctx.Activity("process-payment").Run(ProcessPayment, orderID).Get(); err != nil {
        return err
    }
    
    if err := ctx.Activity("ship-order").Run(ShipOrder, orderID).Get(); err != nil {
        return err
    }
    
    return nil
}
```

### Activities

**What is an Activity?**
An activity is a non-deterministic unit of work that performs a specific task, typically involving interaction with external systems or resources. Activities are the "doers" in the workflow architecture.

**Architectural Role:**
- **Integration Layer**: Activities connect workflows to external systems
- **Failure Domain Isolation**: They contain non-deterministic operations outside the workflow's deterministic environment
- **Side-Effect Boundary**: All external effects (database writes, API calls) belong in activities
- **Retry Boundary**: Activities define a natural boundary for retry logic

**Key Properties:**
- **Non-deterministic**: May produce different results when called multiple times
- **Idempotence-aware**: Should be designed with potential retries in mind
- **Self-contained**: Each activity should perform a single logical operation
- **Independently retriable**: Activities can be retried without affecting other parts of the workflow

**When to Use:**
Use activities whenever your workflow needs to:
- Call external APIs
- Read/write from databases
- Process files or perform I/O
- Perform computationally intensive work
- Interact with external resources

**Implementation Pattern:**
```go
func ProcessPayment(ctx tempolite.ActivityContext, orderID string, amount float64) error {
    // Activities should do one specific thing and handle all possible failures
    logger := ctx.Logger()
    logger.Info("Processing payment", "order", orderID, "amount", amount)
    
    // Call payment gateway API
    paymentResult, err := paymentGateway.ProcessPayment(orderID, amount)
    if err != nil {
        if isRetryableError(err) {
            // Return error for automatic retry
            return err
        }
        // For non-retryable errors, you might return a domain-specific error
        return fmt.Errorf("payment failed permanently: %w", err)
    }
    
    // Record payment in database
    if err := database.SavePayment(orderID, paymentResult.TransactionID); err != nil {
        return err
    }
    
    return nil
}
```

### Side Effects

**What is a Side Effect?**
A side effect is a mechanism for managing non-deterministic values in a deterministic way within workflows. It records the result of non-deterministic operations during the first execution and replays that result in subsequent executions.

**Architectural Role:**
- **Deterministic Wrapper**: Makes non-deterministic operations deterministic within workflows
- **Replay Safety**: Ensures consistent results during workflow replays
- **State Preservation**: Records values for recovery during failures

**Key Properties:**
- **First-execution recording**: Records the result of the non-deterministic operation
- **Replay consistency**: Returns the recorded value during subsequent executions
- **Lightweight**: Typically used for small, simple operations

**When to Use:**
Use side effects for non-deterministic operations that:
- Generate random values (UUIDs, random numbers)
- Get current time or timestamps
- Need consistent values across workflow replays
- Are lightweight and don't require retries

**Implementation Pattern:**
```go
func GenerateOrderID(ctx tempolite.SideEffectContext) string {
    // This will execute once and record the result
    // On replay, the recorded value will be returned instead
    return uuid.New().String()
}

// In workflow
var orderID string
if err := ctx.SideEffect("generate-order-id")
    .Run(GenerateOrderID)
    .Get(&orderID); err != nil {
    return err
}
```

**Side Effect vs. Activity:**
- Use activities for operations that may fail and need retries
- Use side effects for operations that always succeed but produce non-deterministic results

### Signals

**What is a Signal?**
A signal is an asynchronous message sent to a running workflow from outside the workflow. It allows external systems to communicate with running workflows without requiring them to poll.

**Architectural Role:**
- **External Communication Channel**: Enables external systems to communicate with running workflows
- **Event Notification System**: Informs workflows of external events
- **Wait-Point Mechanism**: Allows workflows to pause execution until receiving external input

**Key Properties:**
- **Asynchronous**: Signals don't require immediate handling
- **Persistent**: Signals are durably stored until consumed
- **Targeted**: Signals are delivered to specific workflow instances
- **Typed**: Signals can carry typed data payloads

**When to Use:**
Use signals when:
- Workflows need to wait for external events
- Human intervention is required (approvals, reviews)
- External systems need to notify workflows of status changes
- Long-running processes need to be interruptible

**Implementation Patterns:**
```go
// In workflow - waiting for approval
var approval bool
if err := ctx.WaitForSignal("approval")
    .WithTimeout(24 * time.Hour)
    .Get(&approval); err != nil {
    if errors.Is(err, tempolite.ErrSignalTimeout) {
        return fmt.Errorf("approval timed out")
    }
    return err
}

// External system sending approval
signalFuture := tempolite.Signal("approval")
    .WithPayload(true)
    .ToWorkflow(wfID)
    .Send()
```

### Sagas

**What is a Saga?**
A saga is a pattern for managing distributed transactions with compensating actions. It allows a sequence of operations to be executed with the guarantee that either all succeed or compensating actions are run for completed operations.

**Architectural Role:**
- **Distributed Transaction Coordinator**: Manages transactional integrity across multiple operations
- **Compensation Orchestrator**: Handles cleanup for partial failures
- **Rollback Mechanism**: Provides "undo" functionality for distributed operations

**Key Properties:**
- **Step-based execution**: Executes a series of transaction steps in sequence
- **Automatic compensation**: Rolls back completed steps if a later step fails
- **Transactional boundary**: Creates a logical transaction across multiple activities
- **Recovery-aware**: Can resume compensation after failures

**When to Use:**
Use sagas when:
- Operations span multiple services or resources
- Traditional ACID transactions aren't possible
- You need a way to "undo" operations if something fails
- You need transactional semantics in distributed systems

**Implementation Pattern:**
```go
// Transaction function
func ReserveFunds(ctx tempolite.SagaTransactionContext, orderID string, amount float64) (string, error) {
    reservationID, err := paymentService.Reserve(orderID, amount)
    if err != nil {
        return "", err
    }
    return reservationID, nil
}

// Compensation function
func CancelReservation(ctx tempolite.SagaCompensationContext, orderID string) error {
    // Get the reservation ID from the transaction result
    var reservationID string
    if err := ctx.GetTransactionResult(&reservationID); err != nil {
        return err
    }
    
    return paymentService.CancelReservation(reservationID)
}

// In workflow
saga := ctx.Saga("payment")
saga.AddStep("reserve", ReserveFunds, CancelReservation, orderID, amount)
saga.AddStep("charge", ChargeCard, RefundCharge, paymentID)

if err := saga.Execute(); err != nil {
    // All completed steps were compensated
    return err
}
```

### Child Workflows

**What is a Child Workflow?**
A child workflow is a workflow that is started by another (parent) workflow. It enables modular composition of workflows and allows complex processes to be broken down into manageable, reusable components.

**Architectural Role:**
- **Modularity Mechanism**: Enables breaking down complex workflows
- **Reusability Facilitator**: Allows common workflows to be reused across the system
- **Domain Boundary**: Can separate different business domains
- **Resource Management**: Can isolate resource-intensive operations

**Key Properties:**
- **Parent-Child Relationship**: Child workflows are tracked in relation to their parent
- **Independent Execution**: Child workflows have their own execution context
- **Result Propagation**: Results can be returned to the parent workflow
- **Hierarchical Structure**: Creates a tree-like structure of workflow executions

**When to Use:**
Use child workflows when:
- Breaking down complex workflows into manageable pieces
- Reusing common workflow patterns
- Separating different business domains
- Requiring different retry or timeout policies for a subset of operations

**Implementation Pattern:**
```go
func ParentWorkflow(ctx tempolite.WorkflowContext, orderID string) error {
    // Start child workflow for payment processing
    var paymentResult PaymentResult
    if err := ctx.Workflow("process-payment")
        .WithRetries(RetryPolicy{MaxAttempts: 3})
        .Run(PaymentWorkflow, orderID)
        .Get(&paymentResult); err != nil {
        return err
    }
    
    // Start child workflow for shipping
    var trackingID string
    if err := ctx.Workflow("ship-order")
        .Run(ShippingWorkflow, orderID, paymentResult.TransactionID)
        .Get(&trackingID); err != nil {
        return err
    }
    
    return nil
}
```

### ContinueAsNew

**What is ContinueAsNew?**
ContinueAsNew is a pattern that allows a workflow to restart itself with new parameters while maintaining the same logical identity. It's used to prevent the unlimited growth of execution history for long-running workflows.

**Architectural Role:**
- **History Pruning Mechanism**: Prevents unbounded growth of workflow history
- **Logical Continuity Preserver**: Maintains workflow identity across executions
- **Long-Running Support**: Enables workflows to run indefinitely

**Key Properties:**
- **Same Identity**: Preserves the same workflow ID/entity
- **Fresh History**: Starts with a clean execution history
- **New Parameters**: Can take new input parameters
- **Logical Continuation**: Represents a logical continuation of the same workflow

**When to Use:**
Use ContinueAsNew when:
- Workflows need to run for a long time (days, months, years)
- Execution history would grow too large
- Workflows need to process data in chunks
- Implementing polling or periodic workflows

**Implementation Pattern:**
```go
func PeriodicCheckerWorkflow(ctx tempolite.WorkflowContext, resourceID string, iteration int) error {
    // Check the resource status
    var status string
    if err := ctx.Activity("check-status")
        .Run(CheckResourceStatus, resourceID)
        .Get(&status); err != nil {
        return err
    }
    
    if status == "completed" {
        return nil // Workflow completes
    }
    
    // Sleep before next check
    if err := ctx.Sleep(5 * time.Minute); err != nil {
        return err
    }
    
    // Continue as new with next iteration
    return ctx.ContinueAsNew(resourceID, iteration + 1)
}
```

### Parallel Group

**What is a Parallel Group?**
A parallel group is a mechanism for executing multiple operations concurrently and waiting for all to complete. It improves performance by running independent operations in parallel rather than sequentially.

**Architectural Role:**
- **Concurrency Orchestrator**: Manages execution of multiple concurrent operations
- **Performance Optimizer**: Reduces total execution time for independent operations
- **Error Aggregator**: Collects and reports errors from parallel operations

**Key Properties:**
- **Concurrent Execution**: Runs operations in parallel
- **Structured Waiting**: Provides organized way to wait for completion
- **Result Collection**: Allows retrieving results from all operations
- **Error Handling**: Aggregates errors from failed operations

**When to Use:**
Use parallel groups when:
- Multiple independent operations need to be performed
- Operations can safely execute concurrently
- Reducing total execution time is important
- You need to wait for multiple operations to complete before proceeding

**Implementation Pattern:**
```go
func ProcessOrderWorkflow(ctx tempolite.WorkflowContext, orderID string) error {
    group := ctx.ParallelGroup()
    
    // Add activities to run in parallel
    validateResult := group.Activity("validate").Run(ValidateOrder, orderID)
    paymentResult := group.Activity("payment").Run(ProcessPayment, orderID)
    inventoryResult := group.Activity("inventory").Run(ReserveInventory, orderID)
    
    // Wait for all to complete
    if err := group.Wait(); err != nil {
        return err
    }
    
    // Process results
    var isValid bool
    if err := validateResult.Get(&isValid); err != nil {
        return err
    }
    
    if !isValid {
        return errors.New("invalid order")
    }
    
    // Process other results...
    return nil
}
```

## Architectural Guidelines

### Component Relationships

1. **Workflows are Orchestrators**
   - Workflows should contain the "logic" but delegate the "work" to activities
   - Workflows should be deterministic, with all non-determinism isolated in activities and side effects
   - Workflows should focus on the "what" and "when", activities on the "how"

2. **Activities are Workers**
   - Activities should focus on a single logical operation
   - Activities should handle all potential failures and be idempotent where possible
   - Activities should contain all integration with external systems

3. **Side Effects are Deterministic Wrappers**
   - Side effects should be used sparingly for truly non-deterministic operations
   - Side effects should be lightweight and fast
   - Side effects should not contain critical business logic

4. **Signals are External Communication Channels**
   - Signals should be used for asynchronous communication with external systems
   - Workflows should use signals for long-waiting operations
   - Signals should carry minimal, focused data payloads

5. **Sagas are Transactional Boundaries**
   - Sagas should group related operations that need transactional semantics
   - Compensation functions should be carefully designed to properly undo transactions
   - Sagas should be used when traditional ACID transactions aren't possible

### Anti-Patterns to Avoid

1. **Bloated Workflows**: Don't put too much logic in a single workflow. Break complex processes into child workflows.

2. **Activities with Side Effects Only**: Activities should do meaningful work. Don't create activities just to wrap side effects.

3. **Non-Deterministic Workflows**: Keep workflows deterministic. Move all non-deterministic operations to activities or side effects.

4. **Heavy Side Effects**: Side effects should be lightweight. Use activities for complex or potentially failing operations.

5. **Over-Signaling**: Don't use too many fine-grained signals. Group related data into structured signal payloads.

6. **Direct Database Access in Workflows**: Workflows should not directly access databases or external systems. Use activities for all external interactions.

## Core Architecture

### Workflow Engine Overview

Tempolite is a deterministic workflow engine designed for local execution with persistence. It guarantees that business processes execute reliably despite failures by tracking state and enabling resumability.

> **Implementation Note:** All new code will be implemented under the `v2` folder to separate it from the existing codebase, ensuring a clean implementation while allowing for gradual migration.

#### Key Components Interaction

```
┌─────────────────┐      ┌─────────────────┐      ┌─────────────────┐
│                 │      │                 │      │                 │
│  Client Code    │──┬──▶│  Workflow API   │──┬──▶│   Persistence   │
│                 │  │   │                 │  │   │                 │
└─────────────────┘  │   └─────────────────┘  │   └─────────────────┘
                     │                        │
                     │   ┌─────────────────┐  │
                     │   │                 │  │
                     └──▶│  Worker System  │──┘
                         │                 │
                         └─────────────────┘
```

The engine consists of:
1. **Client API**: External interface for submitting/querying workflows
2. **Workflow API**: Internal context objects for workflow/activity execution
3. **Worker System**: Execution infrastructure for running workflows
4. **Persistence Layer**: SQLite-based state storage

### SQLite Schema Definition

The persistence layer uses SQLite with the following comprehensive set of tables:

#### Core Tables
```sql
-- Runs table tracks the overall execution tree
CREATE TABLE runs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    status TEXT NOT NULL CHECK (status IN ('Pending', 'Running', 'Paused', 'Cancelled', 'Completed', 'Failed')),
    created_at TEXT NOT NULL DEFAULT (DATETIME('now')),
    updated_at TEXT NOT NULL DEFAULT (DATETIME('now')),
    
    INDEX idx_runs_status (status)
);

-- Queue configuration table
CREATE TABLE queues (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    created_at TEXT NOT NULL DEFAULT (DATETIME('now')),
    updated_at TEXT NOT NULL DEFAULT (DATETIME('now'))
);

-- Hierarchy tracks parent-child relationships
CREATE TABLE hierarchies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id INTEGER NOT NULL,
    parent_id INTEGER, -- NULL for root workflows
    child_id INTEGER NOT NULL,
    child_type TEXT NOT NULL, -- workflow, activity, saga, etc.
    
    FOREIGN KEY (run_id) REFERENCES runs(id) ON DELETE CASCADE,
    FOREIGN KEY (parent_id) REFERENCES workflow_entities(id) ON DELETE CASCADE
);

-- Event log for debugging
CREATE TABLE event_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id INTEGER NOT NULL, 
    entity_id INTEGER,
    entity_type TEXT,
    event_type TEXT NOT NULL,
    details TEXT, -- JSON
    created_at TEXT NOT NULL DEFAULT (DATETIME('now')),
    
    FOREIGN KEY (run_id) REFERENCES runs(id) ON DELETE CASCADE
);
```

#### Workflow Tables
```sql
-- Workflow entity definition
CREATE TABLE workflow_entities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id INTEGER NOT NULL,
    handler_name TEXT NOT NULL,
    step_id TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL CHECK (status IN ('Pending', 'Queued', 'Running', 'Paused', 'Cancelled', 'Completed', 'Failed')),
    retry_policy TEXT NOT NULL, -- JSON: {"max_attempts": 3, "max_interval": 1000}
    retry_state TEXT NOT NULL, -- JSON: {"attempts": 0, "timeout": 0}
    created_at TEXT NOT NULL DEFAULT (DATETIME('now')),
    updated_at TEXT NOT NULL DEFAULT (DATETIME('now')),
    
    FOREIGN KEY (run_id) REFERENCES runs(id) ON DELETE CASCADE,
    INDEX idx_workflow_entities_run (run_id),
    INDEX idx_workflow_entities_status (status),
    INDEX idx_workflow_entities_step (step_id)
);

-- Workflow execution attempts
CREATE TABLE workflow_executions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    entity_id INTEGER NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('Pending', 'Queued', 'Running', 'Retried', 'Paused', 'Cancelled', 'Completed', 'Failed')),
    created_at TEXT NOT NULL DEFAULT (DATETIME('now')),
    updated_at TEXT NOT NULL DEFAULT (DATETIME('now')),
    
    FOREIGN KEY (entity_id) REFERENCES workflow_entities(id) ON DELETE CASCADE,
    INDEX idx_workflow_executions_entity (entity_id),
    INDEX idx_workflow_executions_status (status)
);

-- Workflow input data
CREATE TABLE workflow_data (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    entity_id INTEGER NOT NULL,
    input_data TEXT, -- JSON serialized inputs
    
    FOREIGN KEY (entity_id) REFERENCES workflow_entities(id) ON DELETE CASCADE
);

-- Workflow execution output data
CREATE TABLE workflow_execution_data (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    execution_id INTEGER NOT NULL,
    output_data TEXT, -- JSON serialized outputs
    error TEXT, -- serialized error if failed
    
    FOREIGN KEY (execution_id) REFERENCES workflow_executions(id) ON DELETE CASCADE
);
```

#### Activity Tables
```sql
-- Activity entity definition
CREATE TABLE activity_entities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    workflow_id INTEGER NOT NULL,
    step_id TEXT NOT NULL UNIQUE,
    handler_name TEXT NOT NULL,
    status TEXT NOT NULL,
    retry_policy TEXT NOT NULL, -- JSON
    retry_state TEXT NOT NULL, -- JSON
    created_at TEXT NOT NULL DEFAULT (DATETIME('now')),
    updated_at TEXT NOT NULL DEFAULT (DATETIME('now')),
    
    FOREIGN KEY (workflow_id) REFERENCES workflow_entities(id) ON DELETE CASCADE
);

-- Activity execution attempts
CREATE TABLE activity_executions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    entity_id INTEGER NOT NULL,
    status TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (DATETIME('now')),
    updated_at TEXT NOT NULL DEFAULT (DATETIME('now')),
    
    FOREIGN KEY (entity_id) REFERENCES activity_entities(id) ON DELETE CASCADE
);

-- Activity input data
CREATE TABLE activity_data (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    entity_id INTEGER NOT NULL,
    input_data TEXT, -- JSON serialized inputs
    
    FOREIGN KEY (entity_id) REFERENCES activity_entities(id) ON DELETE CASCADE
);

-- Activity execution output data
CREATE TABLE activity_execution_data (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    execution_id INTEGER NOT NULL,
    output_data TEXT, -- JSON serialized outputs
    error TEXT, -- serialized error if failed
    
    FOREIGN KEY (execution_id) REFERENCES activity_executions(id) ON DELETE CASCADE
);
```

#### Side Effect Tables
```sql
-- Side effect entity definition
CREATE TABLE side_effect_entities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    workflow_id INTEGER NOT NULL,
    step_id TEXT NOT NULL UNIQUE,
    handler_name TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (DATETIME('now')),
    updated_at TEXT NOT NULL DEFAULT (DATETIME('now')),
    
    FOREIGN KEY (workflow_id) REFERENCES workflow_entities(id) ON DELETE CASCADE
);

-- Side effect execution attempts
CREATE TABLE side_effect_executions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    entity_id INTEGER NOT NULL,
    status TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (DATETIME('now')),
    updated_at TEXT NOT NULL DEFAULT (DATETIME('now')),
    
    FOREIGN KEY (entity_id) REFERENCES side_effect_entities(id) ON DELETE CASCADE
);

-- Side effect input data
CREATE TABLE side_effect_data (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    entity_id INTEGER NOT NULL,
    input_data TEXT, -- JSON serialized inputs
    
    FOREIGN KEY (entity_id) REFERENCES side_effect_entities(id) ON DELETE CASCADE
);

-- Side effect execution output data
CREATE TABLE side_effect_execution_data (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    execution_id INTEGER NOT NULL,
    output_data TEXT, -- JSON serialized outputs
    error TEXT, -- serialized error if failed
    
    FOREIGN KEY (execution_id) REFERENCES side_effect_executions(id) ON DELETE CASCADE
);
```

#### Signal Tables
```sql
-- Signal entity definition
CREATE TABLE signal_entities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    workflow_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (DATETIME('now')),
    updated_at TEXT NOT NULL DEFAULT (DATETIME('now')),
    
    FOREIGN KEY (workflow_id) REFERENCES workflow_entities(id) ON DELETE CASCADE,
    UNIQUE (workflow_id, name)
);

-- Signal execution data
CREATE TABLE signal_executions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    entity_id INTEGER NOT NULL,
    status TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (DATETIME('now')),
    updated_at TEXT NOT NULL DEFAULT (DATETIME('now')),
    
    FOREIGN KEY (entity_id) REFERENCES signal_entities(id) ON DELETE CASCADE
);

-- Signal payload data
CREATE TABLE signal_data (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    entity_id INTEGER NOT NULL,
    payload TEXT, -- JSON serialized signal data
    
    FOREIGN KEY (entity_id) REFERENCES signal_entities(id) ON DELETE CASCADE
);
```

#### Saga Tables
```sql
-- Saga entity definition
CREATE TABLE saga_entities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    workflow_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (DATETIME('now')),
    updated_at TEXT NOT NULL DEFAULT (DATETIME('now')),
    
    FOREIGN KEY (workflow_id) REFERENCES workflow_entities(id) ON DELETE CASCADE
);

-- Saga execution attempts (transaction/compensation)
CREATE TABLE saga_executions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    entity_id INTEGER NOT NULL,
    type TEXT NOT NULL, -- "transaction" or "compensation"
    status TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (DATETIME('now')),
    updated_at TEXT NOT NULL DEFAULT (DATETIME('now')),
    
    FOREIGN KEY (entity_id) REFERENCES saga_entities(id) ON DELETE CASCADE
);

-- Saga step values for compensation
CREATE TABLE saga_values (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    saga_id INTEGER NOT NULL,
    step_name TEXT NOT NULL,
    key TEXT NOT NULL,
    value TEXT, -- JSON serialized value
    
    FOREIGN KEY (saga_id) REFERENCES saga_entities(id) ON DELETE CASCADE,
    UNIQUE (saga_id, step_name, key)
);
```

## Client API

### Initialization

The Tempolite client is initialized with a configuration that specifies:
- Database connection
- Worker pools
- Queue configurations
- Logging options

```go
// Initialize with SQLite
tempolite, err := tempolite.New(
    context.Background(),
    tempolite.WithSQLite("tempolite.db"),
    tempolite.WithWorkers(10),
    tempolite.WithQueue(QueueConfig{
        Name:    "high-priority",
        Workers: 5,
    }),
    tempolite.WithLogger(logger),
)
```

### Workflow Execution

Workflows are executed through the client API:

```go
// Define a workflow
func OrderProcessWorkflow(ctx tempolite.WorkflowContext, orderID string) error {
    // Validate order using builder pattern
    var isValid bool
    if err := ctx.Activity("validate")
        .WithRetries(RetryPolicy{MaxAttempts: 3})
        .WithTimeout(30 * time.Second)
        .Run(ValidateOrder, orderID)
        .Get(&isValid); err != nil {
        return err
    }
    
    if !isValid {
        return errors.New("invalid order")
    }
    
    // Process payment using builder pattern
    if err := ctx.Activity("payment")
        .WithRetries(RetryPolicy{MaxAttempts: 5})
        .Run(ProcessPayment, orderID)
        .Get(); err != nil {
        return err
    }
    
    return nil
}

// Execute the workflow using builder pattern
wfID := tempolite.NewWorkflow(OrderProcessWorkflow)
    .WithRetries(RetryPolicy{MaxAttempts: 3})
    .WithTimeout(1 * time.Hour)
    .WithQueue("high-priority")
    .Execute("order-123")

// Get the future for waiting on results
future := tempolite.GetWorkflowFuture(wfID)

// Wait for result synchronously
if err := future.Get(); err != nil {
    log.Fatal(err)
}

// Or wait asynchronously
go func() {
    <-future.OnResults()
    if err := future.Error(); err != nil {
        log.Printf("Error: %v", err)
    }
}()
```

### Workflow Control

The client API provides workflow control operations using the workflow ID:

```go
// Execute workflow and get its ID
wfID := tempolite.NewWorkflow(OrderProcessWorkflow)
    .WithRetries(RetryPolicy{MaxAttempts: 3})
    .Execute("order-123")

// Pause a workflow and get a future
pauseFuture := tempolite.PauseWorkflow(wfID)

// Simple synchronous wait
if err := pauseFuture.Wait(); err != nil {
    log.Fatal(err)
}

// Or use the channel directly for more control
select {
case <-pauseFuture.Done():
    if err := pauseFuture.Err(); err != nil {
        log.Fatal(err)
    }
    fmt.Println("Workflow paused successfully")
case <-time.After(5 * time.Second):
    fmt.Println("Pause operation timed out, but might still complete")
}

// Resume a workflow with channel-based waiting
resumeFuture := tempolite.ResumeWorkflow(wfID)
select {
case <-resumeFuture.Done():
    if err := resumeFuture.Err(); err != nil {
        log.Fatal(err)
    }
case <-ctx.Done():
    log.Println("Context cancelled, but resume may still complete")
}

// Multiple operations in parallel
cancelFuture1 := tempolite.CancelWorkflow(wfID1)
cancelFuture2 := tempolite.CancelWorkflow(wfID2)

// Wait for both to complete
for _, future := range []Future{cancelFuture1, cancelFuture2} {
    if err := future.Wait(); err != nil {
        log.Printf("Cancel failed: %v", err)
    }
}

// Query workflow status using the returned ID
status, err := tempolite.GetWorkflowStatus(wfID)
if err != nil {
    log.Fatal(err)
}
```

### Signal Delivery

The client API can send signals to workflows with enhanced control:

```go
// Execute workflow and get its ID
wfID := tempolite.NewWorkflow(OrderProcessWorkflow)
    .WithQueue("signals-queue")
    .Execute("order-123")

// Basic signal delivery with builder pattern
signalFuture := tempolite.Signal("approval-signal")
    .WithPayload(true)
    .WithCorrelationID("approval-request-123") // Optional correlation ID
    .ToWorkflow(wfID)
    .Send()

// Wait for the signal to be delivered
if err := signalFuture.Wait(); err != nil {
    log.Fatal(err)
}

// Or use channels for more control
select {
case <-signalFuture.Done():
    if err := signalFuture.Err(); err != nil {
        log.Fatal(err)
    }
    fmt.Println("Signal delivered successfully")
case <-time.After(5 * time.Second):
    fmt.Println("Signal delivery taking longer than expected")
}

// Send signals to multiple workflows at once
batchSignalFuture := tempolite.Signal("status-update")
    .WithPayload(StatusUpdate{Status: "approved"})
    .ToWorkflows([]WorkflowID{wfID1, wfID2, wfID3})
    .Send()

// Check delivery status per workflow
results := batchSignalFuture.Results()
for wfID, err := range results {
    if err != nil {
        log.Printf("Failed to deliver signal to workflow %v: %v", wfID, err)
    }
}

// Signal with typed payload (compile-time type safety)
type ApprovalSignal struct {
    Approved bool
    Comment  string
    ApprovedBy string
}

tempolite.SignalTyped(wfID, "approval", ApprovalSignal{
    Approved: true,
    Comment: "Looks good",
    ApprovedBy: "Manager",
})
```

The workflow API for consuming signals is also enhanced:

```go
// Handling signals in a workflow

// Wait for a single signal with timeout
var approval bool
if err := ctx.WaitForSignal("approval")
    .WithTimeout(24 * time.Hour)
    .Get(&approval); err != nil {
    if errors.Is(err, tempolite.ErrSignalTimeout) {
        return fmt.Errorf("approval timed out")
    }
    return err
}

// Wait for a typed signal
var approvalData ApprovalSignal
if err := ctx.WaitForSignalTyped("approval", &approvalData); err != nil {
    return err
}

// Process based on approval data
if approvalData.Approved {
    // Process approval
    ctx.Logger().Info("Approved by", "user", approvalData.ApprovedBy)
} else {
    // Handle rejection
    return fmt.Errorf("order rejected: %s", approvalData.Comment)
}

// Wait for any of multiple signals
signalChan := ctx.WaitForAnySignal([]string{"approval", "cancellation", "update"})
select {
case signal := <-signalChan:
    switch signal.Name() {
    case "approval":
        var approved bool
        if err := signal.Get(&approved); err != nil {
            return err
        }
        // Handle approval
    case "cancellation":
        // Handle cancellation
        return errors.New("workflow cancelled")
    case "update":
        var update StatusUpdate
        if err := signal.Get(&update); err != nil {
            return err
        }
        // Handle update
    }
case <-ctx.Deadline():
    return errors.New("workflow timed out")
}

// Wait for all required signals
requiredSignals := ctx.WaitForAllSignals([]string{"approval", "payment"})
if err := requiredSignals.Wait(24 * time.Hour); err != nil {
    return err
}

var approval bool
var payment PaymentInfo
if err := requiredSignals.Get("approval", &approval); err != nil {
    return err
}
if err := requiredSignals.Get("payment", &payment); err != nil {
    return err
}
```

## Workflow API

### Workflow Context

The `WorkflowContext` is the primary interface for workflow functions:

```go
type WorkflowContext interface {
    // Context operations
    WithTimeout(timeout time.Duration) WorkflowContext
    WithRetryPolicy(policy RetryPolicy) WorkflowContext
    WithOptions(opts WorkflowOptions) WorkflowContext
    
    // Core operations
    Activity(name string) ActivityBuilder
    Workflow(name string) WorkflowBuilder
    SideEffect(name string) SideEffectBuilder
    WaitForSignal(name string) SignalBuilder
    Saga(name string) SagaBuilder
    
    // Workflow control
    Sleep(duration time.Duration) error
    ContinueAsNew(args ...interface{}) error
    
    // Logging
    Logger() Logger
    
    // Versioning
    Version() int
    WithVersion(version int, fn func(WorkflowContext) error) VersionExecutor
    
    // Utilities
    ParallelGroup() ParallelGroup
    RegisterQuery(name string, fn interface{}) error
}
```

### Builder Pattern Implementation

The builder pattern is implemented for all operations:

#### ActivityBuilder
```go
type ActivityBuilder interface {
    // Configuration
    WithRetries(policy RetryPolicy) ActivityBuilder
    WithTimeout(timeout time.Duration) ActivityBuilder
    WithHeartbeat(interval time.Duration) ActivityBuilder
    
    // Execution
    Run(fn interface{}, args ...interface{}) ActivityFuture
    RunIf(condition bool, fn interface{}, args ...interface{}) ActivityFuture
    RunWhen(predicate func() bool, fn interface{}, args ...interface{}) ActivityFuture
}

type ActivityFuture interface {
    // Result handling
    Get(results ...interface{}) error
    OnResults() <-chan struct{}
    Cancel() error
    IsDone() bool
    
    // Status information
    Status() ActivityStatus
    Error() error
}
```

Similar interfaces exist for `WorkflowBuilder`, `SideEffectBuilder`, `SignalBuilder`, and `SagaBuilder`.

### Parallel Execution

The parallel execution pattern is implemented through:

```go
type ParallelGroup interface {
    // Add operations to group
    Activity(name string) ActivityBuilder
    Workflow(name string) WorkflowBuilder
    
    // Execution control
    Wait() error
    WaitWithTimeout(timeout time.Duration) error
    Cancel()
}
```

Example usage:
```go
group := ctx.ParallelGroup()

// Add activities to run in parallel
validateResult := group.Activity("validate").Run(ValidateOrder, orderID)
paymentResult := group.Activity("payment").Run(ProcessPayment, orderID)
shippingResult := group.Activity("shipping").Run(ArrangeShipping, orderID)

// Add child workflows to run in parallel
inventoryResult := group.Workflow("check-inventory").Run(CheckInventoryWorkflow, orderID)

// Wait for all to complete
if err := group.Wait(); err != nil {
    return err
}

// Process results
var validationResult bool
if err := validateResult.Get(&validationResult); err != nil {
    return err
}

// Process other results...
```

### Signal Handling

Signals can be waited for individually or as a group:

```go
// Wait for a single signal
if err := ctx.WaitForSignal("approval")
    .WithTimeout(24 * time.Hour)
    .Get(&approvalData); err != nil {
    return err
}

// Wait for multiple signals
signals := ctx.WaitForSignals([]string{"approval", "payment", "shipping"})

// Get the first signal that arrives
signal, err := signals.GetFirst(5 * time.Minute)
if err != nil {
    return err
}

// Handle based on which signal arrived
switch signal.Name() {
case "approval":
    var approval bool
    signal.Get(&approval)
    // Handle approval
}
```

### Saga Implementation

The Saga pattern is implemented with:

```go
type SagaBuilder interface {
    // Configuration
    WithOptions(opts SagaOptions) SagaBuilder
    OnCompensation(handler func(error)) SagaBuilder
    
    // Step definition
    AddStep(name string, txFn interface{}, compFn interface{}, args ...interface{}) SagaBuilder
    Transaction(fn interface{}) SagaTransactionBuilder
    Compensation(fn interface{}) SagaCompensationBuilder
    
    // Execution
    Execute() SagaResult
}

type SagaResult interface {
    // Status checks
    IsCompleted() bool
    IsCompensated() bool
    Error() error
    
    // Result access
    StepResults() map[string]interface{}
}
```

Example usage:
```go
saga := ctx.Saga("payment")
    .WithOptions(SagaOptions{Parallel: false, StrictMode: true})
    .OnCompensation(func(err error) {
        ctx.Logger().Error("Saga failed", "error", err)
    })

// Add steps with transaction and compensation
saga.AddStep("reserve", 
    saga.Transaction(ReserveFunds).WithArgs(orderID, amount), 
    saga.Compensation(ReleaseReservation).WithArgs(orderID))

saga.AddStep("charge", 
    saga.Transaction(ChargeCard).WithArgs(paymentID), 
    saga.Compensation(RefundCharge).WithArgs(paymentID))

// Execute
result := saga.Execute()
if result.IsCompensated() {
    return fmt.Errorf("payment failed and was rolled back: %v", result.Error())
}
```

### Versioning

The versioning API enables handling workflow evolution:

```go
// Check version directly
if ctx.Version() < 2 {
    // Old logic
} else {
    // New logic
}

// Or use version-specific handlers
ctx.WithVersion(1, func(ctx tempolite.WorkflowContext) error {
    // Version 1 implementation
    return nil
}).WithVersion(2, func(ctx tempolite.WorkflowContext) error {
    // Version 2 implementation
    return nil
}).Execute()
```

## Internal Implementation Details

### Worker System

The worker system uses a pool of goroutines to execute workflow tasks:

```go
type WorkerPool struct {
    mu         sync.RWMutex
    workers    int
    tasks      chan Task
    wg         sync.WaitGroup
    isShutdown bool
}

func NewWorkerPool(workers int) *WorkerPool {
    pool := &WorkerPool{
        workers: workers,
        tasks:   make(chan Task, workers*10), // Buffer for tasks
    }
    pool.Start()
    return pool
}

func (p *WorkerPool) Start() {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    for i := 0; i < p.workers; i++ {
        p.wg.Add(1)
        go func() {
            defer p.wg.Done()
            for task := range p.tasks {
                task.Execute()
            }
        }()
    }
}

func (p *WorkerPool) Submit(task Task) error {
    p.mu.RLock()
    defer p.mu.RUnlock()
    
    if p.isShutdown {
        return errors.New("worker pool is shutdown")
    }
    
    p.tasks <- task
    return nil
}

func (p *WorkerPool) Shutdown(gracePeriod time.Duration) {
    p.mu.Lock()
    p.isShutdown = true
    close(p.tasks)
    p.mu.Unlock()
    
    // Wait with timeout
    c := make(chan struct{})
    go func() {
        p.wg.Wait()
        close(c)
    }()
    
    select {
    case <-c:
        // Normal shutdown
    case <-time.After(gracePeriod):
        // Timeout
    }
}
```

### State Machine

Workflow execution follows a state machine pattern:

```
┌─────────┐
│         │
│ Pending ├───────┐
│         │       │
└────┬────┘       │
     │            │
     ▼            ▼
┌─────────┐  ┌─────────┐     ┌─────────┐
│         │  │         │     │         │
│ Queued  ├─▶│ Running ├────▶│ Paused  │
│         │  │         │     │         │
└─────────┘  └────┬────┘     └────┬────┘
                  │               │
                  │               │
                  ▼               │
┌─────────┐  ┌─────────┐         │
│         │  │         │         │
│ Failed  │◀─┤Completed│◀────────┘
│         │  │         │
└─────────┘  └─────────┘
```

Each entity (workflow, activity, etc.) transitions through these states:
- **Pending**: Entity created but not yet queued
- **Queued**: Entity waiting for execution
- **Running**: Entity executing
- **Paused**: Execution paused (can resume)
- **Completed**: Execution successful
- **Failed**: Execution failed (may retry)

### Database Layer Implementation

The SQLite database layer handles all CRUD operations:

```go
type SQLiteDatabase struct {
    db *sql.DB
    mu sync.RWMutex
}

// Example implementation of AddWorkflowEntity
func (db *SQLiteDatabase) AddWorkflowEntity(runID RunID, entity *WorkflowEntity) (WorkflowEntityID, error) {
    db.mu.Lock()
    defer db.mu.Unlock()
    
    tx, err := db.db.Begin()
    if err != nil {
        return 0, err
    }
    defer tx.Rollback()
    
    // Serialize retry policy and state
    retryPolicyJSON, err := json.Marshal(entity.RetryPolicy)
    if err != nil {
        return 0, err
    }
    
    retryStateJSON, err := json.Marshal(entity.RetryState)
    if err != nil {
        return 0, err
    }
    
    // Insert the entity
    result, err := tx.Exec(
        `INSERT INTO workflow_entities 
         (run_id, handler_name, step_id, status, retry_policy, retry_state, created_at, updated_at) 
         VALUES (?, ?, ?, ?, ?, ?, DATETIME('now'), DATETIME('now'))`,
        runID, entity.HandlerName, entity.StepID, entity.Status, 
        retryPolicyJSON, retryStateJSON,
    )
    if err != nil {
        return 0, err
    }
    
    id, err := result.LastInsertId()
    if err != nil {
        return 0, err
    }
    
    if err := tx.Commit(); err != nil {
        return 0, err
    }
    
    return WorkflowEntityID(id), nil
}
```

### Serialization

All data is serialized to JSON for storage:

```go
func serializeInputs(args []interface{}) ([]byte, error) {
    // Serialize each argument
    serialized := make([]interface{}, len(args))
    for i, arg := range args {
        // Get the type name
        t := reflect.TypeOf(arg)
        typeName := ""
        if t != nil {
            typeName = t.String()
        }
        
        // Convert to JSON
        data, err := json.Marshal(arg)
        if err != nil {
            return nil, fmt.Errorf("failed to serialize argument %d: %w", i, err)
        }
        
        // Store with type information
        serialized[i] = map[string]interface{}{
            "type": typeName,
            "data": string(data),
        }
    }
    
    // Serialize the entire array
    return json.Marshal(serialized)
}

func deserializeOutputs(data []byte, results []interface{}) error {
    // Parse the JSON
    var serialized []map[string]interface{}
    if err := json.Unmarshal(data, &serialized); err != nil {
        return err
    }
    
    // Check length match
    if len(serialized) != len(results) {
        return fmt.Errorf("result count mismatch: expected %d, got %d", 
                         len(results), len(serialized))
    }
    
    // Deserialize each result
    for i, result := range results {
        resultValue := reflect.ValueOf(result)
        if resultValue.Kind() != reflect.Ptr || resultValue.IsNil() {
            return fmt.Errorf("result %d must be a non-nil pointer", i)
        }
        
        // Get the JSON data
        jsonData := serialized[i]["data"].(string)
        
        // Unmarshal into the pointer
        if err := json.Unmarshal([]byte(jsonData), result); err != nil {
            return fmt.Errorf("failed to deserialize result %d: %w", i, err)
        }
    }
    
    return nil
}
```

## Activity Implementation

Activities execute outside the workflow's deterministic environment:

```go
// ActivityContext provides the API for activities
type ActivityContext interface {
    // Context operations
    WithTimeout(timeout time.Duration) ActivityContext
    WithHeartbeat(interval time.Duration) ActivityContext
    
    // Heartbeat operations
    RecordHeartbeat(details ...interface{}) error
    
    // Information access
    Info() ActivityInfo
    Logger() Logger
}

type ActivityInfo struct {
    ID          ActivityEntityID
    ExecutionID ActivityExecutionID
    Attempt     int
    WorkflowID  WorkflowEntityID
    Deadline    *time.Time
}
```

Example activity implementation:
```go
func ProcessPayment(ctx tempolite.ActivityContext, orderID string, amount float64) error {
    logger := ctx.Logger()
    logger.Info("Processing payment", "order", orderID, "amount", amount)
    
    // Long-running operation with heartbeats
    for i := 0; i < 10; i++ {
        // Do work
        if err := processStep(i, orderID, amount); err != nil {
            return err
        }
        
        // Record heartbeat
        if err := ctx.RecordHeartbeat(i, "Processing"); err != nil {
            return err
        }
        
        time.Sleep(time.Second)
    }
    
    return nil
}
```

## Side Effect Implementation

Side effects handle non-deterministic operations:

```go
// SideEffectContext provides API for side effects
type SideEffectContext interface {
    // Information access
    Info() SideEffectInfo
    Logger() Logger
}

type SideEffectInfo struct {
    ID          SideEffectEntityID
    ExecutionID SideEffectExecutionID
    WorkflowID  WorkflowEntityID
}
```

Example side effect implementation:
```go
func GenerateOrderID(ctx tempolite.SideEffectContext) string {
    // This non-deterministic operation will be recorded
    // and replayed with the same result during workflow replay
    return uuid.New().String()
}

// In workflow
var orderID string
if err := ctx.SideEffect("generate-order-id")
    .Run(GenerateOrderID)
    .Get(&orderID); err != nil {
    return err
}
```

## Signal Implementation

Signals enable external communication with running workflows:

```go
// SideEffectContext provides API for side effects
type SideEffectContext interface {
    // Information access
    Info() SideEffectInfo
    Logger() Logger
}

type SideEffectInfo struct {
    ID          SideEffectEntityID
    ExecutionID SideEffectExecutionID
    WorkflowID  WorkflowEntityID
}
```

Example signal handling in a workflow:
```go
func OrderApprovalWorkflow(ctx tempolite.WorkflowContext, orderID string) error {
    // Process order
    if err := ctx.Activity("process")
        .Run(ProcessOrder, orderID)
        .Get(); err != nil {
        return err
    }
    
    // Wait for approval signal with rich handling
    ctx.Logger().Info("Waiting for approval signal")
    
    // Option 1: Simple signal waiting with timeout
    var approval bool
    if err := ctx.WaitForSignal("approval")
        .WithTimeout(24 * time.Hour)
        .Get(&approval); err != nil {
        if errors.Is(err, tempolite.ErrSignalTimeout) {
            return ctx.Activity("timeout-handler")
                .Run(HandleApprovalTimeout, orderID)
                .Get()
        }
        return err
    }
    
    if !approval {
        return ctx.Activity("rejection-handler")
            .Run(HandleRejection, orderID)
            .Get()
    }
    
    // Option 2: Wait for any of multiple possible signals
    signalChan := ctx.WaitForAnySignal([]string{"expedite", "cancel", "discount"})
    
    // Set up a timer for reminder
    timerChan := ctx.NewTimer(12 * time.Hour).Chan()
    
    select {
    case signal := <-signalChan:
        switch signal.Name() {
        case "expedite":
            var expediteLevel int
            signal.Get(&expediteLevel)
            return ctx.Activity("expedite")
                .Run(ExpediteOrder, orderID, expediteLevel)
                .Get()
        case "cancel":
            return ctx.Activity("cancel")
                .Run(CancelOrder, orderID)
                .Get()
        case "discount":
            var discountPercent float64
            signal.Get(&discountPercent)
            return ctx.Activity("apply-discount")
                .Run(ApplyDiscount, orderID, discountPercent)
                .Get()
        }
    case <-timerChan:
        // Send reminder but keep waiting
        _ = ctx.Activity("send-reminder")
            .Run(SendApprovalReminder, orderID)
            .Get()
        
        // Continue waiting for signals
        signal := <-signalChan
        // Handle signal...
    }
    
    return nil
}
```

External signal sending with the enhanced API:
```go
// Get workflow ID
wfID := tempolite.NewWorkflow(OrderApprovalWorkflow)
    .Execute("order-123")

// Send simple approval
signalFuture := tempolite.Signal("approval")
    .WithPayload(true)
    .ToWorkflow(wfID)
    .Send()

// Wait for delivery confirmation
if err := signalFuture.Wait(); err != nil {
    log.Printf("Failed to deliver signal: %v", err)
}

// Send structured data
type DiscountSignal struct {
    Percent     float64
    Reason      string
    ApprovedBy  string
    ExpiresAt   time.Time
}

tempolite.SignalTyped(wfID, "discount", DiscountSignal{
    Percent:    15.0,
    Reason:     "Loyalty customer",
    ApprovedBy: "Sales Manager",
    ExpiresAt:  time.Now().Add(48 * time.Hour),
})
```

## Saga Pattern Implementation

Sagas coordinate transactions with compensation:

```go
// Transaction function
func ReserveFunds(ctx tempolite.SagaTransactionContext, orderID string, amount float64) (string, error) {
    reservationID, err := paymentService.Reserve(orderID, amount)
    if err != nil {
        return "", err
    }
    return reservationID, nil
}

// Compensation function
func CancelReservation(ctx tempolite.SagaCompensationContext, orderID string) error {
    // Get the reservation ID from the transaction result
    var reservationID string
    if err := ctx.GetTransactionResult(&reservationID); err != nil {
        return err
    }
    
    return paymentService.CancelReservation(reservationID)
}

// In workflow
saga := ctx.Saga("payment")

// Add steps with transaction and compensation
saga.AddStep("reserve", ReserveFunds, CancelReservation, orderID, amount)
saga.AddStep("charge", ChargeCard, RefundCharge, paymentID)

// Execute
if err := saga.Execute(); err != nil {
    // All completed steps were compensated
    return err
}
```

## ContinueAsNew Implementation

ContinueAsNew restarts workflows with new parameters:

```go
func LongRunningWorkflow(ctx tempolite.WorkflowContext, counter int) error {
    // Check termination condition
    if counter >= 100 {
        return nil
    }
    
    // Do some work
    if err := ctx.Activity("process-batch")
        .Run(ProcessBatch, counter)
        .Get(); err != nil {
        return err
    }
    
    // Continue as new with incremented counter
    return ctx.ContinueAsNew(counter + 1)
}
```

Internally, ContinueAsNew:
1. Completes the current workflow execution
2. Creates a new workflow execution with the same entity ID
3. Preserves signals and workflow identity
4. Starts the workflow from the beginning with new arguments

## Child Workflow Implementation

Child workflows enable modular composition:

```go
func ParentWorkflow(ctx tempolite.WorkflowContext, orderID string) error {
    // Start child workflow
    var result string
    if err := ctx.Workflow("process-payment")
        .WithRetries(RetryPolicy{MaxAttempts: 3})
        .Run(PaymentWorkflow, orderID)
        .Get(&result); err != nil {
        return err
    }
    
    logger.Info("Payment processed", "result", result)
    return nil
}

func PaymentWorkflow(ctx tempolite.WorkflowContext, orderID string) (string, error) {
    // Child workflow implementation
    return "Payment processed successfully", nil
}
```

Internally, child workflows:
1. Create a new workflow entity with its own ID
2. Establish parent-child relationship in the database
3. Execute independently but track hierarchy
4. Return results to parent workflow

## Advanced Patterns

### Workflow Versioning

Handle workflow evolution with version checks:

```go
func MyWorkflow(ctx tempolite.WorkflowContext, input string) error {
    // Handle different versions
    return ctx.WithVersion(1, func(ctx tempolite.WorkflowContext) error {
        // Original implementation
        return ctx.Activity("v1-activity").Run(ActivityV1, input).Get()
    }).WithVersion(2, func(ctx tempolite.WorkflowContext) error {
        // New implementation with additional steps
        if err := ctx.Activity("v2-activity").Run(ActivityV2, input).Get(); err != nil {
            return err
        }
        return ctx.Activity("v2-extra").Run(ExtraActivity, input).Get()
    }).Execute()
}
```

### Parallel Activity Execution

Execute activities in parallel:

```go
func OrderProcessWorkflow(ctx tempolite.WorkflowContext, orderID string) error {
    // Create parallel group
    group := ctx.ParallelGroup()
    
    // Add activities to run in parallel
    validateFuture := group.Activity("validate").Run(ValidateOrder, orderID)
    paymentFuture := group.Activity("payment").Run(ProcessPayment, orderID)
    shippingFuture := group.Activity("shipping").Run(ArrangeShipping, orderID)
    
    // Wait for all to complete
    if err := group.Wait(); err != nil {
        return err
    }
    
    // Process results
    var validationResult bool
    if err := validateFuture.Get(&validationResult); err != nil {
        return err
    }
    
    if !validationResult {
        return errors.New("order validation failed")
    }
    
    // Process other results
    // ...
    
    return nil
}
```

### Conditional Workflow Execution

Execute workflow logic conditionally:

```go
func ProcessOrderWorkflow(ctx tempolite.WorkflowContext, order Order) error {
    // Validate first
    var isValid bool
    if err := ctx.Activity("validate").Run(ValidateOrder, order).Get(&isValid); err != nil {
        return err
    }
    
    if !isValid {
        return errors.New("invalid order")
    }
    
    // Process based on order type
    if order.Value > 1000 {
        // High-value order process
        return ctx.Workflow("high-value-process")
            .Run(HighValueProcess, order)
            .Get()
    } else {
        // Standard order process
        return ctx.Workflow("standard-process")
            .Run(StandardProcess, order)
            .Get()
    }
}
```

### Query Handlers

Implement query handlers for workflow state inspection:

```go
func OrderProcessWorkflow(ctx tempolite.WorkflowContext, orderID string) error {
    // Workflow state
    currentStatus := "initializing"
    
    // Register query handlers
    ctx.RegisterQuery("status", func() string {
        return currentStatus
    })
    
    ctx.RegisterQuery("details", func() map[string]interface{} {
        return map[string]interface{}{
            "order_id": orderID,
            "status":   currentStatus,
            "started":  ctx.Info().StartTime,
        }
    })
    
    // Update status as workflow progresses
    currentStatus = "validating"
    // ...
    
    return nil
}

// External query using workflow ID
wfID := tempolite.NewWorkflow(OrderProcessWorkflow)
    .Execute("order-123")

status, err := tempolite.QueryWorkflow(wfID, "status")
if err != nil {
    log.Fatal(err)
}
fmt.Println("Workflow status:", status)
```

This comprehensive Tempolite Technical Bible provides the detailed documentation needed for implementing all aspects of the workflow engine, from the core architecture to advanced patterns and specific component behavior. Each section includes concrete implementation details, API patterns, and examples to ensure a complete understanding of how the system works.
