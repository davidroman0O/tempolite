package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/davidroman0O/tempolite"
	tempoliteContext "github.com/davidroman0O/tempolite/internal/engine/context"
	"github.com/davidroman0O/tempolite/internal/engine/registry"
)

func TestBasic(t *testing.T) {

	failure := true

	wrk := func(ctx tempoliteContext.WorkflowContext) (int, error) {
		if failure {
			failure = false
			return 0, fmt.Errorf("error on purpose")
		}
		fmt.Println("workflow executed")
		return 420, nil
	}

	ctx := context.Background()

	tp, err := tempolite.New(
		ctx,
		registry.
			New().
			Workflow(wrk).
			Build(),
		tempolite.WithPath("tempolite-test.db"),
		tempolite.WithDestructive(),
	)

	if err != nil {
		t.Fatal(err)
	}

	info := tp.Workflow(wrk, nil)

	fmt.Println("Info", info.Get())

	// fmt.Println("scale up")
	// tp.Scale("default", map[string]int{
	// 	"workflows":   4,
	// 	"activities":  2,
	// 	"sideEffects": 2,
	// 	"sagas":       2,
	// })

	// <-time.After(1 * time.Second)

	// fmt.Println("scale down")
	// tp.Scale("default", map[string]int{
	// 	"workflows":   1,
	// 	"activities":  0,
	// 	"sideEffects": 0,
	// 	"sagas":       0,
	// })

	// <-time.After(1 * time.Second)

	// fmt.Println("scale up")
	// tp.Scale("default", map[string]int{
	// 	"workflows":   2,
	// 	"activities":  2,
	// 	"sideEffects": 2,
	// 	"sagas":       2,
	// })

	<-time.After(1 * time.Second)

	if err = tp.Shutdown(); err != nil {
		if err != context.Canceled {
			t.Fatal(err)
		}
	}
}
