package main

import (
	"log"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

func main() {
	// Load the custom template directory
	err := entc.Generate("./ent/schema", &gen.Config{
		Templates: []*gen.Template{
			gen.MustParse(gen.NewTemplate("node_hashcode").
				ParseFiles("./ent/template/node.tmpl")),
		},
	})

	if err != nil {
		log.Fatalf("running ent codegen: %v", err)
	}
}
