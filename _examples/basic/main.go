package main

import (
	"context"
	"fmt"

	tempolite "github.com/davidroman0O/tempolite"
)

func SimpleWorkflow(ctx tempolite.WorkflowContext, a int) (int, error) {
	fmt.Println("SimpleWorkflow", a)
	return a * 42, nil
}

func main() {

	ctx := context.Background()
	db := tempolite.NewMemoryDatabase()

	tp, err := tempolite.New(ctx, db)
	if err != nil {
		panic(err)
	}

	future, err := tp.ExecuteDefault(SimpleWorkflow, nil, 42)
	if err != nil {
		panic(err)
	}

	var value int
	if err := future.Get(&value); err != nil {
		panic(err)
	}

	fmt.Println("Workflow result:", value)

}
