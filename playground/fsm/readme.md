

I completly change the idea of queues as it was, in theory it is useful to constraint a certain amount of workflows/activities into a fixed amount of workers which is quite useful to force bottleneck

With the recent addition of ProcessedNotification and QueuedNotification, you can have a quite similar process.

Since i plan to use retrypool to wrap that orchestrators, you will be able to process X amount of workflows as there are workers. We should see queues as another retrypool that contains orchestrator that that queue. Which mean that two retrypool will share the same database implementation as a constraint so one workflow in the default queue can wait for the completion of other workflows in the other retrypool by interrogating the database.

We need also to have a parent callback shared by retrypool/orchestrators so they can dispatch a workflow to another queue pool. 

That mean we need to register root workflows per each queue pool which will register them into orchestrators. We can only dispatch root workflows.

