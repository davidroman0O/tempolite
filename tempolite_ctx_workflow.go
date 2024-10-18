package tempolite

import (
	"fmt"
	"log"

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
		return fmt.Errorf("error fetching workflow: %w", err)
	}
	if workflow.IsPaused {
		return errWorkflowPaused
	}
	return nil
}

func (w WorkflowContext[T]) setExecutionAsPaused() error {
	_, err := w.tp.client.Workflow.UpdateOneID(w.workflowID).SetStatus(workflow.StatusPaused).Save(w.tp.ctx)
	return err
}

func (w WorkflowContext[T]) GetVersion(changeID string, minSupported, maxSupported int) int {
	log.Printf("GetVersion called for workflowType: %s, workflowID: %s, changeID: %s, min: %d, max: %d", w.workflowType, w.workflowID, changeID, minSupported, maxSupported)
	version, err := w.tp.getOrCreateVersion(w.workflowType, w.workflowID, changeID, minSupported, maxSupported)
	if err != nil {
		log.Printf("Error getting version for workflowType %s, changeID %s: %v", w.workflowType, changeID, err)
		return minSupported
	}
	log.Printf("GetVersion returned %d for workflowType: %s, changeID: %s", version, w.workflowType, changeID)
	return version
}

func (w WorkflowContext[T]) ContinueAsNew(ctx WorkflowContext[T], values ...any) error {
	// todo: implement
	return nil
}

func (w WorkflowContext[T]) GetWorkflow(id WorkflowExecutionID) *WorkflowExecutionInfo[T] {
	return w.tp.getWorkflowExecution(w, id, nil)
}

func (w WorkflowContext[T]) SideEffect(stepID T, handler interface{}) *SideEffectInfo[T] {
	if err := w.checkIfPaused(); err != nil {
		if uerr := w.setExecutionAsPaused(); uerr != nil {
			log.Printf("Error setting workflow as paused: %v", uerr)
			return &SideEffectInfo[T]{err: uerr}
		}
		return &SideEffectInfo[T]{err: err}
	}
	id, err := w.tp.enqueueSubSideEffect(w, stepID, handler)
	if err != nil {
		log.Printf("Error enqueuing side effect: %v", err)
	}
	return w.tp.getSideEffect(w, id, err)
}

func (w WorkflowContext[T]) Workflow(stepID T, handler interface{}, inputs ...any) *WorkflowInfo[T] {
	if err := w.checkIfPaused(); err != nil {
		if uerr := w.setExecutionAsPaused(); uerr != nil {
			log.Printf("Error setting workflow as paused: %v", uerr)
			return &WorkflowInfo[T]{err: uerr}
		}
		return &WorkflowInfo[T]{err: err}
	}
	id, err := w.tp.enqueueSubWorkflow(w, stepID, handler, inputs...)
	return w.tp.getWorkflow(w, id, err)
}

func (w WorkflowContext[T]) ActivityFunc(stepID T, handler interface{}, inputs ...any) *ActivityInfo[T] {
	if err := w.checkIfPaused(); err != nil {
		if uerr := w.setExecutionAsPaused(); uerr != nil {
			log.Printf("Error setting workflow as paused: %v", uerr)
			return &ActivityInfo[T]{err: uerr}
		}
		return &ActivityInfo[T]{err: err}
	}
	id, err := w.tp.enqueueSubActivtyFunc(w, stepID, handler, inputs...)
	if err != nil {
		log.Printf("Error enqueuing activity execution: %v", err)
	}
	// fmt.Println("\t \t activity execution id", id, err)
	return w.tp.getActivity(w, id, err)
}

func (w WorkflowContext[T]) ExecuteActivity(stepID T, name HandlerIdentity, inputs ...any) *ActivityInfo[T] {
	if err := w.checkIfPaused(); err != nil {
		if uerr := w.setExecutionAsPaused(); uerr != nil {
			log.Printf("Error setting workflow as paused: %v", uerr)
			return &ActivityInfo[T]{err: uerr}
		}
		return &ActivityInfo[T]{err: err}
	}
	id, err := w.tp.enqueueSubActivityExecution(w, stepID, name, inputs...)
	if err != nil {
		log.Printf("Error enqueuing activity execution: %v", err)
	}
	// fmt.Println("\t \t activity execution id", id, err)
	return w.tp.getActivity(w, id, err)
}
