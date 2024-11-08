package context

import (
	"github.com/davidroman0O/tempolite/internal/persistence/repository"
	"github.com/davidroman0O/tempolite/internal/types"
)

type WorkflowInfo struct {
	db       repository.Repository
	entityID types.WorkflowID
	err      error
}

func NewWorkflowInfo(id types.WorkflowID, db repository.Repository) *WorkflowInfo {
	return &WorkflowInfo{
		db:       db,
		entityID: id,
	}
}

func NewWorkflowInfoWithError(err error) *WorkflowInfo {
	return &WorkflowInfo{
		err: err,
	}
}

func (w *WorkflowInfo) Get(output ...any) error {
	if w.err != nil {
		return w.err
	}
	return nil
}
