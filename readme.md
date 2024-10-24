# Tempolite 🚀

Tempolite is a lightweight, SQLite-based workflow engine for Go that provides deterministic execution of workflows with support for activities, side effects, sagas, signals, and versioning. It's designed to be a simpler alternative to complex workflow engines while maintaining essential features for reliable business process automation.

> Not ready for prime time but good enough for playing with it and small applications.

## Features

### 🔄 Workflows
- Deterministic execution with automatic retries
- Support for sub-workflows
- Version management for handling code changes
- Pause/Resume capabilities
- ContinueAsNew for long-running workflows
- Replay and retry mechanisms for debugging and recovery
- Automatic state persistence

### 🛠 Activities
- Non-deterministic operations isolation
- Automatic retries on failure
- Support for both function-based and struct-based activities
- Error handling and recovery

### 📡 Side Effects
- Management of non-deterministic operations within workflows
- Perfect for handling random numbers, timestamps, or UUIDs
- Consistent replay behavior

### ⚡ Signals
- Asynchronous workflow communication
- Wait for external events
- Perfect for human interactions or system integrations

### 🔄 Sagas
- Transaction coordination with compensation logic
- Automatic rollback of completed steps on failure
- Step-by-step transaction execution
- Built-in error handling and recovery

### 📦 Database Management
- Automatic database rotation based on size or page count
- Built-in pool management for high availability
- SQLite-based for simplicity and reliability

## Getting Started

```bash
go get github.com/davidroman0O/tempolite
```

Let's create a simple workflow that processes an order:

```go

func ProcessOrderWorkflow(ctx tempolite.WorkflowContext, orderID string) error {
    // Execute an activity to validate the order
    var isValid bool
    if err := ctx.Activity("validate", ValidateOrder, orderID).Get(&isValid); err != nil {
        return err
    }

    if !isValid {
        return fmt.Errorf("invalid order: %s", orderID)
    }

    return nil
}

func main() {
    tp, err := tempolite.New(
        context.Background(),
        tempolite.NewRegistry().
            Workflow(ProcessOrderWorkflow).
            Activity(ValidateOrder).
            Build(),
        tempolite.WithPath("./workflows.db"), // optional, otherwise memory
    )
    if err != nil {
        log.Fatal(err)
    }
    defer tp.Close()

    if err := tp.Workflow("process-order", ProcessOrderWorkflow, nil, "order-123").Get(); err != nil {
        log.Fatal(err)
    }
}
```

## Working with Workflow Components: The Info Pattern

When you trigger any operation in a workflow (activities, side effects, signals, or sagas), Tempolite returns an Info struct with a `Get` method. This consistent pattern helps you handle results and errors:

```go
func OrderWorkflow(ctx tempolite.WorkflowContext, orderID string) error {
    // ActivityInfo
    var total float64
    err := ctx.Activity("calculate", CalculateTotal, orderID).Get(&total)

    // SideEffectInfo
    var trackingNumber string
    err = ctx.SideEffect("tracking", GenerateTrackingNumber).Get(&trackingNumber)

    // SignalInfo
    var approval bool
    err = ctx.Signal("approval").Get(&approval)

    // SagaInfo
    err = ctx.Saga("process", sagaDef).Get() // Sagas don't return values

    // WorkflowInfo (when starting workflows)
    var result string
    err = ctx.Workflow("sub-process", SubWorkflow, nil, "data").Get(&result)

    return err
}
```

### Understanding the Info Pattern

Each Info struct (`ActivityInfo`, `SideEffectInfo`, etc.) follows the same principles:

1. They're returned immediately when you trigger the operation
2. The `Get` method blocks until the operation completes
3. `Get` accepts pointer arguments to store results
4. The number of pointer arguments must match the operation's return values

Here's a more detailed example:

```go
// An activity that returns multiple values
func ProcessOrder(ctx tempolite.ActivityContext, orderID string) (float64, string, error) {
    return 99.99, "processed", nil
}

func WorkflowWithMultipleReturns(ctx tempolite.WorkflowContext, orderID string) error {
    var (
        amount  float64
        status  string
    )

    // Get accepts multiple pointers matching the activity's return values
    // (excluding the error which is returned by Get itself)
    if err := ctx.Activity("process", ProcessOrder, orderID).Get(&amount, &status); err != nil {
        return fmt.Errorf("process failed: %w", err)
    }

    log.Printf("Processed order: amount=%f, status=%s", amount, status)
    return nil
}
```

### Working with Results

The Info pattern helps you handle operation results in a clean way:

```go
// Store the info struct for later use
activityInfo := ctx.Activity("process", ProcessOrder, orderID)

// Do other work...

// Get the results when you need them
var amount float64
var status string
if err := activityInfo.Get(&amount, &status); err != nil {
    return err
}
```

This pattern is especially useful when working with multiple operations:

```go
// Trigger multiple activities
validateInfo := ctx.Activity("validate", ValidateOrder, orderID)
paymentInfo := ctx.Activity("payment", ProcessPayment, orderID)
shippingInfo := ctx.Activity("shipping", ArrangeShipping, orderID)

// Get results in any order
var validationResult bool
if err := validateInfo.Get(&validationResult); err != nil {
    return err
}

var shippingLabel string
if err := shippingInfo.Get(&shippingLabel); err != nil {
    return err
}

var paymentRef string
if err := paymentInfo.Get(&paymentRef); err != nil {
    return err
}
```

### During Replay

The Info pattern handles replay seamlessly. During replay or retry:
1. Previously successful operations return their original results instantly
2. Failed operations are re-executed
3. The `Get` method behavior remains consistent

This makes your workflow code clean and predictable, whether it's running for the first time or being replayed.

## Activities: Two Ways to Handle External Work

Activities in Tempolite handle all your non-deterministic operations - think API calls, database operations, or any external work. There are two ways to create activities, each suited for different needs.

### Function-Based Activities

The simplest way is to just write a function. This is perfect for straightforward operations:

```go
func ProcessPayment(ctx tempolite.ActivityContext, amount float64) error {
    // Your payment logic here
    return nil
}

// Register it in your registry
registry := tempolite.NewRegistry().
    Activity(ProcessPayment).
    Build()
```

### Worker-Based Activities

Sometimes you need more complex activities that share resources or maintain state. For these cases, you can create a worker structure:

```go
type PaymentService struct {
    client *PaymentClient
    config *Config
}

func (ps *PaymentService) Run(ctx tempolite.ActivityContext, amount float64) error {
    // Use ps.client and ps.config here
    return ps.client.Charge(amount)
}

// Create an instance and register its Run method
paymentService := &PaymentService{
    client: newPaymentClient(),
    config: loadConfig(),
}

registry := tempolite.NewRegistry().
    Activity(paymentService.Run).  // Register the Run method, not the struct!
    Build()
```

The key difference? Worker-based activities can share resources and maintain state between executions, while function-based activities are stateless.

### Understanding Activity Inputs and Outputs

Activities are flexible with their inputs and outputs, following these rules:

1. First parameter must be the activity context
2. Can have any number of additional input parameters
3. Must return at least an error as the last return value
4. Can return any number of values before the error

Here's an example showing different activity signatures:

```go
// Minimal activity: just context and error
func SimpleActivity(ctx tempolite.ActivityContext) error {
    return nil
}

// Single input, single output + error
func CalculateTotal(ctx tempolite.ActivityContext, orderID string) (float64, error) {
    return 99.99, nil
}

// Multiple inputs, multiple outputs + error
func ProcessOrder(
    ctx tempolite.ActivityContext,
    orderID string,
    amount float64,
    options mapstring,
) (string, float64, []string, error) {
    return "processed", 99.99, []string{"item1", "item2"}, nil
}
```

When using activities, you must:
1. Provide all input parameters when calling the activity
2. Provide pointers for all outputs (except error) when calling Get

```go
func OrderWorkflow(ctx tempolite.WorkflowContext, orderID string) error {
    // Activity with multiple inputs
    activityInfo := ctx.Activity(
        "process",
        ProcessOrder,
        orderID,
        99.99,
        mapstring{"priority": "high"},
    )

    // Must provide pointers for ALL outputs
    var (
        status string
        finalAmount float64
        items []string
    )
    if err := activityInfo.Get(&status, &finalAmount, &items); err != nil {
        return err
    }

    return nil
}
```

The same pattern applies to workflows:

```go
// Workflow can have multiple inputs and outputs
func ComplexWorkflow(
    ctx tempolite.WorkflowContext,
    orderID string,
    amount float64,
) (string, float64, error) {
    // Workflow logic here
    return "completed", amount * 1.1, nil
}

// When starting the workflow
workflowInfo := tp.Workflow(
    "complex",
    ComplexWorkflow,
    nil,
    "order-123",
    99.99,
)

// When getting results
var (
    status string
    finalAmount float64
)
if err := workflowInfo.Get(&status, &finalAmount); err != nil {
    log.Fatal(err)
}
```

If you don't provide enough pointer arguments to `Get()` or provide the wrong types, Tempolite will return an error. This helps ensure type safety and correct handling of all outputs.

## Smart Execution: How Tempolite Handles Retries and Replays

When a workflow fails or is replayed, Tempolite is smart about re-execution. Let's say you have this workflow:

```go
func OrderWorkflow(ctx tempolite.WorkflowContext, orderID string) error {
    // Generate tracking number
    var trackingNum string
    if err := ctx.SideEffect("tracking", GenerateTrackingNumber).Get(&trackingNum); err != nil {
        return err
    }

    // Process payment
    if err := ctx.Activity("payment", ProcessPayment, 100.0).Get(); err != nil {
        return err
    }

    // This activity fails
    if err := ctx.Activity("shipping", ShipOrder, orderID).Get(); err != nil {
        return err
    }
    
    return nil
}
```

If the shipping activity fails and the workflow retries, Tempolite won't re-execute the successful side effect or payment activity. Instead, it'll use the stored results from the previous execution. This ensures:
- No double-charging customers
- Consistent tracking numbers
- Preservation of successful work
- Deterministic execution on replay

The same behavior applies when you manually replay a workflow using `ReplayWorkflow` - all previously successful operations return their original results.

## Retry vs Replay: Understanding the Difference

Tempolite offers two ways to handle workflow recovery:

### Retry Workflow

When you use `RetryWorkflow`, you're saying "start this workflow fresh with the same inputs." It creates a completely new workflow instance:

```go
// Start a fresh workflow with the same inputs
newWorkflowInfo := tp.RetryWorkflow(workflowID)
```

This is perfect when:
- You want to start from scratch
- Previous results should be discarded
- You're debugging issues
- You need different execution parameters

### Replay Workflow

`ReplayWorkflow` is different - it creates a new execution of the existing workflow, continuing from where it left off:

```go
// Continue the existing workflow
replayInfo := tp.ReplayWorkflow(workflowID)
```

Use replay when:
- You want to continue a paused workflow
- You're debugging workflow behavior
- You want to preserve the original workflow's decisions
- You need to verify deterministic execution

## Long-Running Workflows with ContinueAsNew

Some workflows run for a long time or loop indefinitely. For these cases, use `ContinueAsNew` to create a fresh workflow instance while maintaining logical continuity:

```go
func MonitoringWorkflow(ctx tempolite.WorkflowContext, iteration int) error {
    // Do some monitoring work...
    
    if iteration >= 1000 {
        // Create a fresh workflow instance with a clean history
        return ctx.ContinueAsNew(ctx, "continue", iteration + 1)
    }
    
    return nil
}
```

Why use ContinueAsNew?
- Prevents workflow history from growing too large
- Maintains logical workflow continuity
- Perfect for recurring tasks
- Helps with workflow versioning

## Signals: Handling External Events

Signals let your workflows communicate with the outside world:

```go
func ApprovalWorkflow(ctx tempolite.WorkflowContext, orderID string) error {
    // Wait for manager approval
    var approved bool
    if err := ctx.Signal("approval").Receive(ctx, &approved); err != nil {
        return err
    }

    if approved {
        return ctx.Activity("process", ProcessOrder, orderID).Get()
    }
    return nil
}

// In your manager approval service:
tp.PublishSignal(workflowID, "approval", true)
```

## Sagas: Managing Complex Transactions

Sagas help you manage sequences of operations where each step might need to be undone:

```go
type OrderStep struct {
    OrderID string
}

func (s OrderStep) Transaction(ctx tempolite.TransactionContext) (interface{}, error) {
    // Process the order
    return "Order processed", nil
}

func (s OrderStep) Compensation(ctx tempolite.CompensationContext) (interface{}, error) {
    // Undo the order processing
    return "Order cancelled", nil
}

func OrderWorkflow(ctx tempolite.WorkflowContext, orderID string) error {
    saga := tempolite.NewSaga[OrderID]().
        AddStep(OrderStep{OrderID: orderID}).
        Build()

    return ctx.Saga("process-order", saga).Get()
}
```

## Database Management with TempolitePool

The TempolitePool helps you manage your workflow database as it grows. Think of it as a smart database manager that:
- Creates new database files when the current one gets too big
- Seamlessly transitions workflows to new database files
- Keeps track of all your workflow history
- Handles database connections and cleanup

Here's how you might use it:

```go
pool, err := tempolite.NewTempolitePool(
    context.Background(),
    registry,
    tempolite.TempolitePoolOptions{
        MaxFileSize:  50 * 1024 * 1024,  // Start new DB after 50MB
        BaseFolder:   "./db",
        BaseName:     "workflows",
    },
)

// Use it just like regular Tempolite
workflowInfo := pool.Workflow("process-order", ProcessOrderWorkflow, nil, "order-123")
```

The pool takes care of all the database management details, so you can focus on your business logic. When a database file gets too large, it'll automatically create a new one and move new workflows there, while maintaining access to your historical workflow data.

## Best Practices

1. **Keep Workflows Deterministic**
   - Move non-deterministic operations to activities or side effects
   - Avoid time-based decisions in workflows
   - Use signals for external dependencies

2. **Handle Errors Properly**
   - Implement retry logic for activities
   - Use sagas for complex transactions
   - Plan compensation steps
   - Monitor workflow states

3. **Version Management**
   - Use GetVersion for workflow updates
   - Plan migration strategies
   - Test version transitions
   - Document version changes

4. **Database Management**
   - Configure appropriate rotation policies
   - Monitor database sizes
   - Plan cleanup strategies
   - Backup important workflows

5. **Testing**
   - Test workflow replay behavior
   - Verify compensation logic
   - Test version transitions
   - Simulate failures

# Roadmap

- More utility functions
- Adding Select

```go
ctx.Select(
    ctx.Workflow( /* ... */),
    ctx.Workflow( /* ... */)
)
```

- Timers
- Maybe providing Postges and Mysql support
- OpenTelemetry Tracing
- Web UI for diagnostic and administration

# If you like it you also might like

- [`github.com/davidroman0O/comfylite3`](https://github.com/davidroman0O/comfylite3): A `github.com/mattn/go-sqlite3` wrapper that prevent "database is locked" error when using multiple goroutines.
- [`github.com/davidroman0O/retrypool`](https://github.com/davidroman0O/retrypool): A powerful worker pool library not like you're used to see!