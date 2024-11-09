package info

import (
	"fmt"
	"reflect"

	"github.com/davidroman0O/tempolite/internal/engine/io"
	"github.com/davidroman0O/tempolite/internal/persistence/ent"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/entity"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/execution"
	"github.com/davidroman0O/tempolite/internal/persistence/repository"
	"github.com/davidroman0O/tempolite/internal/types"
)

type workflowGetResponse struct {
	outputs [][]byte
	err     error
}

type WorkflowInfo struct {
	db          repository.Repository
	handler     types.HandlerInfo
	entityID    types.WorkflowID
	responseChn chan workflowGetResponse
	done        func()
}

func NewWorkflowInfo(id types.WorkflowID, handler types.HandlerInfo, db repository.Repository, info *InfoClock) *WorkflowInfo {
	wi := &WorkflowInfo{
		db:          db,
		handler:     handler,
		entityID:    id,
		responseChn: make(chan workflowGetResponse, 1),
	}

	info.AddInfo(wi.entityID.ID(), wi)

	wi.done = func() {
		<-info.Remove(wi.entityID.ID())
		close(wi.responseChn)
	}

	return wi
}

func NewWorkflowInfoWithError(err error) *WorkflowInfo {
	wi := &WorkflowInfo{
		responseChn: make(chan workflowGetResponse, 1),
	}
	wi.responseChn <- workflowGetResponse{
		err: err,
	}
	close(wi.responseChn)
	return wi
}

func (w *WorkflowInfo) Tick() error {

	tx, err := w.db.Tx()
	if err != nil {
		return err
	}

	entityObj, err := w.db.Workflows().Get(tx, int(w.entityID))
	if err != nil {
		if ent.IsNotFound(err) {
			return err
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	switch entity.Status(entityObj.Status) {
	case entity.StatusCompleted, entity.StatusFailed, entity.StatusCancelled:
		switch execution.Status(entityObj.Execution.Status) {
		case execution.StatusCompleted:
			w.responseChn <- workflowGetResponse{
				outputs: entityObj.Execution.WorkflowExecutionData.Output,
			}
			w.done()
		case execution.StatusFailed:
			w.responseChn <- workflowGetResponse{
				err: fmt.Errorf(entityObj.Execution.WorkflowExecutionData.ErrMsg),
			}
			w.done()
		case execution.StatusCancelled:
			w.responseChn <- workflowGetResponse{
				err: fmt.Errorf("workflow was cancelled"),
			}
			w.done()
		case execution.StatusRetried:
			w.responseChn <- workflowGetResponse{
				err: fmt.Errorf("workflow execution was retried"),
			}
			w.done()
		default:
			return nil
		}
	default:
		return nil
	}

	return nil
}

func (w *WorkflowInfo) Get(output ...any) error {

	response := <-w.responseChn

	if response.err != nil {
		fmt.Println("error", response.err)
		return response.err
	}

	realOutputs, err := io.ConvertOutputsFromSerialization(w.handler, response.outputs)
	if err != nil {
		return err
	}

	for idx, outPtr := range output {
		outVal := reflect.ValueOf(outPtr).Elem()
		outputVal := reflect.ValueOf(realOutputs[idx])

		if outVal.Type() != outputVal.Type() {
			return fmt.Errorf("type mismatch at index %d: expected %v, got %v", idx, outVal.Type(), outputVal.Type())
		}

		outVal.Set(outputVal)
	}

	return nil
}
