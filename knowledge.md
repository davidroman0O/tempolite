
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

## Things I changed

### Versioning

When using Temporal, they have a workflow.GetVersion to check the version of a workflow, so if you inverse two side effect because you were working or if you have a hot fix to do an important workflow in production you're simply screwed. 

I don't think I like having to do that```func determineVersion(ctx workflow.Context) (int, int) {
    version1 := workflow.GetVersion("side-effect-1", workflow.DefaultVersion, 1)
    version2 := workflow.GetVersion("side-effect-2", workflow.DefaultVersion, 1)
    return version1, version2
}

func Workflow(ctx workflow.Context) error {
    version1, version2 := determineVersion(ctx)

    if version1 == workflow.DefaultVersion {
        // Logic for side effect 1, old version
    } else {
        // Logic for side effect 1, new version
    }

    if version2 == workflow.DefaultVersion {
        // Logic for side effect 2, old version
    } else {
        // Logic for side effect 2, new version
    }

    return nil
}
```

I FUCKING HATE IT HOW THEY DO IT

Instead I will use the folder hierarchy to provide the versioning, for example you can have `github.com/davidroman0O/go-tempolite/examples/todo/tempolite/v1/activities/addTodo` as reflect.TypeOf(v).PkgPath() which clearly identify the activity. 

```
tempolite/
├── v1/
│   ├── workflow/
│   │   └── workflow.go
│   ├── activities/
│   │   └── activities.go
│   ├── side_effects/
│   │   └── side_effects.go
│   └── saga/
│       └── saga.go
├── v2/
│   ├── workflow/
│   │   └── workflow.go
│   ├── activities/
│   │   └── activities.go
│   ├── side_effects/
│   │   └── side_effects.go
│   └── saga/
│       └── saga.go
└── main.go
```

Benefits from use that structure:

- Enhanced Modularity:
   - Each version has its own folder, and within that version, each type of component (workflows, activities, side effects, sagas) is further isolated.
   - This allows for better modularity, where each component can be updated or reviewed independently without worrying about the others.

- Improved Readability:
   - With sub-folders, it's easier for developers to understand which logic belongs to workflows, activities, side effects, or sagas.
   - It also helps when onboarding new team members, as they can easily find the logic related to specific parts of a workflow version.

- Cleaner Dependency Management:
   - If activities or side effects have dependencies specific to a version, they can be kept isolated in their respective sub-folders.
   - This reduces the risk of version-specific logic or imports bleeding into other parts of the codebase.

- Simpler Refactoring:
   - Should you need to move an entire version or refactor one component (e.g., activities in `v2`), you can do so without impacting the rest of the workflow logic.
   - It also makes it easier to manage changes across versions since each version’s components are self-contained.


Which should look like 

```go

import (
    v1Workflows "github.com/dude/project/tempolite/v1/workflow"
)

func main() {
    tp, _ := tempolite.New()

    tp.RegisterWorkflow(v1Workflows.Something)
    // or even with optional an personal preference
    tp.RegisterWorkflow(v1Workflows.Something, "v1")
    
    // Or 
    // I want to setup my workflow just I would setup an http api
    v1Group := tp.WorkflowGroup("v1")
    tp.RegisterWorkflow(v1Workflows.Something)

}

```

