package tempolite

import (
	"fmt"
	"testing"

	"github.com/k0kubun/pp/v3"
)

func TestBasicDatabase(t *testing.T) {
	db := NewDefaultDatabase()
	if db == nil {
		t.Error("Failed to create database")
	}

	id, err := db.AddWorkflowEntity(&WorkflowEntity{
		BaseEntity: BaseEntity{
			HandlerName: "test",
			Type:        EntityTypeWorkflow,
		},
		WorkflowData: &WorkflowData{
			Inputs: [][]byte{},
		},
	})
	if err != nil {
		t.Error(err)
	}

	fmt.Println("Added workflow entity with ID:", id)

	data, err := db.GetWorkflowEntity(id)
	if err != nil {
		t.Error(err)
	}

	pp.Println(data)
}
