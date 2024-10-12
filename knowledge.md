
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