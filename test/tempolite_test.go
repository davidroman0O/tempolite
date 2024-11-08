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

	wrk := func(ctx tempoliteContext.WorkflowContext) error {
		fmt.Println("workflow executed")
		return nil
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

	<-time.After(1 * time.Second)

	if err = tp.Shutdown(); err != nil {
		if err != context.Canceled {
			t.Fatal(err)
		}
	}
}
