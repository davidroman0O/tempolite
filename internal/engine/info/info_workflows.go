package info

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/davidroman0O/retrypool"
	"github.com/davidroman0O/tempolite/internal/clock"
	"github.com/davidroman0O/tempolite/internal/engine/io"
	"github.com/davidroman0O/tempolite/internal/persistence/ent"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/entity"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/execution"
	"github.com/davidroman0O/tempolite/internal/persistence/repository"
	"github.com/davidroman0O/tempolite/internal/types"
	"github.com/davidroman0O/tempolite/pkg/logs"
)

type WorkflowInfo struct {
	ctx            context.Context
	cancel         context.CancelFunc
	db             repository.Repository
	handler        types.HandlerInfo
	entityID       types.WorkflowID
	requestReponse *retrypool.RequestResponse[struct{}, [][]byte]
	done           func()

	c        *clock.Clock // todo: remove
	tickerID interface{}
}

func NewWorkflowInfo(ctx context.Context, id types.WorkflowID, handler types.HandlerInfo, db repository.Repository, c *clock.Clock) *WorkflowInfo {
	wi := &WorkflowInfo{
		db:             db,
		handler:        handler,
		entityID:       id,
		c:              c,
		requestReponse: retrypool.NewRequestResponse[struct{}, [][]byte](struct{}{}),
	}
	wi.ctx, wi.cancel = context.WithCancel(ctx)
	wi.tickerID = fmt.Sprintf("workflow-%v", wi.entityID.ID())

	if !wi.c.HasTickerID(wi.tickerID) {
		wi.prepareClock()
	}

	return wi
}

func (wi *WorkflowInfo) prepareClock() {

	wi.done = func() {
		logs.Debug(wi.ctx, "Removing workflow info", "handlerName", wi.handler.HandlerName, "workflowID", wi.entityID)
		wi.cancel()
		wi.c.RemoveTicker(wi.tickerID) // should trigger the clean up
		logs.Debug(wi.ctx, "Removed workflow info", "handlerName", wi.handler.HandlerName, "workflowID", wi.entityID)
	}

	logs.Debug(wi.ctx, "Adding workflow info", "handlerName", wi.handler.HandlerName, "workflowID", wi.entityID)

	wi.c.AddTicker(
		wi.tickerID,
		wi,
		clock.WithCleanup(func() {
			// Rule: we might be a sub-workflow, during a shutdown, our parent was cancelled but we are still running trying to find the answer too because we were triggered
			// it's time for us to sleep.
			// fmt.Println("REMOVE WORKFLOW INFO", wi.entityID.ID())
			wi.cancel()
		}),
	)
}

func NewWorkflowInfoWithError(ctx context.Context, err error) *WorkflowInfo {
	wi := &WorkflowInfo{
		ctx:            ctx,
		requestReponse: retrypool.NewRequestResponse[struct{}, [][]byte](struct{}{}),
	}
	logs.Debug(ctx, "Adding workflow info with error", "error", err)
	wi.requestReponse.CompleteWithError(err)
	return wi
}

func (w *WorkflowInfo) WorkflowID() types.WorkflowID {
	return w.entityID
}

func (w *WorkflowInfo) Tick(ctx context.Context) error {

	tx, err := w.db.Tx()
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil
		}
		logs.Error(w.ctx, "WorkflowInfo Tick error creating transaction", "error", err, "handlerName", w.handler.HandlerName)
		return err
	}

	if ctx.Err() != nil {
		logs.Debug(w.ctx, "WorkflowInfo Tick context canceled", "error", ctx.Err(), "handlerName", w.handler.HandlerName)
		return ctx.Err()
	}

	// logs.Debug(w.ctx, "WorkflowInfo Tick getting workflow", "handlerName", w.handler.HandlerName, "workflowID", w.entityID)
	entityObj, err := w.db.Workflows().Get(tx, w.entityID.ID())
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return nil
		}
		if ent.IsNotFound(err) {
			logs.Error(w.ctx, "WorkflowInfo Tick workflow not found", "workflowID", w.entityID, "error", err, "handlerName", w.handler.HandlerName)
			return err
		}
		logs.Error(w.ctx, "WorkflowInfo Tick error getting workflow", "workflowID", w.entityID, "error", err, "handlerName", w.handler.HandlerName)
		return err
	}

	if ctx.Err() != nil {
		logs.Debug(w.ctx, "WorkflowInfo Tick context canceled", "error", ctx.Err(), "handlerName", w.handler.HandlerName)
		tx.Rollback()
		return ctx.Err()
	}

	if err := tx.Commit(); err != nil {
		if errors.Is(err, context.Canceled) {
			return nil
		}
		logs.Error(w.ctx, "WorkflowInfo Tick error committing transaction", "error", err, "handlerName", w.handler.HandlerName)
		return err
	}

	logs.Debug(w.ctx, "WorkflowInfo Tick workflow found", "handlerName", w.handler.HandlerName, "workflowID", w.entityID, "executionID", entityObj.Execution.ID, "status", entityObj.Status, "executionStatus", entityObj.Execution.Status)

	switch entity.Status(entityObj.Status) {
	case entity.StatusCompleted, entity.StatusFailed, entity.StatusCancelled:
		switch execution.Status(entityObj.Execution.Status) {
		case execution.StatusCompleted:
			fmt.Println("OUTPUT COMPLETED", entityObj.Execution.WorkflowExecutionData.Output)
			w.requestReponse.Complete(entityObj.Execution.WorkflowExecutionData.Output)
			w.done()
			logs.Debug(w.ctx, "WorkflowInfo Tick workflow completed", "handlerName", w.handler.HandlerName, "workflowID", w.entityID)
		case execution.StatusFailed:
			w.requestReponse.CompleteWithError(fmt.Errorf(entityObj.Execution.WorkflowExecutionData.ErrMsg))
			w.done()
			logs.Debug(w.ctx, "WorkflowInfo Tick workflow failed", "handlerName", w.handler.HandlerName, "workflowID", w.entityID)
		case execution.StatusCancelled:
			w.requestReponse.CompleteWithError(fmt.Errorf("workflow was cancelled"))
			w.done()
			logs.Debug(w.ctx, "WorkflowInfo Tick workflow cancelled", "handlerName", w.handler.HandlerName, "workflowID", w.entityID)
		case execution.StatusRetried:
			w.requestReponse.CompleteWithError(fmt.Errorf("workflow execution was retried"))
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

	logs.Debug(w.ctx, "WorkflowInfo Get asked to listen to", "workflowID", w.entityID.ID(), "handlerName", w.handler.HandlerName, "workflowID", w.entityID)

	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case <-w.ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				logs.Debug(w.ctx, "WorkflowInfo Get waiting for response", "workflowID", w.entityID.ID(), "handlerName", w.handler.HandlerName, "workflowID", w.entityID, "ctx", w.ctx.Err())
				// tickers := w.c.ListSubscribers()
				// for idx, v := range tickers {
				// 	logs.Debug(w.ctx, "WorkflowInfo List Tickers ", "idx", idx, "TickerID", v)
				// }
			}
		}
	}()

	// avoid being blocked
	select {
	case <-w.ctx.Done():
		fmt.Println("CONTEXT DONE", w.entityID)
		ticker.Stop()
		return w.ctx.Err()
	case <-w.requestReponse.Done():
		ticker.Stop()
	}

	logs.Debug(w.ctx, "WorkflowInfo Get received response", "workflowID", w.entityID.ID(), "handlerName", w.handler.HandlerName, "workflowID", w.entityID)

	var ouputs [][]byte = [][]byte{}
	var err error

	if ouputs, err = w.requestReponse.Wait(w.ctx); err != nil {
		logs.Error(w.ctx, "WorkflowInfo Get error getting response", "workflowID", w.entityID.ID(), "handlerName", w.handler.HandlerName, "workflowID", w.entityID, "error", err)
		return err
	}

	fmt.Println("OUTPUTS", ouputs)
	realOutputs, err := io.ConvertOutputsFromSerialization(w.handler, ouputs)
	if err != nil {
		logs.Error(w.ctx, "WorkflowInfo Get error converting outputs", "workflowID", w.entityID.ID(), "handlerName", w.handler.HandlerName, "workflowID", w.entityID, "error", err)
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

	logs.Debug(w.ctx, "WorkflowInfo Get outputs set", "workflowID", w.entityID.ID(), "handlerName", w.handler.HandlerName, "workflowID", w.entityID)

	return nil
}
