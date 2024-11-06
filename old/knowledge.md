
It seems there is a clear distinct between WorkflowExecution and Workflow, like ActivityExecution and Activity, what's happening within the execution model only concern them for example, if I execute an activity from within a workflow and want to wait for it then i will have the execution id to wait for that specific execution, but if the workflow is a root resource that mean if i use the client to get the info and wait for that workflow i will have the workflow ID and not the workflow execution id with would only concern whats inside that execution model since a workflow might anyway retry multiple times and have multiple executions within the execution model.

> Execution model vs Logical model

There is clearly a distinction, a Workflow logical model might have multiple workflow executions within the execution model. If we want to know when the workflow is finished we only care about the logical model data, if we are within the workflow that mean we are within the execution model and if we want to wait for a sub-workflow that mean we wait for its execution resource within the execution model.

## {Execution}Info vs Info

There is the Info from within the execution that will wait for the execution id of the resource
There is the Info from outside the execution that will wait for the resource 

Two very different things

## Resource status vs execution status

What happen within the execution is only execution related, what happen outside of tempolite is just resource related.

## TLDR

Workflow
├── Activity
│   └── Workflows call activities to perform external, non-deterministic tasks.
│       This could include API calls, database operations, or other I/O tasks.
│       Activities are separate from the workflow’s deterministic execution,
│       meaning they don’t maintain history and are retried according to their
│       own retry policies.
│       
├── SideEffect
│   └── Workflows use SideEffects for capturing non-deterministic values like
│       random numbers or timestamps that need to remain consistent upon replay.
│       SideEffects allow a workflow to record a value in its history so that
│       on replay, the recorded value is used rather than recalculating it.
│       This keeps the workflow deterministic.
│       
├── Sub-Workflow
│   └── Workflows can orchestrate other workflows as Sub-Workflows when they
│       need to break down a complex process into smaller, more manageable
│       units. Sub-Workflows have their own histories and can run independently,
│       but their execution is still controlled by the parent workflow. 
│       The parent workflow waits for the Sub-Workflow’s completion or handles
│       any errors that might occur during its execution.
│       
└── Saga
    ├── Directly Manages Sub-Activities 
    │   └── A specialized Saga function directly manage
    │       sub-activities due to the non-deterministic nature of its tasks.
    │       This would allow the Saga to perform complex compensating actions
    │       as a series of smaller, discrete tasks.
    │
    └── A Workflow can implement the Saga pattern to coordinate long-running
        business processes with compensation logic. A Saga is a sequence of
        activities where each step might require a compensating action if
        a failure occurs. The Workflow manages these compensations, ensuring
        that the overall state remains consistent, even if parts of the process
        fail.

Tempolite is NOT Temporal
It can support similar concepts but not similar features and I do not intent to have a full on-complete workflow engine solution AT ALL.
It's a very lightweight version of the concepts of Temporal for small scoped applications.

Therefore, i can't have the full on history of events with full-on backend detection of the order of the activities. Because it would be too much code to produce and too much bloat when we could have the bare minimum just to be comfortable with small applications

Maybe we need to add names on each "Execute" something like tp.Workflow("name workflow", thefunction, someParam{}) or even ctx.Activity("second-one", As[someStructActivity](), paramparam{}) so we could really "downgrade" but keep a reliable experience using it 
even maybe add a generic parameter to tempolite which has to be a primitive interface 

```go
type primitive interface {
   string | int 
}
```

something like that, so the dev can put whatever it want as identifier of its app for each workflow or activity or sideeffect or saga calls so when it retries then tempolite knows which one is which

## Activities vs Side Effects

I had a hard time to truely and deeply understand the difference between both.

- Activities handle external non-determinism (like network requests)
- SideEffects handle internal non-determinism (like generating timestamps)

> but but you can just use Activities then it's the same with retries

- Activities often represent expensive and long-running tasks that are external
- SideEffects are lightweight and used for quick operations within the workflow logic that need to be captured as-is

Basically if you need to enhance the logic of the workflow, use a SideEffect, if you need to feed data to the workflow then it's an Activity. 

> SideEffects are for enhancing or modifying the internal workflow logic (like introducing non-deterministic behavior within the workflow), while Activities are used to bring in external data or interact with systems outside the workflow.

SideEffects are designed to be simple, synchronous operations that directly use the workflow's scope, allowing them to integrate seamlessly with the workflow's internal logic. They are immediately recorded in the workflow history because they introduce non-deterministic behavior (like accessing the current time or generating a random number) that needs to be captured for determinism during replay.

On the other hand, Activities are external calls and do not share the workflow's scope because they operate outside the workflow’s internal logic. They are more complex, potentially long-running, and involve external systems, so they need to be treated differently — with retries, failure handling, and result persistence in history after successful execution.

This is why SideEffects are lightweight and use the workflow’s context directly, while Activities are more independent and operate as separate units with their own lifecycle.

