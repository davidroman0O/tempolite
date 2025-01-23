package tests

import (
	"context"
	"testing"

	"github.com/davidroman0O/tempolite/data"
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

	// fmt.Println(d.GetRunByID(1))
	defer d.Close()
}
