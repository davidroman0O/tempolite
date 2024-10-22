# Tempolite 🚀

Tempolite is a lightweight, SQLite-based workflow engine for Go that provides deterministic execution of workflows with support for activities, side effects, sagas, signals, and versioning. It's designed to be a simpler alternative to complex workflow engines while maintaining essential features for reliable business process automation.

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
go get github.com/davidroman0O/go-tempolite
```

Let's create a simple workflow that processes an order:

```go
type OrderID string

func ProcessOrderWorkflow(ctx tempolite.WorkflowContext[OrderID], orderID string) error {
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
    tp, err := tempolite.New[OrderID](
        context.Background(),
        tempolite.NewRegistry[OrderID]().
            Workflow(ProcessOrderWorkflow).
            Activity(ValidateOrder).
            Build(),
        tempolite.WithPath("./workflows.db"),
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
func OrderWorkflow(ctx tempolite.WorkflowContext[OrderID], orderID string) error {
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
func ProcessOrder(ctx tempolite.ActivityContext[OrderID], orderID string) (float64, string, error) {
    return 99.99, "processed", nil
}

func WorkflowWithMultipleReturns(ctx tempolite.WorkflowContext[OrderID], orderID string) error {
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
func ProcessPayment(ctx tempolite.ActivityContext[OrderID], amount float64) error {
    // Your payment logic here
    return nil
}

// Register it in your registry
registry := tempolite.NewRegistry[OrderID]().
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

func (ps *PaymentService) Run(ctx tempolite.ActivityContext[OrderID], amount float64) error {
    // Use ps.client and ps.config here
    return ps.client.Charge(amount)
}

// Create an instance and register its Run method
paymentService := &PaymentService{
    client: newPaymentClient(),
    config: loadConfig(),
}

registry := tempolite.NewRegistry[OrderID]().
    Activity(paymentService.Run).  // Register the Run method, not the struct!
    Build()
```

The key difference? Worker-based activities can share resources and maintain state between executions, while function-based activities are stateless.
