package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	tempolite "github.com/davidroman0O/tempolite"
)

func TestTempoliteScale(t *testing.T) {

	database := tempolite.NewMemoryDatabase()
	defer database.SaveAsJSON("./jsons/tempolite_workflows_execute.json")
	ctx := context.Background()

	tp, err := tempolite.New(
		ctx,
		database,
	)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("Scale up")
	if err := tp.ScaleQueue(tempolite.DefaultQueue, 2); err != nil {
		t.Fatal(err)
	}

	<-time.After(time.Second)

	fmt.Println("Scale up")
	if err := tp.ScaleQueue(tempolite.DefaultQueue, 4); err != nil {
		t.Fatal(err)
	}

	<-time.After(time.Second)

	fmt.Println("Scale down")
	if err := tp.ScaleQueue(tempolite.DefaultQueue, 0); err != nil {
		t.Fatal(err)
	}

	<-time.After(time.Second * 2)

	tp.Close()

}

func TestTempoliteScaleConcurrentWorkflows(t *testing.T) {

	database := tempolite.NewMemoryDatabase()
	defer database.SaveAsJSON("./jsons/tempolite_workflows_execute.json")
	ctx := context.Background()

	tp, err := tempolite.New(
		ctx,
		database,
	)
	if err != nil {
		t.Fatal(err)
	}

	workflowFunc := func(ctx tempolite.WorkflowContext) (int, error) {
		return 1, nil
	}

	go func() {
		for i := 0; i < 100; i++ {
			if _, err := tp.ExecuteDefault(workflowFunc, nil); err != nil {
				t.Fatal(err)
			}
		}
	}()

	fmt.Println("Scale up")
	if err := tp.ScaleQueue(tempolite.DefaultQueue, 2); err != nil {
		t.Fatal(err)
	}

	<-time.After(time.Second)

	fmt.Println("Scale up")
	if err := tp.ScaleQueue(tempolite.DefaultQueue, 4); err != nil {
		t.Fatal(err)
	}

	<-time.After(time.Second)

	fmt.Println("Scale down")
	if err := tp.ScaleQueue(tempolite.DefaultQueue, 1); err != nil {
		t.Fatal(err)
	}

	<-time.After(time.Second * 5)

	fmt.Println("Scale Up")
	if err := tp.ScaleQueue(tempolite.DefaultQueue, 2); err != nil {
		t.Fatal(err)
	}

	<-time.After(time.Second)

	if err := tp.Wait(); err != nil {
		t.Fatal(err)
	}

	tp.Close()

}
