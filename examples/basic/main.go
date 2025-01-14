package main

import (
	"context"
	"fmt"

	tempolite "github.com/davidroman0O/tempolite"
	"github.com/k0kubun/pp/v3"
)

func SimpleWorkflow(ctx tempolite.WorkflowContext, a int) (int, error) {
	fmt.Println("SimpleWorkflow", a)
	return a * 42, nil
}

func main() {

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	db := tempolite.NewMemoryDatabase()
	var tp *tempolite.Tempolite
	var err error

	tp, err = tempolite.New(
		ctx,
		db,
		tempolite.WithChangeHandler(func() {
			pp.Println(tp.Metrics())
		}),
	)
	if err != nil {
		panic(err)
	}

	// go func() {
	// 	ticker := time.NewTicker(time.Second / 32)
	// 	for {
	// 		select {
	// 		case <-ctx.Done():
	// 			return
	// 		case <-ticker.C:
	// 			pp.Println(tp.Metrics())
	// 		}
	// 	}
	// }()

	go func() {
		for range 10 {
			tp.ExecuteDefault(SimpleWorkflow, nil, 42)
		}
	}()

	// var value int
	// if err := future.Get(&value); err != nil {
	// 	panic(err)
	// }

	tp.Wait()

	// fmt.Println("Workflow result:", value)

	cancel()

}
