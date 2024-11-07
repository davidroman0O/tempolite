package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/davidroman0O/tempolite"
	tempoliteContext "github.com/davidroman0O/tempolite/internal/engine/context"
	"github.com/davidroman0O/tempolite/internal/engine/registry"
)

func TestBasic(t *testing.T) {

	wrk := func(ctx tempoliteContext.WorkflowContext) error {
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

	if err = tp.Shutdown(); err != nil {
		t.Fatal(err)
	}
}
