package tests

import (
	"context"
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

	if err := tp.ScaleQueue(tempolite.DefaultQueue, 2); err != nil {
		t.Fatal(err)
	}

	<-time.After(time.Second)

	if err := tp.ScaleQueue(tempolite.DefaultQueue, 4); err != nil {
		t.Fatal(err)
	}

	<-time.After(time.Second)

	if err := tp.ScaleQueue(tempolite.DefaultQueue, 1); err != nil {
		t.Fatal(err)
	}

	<-time.After(time.Second)

	tp.Close()

}
