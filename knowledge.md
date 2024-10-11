
It seems there is a clear distinct between WorkflowExecution and Workflow, like ActivityExecution and Activity, what's happening within the execution model only concern them for example, if I execute an activity from within a workflow and want to wait for it then i will have the execution id to wait for that specific execution, but if the workflow is a root resource that mean if i use the client to get the info and wait for that workflow i will have the workflow ID and not the workflow execution id with would only concern whats inside that execution model since a workflow might anyway retry multiple times and have multiple executions within the execution model.

> Execution model vs Logical model

There is clearly a distinction, a Workflow logical model might have multiple workflow executions within the execution model. If we want to know when the workflow is finished we only care about the logical model data, if we are within the workflow that mean we are within the execution model and if we want to wait for a sub-workflow that mean we wait for its execution resource within the execution model.

## {Execution}Info vs Info

There is the Info from within the execution that will wait for the execution id of the resource
There is the Info from outside the execution that will wait for the resource 

Two very different things

## Resource status vs execution status

What happen within the execution is only execution related, what happen outside of tempolite is just resource related.

