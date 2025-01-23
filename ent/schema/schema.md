
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


| **Scenario**                | **WorkflowEntityID** | **RunID**       | **Execution Tree**  |
|-----------------------------|----------------------|-----------------|---------------------|
| **ContinueAsNew**           | ✅ Same              | ✅ Same         | 🟢 Same tree        |
| **StartChildWorkflow**      | 🔄 New              | ✅ Same         | 🔄 New sub-tree     |
| **StartWorkflow (New ID)**  | 🔄 New              | 🔄 New         | 🆕 New tree        |

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

