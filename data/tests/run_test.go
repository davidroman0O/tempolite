package tests

import (
	"context"
	"testing"

	"github.com/davidroman0O/tempolite/data"
	"github.com/k0kubun/pp/v3"
)

func TestRunCreate(t *testing.T) {
	ctx := context.Background()
	d, err := data.New(
		ctx,
		data.WithFilePath("./db/test_run_create.db"),
	)
	if err != nil {
		t.Fatal(err)
	}
	// if err := d.CreateNewWorkflow(data.CreateNewWorkflow{}); err != nil {
	// 	t.Fatal(err)
	// }

	work := func(ctx data.WorkflowContext) error {
		return nil
	}

	entity, err := d.NewWorkflow(
		work,
		data.WorkflowOptions{},
		data.NewWorkflowOptions(),
	)
	if err != nil {
		t.Fatal(err)
	}

	if entity.ID != 1 {
		t.Fatal("ID not equal to 1")
	}

	full, err := d.GetWorkflowEntityByID(entity.ID)
	if err != nil {
		t.Fatal(err)
	}

	pp.Println(full)

	// fmt.Println(d.GetRunByID(1))
	defer d.Close()
}
