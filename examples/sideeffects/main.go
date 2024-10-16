package main

import (
	"context"
	"fmt"
	"log"

	"github.com/davidroman0O/go-tempolite"
)

func Workflow(ctx tempolite.WorkflowContext[string]) error {

	var pathAOrB bool
	if err := ctx.SideEffect("switch", func(ctx tempolite.SideEffectContext[string]) bool {
		return true
	}).Get(&pathAOrB); err != nil {
		return err
	}

	if pathAOrB {
		if err := ctx.ActivityFunc("activityA", ActivityA).Get(); err != nil {
			return err
		}
	} else {
		if err := ctx.ActivityFunc("activityb", ActivityB).Get(); err != nil {
			return err
		}
	}

	return nil
}

func ActivityA(ctx tempolite.ActivityContext[string]) error {
	fmt.Println("Activity A")
	return nil
}

func ActivityB(ctx tempolite.ActivityContext[string]) error {
	fmt.Println("Activity B")
	return nil
}

func main() {
	ctx := context.Background()
	tp, err := tempolite.New[string](
		ctx,
		tempolite.NewRegistry[string]().
			Workflow(Workflow).
			ActivityFunc(ActivityA).
			ActivityFunc(ActivityB).
			Build(),
	)
	if err != nil {
		log.Fatalf("Failed to create Tempolite instance: %v", err)
	}

	if err := tp.Workflow("workflow", Workflow).Get(); err != nil {
		log.Fatalf("Failed to enqueue workflow: %v", err)
	}

	if err := tp.Wait(); err != nil {
		log.Fatalf("Failed to wait for Tempolite instance: %v", err)
	}

	tp.Close()
}
