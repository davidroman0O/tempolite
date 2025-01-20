package main

import (
	"context"
	"fmt"
	"time"

	"github.com/k0kubun/pp/v3"

	"github.com/davidroman0O/tempolite"
)

func Workflow(ctx tempolite.WorkflowContext, _id int) ([]tempolite.WorkflowEntityID, error) {

	var waitFor []tempolite.Future
	for trackID := range 100 {
		waitFor = append(waitFor,
			ctx.Workflow(
				fmt.Sprintf("track-%d", trackID),
				Track,
				&tempolite.WorkflowOptions{
					Queue:          "track-queue",
					DeferExecution: true,
				},
				trackID))
	}

	ids := []tempolite.WorkflowEntityID{}
	for _, futu := range waitFor {
		ids = append(ids, futu.WorkflowID())
	}

	return ids, nil
}

func Track(ctx tempolite.WorkflowContext, a tempolite.WorkflowEntityID) error {

	if err := ctx.Activity("discovery", Discovery, nil, a).Get(); err != nil {
		return err
	}

	if err := ctx.Activity("download", Download, nil, a).Get(); err != nil {
		return err
	}

	if err := ctx.Activity("encode", Encode, nil, a).Get(); err != nil {
		return err
	}

	if err := ctx.Activity("metadata", Metadata, nil, a).Get(); err != nil {
		return err
	}

	return nil
}

func Discovery(ctx tempolite.ActivityContext, a tempolite.WorkflowEntityID) error {
	return nil
}

func Download(ctx tempolite.ActivityContext, a int) error {

	// wait random between 10 to 55 seconds
	// <-time.After(time.Duration(rand.Intn(5)+2) * time.Millisecond)
	<-time.After(time.Millisecond * 100)

	return nil
}

func Encode(ctx tempolite.ActivityContext, a int) error {

	// wait random between 10 to 55 seconds
	// <-time.After(time.Duration(rand.Intn(3)+2) * time.Millisecond)
	<-time.After(time.Millisecond * 100)

	return nil
}

func Metadata(ctx tempolite.ActivityContext, a int) error {

	// wait random between 10 to 35 seconds
	// <-time.After(time.Duration(rand.Intn(2)+2) * time.Millisecond)
	<-time.After(time.Millisecond * 100)

	return nil
}

func main() {

	ctx := context.Background()
	var tp *tempolite.Tempolite

	var err error

	tp, err = tempolite.New(
		ctx,
		tempolite.NewMemoryDatabase(),
		tempolite.WithQueue(tempolite.QueueConfig{
			Name:         "track-queue",
			MaxRuns:      5,
			MaxWorkflows: 5,
			Debug:        true,
		}),
		tempolite.WithDefaultQueueWorkers(5, 5),
		tempolite.WithChangeHandler(func() {
			// pp.Println("change::", tp.Metrics())
		}),
	)
	if err != nil {
		panic(err)
	}

	var future tempolite.Future
	if future, err = tp.ExecuteDefault(
		Workflow,
		nil,
		1,
	); err != nil {
		panic(err)
	}

	list := []tempolite.WorkflowEntityID{}
	if err := future.Get(&list); err != nil {
		panic(err)
	}

	go func() {
		for {
			<-time.After(time.Second / 4)
			pp.Println("Metrics::", tp.Metrics())
		}
	}()

	for _, id := range list {
		if _, err := tp.Enqueue(id); err != nil {
			panic(err)
		}
	}

	if err := tp.Wait(); err != nil {
		panic(err)
	}

	if err := tp.Close(); err != nil {
		panic(err)
	}
}
