package tempolite

import (
	"fmt"

	"github.com/davidroman0O/go-tempolite/ent/workflow"
)

type WorkflowContext[T Identifier] struct {
	TempoliteContext
	tp           *Tempolite[T]
	workflowID   string
	executionID  string
	runID        string
	workflowType string
	stepID       string
}

func (w WorkflowContext[T]) ContinueAsNew(ctx WorkflowContext[T], values ...any) error {
	// todo: implement continue as new
	return nil
}

func (w WorkflowContext[T]) RunID() string {
	return w.runID
}

func (w WorkflowContext[T]) EntityID() string {
	return w.workflowID
}

func (w WorkflowContext[T]) ExecutionID() string {
	return w.executionID
}

func (w WorkflowContext[T]) StepID() string {
	return w.stepID
}

func (w WorkflowContext[T]) EntityType() string {
	return "workflow"
}

func (w WorkflowContext[T]) checkIfPaused() error {
	workflow, err := w.tp.client.Workflow.Get(w.tp.ctx, w.workflowID)
	if err != nil {
		w.tp.logger.Error(w.tp.ctx, "Error fetching workflow", "workflowID", w.workflowID, "error", err)
		return fmt.Errorf("error fetching workflow: %w", err)
	}
	if workflow.IsPaused {
		return errWorkflowPaused
	}
	return nil
}

func (w WorkflowContext[T]) setExecutionAsPaused() error {
	_, err := w.tp.client.Workflow.UpdateOneID(w.workflowID).SetStatus(workflow.StatusPaused).Save(w.tp.ctx)
	if err != nil {
		w.tp.logger.Error(w.tp.ctx, "Error setting workflow as paused", "workflowID", w.workflowID, "error", err)
	}
	return err
}

func (w WorkflowContext[T]) GetVersion(changeID string, minSupported, maxSupported int) int {
	w.tp.logger.Debug(w.tp.ctx, "GetVersion", "workflowType", w.workflowType, "workflowID", w.workflowID, "changeID", changeID, "minSupported", minSupported, "maxSupported", maxSupported)
	version, err := w.tp.getOrCreateVersion(w.workflowType, w.workflowID, changeID, minSupported, maxSupported)
	if err != nil {
		w.tp.logger.Error(w.tp.ctx, "Error getting version", "workflowType", w.workflowType, "workflowID", w.workflowID, "changeID", changeID, "error", err)
		return minSupported
	}
	w.tp.logger.Debug(w.tp.ctx, "GetVersion", "workflowType", w.workflowType, "workflowID", w.workflowID, "changeID", changeID, "version", version)
	return version
}

func (w WorkflowContext[T]) GetWorkflow(id WorkflowExecutionID) *WorkflowExecutionInfo[T] {
	return w.tp.getWorkflowExecution(w, id, nil)
}

func (w WorkflowContext[T]) SideEffect(stepID T, handler interface{}) *SideEffectInfo[T] {
	if err := w.checkIfPaused(); err != nil {
		if uerr := w.setExecutionAsPaused(); uerr != nil {
			w.tp.logger.Error(w.tp.ctx, "Error setting workflow as paused", "error", uerr)
			return &SideEffectInfo[T]{err: uerr}
		}
		return &SideEffectInfo[T]{err: err}
	}
	id, err := w.tp.enqueueSideEffect(w, stepID, handler)
	if err != nil {
		w.tp.logger.Error(w.tp.ctx, "Error enqueuing side effect", "error", err)
	}
	return w.tp.getSideEffect(w, id, err)
}

func (w WorkflowContext[T]) Workflow(stepID T, handler interface{}, inputs ...any) *WorkflowInfo[T] {
	if err := w.checkIfPaused(); err != nil {
		if uerr := w.setExecutionAsPaused(); uerr != nil {
			w.tp.logger.Error(w.tp.ctx, "Error setting workflow as paused", "error", uerr)
			return &WorkflowInfo[T]{err: uerr}
		}
		return &WorkflowInfo[T]{err: err}
	}
	id, err := w.tp.enqueueWorkflow(w, stepID, handler, inputs...)
	if err != nil {
		w.tp.logger.Error(w.tp.ctx, "Error enqueuing workflow", "error", err)
	}
	return w.tp.getWorkflow(w, id, err)
}

func (w WorkflowContext[T]) ActivityFunc(stepID T, handler interface{}, inputs ...any) *ActivityInfo[T] {
	if err := w.checkIfPaused(); err != nil {
		if uerr := w.setExecutionAsPaused(); uerr != nil {
			w.tp.logger.Error(w.tp.ctx, "Error setting workflow as paused", "error", uerr)
			return &ActivityInfo[T]{err: uerr}
		}
		return &ActivityInfo[T]{err: err}
	}
	id, err := w.tp.enqueueActivityFunc(w, stepID, handler, inputs...)
	if err != nil {
		w.tp.logger.Error(w.tp.ctx, "Error enqueuing activity function", "error", err)
	}
	return w.tp.getActivity(w, id, err)
}

func (w WorkflowContext[T]) ExecuteActivity(stepID T, name HandlerIdentity, inputs ...any) *ActivityInfo[T] {
	if err := w.checkIfPaused(); err != nil {
		if uerr := w.setExecutionAsPaused(); uerr != nil {
			w.tp.logger.Error(w.tp.ctx, "Error setting workflow as paused", "error", uerr)
			return &ActivityInfo[T]{err: uerr}
		}
		return &ActivityInfo[T]{err: err}
	}
	id, err := w.tp.enqueueActivity(w, stepID, name, inputs...)
	if err != nil {
		w.tp.logger.Error(w.tp.ctx, "Error enqueuing activity", "error", err)
	}
	return w.tp.getActivity(w, id, err)
}
