package info

import (
	"context"
	"fmt"
	"reflect"

	"github.com/davidroman0O/tempolite/internal/engine/io"
	"github.com/davidroman0O/tempolite/internal/persistence/ent"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/entity"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/execution"
	"github.com/davidroman0O/tempolite/internal/persistence/repository"
	"github.com/davidroman0O/tempolite/internal/types"
	"github.com/davidroman0O/tempolite/pkg/logs"
)

type workflowGetResponse struct {
	outputs [][]byte
	err     error
}

type WorkflowInfo struct {
	ctx         context.Context
	db          repository.Repository
	handler     types.HandlerInfo
	entityID    types.WorkflowID
	responseChn chan workflowGetResponse
	done        func()
}

func NewWorkflowInfo(ctx context.Context, id types.WorkflowID, handler types.HandlerInfo, db repository.Repository, info *InfoClock) *WorkflowInfo {
	wi := &WorkflowInfo{
		ctx:         ctx,
		db:          db,
		handler:     handler,
		entityID:    id,
		responseChn: make(chan workflowGetResponse, 1),
	}

	logs.Debug(ctx, "Adding workflow info", "handlerName", handler.HandlerName, "workflowID", id)
	info.AddInfo(wi.entityID.ID(), wi)

	wi.done = func() {
		logs.Debug(ctx, "Removing workflow info", "handlerName", handler.HandlerName, "workflowID", id)
		<-info.Remove(wi.entityID.ID())
		close(wi.responseChn)
	}

	return wi
}

func NewWorkflowInfoWithError(ctx context.Context, err error) *WorkflowInfo {
	wi := &WorkflowInfo{
		ctx:         ctx,
		responseChn: make(chan workflowGetResponse, 1),
	}
	logs.Debug(ctx, "Adding workflow info with error", "error", err)
	wi.responseChn <- workflowGetResponse{
		err: err,
	}
	close(wi.responseChn)
	return wi
}

func (w *WorkflowInfo) Tick() error {

	tx, err := w.db.Tx()
	if err != nil {
		logs.Error(w.ctx, "WorkflowInfo Tick error creating transaction", "error", err, "handlerName", w.handler.HandlerName)
		return err
	}

	entityObj, err := w.db.Workflows().Get(tx, int(w.entityID))
	if err != nil {
		if ent.IsNotFound(err) {
			logs.Error(w.ctx, "WorkflowInfo Tick workflow not found", "workflowID", w.entityID, "error", err, "handlerName", w.handler.HandlerName)
			return err
		}
		logs.Error(w.ctx, "WorkflowInfo Tick error getting workflow", "workflowID", w.entityID, "error", err, "handlerName", w.handler.HandlerName)
		return err
	}

	if err := tx.Commit(); err != nil {
		logs.Error(w.ctx, "WorkflowInfo Tick error committing transaction", "error", err, "handlerName", w.handler.HandlerName)
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
			logs.Debug(w.ctx, "WorkflowInfo Tick workflow completed", "handlerName", w.handler.HandlerName, "workflowID", w.entityID)
		case execution.StatusFailed:
			w.responseChn <- workflowGetResponse{
				err: fmt.Errorf(entityObj.Execution.WorkflowExecutionData.ErrMsg),
			}
			w.done()
			logs.Debug(w.ctx, "WorkflowInfo Tick workflow failed", "handlerName", w.handler.HandlerName, "workflowID", w.entityID)
		case execution.StatusCancelled:
			w.responseChn <- workflowGetResponse{
				err: fmt.Errorf("workflow was cancelled"),
			}
			w.done()
			logs.Debug(w.ctx, "WorkflowInfo Tick workflow cancelled", "handlerName", w.handler.HandlerName, "workflowID", w.entityID)
		case execution.StatusRetried:
			w.responseChn <- workflowGetResponse{
				err: fmt.Errorf("workflow execution was retried"),
			}
			w.done()
			logs.Debug(w.ctx, "WorkflowInfo Tick workflow retried", "handlerName", w.handler.HandlerName, "workflowID", w.entityID)
		default:
			return nil
		}
	default:
		return nil
	}

	return nil
}

func (w *WorkflowInfo) Get(output ...any) error {

	logs.Debug(w.ctx, "WorkflowInfo Get waiting for response", "handlerName", w.handler.HandlerName, "workflowID", w.entityID)
	response := <-w.responseChn
	logs.Debug(w.ctx, "WorkflowInfo Get received response", "handlerName", w.handler.HandlerName, "workflowID", w.entityID)

	if response.err != nil {
		logs.Error(w.ctx, "WorkflowInfo Get error getting response", "handlerName", w.handler.HandlerName, "workflowID", w.entityID, "error", response.err)
		return response.err
	}

	realOutputs, err := io.ConvertOutputsFromSerialization(w.handler, response.outputs)
	if err != nil {
		logs.Error(w.ctx, "WorkflowInfo Get error converting outputs", "handlerName", w.handler.HandlerName, "workflowID", w.entityID, "error", err)
		return err
	}

	for idx, outPtr := range output {
		outVal := reflect.ValueOf(outPtr).Elem()
		outputVal := reflect.ValueOf(realOutputs[idx])

		if outVal.Type() != outputVal.Type() {
			logs.Error(w.ctx, "WorkflowInfo Get type mismatch", "handlerName", w.handler.HandlerName, "workflowID", w.entityID, "index", idx, "expected", outVal.Type(), "got", outputVal.Type())
			return fmt.Errorf("type mismatch at index %d: expected %v, got %v", idx, outVal.Type(), outputVal.Type())
		}

		outVal.Set(outputVal)
	}

	logs.Debug(w.ctx, "WorkflowInfo Get outputs set", "handlerName", w.handler.HandlerName, "workflowID", w.entityID)

	return nil
}
