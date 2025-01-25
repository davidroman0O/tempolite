
This document represent all the final design of my workflow engine.



Having an ID as a stepID for my system fits, since my pocket workflow engine will run on a single machine, I don’t need strict deterministic replay like Temporal. I should focus on state persistence and resumability.

- Step Execution Tracking via SQLite

Store each stepID and its execution status/result in SQLite.
When a workflow runs, check SQLite first to see if a step needs to execute.

- Local Execution Per Worker

Even if you later introduce distributed job queues, each worker will execute entire workflows locally.
No need for event-driven replay across multiple nodes.

- Resumable & Modifiable Workflows

If a workflow fails, you retry from the failed step, not from the beginning.
Developers can reorder steps without breaking existing workflow state.

- No Event Replay or Temporal-Like History

No deterministic replay needed—SQLite stores the last known good state.
No need to replay past events; just fetch completed steps from SQLite.


```
RunID (Execution Tree)
├── WorkflowEntity (Parent)
│   ├── WorkflowData
│   ├── WorkflowExecution #1
│   │   └── WorkflowExecutionData
│   ├── WorkflowExecution #2 (Retry)
│   │   └── WorkflowExecutionData
│   │
│   ├── ActivityEntity
│   │   ├── ActivityData
│   │   ├── ActivityExecution #1
│   │   │   └── ActivityExecutionData
│   │   └── ActivityExecution #2 (Retry)
│   │       └── ActivityExecutionData
│   │
│   ├── SideEffectEntity
│   │   ├── SideEffectData
│   │   └── SideEffectExecution
│   │       └── SideEffectExecutionData
│   │
│   ├── SagaEntity
│   │   ├── SagaData
│   │   ├── SagaExecution (Transaction)
│   │   │   └── SagaExecutionData
│   │   └── SagaExecution (Compensation)
│   │       └── SagaExecutionData
│   │
│   ├── ChildWorkflow (Cross-Queue)
│   │   ├── WorkflowExecution #1
│   │   └── WorkflowExecution #2 (Retry)
```

Behaviours:

| **Scenario**                | **WorkflowEntityID** | **RunID**       | **Execution Tree**  |
|-----------------------------|----------------------|-----------------|---------------------|
| **ContinueAsNew**           | ✅ Same              | ✅ Same         | 🟢 Same tree        |
| **Sub-Workflow**      | 🔄 New              | ✅ Same         | 🔄 New sub-tree     |
| **Cross-Queue Workflow**      | 🔄 New              | ✅ Same         | 🔄 New sub-tree     |
| **NewWorkflow (New ID)**  | 🔄 New              | 🔄 New         | 🆕 New tree        |

Usecases:

| **Workflow Type**            | **Execution Context**                      | **RunID**    | **Parent Waits?** | **Use Cases** |
|-----------------------------|------------------------------------------|-------------|-----------------|--------------|
| **ContinueAsNew**           | Same worker/goroutine                   | ✅ Same     | ✅ Yes          | Looping workflows, avoiding event history growth. |
| **Sub-workflow**            | Same worker/goroutine                   | ✅ Same     | ✅ Yes          | Modular workflow logic, breaking down large workflows. |
| **Cross-Queue Workflow**    | Different worker (same machine)         | ✅ Same     | ✅ Yes          | Concurrency control (e.g., limiting API key usage). |
| **New Workflow (Detached)** | New worker pool (same machine)          | ❌ New      | ❌ No           | Fire-and-forget tasks, batch processing, lifecycle workflows, external system orchestration, independent retries. |

This table keeps it **clear and structured** for your engine’s design. 🚀

---

Execution Scenarios in Tempolite

# Root RunID - Execution Tree

```
RunID-1
├── WorkflowEntity-1 (Parent Workflow)
│   ├── WorkflowExecution-1
│   │   ├── ActivityEntity-1
│   │   │   ├── ActivityExecution-1
│   │   │   └── ActivityExecution-2 (Retry)
│   │   ├── SideEffectEntity-1
│   │   │   ├── SideEffectExecution-1
│   │   ├── SagaEntity-1
│   │   │   ├── SagaExecution-1 (Transaction)
│   │   │   ├── SagaExecution-2 (Compensation)
│   │   ├── StartChildWorkflow → WorkflowEntity-2 (Same RunID, New WorkflowEntityID)
│   │   │   ├── WorkflowExecution-1
│   │   │   ├── WorkflowExecution-2 (Retry)
│   │   ├── ContinueAsNew → WorkflowExecution-2 (Same WorkflowEntityID)
│   │
│   ├── WorkflowExecution-2 (Retry)
│   │   ├── ActivityEntity-2
│   │   │   ├── ActivityExecution-1
│   │   ├── SideEffectEntity-2
│   │   │   ├── SideEffectExecution-1
│   │   ├── SagaEntity-2
│   │   │   ├── SagaExecution-1 (Transaction)
│   │   │   ├── SagaExecution-2 (Compensation)
│   │
│   ├── ContinueAsNew → WorkflowExecution-3 (Same WorkflowEntityID)
│   │
│   ├── Cross-Queue Workflow → WorkflowEntity-3 (Different Queue, Same RunID)
│   │   ├── WorkflowExecution-1
│   │   ├── WorkflowExecution-2 (Retry)
│   │
│   ├── StartWorkflow (Independent Execution) → RunID-2 (New Execution Tree)
│       ├── WorkflowEntity-4
│       │   ├── WorkflowExecution-1
│       │   ├── WorkflowExecution-2 (Retry)
│       │   ├── ActivityEntity-3
│       │   │   ├── ActivityExecution-1
│       │   ├── SideEffectEntity-3
│       │   │   ├── SideEffectExecution-1
│       │   ├── SagaEntity-3
│       │   │   ├── SagaExecution-1 (Transaction)
│       │   │   ├── SagaExecution-2 (Compensation)
```

### **Explanation of Execution Flows**

1. **Normal Execution**
   - `WorkflowEntity-1` runs under `RunID-1`.
   - It executes activities, side effects, and sagas.

2. **Retries**
   - If any activity, saga, or workflow execution fails, a **new execution ID** is created while keeping the same entity ID.

3. **ContinueAsNew**
   - A workflow clears its execution history but **keeps the same WorkflowEntityID**.
   - It starts a **new WorkflowExecutionID**.

4. **StartChildWorkflow**
   - Creates a **new WorkflowEntityID** but remains **under the same RunID**.
   - Parent can track and wait for completion.

5. **Cross-Queue Workflows**
   - Moves execution to another queue but stays in the same execution tree (`RunID` remains the same).

6. **StartWorkflow (New Execution Tree)**
   - Creates a **new RunID**, forming an independent execution tree.
   - No parent-child relationship is maintained.

This structure fully models **workflow executions, retries, child workflows, cross-queue executions, and independent workflow trees.** 🚀 Let me know if you need any refinements!

---

### Context changes 

We need to change the old API of options and context with a context builder as follow:

```go

type WorkflowContext struct {
    id uint // unique ID at runtime
}

func (ctx *WorkflowContext) WithRetries(policy RetryPolicy) *WorkflowOptionBuilder {
    return WorkflowOptionBuilder{}
}

type WorkflowOptionBuilder struct {}

func (b *WorkflowOptionBuilder) WithID(id string) *WorkflowOptionBuilder {}
func (b *WorkflowOptionBuilder) WithRetries(policy RetryPolicy) *WorkflowOptionBuilder {}

func Workflow(ctx tempolite.WorkflowContext) error {

    // question is what is the real params?
    
    if err := ctx.Activity(ctx.WithRetries(RetryPolicy{})).Get(); err != nil {
        return err
    }

    //  with nothing, you can pass the same context
    if err := ctx.Workflow(ctx, Workflow).Get(); err != nil {
        return err
    }

    return nil
}

func Activity(ctx tempolite.ActivityContext) error {
    return nil
}

```

The goal is to simplify the API and making it easier to use.
