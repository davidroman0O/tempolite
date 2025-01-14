package main

import (
	"context"
	"fmt"
	"time"

	"math/rand"

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
	<-time.After(time.Duration(rand.Intn(45)+10) * time.Second)

	return nil
}

func Encode(ctx tempolite.ActivityContext, a int) error {

	// wait random between 10 to 55 seconds
	<-time.After(time.Duration(rand.Intn(45)+10) * time.Second)

	return nil
}

func Metadata(ctx tempolite.ActivityContext, a int) error {

	// wait random between 10 to 35 seconds
	<-time.After(time.Duration(rand.Intn(25)+10) * time.Second)

	return nil
}

func main() {

	ctx := context.Background()

	tp, err := tempolite.New(
		ctx,
		tempolite.NewMemoryDatabase(),
		tempolite.WithQueue(tempolite.QueueConfig{
			Name:         "track-queue",
			MaxRuns:      1,
			MaxWorkflows: 5,
			Debug:        true,
		}),
		tempolite.WithDefaultQueueWorkers(1, 5),
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
