package tempolite

import (
	"fmt"

	"github.com/davidroman0O/tempolite/ent"
	"github.com/davidroman0O/tempolite/ent/workflow"
	"github.com/davidroman0O/tempolite/ent/workflowexecution"
)

type WorkflowContext struct {
	TempoliteContext
	tp              *Tempolite
	workflowID      string
	executionID     string
	runID           string
	workflowType    string
	stepID          string
	handlerIdentity HandlerIdentity
	queueName       string
}

func (w WorkflowContext) QueueName() string {
	if w.queueName == "" {
		return "default"
	}
	return w.queueName
}

func (w WorkflowContext) ContinueAsNew(ctx WorkflowContext, stepID string, options tempoliteWorkflowOptions, values ...any) error {
	// Check if the workflow is paused
	if err := w.checkIfPaused(); err != nil {
		if uerr := w.setExecutionAsPaused(); uerr != nil {
			w.tp.logger.Error(w.tp.ctx, "Error setting workflow as paused", "error", uerr)
			return uerr
		}
		return err
	}

	executionCurrent, err := w.tp.client.WorkflowExecution.Get(w.tp.ctx, w.executionID)
	if err != nil {
		w.tp.logger.Error(w.tp.ctx, "Error fetching current execution", "error", err)
		return fmt.Errorf("failed to fetch current execution: %w", err)
	}

	// If you replay a workflow, we stop right here
	if executionCurrent.IsReplay {
		return nil
	}

	if value, ok := w.tp.workflows.Load(w.handlerIdentity); ok {
		handler := value.(Workflow)
		newWorkflowID, err := w.tp.enqueueWorkflow(w, stepID, handler.Handler, options, values...)
		if err != nil {
			w.tp.logger.Error(w.tp.ctx, "Error enqueuing workflow", "error", err)
			return fmt.Errorf("failed to enqueue new workflow: %w", err)
		}

		// Start a transaction
		tx, err := w.tp.client.Tx(w.tp.ctx)
		if err != nil {
			w.tp.logger.Error(w.tp.ctx, "Error starting transaction", "error", err)
			return fmt.Errorf("failed to start transaction: %w", err)
		}
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		// Update the new workflow with the continuation relationship
		if err := tx.Workflow.UpdateOneID(string(newWorkflowID)).
			SetContinuedFromID(w.workflowID).
			Exec(w.tp.ctx); err != nil {
			tx.Rollback()
			w.tp.logger.Error(w.tp.ctx, "Error updating new workflow", "error", err)
			return fmt.Errorf("failed to update new workflow: %w", err)
		}

		// Commit the transaction
		if err := tx.Commit(); err != nil {
			w.tp.logger.Error(w.tp.ctx, "Error committing transaction", "error", err)
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		w.tp.logger.Info(w.tp.ctx, "Workflow continued as new", "oldWorkflowID", w.workflowID, "newWorkflowID", newWorkflowID)
		return nil
	}

	w.tp.logger.Error(w.tp.ctx, "Failed to continue as new: workflow not found", "handlerIdentity", w.handlerIdentity)
	return fmt.Errorf("failed to continue as new: workflow not found")
}

func (w WorkflowContext) RunID() string {
	return w.runID
}

func (w WorkflowContext) EntityID() string {
	return w.workflowID
}

func (w WorkflowContext) ExecutionID() string {
	return w.executionID
}

func (w WorkflowContext) StepID() string {
	return w.stepID
}

func (w WorkflowContext) EntityType() string {
	return "workflow"
}

func (w WorkflowContext) checkIfPaused() error {
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

func (w WorkflowContext) setExecutionAsPaused() error {
	_, err := w.tp.client.Workflow.UpdateOneID(w.workflowID).SetStatus(workflow.StatusPaused).Save(w.tp.ctx)
	if err != nil {
		w.tp.logger.Error(w.tp.ctx, "Error setting workflow as paused", "workflowID", w.workflowID, "error", err)
	}

	// Get latest workflow execution
	wfExec, err := w.tp.client.WorkflowExecution.Query().
		Where(workflowexecution.HasWorkflowWith(workflow.ID(w.workflowID))).
		Order(ent.Desc(workflowexecution.FieldStartedAt)).
		First(w.tp.ctx)
	if err != nil {
		w.tp.logger.Error(w.tp.ctx, "Error fetching latest workflow execution", "workflowID", w.workflowID, "error", err)
		return err
	}

	// Set it to pause too
	_, err = w.tp.client.WorkflowExecution.UpdateOneID(wfExec.ID).
		SetStatus(workflowexecution.StatusPaused).
		Save(w.tp.ctx)
	if err != nil {
		w.tp.logger.Error(w.tp.ctx, "Error pausing latest workflow execution", "workflowID", w.workflowID, "workflowExecutionID", wfExec.ID, "error", err)
		// if rerr := tx.Rollback(); rerr != nil {
		// 	w.tp.logger.Error(w.tp.ctx, "Error rolling back transaction", "error", rerr)
		// }
		return err
	}

	return err
}

func (w WorkflowContext) GetVersion(changeID string, minSupported, maxSupported int) int {
	w.tp.logger.Debug(w.tp.ctx, "GetVersion", "workflowType", w.workflowType, "workflowID", w.workflowID, "changeID", changeID, "minSupported", minSupported, "maxSupported", maxSupported)
	version, err := w.tp.getOrCreateVersion(w.workflowType, w.workflowID, changeID, minSupported, maxSupported)
	if err != nil {
		w.tp.logger.Error(w.tp.ctx, "Error getting version", "workflowType", w.workflowType, "workflowID", w.workflowID, "changeID", changeID, "error", err)
		return minSupported
	}
	w.tp.logger.Debug(w.tp.ctx, "GetVersion", "workflowType", w.workflowType, "workflowID", w.workflowID, "changeID", changeID, "version", version)
	return version
}

func (w WorkflowContext) GetWorkflow(id WorkflowExecutionID) *WorkflowExecutionInfo {
	return w.tp.getWorkflowExecution(w, id, nil)
}

func (w WorkflowContext) SideEffect(stepID string, handler interface{}) *SideEffectInfo {
	if err := w.checkIfPaused(); err != nil {
		if uerr := w.setExecutionAsPaused(); uerr != nil {
			w.tp.logger.Error(w.tp.ctx, "Error setting workflow as paused", "error", uerr)
			return &SideEffectInfo{tp: w.tp, err: uerr}
		}
		return &SideEffectInfo{tp: w.tp, err: err}
	}
	id, err := w.tp.enqueueSideEffect(w, stepID, handler)
	if err != nil {
		w.tp.logger.Error(w.tp.ctx, "Error enqueuing side effect", "error", err)
	}
	return w.tp.getSideEffect(w, id, err)
}

// Original Activity with queue inheritance
func (w WorkflowContext) Activity(stepID string, handler interface{}, options tempoliteActivityOptions, inputs ...any) *ActivityInfo {
	if err := w.checkIfPaused(); err != nil {
		if uerr := w.setExecutionAsPaused(); uerr != nil {
			w.tp.logger.Error(w.tp.ctx, "Error setting workflow as paused", "error", uerr)
			return &ActivityInfo{tp: w.tp, err: uerr}
		}
		return &ActivityInfo{tp: w.tp, err: err}
	}
	// Use parent workflow's queue
	id, err := w.tp.enqueueActivityFunc(w, stepID, handler, options, inputs...)
	if err != nil {
		w.tp.logger.Error(w.tp.ctx, "Error enqueuing activity function", "error", err)
	}
	return w.tp.getActivity(w, id, err)
}

// Original Workflow with queue inheritance
func (w WorkflowContext) Workflow(stepID string, handler interface{}, options tempoliteWorkflowOptions, inputs ...any) *WorkflowInfo {
	if err := w.checkIfPaused(); err != nil {
		if uerr := w.setExecutionAsPaused(); uerr != nil {
			w.tp.logger.Error(w.tp.ctx, "Error setting workflow as paused", "error", uerr)
			return &WorkflowInfo{tp: w.tp, err: uerr}
		}
		return &WorkflowInfo{tp: w.tp, err: err}
	}
	// Use parent workflow's queue by default
	id, err := w.tp.enqueueWorkflow(w, stepID, handler, options, inputs...)
	if err != nil {
		w.tp.logger.Error(w.tp.ctx, "Error enqueuing workflow", "error", err)
	}
	return w.tp.getWorkflow(w, id, err)
}

// // Optional queue override for workflows
// func (w WorkflowContext) WorkflowInQueue(stepID string, handler interface{}, inputs ...any) *WorkflowInfo {
// 	if err := w.checkIfPaused(); err != nil {
// 		if uerr := w.setExecutionAsPaused(); uerr != nil {
// 			w.tp.logger.Error(w.tp.ctx, "Error setting workflow as paused", "error", uerr)
// 			return &WorkflowInfo{ tp:w.tp, err: uerr}
// 		}
// 		return &WorkflowInfo{tp: w.tp, err: err}
// 	}
// 	// Create new context with specified queue
// 	queueCtx := w
// 	queueCtx.queueName = queueName
// 	id, err := w.tp.enqueueWorkflow(queueCtx, stepID, handler, inputs...)
// 	if err != nil {
// 		w.tp.logger.Error(w.tp.ctx, "Error enqueuing workflow", "error", err)
// 	}
// 	return w.tp.getWorkflow(w, id, err)
// }
