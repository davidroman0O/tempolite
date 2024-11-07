package context

import "github.com/davidroman0O/tempolite/internal/types"

type WorkflowInfo struct {
	EntityID types.WorkflowID
	err      error
}

func NewWorkflowInfo(id types.WorkflowID) *WorkflowInfo {
	return &WorkflowInfo{
		EntityID: id,
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
