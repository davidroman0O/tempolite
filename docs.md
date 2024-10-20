
The idea is to make an autonomous runtime that have workflows capabilities. 

Later on, I want that library to be controlled by another that will use the Raft protocol for distributed nodes. Each node will have a fixed set of handlers (with parameters), therefore specific responsiblity.


# Listing of all modules, rules, tricks and code

One retrypool per component
- handler
- saga handler
- compensation
- sideeffect

Specify how many workers per pool.

Each pool got their own worker for each component.


Here a general idea of the API:

```go 

func main() {
    // automatically detects the first parameter to be HandlerContext or HandlerSagaContext
    // put them into different maps
	tempolite.RegisterHandler(sagaHandler)
	tempolite.RegisterHandler(handler)
	tempolite.RegisterHandler(handlerSide)

    db, err := sql.Open("sqlite3", "tasks.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

    // by default the workers will have 1 instance
	tp, err := tempolite.New(
        ctx, 
        db, 
        // Later on we will add lot more options
        tempolite.WithSagaWorkers(3),
        tempolite.WithMaxDelay(time.Second*5), // we will have retrypool options and tempolite's options wrapped
    )
	if err != nil {
		log.Fatalf("Failed to initialize WorkflowBox: %v", err)
	}
    defer tp.Close()

    // tp.GetInfo(ctx context.Context, id string) (interface{}, error):  returns either saga information or handler information or side effect information, etc
    // tp.GetExecutionTree : return the whole tree of execution of a task
    // tp.SendSignal(ctx context.Context, id string, name string, payload interface{}) error 
    // tp.ReceiveSignal(ctx context.Context, id string, name string) (<-chan interface{}, error) 
    // tp.Cancel(ctx context.Context, id string) error
    // tp.Terminate(ctx context.Context, id string) error

    ctx := context.Background()
    // eventually, that context depends on the tempolite's context, if tempolite's context timout earlier then we will cancel this one too
    ctx, cancel := context.WithTimeout(30 * time.Second) 
    defer cancel()

    var id string
    var err error
    if id, err = tp.Enqueue(
        ctx, 
        handler, 
        12, 
        tempolite.EnqueueWithMaxDuration(time.Second*15), // we will have retrypool options and tempolite's options wrapped
    ); err != nil {
        panic(err)
    }

    fmt.Println(id)

    tp.Wait(func(info TempoliteInfo) bool {
        return info.Tasks == 0 && info.SagaTasks == 0 && info.CompensationTask == 0
    }, time.Second * 1)
}

func handlerSide(ctx tempolite.HandlerContext, param int) (interface{}, error) {
    return 40 + param, nil
}

type SomeSagaStep struct {
    // might hold transitory data that is kept as an instance during the lifecycle of the sagastep
}

// put whatever parmaeter you want
func NewSomeSagaStep() SomeSagaStep {
    return SomeSagaStep{}
}

func (s SomeSagaStep) Transaction(ctx tempolite.TransactionContext) error {
    return nil
}

func (s SomeSagaStep) Compensation(ctx tempolite.CompensationContext) error {
    return nil
}

func sagaHandler(ctx tempolite.HandlerSagaContext, param int) error {
    var err error
    var sideEffectResult interface{}
    // even if we restart the same sagaHandler when we retry, if successful, we will keep the result in database
    sideEffectResult, err = ctx.SideEffect("unique_key", func() (interface{}, error) {
        var err error
        var id string
        if id, err = ctx.Enqueue(handlerSide, 2); err != nil {
            return nil, err
        }
        var value interface{}
        if value, err = ctx.WaitForCompletion(id); err != nil {
            return nil, err
        }
        return value, nil
    })

    // we don't want people to use functions directly because they will mix-match the context and misuse the api
    // at least they will do less mistakes and have isolation of their steps
    if err = ctx.Step(NewSomeSagaStep()); err != nil {
        return err 
    }

    return nil
}


type SomeSideEffect struct {}

func (s SomeSideEffect) Run(ctx tempolite.SideEffectContext) (interface{}, error) {
    // if it were to use Enqueue a saga or handler, we will still wait anyway, do it doesn't really matter
    // we will have to return a result at some point
    // we can use ctx.Enqueue and we will have a guarantee to use the correct ctx
    return "result!!!", nil
}

func handler(ctx tempolite.HandlerContext, param int) (interface{}, error) {
    var err error
    var sagaID string
    // automatically create relationship parent-children
    // the context got the contextID automatically
    // enqueue is asynchronous
    if sagaID, err = ctx.Enqueue(sagaHandler, 42); err != nil {
        return err
    }

    if _, err = ctx.WaitForCompletion(sagaID); err != nil {
        return nil
    }

    var sideEffectResult interface{}

    // this is blocking, we don't need a wait api, it will wait for the function to complete
    // again, forcing the interface helps to prevent developer mistakes to use the handler context
    // and if offers benefits for me to maintain flat code + hierarchy + giving special context
    // the dev will be happy to have isolation and stability
    sideEffectResult, err = ctx.SideEffect("unique_key", SomeSideEffect{})

    signalCh, err := ctx.ReceiveSignal("wait-input")
    if err != nil {
        return nil, err
    }

    payload, ok := <-signalCh
    if !ok {
        return nil, fmt.Errorf("ooooh not ok")
    }

    //  handler or saga handler can or not return a result
    // `return data, err` is valid as `return err` is valid too
    return fmt.Sprintf("something with %v", payload), nil
}

```

The Enqueue function will have to detect if the handler is a saga or not
