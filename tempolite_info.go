package tempolite

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"runtime"
	"time"

	"github.com/davidroman0O/go-tempolite/ent"
	"github.com/davidroman0O/go-tempolite/ent/activity"
	"github.com/davidroman0O/go-tempolite/ent/activityexecution"
	"github.com/davidroman0O/go-tempolite/ent/saga"
	"github.com/davidroman0O/go-tempolite/ent/sideeffect"
	"github.com/davidroman0O/go-tempolite/ent/sideeffectexecution"
	"github.com/davidroman0O/go-tempolite/ent/workflow"
	"github.com/davidroman0O/go-tempolite/ent/workflowexecution"
)

type WorkflowInfo[T Identifier] struct {
	tp         *Tempolite[T]
	WorkflowID WorkflowID
	err        error
}

// Try to find the latest workflow execution until it reaches a final state
func (i *WorkflowInfo[T]) Get(output ...interface{}) error {
	defer func() {
		log.Printf("WorkflowInfo.Get: %v %v", i.WorkflowID, i.err)
	}()
	if i.err != nil {
		return i.err
	}
	// Check if all output parameters are pointers
	for idx, out := range output {
		if reflect.TypeOf(out).Kind() != reflect.Ptr {
			return fmt.Errorf("output parameter at index %d is not a pointer", idx)
		}
	}
	ticker := time.NewTicker(time.Second / 16)
	defer ticker.Stop()
	var value any
	var ok bool
	var workflowHandlerInfo Workflow
	for {
		select {
		case <-i.tp.ctx.Done():
			return i.tp.ctx.Err()
		case <-ticker.C:
			workflowEntity, err := i.tp.client.Workflow.Query().Where(workflow.IDEQ(i.WorkflowID.String())).Only(i.tp.ctx)
			if err != nil {
				return err
			}
			if value, ok = i.tp.workflows.Load(HandlerIdentity(workflowEntity.Identity)); ok {
				if workflowHandlerInfo, ok = value.(Workflow); !ok {
					log.Printf("scheduler: workflow %s is not handler info", i.WorkflowID)
					return errors.New("workflow is not handler info")
				}
				switch workflowEntity.Status {
				// wait for the confirmation that the workflow entity reached a final state
				case workflow.StatusCompleted, workflow.StatusFailed, workflow.StatusCancelled:
					// Then only get the latest workflow execution
					// Simply because eventually my children can have retries and i need to let them finish
					latestExec, err := i.tp.client.WorkflowExecution.Query().
						Where(workflowexecution.HasWorkflowWith(workflow.IDEQ(i.WorkflowID.String()))).
						Order(ent.Desc(workflowexecution.FieldStartedAt)).
						First(i.tp.ctx)
					if err != nil {
						return err
					}
					switch latestExec.Status {
					case workflowexecution.StatusCompleted:
						outputs, err := i.tp.convertOuputs(HandlerInfo(workflowHandlerInfo), latestExec.Output)
						if err != nil {
							return err
						}
						if len(output) != len(outputs) {
							return fmt.Errorf("output length mismatch: expected %d, got %d", len(outputs), len(output))
						}

						for idx, outPtr := range output {
							outVal := reflect.ValueOf(outPtr).Elem()
							outputVal := reflect.ValueOf(outputs[idx])

							if outVal.Type() != outputVal.Type() {
								return fmt.Errorf("type mismatch at index %d: expected %v, got %v", idx, outVal.Type(), outputVal.Type())
							}

							outVal.Set(outputVal)
						}
						return nil
					case workflowexecution.StatusCancelled:
						return fmt.Errorf("workflow %s was cancelled", i.WorkflowID)
					case workflowexecution.StatusFailed:
						return errors.New(latestExec.Error)
					case workflowexecution.StatusRetried:
						return errors.New("workflow was retried")
					case workflowexecution.StatusPending, workflowexecution.StatusRunning:
						runtime.Gosched()
						continue
					}
				default:
					runtime.Gosched()
					continue
				}
			}
			runtime.Gosched()
		}
	}
}

type WorkflowExecutionInfo[T Identifier] struct {
	tp          *Tempolite[T]
	ExecutionID WorkflowExecutionID
	err         error
}

// Try to find the workflow execution until it reaches a final state
func (i *WorkflowExecutionInfo[T]) Get(output ...interface{}) error {

	if i.err != nil {
		return i.err
	}

	for idx, out := range output {
		if reflect.TypeOf(out).Kind() != reflect.Ptr {
			return fmt.Errorf("output parameter at index %d is not a pointer", idx)
		}
	}

	ticker := time.NewTicker(time.Second / 16)
	defer ticker.Stop()
	var value any
	var ok bool
	var workflowHandlerInfo Workflow
	for {
		select {
		case <-i.tp.ctx.Done():
			return i.tp.ctx.Err()
		case <-ticker.C:
			workflowExecEntity, err := i.tp.client.WorkflowExecution.Query().Where(workflowexecution.IDEQ(i.ExecutionID.String())).WithWorkflow().Only(i.tp.ctx)
			if err != nil {
				return err
			}
			workflowEntity, err := i.tp.client.Workflow.Query().Where(workflow.IDEQ(workflowExecEntity.Edges.Workflow.ID)).Only(i.tp.ctx)
			if err != nil {
				return err
			}
			if value, ok = i.tp.workflows.Load(HandlerIdentity(workflowEntity.Identity)); ok {
				if workflowHandlerInfo, ok = value.(Workflow); !ok {
					log.Printf("scheduler: workflow %s is not handler info", workflowExecEntity.Edges.Workflow.ID)
					return errors.New("workflow is not handler info")
				}
				switch workflowExecEntity.Status {
				case workflowexecution.StatusCompleted:
					outputs, err := i.tp.convertOuputs(HandlerInfo(workflowHandlerInfo), workflowExecEntity.Output)
					if err != nil {
						return err
					}
					if len(output) != len(outputs) {
						return fmt.Errorf("output length mismatch: expected %d, got %d", len(outputs), len(output))
					}

					for idx, outPtr := range output {
						outVal := reflect.ValueOf(outPtr).Elem()
						outputVal := reflect.ValueOf(outputs[idx])

						if outVal.Type() != outputVal.Type() {
							return fmt.Errorf("type mismatch at index %d: expected %v, got %v", idx, outVal.Type(), outputVal.Type())
						}

						outVal.Set(outputVal)
					}
					return nil
				case workflowexecution.StatusCancelled:
					return fmt.Errorf("workflow %s was cancelled", workflowExecEntity.Edges.Workflow.ID)
				case workflowexecution.StatusFailed:
					return errors.New(workflowExecEntity.Error)
				case workflowexecution.StatusRetried:
					return errors.New("workflow was retried")
				case workflowexecution.StatusPending, workflowexecution.StatusRunning:
					runtime.Gosched()
					continue
				}
			}
			runtime.Gosched()
		}
	}
}

func (i *WorkflowExecutionInfo[T]) Cancel() error {
	// todo: implement
	return nil
}

func (i *WorkflowExecutionInfo[T]) Pause() error {
	// todo: implement
	return nil
}

func (i *WorkflowExecutionInfo[T]) Resume() error {
	// todo: implement
	return nil
}

type ActivityInfo[T Identifier] struct {
	tp         *Tempolite[T]
	ActivityID ActivityID
	err        error
}

func (i *ActivityInfo[T]) Get(output ...interface{}) error {
	if i.err != nil {
		return i.err
	}
	for idx, out := range output {
		if reflect.TypeOf(out).Kind() != reflect.Ptr {
			return fmt.Errorf("output parameter at index %d is not a pointer", idx)
		}
	}

	ticker := time.NewTicker(time.Second / 16)
	defer ticker.Stop()
	var value any
	var ok bool
	var activityHandlerInfo Activity
	fmt.Println("searching id", i.ActivityID.String())
	for {
		select {
		case <-i.tp.ctx.Done():
			return i.tp.ctx.Err()
		case <-ticker.C:
			activityEntity, err := i.tp.client.Activity.Query().Where(activity.IDEQ(i.ActivityID.String())).Only(i.tp.ctx)
			if err != nil {
				fmt.Println("error searching activity", err)
				return err
			}
			if value, ok = i.tp.activities.Load(HandlerIdentity(activityEntity.Identity)); ok {
				if activityHandlerInfo, ok = value.(Activity); !ok {
					log.Printf("scheduler: activity %s is not handler info", i.ActivityID.String())
					return errors.New("activity is not handler info")
				}

				switch activityEntity.Status {
				// wait for the confirmation that the workflow entity reached a final state
				case activity.StatusCompleted, activity.StatusFailed, activity.StatusCancelled:
					fmt.Println("searching for activity execution of ", i.ActivityID.String())
					// Then only get the latest activity execution
					// Simply because eventually my children can have retries and i need to let them finish
					latestExec, err := i.tp.client.ActivityExecution.Query().
						Where(
							activityexecution.HasActivityWith(activity.IDEQ(i.ActivityID.String())),
						).
						Order(ent.Desc(activityexecution.FieldStartedAt)).
						First(i.tp.ctx)

					if err != nil {
						if ent.IsNotFound(err) {
							// Handle the case where no execution is found
							log.Printf("No execution found for activity %s", i.ActivityID)
							return fmt.Errorf("no execution found for activity %s", i.ActivityID)
						}
						log.Printf("Error querying activity execution: %v", err)
						return fmt.Errorf("error querying activity execution: %w", err)
					}

					switch latestExec.Status {
					case activityexecution.StatusCompleted:
						outputs, err := i.tp.convertOuputs(HandlerInfo(activityHandlerInfo), latestExec.Output)
						if err != nil {
							return err
						}
						if len(output) != len(outputs) {
							return fmt.Errorf("output length mismatch: expected %d, got %d", len(outputs), len(output))
						}

						for idx, outPtr := range output {
							outVal := reflect.ValueOf(outPtr).Elem()
							outputVal := reflect.ValueOf(outputs[idx])

							if outVal.Type() != outputVal.Type() {
								return fmt.Errorf("type mismatch at index %d: expected %v, got %v", idx, outVal.Type(), outputVal.Type())
							}

							outVal.Set(outputVal)
						}
						return nil
					case activityexecution.StatusFailed:
						return errors.New(latestExec.Error)
					case activityexecution.StatusRetried:
						return errors.New("activity was retried")
					case activityexecution.StatusPending, activityexecution.StatusRunning:
						// The workflow is still in progress
						// return  errors.New("workflow is still in progress")
						runtime.Gosched()
						continue
					}
				default:
					runtime.Gosched()
					continue
				}
			}
			runtime.Gosched()
		}
	}
}

type ActivityExecutionInfo[T Identifier] struct {
	tp          *Tempolite[T]
	ExecutionID ActivityExecutionID
	err         error
}

// Try to find the activity execution until it reaches a final state
func (i *ActivityExecutionInfo[T]) Get(output ...interface{}) error {
	if i.err != nil {
		return i.err
	}
	for idx, out := range output {
		if reflect.TypeOf(out).Kind() != reflect.Ptr {
			return fmt.Errorf("output parameter at index %d is not a pointer", idx)
		}
	}

	ticker := time.NewTicker(time.Second / 16)
	defer ticker.Stop()
	var value any
	var ok bool
	var activityHandlerInfo Activity
	for {
		select {
		case <-i.tp.ctx.Done():
			return i.tp.ctx.Err()
		case <-ticker.C:
			fmt.Println("searching id", i.ExecutionID.String())
			activityExecEntity, err := i.tp.client.ActivityExecution.Query().Where(activityexecution.IDEQ(i.ExecutionID.String())).WithActivity().Only(i.tp.ctx)
			if err != nil {
				return err
			}
			activityEntity, err := i.tp.client.Activity.Query().Where(activity.IDEQ(activityExecEntity.Edges.Activity.ID)).Only(i.tp.ctx)
			if err != nil {
				return err
			}
			if value, ok = i.tp.activities.Load(HandlerIdentity(activityEntity.Identity)); ok {
				if activityHandlerInfo, ok = value.(Activity); !ok {
					log.Printf("scheduler: activity %s is not handler info", activityExecEntity.Edges.Activity.ID)
					return errors.New("activity is not handler info")
				}
				switch activityExecEntity.Status {
				case activityexecution.StatusCompleted:
					outputs, err := i.tp.convertOuputs(HandlerInfo(activityHandlerInfo), activityExecEntity.Output)
					if err != nil {
						return err
					}
					if len(output) != len(outputs) {
						return fmt.Errorf("output length mismatch: expected %d, got %d", len(outputs), len(output))
					}

					for idx, outPtr := range output {
						outVal := reflect.ValueOf(outPtr).Elem()
						outputVal := reflect.ValueOf(outputs[idx])

						if outVal.Type() != outputVal.Type() {
							return fmt.Errorf("type mismatch at index %d: expected %v, got %v", idx, outVal.Type(), outputVal.Type())
						}

						outVal.Set(outputVal)
					}
					return nil
				case activityexecution.StatusFailed:
					return errors.New(activityExecEntity.Error)
				case activityexecution.StatusRetried:
					return errors.New("activity was retried")
				case activityexecution.StatusPending, activityexecution.StatusRunning:
					runtime.Gosched()
					continue
				}
			}
			runtime.Gosched()
		}
	}
}

type SideEffectInfo[T Identifier] struct {
	tp       *Tempolite[T]
	EntityID SideEffectID
	err      error
}

func (i *SideEffectInfo[T]) Get(output ...interface{}) error {
	if i.err != nil {
		return i.err
	}

	ticker := time.NewTicker(time.Second / 16)
	defer ticker.Stop()

	var value any
	var ok bool
	var sideeffectHandlerInfo SideEffect

	for {
		select {
		case <-i.tp.ctx.Done():
			return i.tp.ctx.Err()
		case <-ticker.C:
			sideEffectEntity, err := i.tp.client.SideEffect.Query().
				Where(sideeffect.IDEQ(i.EntityID.String())).
				WithExecutions().
				Only(i.tp.ctx)
			if err != nil {
				return err
			}

			if value, ok = i.tp.sideEffects.Load(sideEffectEntity.ID); ok {
				if sideeffectHandlerInfo, ok = value.(SideEffect); !ok {
					log.Printf("scheduler: sideeffect %s is not handler info", i.EntityID.String())
					return errors.New("sideeffect is not handler info")
				}

				switch sideEffectEntity.Status {
				case sideeffect.StatusCompleted:
					fmt.Println("searching for side effect execution of ", i.EntityID.String())
					latestExec, err := i.tp.client.SideEffectExecution.Query().
						Where(
							sideeffectexecution.HasSideEffect(),
						).
						Order(ent.Desc(sideeffectexecution.FieldStartedAt)).
						First(i.tp.ctx)
					if err != nil {
						if ent.IsNotFound(err) {
							// Handle the case where no execution is found
							log.Printf("No execution found for sideeffect %s", i.EntityID)
							return fmt.Errorf("no execution found for sideeffect %s", i.EntityID)
						}
						log.Printf("Error querying sideeffect execution: %v", err)
						return fmt.Errorf("error querying sideeffect execution: %w", err)
					}

					switch latestExec.Status {
					case sideeffectexecution.StatusCompleted:
						outputs, err := i.tp.convertOuputs(HandlerInfo(sideeffectHandlerInfo), latestExec.Output)
						if err != nil {
							return err
						}

						if len(output) != len(outputs) {
							return fmt.Errorf("output length mismatch: expected %d, got %d", len(outputs), len(output))
						}

						for idx, outPtr := range output {
							outVal := reflect.ValueOf(outPtr).Elem()
							outputVal := reflect.ValueOf(outputs[idx])

							if outVal.Type() != outputVal.Type() {
								return fmt.Errorf("type mismatch at index %d: expected %v, got %v", idx, outVal.Type(), outputVal.Type())
							}

							outVal.Set(outputVal)
						}
						return nil
					case sideeffectexecution.StatusFailed:
						return errors.New(latestExec.Error)
					case sideeffectexecution.StatusPending, sideeffectexecution.StatusRunning:
						// The workflow is still in progress
						// return  errors.New("workflow is still in progress")
						runtime.Gosched()
						continue
					}

				case sideeffect.StatusPending, sideeffect.StatusRunning:
					runtime.Gosched()
					continue
				}
			}
			runtime.Gosched()
		}
	}
}

type SagaInfo[T Identifier] struct {
	tp     *Tempolite[T]
	SagaID SagaID
	err    error
}

func (i *SagaInfo[T]) Get(output ...interface{}) error {
	if i.err != nil {
		return i.err
	}

	ticker := time.NewTicker(time.Second / 16)
	defer ticker.Stop()

	for {
		select {
		case <-i.tp.ctx.Done():
			return i.tp.ctx.Err()
		case <-ticker.C:
			sagaEntity, err := i.tp.client.Saga.Query().
				Where(saga.IDEQ(i.SagaID.String())).
				WithSteps().
				Only(i.tp.ctx)
			if err != nil {
				return err
			}

			switch sagaEntity.Status {
			case saga.StatusCompleted:
				return nil
			case saga.StatusFailed:
				return errors.New(sagaEntity.Error)
			case saga.StatusCompensated:
				return errors.New("saga was compensated")
			case saga.StatusPending, saga.StatusRunning:
				runtime.Gosched()
				continue
			}
			runtime.Gosched()
		}
	}
}

type TempoliteContext interface {
	context.Context
	RunID() string
	EntityID() string
	ExecutionID() string
	EntityType() string
	StepID() string
}

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
		return ErrWorkflowPaused
	}
	return nil
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
		return &WorkflowInfo[T]{err: err}
	}
	id, err := w.tp.enqueueSubWorkflow(w, stepID, handler, inputs...)
	return w.tp.getWorkflow(w, id, err)
}

func (w WorkflowContext[T]) ActivityFunc(stepID T, handler interface{}, inputs ...any) *ActivityInfo[T] {
	if err := w.checkIfPaused(); err != nil {
		return &ActivityInfo[T]{err: err}
	}
	id, err := w.tp.enqueueSubActivtyFunc(w, stepID, handler, inputs...)
	if err != nil {
		log.Printf("Error enqueuing activity execution: %v", err)
	}
	fmt.Println("\t \t activity execution id", id, err)
	return w.tp.getActivity(w, id, err)
}

func (w WorkflowContext[T]) ExecuteActivity(stepID T, name HandlerIdentity, inputs ...any) *ActivityInfo[T] {
	if err := w.checkIfPaused(); err != nil {
		return &ActivityInfo[T]{err: err}
	}
	id, err := w.tp.enqueueSubActivityExecution(w, stepID, name, inputs...)
	if err != nil {
		log.Printf("Error enqueuing activity execution: %v", err)
	}
	fmt.Println("\t \t activity execution id", id, err)
	return w.tp.getActivity(w, id, err)
}

func (w WorkflowContext[T]) GetActivity(id string) (*ActivityInfo[T], error) {
	// todo: implement
	return nil, nil
}

type ActivityContext[T Identifier] struct {
	TempoliteContext
	tp          *Tempolite[T]
	activityID  string
	executionID string
	runID       string
	stepID      string
}

func (w ActivityContext[T]) StepID() string {
	return w.stepID
}

func (w ActivityContext[T]) RunID() string {
	return w.runID
}

func (w ActivityContext[T]) EntityID() string {
	return w.activityID
}

func (w ActivityContext[T]) ExecutionID() string {
	return w.executionID
}

func (w ActivityContext[T]) EntityType() string {
	return "activity"
}

func (w ActivityContext[T]) GetActivity(id ActivityExecutionID) *ActivityExecutionInfo[T] {
	return w.tp.getActivityExecution(w, id, nil)
}

type SideEffectContext[T Identifier] struct {
	TempoliteContext
	tp           *Tempolite[T]
	sideEffectID string
	executionID  string
	runID        string
	stepID       string
}

func (w SideEffectContext[T]) StepID() string {
	return w.stepID
}

func (w SideEffectContext[T]) RunID() string {
	return w.runID
}

func (w SideEffectContext[T]) EntityID() string {
	return w.sideEffectID
}

func (w SideEffectContext[T]) ExecutionID() string {
	return w.executionID
}

func (w SideEffectContext[T]) EntityType() string {
	return "sideEffect"
}

type TransactionContext[T Identifier] struct {
	TempoliteContext
	tp *Tempolite[T]
}

func (w TransactionContext[T]) GetActivity(id string) (*ActivityInfo[T], error) {
	// todo: implement
	return nil, nil
}

func (w TransactionContext[T]) EntityType() string {
	return "transaction"
}

type CompensationContext[T Identifier] struct {
	TempoliteContext
	tp *Tempolite[T]
}

func (w CompensationContext[T]) EntityType() string {
	return "compensation"
}
