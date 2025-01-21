# Workflow Engine Design Document

## 1. Overview

This document outlines the architecture and execution model of the workflow engine, specifically focusing on the handling of **WorkflowEntityIDs, RunIDs, WorkflowExecutionIDs**, and the incorporation of **Activities**, **SideEffects**, **ContinueAsNew**, and **Sub-Workflows across Queues**.

## 2. Workflow Execution Model

### 2.1 Key Identifiers

- **WorkflowEntityID**: Represents a persistent workflow instance. All executions and retries belong to the same WorkflowEntityID.
- **WorkflowExecutionID**: A unique identifier for each execution attempt within the same WorkflowEntityID.
- **RunID**: A unique identifier for each distinct workflow execution tree, including parent and sub-workflows.
- **ActivityEntityID**: Represents a unique persistent activity execution that can be retried independently of its workflow.
- **ActivityExecutionID**: A unique identifier for a single execution attempt of an activity within an activity entity.
- **SideEffectEntityID**: Represents a unique instance of a side effect operation tied to a workflow execution.
- **SideEffectExecutionID**: A unique identifier for a single execution attempt of a side effect.

### 2.2 Execution Flow

1. A **workflow starts** with a **WorkflowEntityID** and generates a new **RunID**.
2. If a workflow **fails and retries**, a new **WorkflowExecutionID** is assigned under the same **WorkflowEntityID**.
3. If the workflow **uses `continueAsNew`**, a new **WorkflowExecutionID** is generated, but it retains the **same WorkflowEntityID**. This prevents history growth while maintaining logical continuity.
4. If a **sub-workflow is started**, it generates a **new WorkflowEntityID and a new RunID**, making it an independent execution instance.
5. Activities executed within a workflow have a unique **ActivityEntityID**, which persists across retries. Each execution attempt of an activity gets a new **ActivityExecutionID**.
6. SideEffects have a **SideEffectEntityID**, which is tied to the workflow execution. A new **SideEffectExecutionID** is generated per execution attempt to ensure determinism.
7. If a **sub-workflow is scheduled on another queue**, it is assigned a **TaskQueue** that determines which worker processes it independently.

### 2.3 Example Structure

```
RunID "run-001"  (Parent Workflow Execution Tree)
└── WorkflowEntityID: "doing-something-001"
    ├── WorkflowExecutionID: "exec-001"  (First execution attempt)
    ├── Failed, retry triggered → WorkflowExecutionID: "exec-002"  
    ├── continueAsNew → WorkflowExecutionID: "exec-003"  (Fresh execution, reset history)
    │
    ├── Activity Execution
    │   ├── ActivityEntityID: "activity-001"
    │   ├── ActivityExecutionID: "activity-exec-001"
    │   └── Retry triggered → ActivityExecutionID: "activity-exec-002"
    │
    ├── SideEffect Execution
    │   ├── SideEffectEntityID: "sideeffect-001"
    │   ├── SideEffectExecutionID: "sideeffect-exec-001"
    │   └── Re-execution triggered → SideEffectExecutionID: "sideeffect-exec-002"
    │
    ├── Sub-workflow (Independent, on separate queue)
    │   ├── RunID: "run-002" (New execution tree)
    │   ├── WorkflowEntityID: "sub-workflow-001"
    │   ├── WorkflowExecutionID: "exec-004" (Sub-workflow execution)
    │   └── TaskQueue: "queue-002"
```

or 


```
RunID (Execution Tree)
├── WorkflowEntity (Parent)
│   ├── WorkflowExecution #1
│   ├── WorkflowExecution #2 (Retry)
│   ├── WorkflowExecution #3 (ContinueAsNew)
│   ├── ActivityEntity
│   │   ├── ActivityExecution #1
│   │   └── ActivityExecution #2 (Retry)
│   ├── SideEffectEntity
│   │   ├── SideEffectExecution
│   ├── SagaEntity
│   │   ├── SagaExecution (Transaction)
│   │   └── SagaExecution (Compensation)
│   ├── ChildWorkflow (Cross-Queue)
│   │   ├── WorkflowExecution #1
│   │   └── WorkflowExecution #2 (Retry)

```

## 4. Better options


```go
func ParentWorkflow(ctx tempolite.WorkflowContext) error {

    swOptions := tempolite.WorkflowOptions{
        TaskQueue: "queue-002",
    }

    ctx = tempolite.WithOptions(ctx, swOptions)
    
    var result string
    err := ctx.Workflow("some name", ChildWorkflow, "some input").Get(ctx, &result)
    if err != nil {
        return err
    }

    return nil
}

func ChildWorkflow(ctx tempolite.WorkflowContext, input string) (string, error) {
    return fmt.Sprintf("Processed: %s", input), nil
}
```

## 3. ContinueAsNew vs. StartChildWorkflow

### 3.1 `ContinueAsNew`
- **Retains the same WorkflowEntityID** but creates a new execution with a fresh WorkflowExecutionID.
- **Clears execution history**, preventing history growth while keeping logical continuity.
- **Resets retry policy**, meaning retries within the previous execution do not carry over.
- **Signals to old runs are lost** unless handled before `continueAsNew` is called.
- **Useful for long-running workflows** that need to periodically reset execution history.

#### Example Implementation
```go
func MyWorkflow(ctx tempolite.WorkflowContext, count int) error {
    if count >= 10 {
        return nil
    }

    ctx.Sleep(ctx, time.Hour) // Simulate long-running logic

    return ctx.NewContinueAsNewError(ctx, MyWorkflow, count+1)
}
```

### 3.2 `StartChildWorkflow`
- **Creates a completely independent workflow** with its own WorkflowEntityID and RunID.
- **Parent workflow can track child workflow progress** and get results when it completes.
- **Useful for workflows that need parallel execution or lifecycle separation.**

#### Example Implementation

```go
func ParentWorkflow(ctx tempolite.WorkflowContext) error {

    swOptions := tempolite.WorkflowOptions{
        TaskQueue: "queue-002",
    }

    ctx = workflow.WithOptions(ctx, swOptions)
    
    var result string
    err := ctx.Workflow(ctx, ChildWorkflow, "some input").Get(ctx, &result)
    if err != nil {
        return err
    }

    return nil
}

func ChildWorkflow(ctx tempolite.WorkflowContext, input string) (string, error) {
    return fmt.Sprintf("Processed: %s", input), nil
}
```

### 3.3 Key Differences

| Feature                | `ContinueAsNew`                                | `StartChildWorkflow`                |
|------------------------|-----------------------------------------------|-------------------------------------|
| WorkflowEntityID      | Same                                        | New                                 |
| RunID                 | New                                        | New                                 |
| Execution History     | Cleared                                    | Independent                         |
| Retry Policy          | Resets                                     | Separate                           |
| Parent-Child Relation | No                                         | Yes                                 |
| Use Case              | Long-running workflows with history cleanup | Parallel execution & lifecycle separation |

## 4. Managing Sub-Workflows Across Queues

### 4.1 Definition

Sub-workflows can be scheduled on **different worker queues**, allowing distributed execution and resource optimization.

### 4.2 How It Works

1. The parent workflow **defines a sub-workflow and assigns it to a specific queue**.
2. The **TaskQueue** determines where the sub-workflow runs.
3. Workers listening to the assigned queue will **pick up the sub-workflow execution**.
4. The sub-workflow execution **operates independently** but maintains a reference to its parent workflow.

## 5. Summary

This structure ensures a **scalable and maintainable workflow engine**, handling retries, `continueAsNew`, `StartChildWorkflow`, and sub-workflows across queues correctly.

