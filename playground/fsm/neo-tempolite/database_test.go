package tempolite

import (
	"fmt"
	"testing"

	"github.com/k0kubun/pp/v3"
)

func TestBasicDatabase(t *testing.T) {
	db := NewMemoryDatabase()
	if db == nil {
		t.Error("Failed to create database")
	}

	id, err := db.AddWorkflowEntity(&WorkflowEntity{
		BaseEntity: BaseEntity{
			HandlerName: "test",
			Type:        EntityWorkflow,
			QueueID:     1,
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
