package main

import (
	"context"

	"github.com/davidroman0O/tempolite/v2"
)

func main() {
	tp, err := tempolite.New(context.Background(), tempolite.WithInMemorySQLite())
	if err != nil {
		panic(err)
	}

	if err := tp.Start(); err != nil {
		panic(err)
	}

}
