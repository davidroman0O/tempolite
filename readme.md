# Tempolite 🚀

Tempolite is a lightweight, workflow engine for Go that provides deterministic execution of workflows with support for activities, side effects, sagas, signals, and versioning. It's designed to be a simpler alternative to complex workflow engines while maintaining essential features for reliable business process automation.

> Work In Progress: Not ready for prime time but good enough for playing with it and small applications.

<!-- TODO: make a section that explain the different patterns, explain what you should or should not do -->
<!-- TODO: explain the whole idea of "it's not temporal but a specialized local pocket workflow engine", we're staking tasks to do in a convinient way -->

## Features

### 🔄 Workflows
- Deterministic execution with automatic retries
- Support for sub-workflows
- Version management for handling code changes
- Pause/Resume capabilities
- ContinueAsNew for long-running workflows
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

<!-- ### 📦 Database Management -->

### TODO:
- Replay and retry mechanisms for debugging and recovery
- SQLite database implementation

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

	database := tempolite.NewMemoryDatabase()

    tp, err := tempolite.New(
        context.Background(),
        database,
    )
    if err != nil {
        log.Fatal(err)
    }
    defer tp.Close()

    if err := tp.ExecuteDefault("process-order", ProcessOrderWorkflow, nil, "order-123").Get(); err != nil {
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
    if err := ctx.Activity("calculate", CalculateTotal, orderID).Get(&total); err != nil {
        return err
    }

    // SideEffectInfo
    var trackingNumber string
    if err := ctx.SideEffect("tracking", GenerateTrackingNumber).Get(&trackingNumber); err != nil {
        return err
    }

    // SignalInfo
    var approval bool
    if err := ctx.Signal("approval", &approval); err != nil {
        return err
    }

    // SagaInfo
    // TODO: I will change that `.Get` to a direct call instead
    if err := ctx.Saga("process", sagaDef).Get(); err != nil {
        return nil
    }

    // WorkflowInfo (when starting workflows)
    var result string
    if err := ctx.Workflow("sub-process", SubWorkflow, nil, "data").Get(&result); err != nil {
        return err
    }

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
activityFuture := ctx.Activity("process", ProcessOrder, orderID)

// Do other work...

// Get the results when you need them
var amount float64
var status string
if err := activityFuture.Get(&amount, &status); err != nil {
    return err
}
```

This pattern is especially useful when working with multiple operations:

```go
// Trigger multiple activities
validateFuture := ctx.Activity("validate", ValidateOrder, orderID)
paymentFuture := ctx.Activity("payment", ProcessPayment, orderID)
shippingFuture := ctx.Activity("shipping", ArrangeShipping, orderID)

// Get results in any order
var validationResult bool
if err := validateFuture.Get(&validationResult); err != nil {
    return err
}

var shippingLabel string
if err := shippingFuture.Get(&shippingLabel); err != nil {
    return err
}

var paymentRef string
if err := paymentFuture.Get(&paymentRef); err != nil {
    return err
}
```

### During Replay

The Info pattern handles replay seamlessly. During replay or retry:
1. Previously successful operations return their original results instantly
2. Failed operations are re-executed
3. The `Get` method behavior remains consistent

This makes your workflow code clean and predictable, whether it's running for the first time or being replayed.

####

TODO: finish the readme.md