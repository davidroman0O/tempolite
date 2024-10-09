# Temporal-Inspired Design Document

## Entities and Their Relationships

### 1. ExecutionContext

The `ExecutionContext` represents the overall workflow execution. It's the top-level entity that groups together all related activities for a single workflow run.

**Fields:**
- `ID`: Unique identifier for the execution context
- `CurrentRunID`: Identifier for the current run of the workflow
- `Status`: Overall status of the workflow (Running, Completed, Failed)
- `StartTime`: When the workflow started
- `EndTime`: When the workflow ended (if completed or failed)

**Relationships:**
- Has many `HandlerExecution`s

### 2. HandlerExecution

The `HandlerExecution` represents a single execution of a handler (activity) within the workflow. It can be thought of as a node in the workflow graph.

**Fields:**
- `ID`: Unique identifier for the handler execution
- `RunID`: Identifier for the current run (matches `ExecutionContext.CurrentRunID`)
- `HandlerName`: Name of the handler being executed
- `Status`: Status of this specific handler execution (Pending, Running, Completed, Failed)
- `StartTime`: When this handler execution started
- `EndTime`: When this handler execution ended (if completed or failed)
- `RetryCount`: Number of times this handler has been retried
- `MaxRetries`: Maximum number of retries allowed for this handler

**Relationships:**
- Belongs to one `ExecutionContext`
- Can have one parent `HandlerExecution` (for nested handlers)
- Can have many child `HandlerExecution`s (for nested handlers)
- Has many `HandlerTask`s (for retries)

### 3. HandlerTask

The `HandlerTask` represents a specific attempt to execute a handler. Each retry creates a new `HandlerTask`.

**Fields:**
- `ID`: Unique identifier for the handler task
- `HandlerName`: Name of the handler to be executed
- `Payload`: Input data for the handler
- `Result`: Output data from the handler (if successful)
- `Error`: Error information (if failed)
- `Status`: Status of this specific task (Pending, In Progress, Completed, Failed)
- `CreatedAt`: When this task was created
- `CompletedAt`: When this task was completed or failed

**Relationships:**
- Belongs to one `HandlerExecution`

## Workflow Execution Process

1. When a workflow is started, an `ExecutionContext` is created.
2. For each handler in the workflow, a `HandlerExecution` is created and associated with the `ExecutionContext`.
3. For each `HandlerExecution`, an initial `HandlerTask` is created.
4. The `HandlerTask` is executed by a worker.
5. If the `HandlerTask` fails and retries are available, a new `HandlerTask` is created for the same `HandlerExecution` with an incremented `RetryCount`.
6. This process continues until the handler succeeds or max retries are reached.
7. Once all `HandlerExecution`s are completed or failed, the `ExecutionContext` is updated accordingly.

## Key Concepts

- **Workflow**: Represented by the `ExecutionContext`, it's the overall process being executed.
- **Activity**: Represented by the `HandlerExecution`, it's a single step in the workflow.
- **Task**: Represented by the `HandlerTask`, it's a single attempt to execute an activity.
- **Retry**: Creating a new `HandlerTask` for an existing `HandlerExecution` when a previous attempt fails.

This design allows for:
- Tracking the overall workflow state (`ExecutionContext`)
- Managing individual activities and their states (`HandlerExecution`)
- Handling retries and maintaining execution history (`HandlerTask`)
- Supporting nested workflows (parent-child relationships between `HandlerExecution`s)