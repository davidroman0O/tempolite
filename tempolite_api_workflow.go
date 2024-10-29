package tempolite

import (
	"fmt"
	"reflect"
	"runtime"

	"github.com/davidroman0O/tempolite/ent"
	"github.com/davidroman0O/tempolite/ent/executionrelationship"
	"github.com/davidroman0O/tempolite/ent/run"
	"github.com/davidroman0O/tempolite/ent/schema"
	"github.com/davidroman0O/tempolite/ent/workflow"
	"github.com/davidroman0O/tempolite/ent/workflowexecution"
	"github.com/google/uuid"
)

// func (tp *Tempolite) Workflow(stepID T, workflowFunc interface{}, opts tempoliteWorkflowConfig, params ...interface{}) *WorkflowInfo {
func (tp *Tempolite) Workflow(workflowFunc interface{}, options tempoliteWorkflowOptions, params ...interface{}) *WorkflowInfo {
	tp.logger.Debug(tp.ctx, "Workflow")
	id, err := tp.executeWorkflow(workflowFunc, options, params...)
	if err != nil {
		tp.logger.Error(tp.ctx, "Error executing workflow", "error", err)
	}
	return tp.getWorkflowRoot(id, err)
}

func (tp *Tempolite) GetWorkflow(id WorkflowID) *WorkflowInfo {
	tp.logger.Debug(tp.ctx, "GetWorkflow", "workflowID", id)
	info := WorkflowInfo{
		tp:         tp,
		WorkflowID: id,
	}
	workflowEntity, err := tp.client.Workflow.Get(tp.ctx, string(id))
	if err == nil {
		if value, ok := tp.workflows.Load(HandlerIdentity(workflowEntity.Identity)); ok {
			workflowHandlerInfo := value.(Workflow)
			info.handler = workflowHandlerInfo.Handler
		} else {
			info.err = fmt.Errorf("workflow handler not found: %w", err)
		}
	} else {
		info.err = fmt.Errorf("error getting workflow: %w", err)
	}
	return &info
}

func (tp *Tempolite) getWorkflowRoot(id WorkflowID, err error) *WorkflowInfo {
	tp.logger.Debug(tp.ctx, "getWorkflowRoot", "workflowID", id, "error", err)
	info := WorkflowInfo{
		tp:         tp,
		WorkflowID: id,
		err:        err,
	}
	if err != nil {
		return &info
	}
	workflowEntity, err := tp.client.Workflow.Get(tp.ctx, string(id))
	if err == nil {
		if value, ok := tp.workflows.Load(HandlerIdentity(workflowEntity.Identity)); ok {
			workflowHandlerInfo := value.(Workflow)
			info.handler = workflowHandlerInfo.Handler
		} else {
			info.err = fmt.Errorf("workflow handler not found: %w", err)
		}
	} else {
		info.err = fmt.Errorf("error getting workflow: %w", err)
	}
	return &info
}

func (tp *Tempolite) getWorkflow(ctx TempoliteContext, id WorkflowID, err error) *WorkflowInfo {
	tp.logger.Debug(tp.ctx, "getWorkflow", "workflowID", id, "error", err)
	info := WorkflowInfo{
		tp:         tp,
		WorkflowID: id,
		err:        err,
	}
	return &info
}

func (tp *Tempolite) getWorkflowExecution(ctx TempoliteContext, id WorkflowExecutionID, err error) *WorkflowExecutionInfo {
	tp.logger.Debug(tp.ctx, "getWorkflowExecution", "workflowExecutionID", id, "error", err)
	info := WorkflowExecutionInfo{
		tp:          tp,
		ExecutionID: id,
		err:         err,
	}
	return &info
}

func (tp *Tempolite) enqueueWorkflow(ctx TempoliteContext, stepID string, workflowFunc interface{}, options tempoliteWorkflowOptions, params ...interface{}) (WorkflowID, error) {
	switch ctx.EntityType() {
	case "workflow":
		// Proceed with sub-workflow creation
	default:
		return "", fmt.Errorf("context entity type %s not supported", ctx.EntityType())
	}

	// Check for existing sub-workflow with the same stepID within the current workflow
	exists, err := tp.client.ExecutionRelationship.Query().
		Where(
			executionrelationship.And(
				executionrelationship.RunID(ctx.RunID()),
				executionrelationship.ParentEntityID(ctx.EntityID()),
				executionrelationship.ChildStepID(stepID),
				executionrelationship.ParentStepID(ctx.StepID()),
			),
			executionrelationship.ChildTypeEQ(executionrelationship.ChildTypeWorkflow),
		).
		First(tp.ctx)
	if err == nil {
		tp.logger.Debug(tp.ctx, "Existing sub-workflow found", "childEntityID", exists.ChildEntityID)
		// todo: is there a way to just get the status?
		act, err := tp.client.Workflow.Get(tp.ctx, exists.ChildEntityID)
		if err != nil {
			tp.logger.Error(tp.ctx, "Error getting workflow", "error", err)
			return "", err
		}
		if act.Status == workflow.StatusCompleted {
			tp.logger.Debug(tp.ctx, "Sub-workflow already completed", "workflowID", exists.ChildEntityID)
			return WorkflowID(exists.ChildEntityID), nil
		}
	} else {
		if !ent.IsNotFound(err) {
			tp.logger.Error(tp.ctx, "Error checking for existing stepID", "error", err)
			return "", fmt.Errorf("error checking for existing stepID: %w", err)
		}
	}

	funcName := runtime.FuncForPC(reflect.ValueOf(workflowFunc).Pointer()).Name()
	handlerIdentity := HandlerIdentity(funcName)

	retryPolicyConfig := schema.RetryPolicy{
		MaximumAttempts: 1,
	}

	queueName := "default"

	if options != nil {
		config := tempoliteWorkflowConfig{}
		for _, opt := range options {
			opt(&config)
		}
		if config.queueName != "" {
			queueName = config.queueName
		}
		if config.retryMaximumAttempts > 0 {
			retryPolicyConfig.MaximumAttempts = config.retryMaximumAttempts
		}
		if config.retryInitialInterval > 0 {
			retryPolicyConfig.InitialInterval = config.retryInitialInterval
		}
		if config.retryBackoffCoefficient > 0 {
			retryPolicyConfig.BackoffCoefficient = config.retryBackoffCoefficient
		}
		if config.maximumInterval > 0 {
			retryPolicyConfig.MaximumInterval = config.maximumInterval
		}
	}

	var value any
	var ok bool
	var tx *ent.Tx

	tp.logger.Debug(tp.ctx, "searching workflow handler", "handlerIdentity", handlerIdentity)
	if value, ok = tp.workflows.Load(handlerIdentity); ok {
		var workflowHandlerInfo Workflow
		if workflowHandlerInfo, ok = value.(Workflow); !ok {
			// could be development bug
			tp.logger.Error(tp.ctx, "Workflow is not handler info", "handlerIdentity", handlerIdentity)
			return "", fmt.Errorf("workflow %s is not handler info", handlerIdentity)
		}

		tp.logger.Debug(tp.ctx, "verifying handler and params", "workflowHandlerInfo", workflowHandlerInfo, "params", params)
		if err := tp.verifyHandlerAndParams(HandlerInfo(workflowHandlerInfo), params); err != nil {
			tp.logger.Error(tp.ctx, "Error verifying handler and params", "error", err)
			return "", err
		}

		if len(params) != workflowHandlerInfo.NumIn {
			tp.logger.Error(tp.ctx, "Parameter count mismatch", "expected", workflowHandlerInfo.NumIn, "got", len(params))
			return "", fmt.Errorf("parameter count mismatch: expected %d, got %d", workflowHandlerInfo.NumIn, len(params))
		}

		for idx, param := range params {
			if reflect.TypeOf(param) != workflowHandlerInfo.ParamTypes[idx] {
				tp.logger.Error(tp.ctx, "Parameter type mismatch", "expected", workflowHandlerInfo.ParamTypes[idx], "got", reflect.TypeOf(param))
				return "", fmt.Errorf("parameter type mismatch: expected %s, got %s", workflowHandlerInfo.ParamTypes[idx], reflect.TypeOf(param))
			}
		}

		serializableParams, err := tp.convertInputsForSerialization(HandlerInfo(workflowHandlerInfo), params)
		if err != nil {
			tp.logger.Error(tp.ctx, "Error converting inputs for serialization", "error", err)
			return "", err
		}

		tp.logger.Debug(tp.ctx, "Creating transaction to create workflow", "handlerIdentity", handlerIdentity)
		if tx, err = tp.client.Tx(tp.ctx); err != nil {
			tp.logger.Error(tp.ctx, "Error creating transaction to create workflow", "error", err)
			return "", err
		}

		tp.logger.Debug(tp.ctx, "Creating workflow entity")
		//	definition of a workflow, it exists but it is nothing without an execution that will be created as long as it retries
		var workflowEntity *ent.Workflow
		if workflowEntity, err = tx.Workflow.
			Create().
			SetID(uuid.NewString()).
			SetStepID(stepID).
			SetIdentity(string(handlerIdentity)).
			SetHandlerName(workflowHandlerInfo.HandlerName).
			SetInput(serializableParams).
			SetRetryPolicy(retryPolicyConfig).
			SetQueueName(queueName).
			Save(tp.ctx); err != nil {
			if err = tx.Rollback(); err != nil {
				tp.logger.Error(tp.ctx, "Error rolling back transaction creating workflow entity", "error", err)
				return "", err
			}
			tp.logger.Error(tp.ctx, "Error creating workflow entity", "error", err)
			return "", err
		}

		tp.logger.Debug(tp.ctx, "Creating workflow execution")
		// instance of the workflow definition which will be used to create a workflow task and match with the in-memory registry
		var workflowExecution *ent.WorkflowExecution
		if workflowExecution, err = tx.WorkflowExecution.
			Create().
			SetID(uuid.NewString()).
			SetRunID(ctx.RunID()).
			SetWorkflow(workflowEntity).
			SetQueueName(queueName).
			Save(tp.ctx); err != nil {
			if err = tx.Rollback(); err != nil {
				tp.logger.Error(tp.ctx, "Error rolling back transaction creating workflow execution", "error", err)
				return "", err
			}
			tp.logger.Error(tp.ctx, "Error creating workflow execution", "error", err)
			return "", err
		}

		tp.logger.Debug(tp.ctx, "Creating workflow relationship")
		if _, err := tx.ExecutionRelationship.Create().
			// run id
			SetRunID(ctx.RunID()).
			// entity
			SetParentEntityID(ctx.EntityID()).
			SetChildEntityID(workflowEntity.ID).
			// execution
			SetParentID(ctx.ExecutionID()).
			SetChildID(workflowExecution.ID).
			// Types
			SetParentType(executionrelationship.ParentTypeWorkflow).
			SetChildType(executionrelationship.ChildTypeWorkflow).
			// steps
			SetParentStepID(ctx.StepID()).
			SetChildStepID(stepID).
			//
			Save(tp.ctx); err != nil {
			if err = tx.Rollback(); err != nil {
				tp.logger.Error(tp.ctx, "Error rolling back transaction creating execution relationship", "error", err)
				return "", err
			}
			tp.logger.Error(tp.ctx, "Error creating execution relationship", "error", err)
			return "", err
		}

		tp.logger.Debug(tp.ctx, "Committing transaction creating workflow", "handlerIdentity", handlerIdentity)
		if err = tx.Commit(); err != nil {
			tp.logger.Error(tp.ctx, "Error committing transaction creating workflow", "error", err)
			return "", err
		}

		tp.logger.Debug(tp.ctx, "Workflow created", "workflowID", workflowEntity.ID, "handlerIdentity", handlerIdentity)
		return WorkflowID(workflowEntity.ID), nil

	} else {
		tp.logger.Error(tp.ctx, "Workflow not found", "handlerIdentity", handlerIdentity)
		return "", fmt.Errorf("workflow %s not found", handlerIdentity)
	}
}

func (tp *Tempolite) executeWorkflow(workflowFunc interface{}, options tempoliteWorkflowOptions, params ...interface{}) (WorkflowID, error) {
	funcName := runtime.FuncForPC(reflect.ValueOf(workflowFunc).Pointer()).Name()
	handlerIdentity := HandlerIdentity(funcName)
	var value any
	var ok bool
	var err error
	var tx *ent.Tx

	tp.logger.Debug(tp.ctx, "searching workflow handler", "handlerIdentity", handlerIdentity)
	if value, ok = tp.workflows.Load(handlerIdentity); ok {
		var workflowHandlerInfo Workflow
		if workflowHandlerInfo, ok = value.(Workflow); !ok {
			tp.logger.Error(tp.ctx, "Workflow is not handler info", "handlerIdentity", handlerIdentity)
			return "", fmt.Errorf("workflow %s is not handler info", handlerIdentity)
		}

		tp.logger.Debug(tp.ctx, "verifying handler and params", "workflowHandlerInfo", workflowHandlerInfo, "params", params)
		if err := tp.verifyHandlerAndParams(HandlerInfo(workflowHandlerInfo), params); err != nil {
			tp.logger.Error(tp.ctx, "Error verifying handler and params", "error", err)
			return "", err
		}

		if len(params) != workflowHandlerInfo.NumIn {
			tp.logger.Error(tp.ctx, "Parameter count mismatch", "handlerName", workflowHandlerInfo.HandlerLongName, "expected", workflowHandlerInfo.NumIn, "got", len(params))
			return "", fmt.Errorf("parameter count mismatch: expected %d, got %d", workflowHandlerInfo.NumIn, len(params))
		}

		for idx, param := range params {
			if reflect.TypeOf(param) != workflowHandlerInfo.ParamTypes[idx] {
				tp.logger.Error(tp.ctx, "Parameter type mismatch", "handlerName", workflowHandlerInfo.HandlerLongName, "expected", workflowHandlerInfo.ParamTypes[idx], "got", reflect.TypeOf(param))
				return "", fmt.Errorf("parameter type mismatch: expected %s, got %s", workflowHandlerInfo.ParamTypes[idx], reflect.TypeOf(param))
			}
		}

		tp.logger.Debug(tp.ctx, "Creating transaction to create workflow", "handlerIdentity", handlerIdentity)
		if tx, err = tp.client.Tx(tp.ctx); err != nil {
			tp.logger.Error(tp.ctx, "Error creating transaction to create workflow", "error", err)
			return "", err
		}

		tp.logger.Debug(tp.ctx, "Creating root run entity")
		var runEntity *ent.Run
		// root Run entity that anchored the workflow entity despite retries
		if runEntity, err = tx.
			Run.
			Create().
			SetID(uuid.NewString()).    // immutable
			SetRunID(uuid.NewString()). // can change if that flow change due to a retry
			SetType(run.TypeWorkflow).
			Save(tp.ctx); err != nil {
			if err = tx.Rollback(); err != nil {
				tp.logger.Error(tp.ctx, "Error rolling back transaction creating root run entity", "error", err)
				return "", err
			}
			tp.logger.Error(tp.ctx, "Error creating root run entity", "error", err)
			return "", err
		}

		retryPolicyConfig := schema.RetryPolicy{
			MaximumAttempts: 1,
		}

		queueName := "default"

		if options != nil {
			config := tempoliteWorkflowConfig{}
			for _, opt := range options {
				opt(&config)
			}
			if config.queueName != "" {
				queueName = config.queueName
			}
			if config.retryMaximumAttempts > 0 {
				retryPolicyConfig.MaximumAttempts = config.retryMaximumAttempts
			}
			if config.retryInitialInterval > 0 {
				retryPolicyConfig.InitialInterval = config.retryInitialInterval
			}
			if config.retryBackoffCoefficient > 0 {
				retryPolicyConfig.BackoffCoefficient = config.retryBackoffCoefficient
			}
			if config.maximumInterval > 0 {
				retryPolicyConfig.MaximumInterval = config.maximumInterval
			}
		}

		serializableParams, err := tp.convertInputsForSerialization(HandlerInfo(workflowHandlerInfo), params)
		if err != nil {
			tp.logger.Error(tp.ctx, "Error converting inputs for serialization", "error", err)
			return "", err
		}

		tp.logger.Debug(tp.ctx, "Creating workflow entity")
		//	definition of a workflow, it exists but it is nothing without an execution that will be created as long as it retries
		var workflowEntity *ent.Workflow
		if workflowEntity, err = tx.Workflow.
			Create().
			SetID(runEntity.RunID).
			SetStatus(workflow.StatusPending).
			SetStepID("root").
			SetIdentity(string(handlerIdentity)).
			SetHandlerName(workflowHandlerInfo.HandlerName).
			SetInput(serializableParams).
			SetRetryPolicy(retryPolicyConfig).
			SetQueueName(queueName).
			Save(tp.ctx); err != nil {
			if err = tx.Rollback(); err != nil {
				tp.logger.Error(tp.ctx, "Error rolling back transaction creating workflow entity", "error", err)
				return "", err
			}
			tp.logger.Error(tp.ctx, "Error creating workflow entity", "error", err)
			return "", err
		}

		tp.logger.Debug(tp.ctx, "Creating workflow execution")
		// instance of the workflow definition which will be used to create a workflow task and match with the in-memory registry
		var workflowExecution *ent.WorkflowExecution
		if workflowExecution, err = tx.WorkflowExecution.
			Create().
			SetID(uuid.NewString()).
			SetRunID(runEntity.ID).
			SetWorkflow(workflowEntity).
			SetQueueName(queueName).
			Save(tp.ctx); err != nil {
			if err = tx.Rollback(); err != nil {
				tp.logger.Error(tp.ctx, "Error rolling back transaction creating workflow execution", "error", err)
				return "", err
			}
			tp.logger.Error(tp.ctx, "Error creating workflow execution", "error", err)
			return "", err
		}

		tp.logger.Debug(tp.ctx, "workflow execution created", "workflowExecutionID", workflowExecution.ID, "handlerIdentity", handlerIdentity)

		tp.logger.Debug(tp.ctx, "Updating run entity with workflow")
		if _, err = tx.Run.UpdateOneID(runEntity.ID).SetWorkflow(workflowEntity).Save(tp.ctx); err != nil {
			if err = tx.Rollback(); err != nil {
				tp.logger.Error(tp.ctx, "Error rolling back transaction updating run entity with workflow", "error", err)
				return "", err
			}
			tp.logger.Error(tp.ctx, "Error updating run entity with workflow", "error", err)
			return "", err
		}

		tp.logger.Debug(tp.ctx, "Commiting transaction creating workflow", "handlerIdentity", handlerIdentity)
		if err = tx.Commit(); err != nil {
			tp.logger.Error(tp.ctx, "Error committing transaction creating workflow", "error", err)
			return "", err
		}

		tp.logger.Debug(tp.ctx, "Workflow created", "workflowID", workflowEntity.ID, "handlerIdentity", handlerIdentity)
		//	we're outside of the execution model, so we care about the workflow entity
		return WorkflowID(workflowEntity.ID), nil
	} else {
		tp.logger.Error(tp.ctx, "Workflow not found", "handlerIdentity", handlerIdentity)
		return "", fmt.Errorf("workflow %s not found", handlerIdentity)
	}
}

// RetryWorkflow initiates the retry process and returns a WorkflowInfo
func (tp *Tempolite) RetryWorkflow(workflowID WorkflowID) *WorkflowInfo {
	tp.logger.Debug(tp.ctx, "Retrying workflow", "workflowID", workflowID)
	id, err := tp.retryWorkflow(workflowID)
	return tp.getWorkflowRoot(id, err)
}

// retryWorkflow handles the actual retry logic
func (tp *Tempolite) retryWorkflow(workflowID WorkflowID) (WorkflowID, error) {
	// Find the original workflow
	originalWf, err := tp.client.Workflow.Get(tp.ctx, workflowID.String())
	if err != nil {
		tp.logger.Error(tp.ctx, "Error finding original workflow", "error", err)
		return "", fmt.Errorf("error finding original workflow: %w", err)
	}

	// Start a transaction
	tx, err := tp.client.Tx(tp.ctx)
	if err != nil {
		tp.logger.Error(tp.ctx, "Error starting transaction", "error", err)
		return "", fmt.Errorf("error starting transaction: %w", err)
	}

	// Create a new workflow as a retry of the original
	newWf, err := tx.Workflow.
		Create().
		SetID(uuid.New().String()).
		SetStepID(originalWf.StepID).
		SetStatus(workflow.StatusPending).
		SetIdentity(originalWf.Identity).
		SetHandlerName(originalWf.HandlerName).
		SetInput(originalWf.Input).
		SetRetryPolicy(originalWf.RetryPolicy).
		SetRetriedFromID(originalWf.ID).
		Save(tp.ctx)
	if err != nil {
		tp.logger.Error(tp.ctx, "Error creating new workflow", "error", err)
		if rerr := tx.Rollback(); rerr != nil {
			tp.logger.Error(tp.ctx, "Error rolling back transaction", "error", rerr)
		}
		return "", fmt.Errorf("error creating new workflow: %w", err)
	}

	// Create a new workflow execution for the new workflow
	_, err = tx.WorkflowExecution.
		Create().
		SetID(uuid.New().String()).
		SetRunID(newWf.ID).
		SetWorkflow(newWf).
		SetStatus(workflowexecution.StatusPending).
		Save(tp.ctx)
	if err != nil {
		tp.logger.Error(tp.ctx, "Error creating new workflow execution", "error", err)
		if rerr := tx.Rollback(); rerr != nil {
			tp.logger.Error(tp.ctx, "Error rolling back transaction", "error", rerr)
		}
		return "", fmt.Errorf("error creating new workflow execution: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		tp.logger.Error(tp.ctx, "Error committing transaction", "error", err)
		return "", fmt.Errorf("error committing transaction: %w", err)
	}

	tp.logger.Info(tp.ctx, "Workflow retry created successfully", "originalWorkflowID", workflowID, "newWorkflowID", newWf.ID)
	return WorkflowID(newWf.ID), nil
}

func (tp *Tempolite) ReplayWorkflow(workflowID WorkflowID) *WorkflowInfo {
	tp.logger.Debug(tp.ctx, "Replaying workflow", "workflowID", workflowID)
	id, err := tp.replayWorkflow(workflowID)
	return tp.getWorkflowRoot(id, err)
}

func (tp *Tempolite) replayWorkflow(workflowID WorkflowID) (WorkflowID, error) {
	// Find the original workflow
	originalWf, err := tp.client.Workflow.Get(tp.ctx, workflowID.String())
	if err != nil {
		tp.logger.Error(tp.ctx, "Error finding original workflow", "error", err)
		return "", fmt.Errorf("error finding original workflow: %w", err)
	}

	// Check if the workflow is in a final state
	if originalWf.Status != workflow.StatusCompleted && originalWf.Status != workflow.StatusFailed {
		tp.logger.Error(tp.ctx, "Cannot replay workflow that is not completed or failed", "status", originalWf.Status)
		return "", fmt.Errorf("cannot replay workflow that is not completed or failed")
	}

	// Find the last execution of the workflow
	lastExecution, err := tp.client.WorkflowExecution.Query().
		Where(workflowexecution.HasWorkflowWith(workflow.ID(originalWf.ID))).
		Order(ent.Desc(workflowexecution.FieldStartedAt)).
		First(tp.ctx)
	if err != nil {
		tp.logger.Error(tp.ctx, "Error finding last execution of workflow", "error", err)
		return "", fmt.Errorf("error finding last execution of workflow: %w", err)
	}

	// Start a transaction
	tx, err := tp.client.Tx(tp.ctx)
	if err != nil {
		tp.logger.Error(tp.ctx, "Error starting transaction", "error", err)
		return "", fmt.Errorf("error starting transaction: %w", err)
	}

	// Create a new workflow execution for the existing workflow
	newExecution, err := tx.WorkflowExecution.
		Create().
		SetID(uuid.New().String()).
		SetRunID(lastExecution.RunID).
		SetWorkflow(originalWf).
		SetStatus(workflowexecution.StatusPending).
		SetIsReplay(true).
		Save(tp.ctx)
	if err != nil {
		tp.logger.Error(tp.ctx, "Error creating new workflow execution", "error", err)
		if rerr := tx.Rollback(); rerr != nil {
			tp.logger.Error(tp.ctx, "Error rolling back transaction", "error", rerr)
		}
		return "", fmt.Errorf("error creating new workflow execution: %w", err)
	}

	// Update the workflow status to running
	_, err = tx.Workflow.UpdateOne(originalWf).
		SetStatus(workflow.StatusRunning).
		Save(tp.ctx)
	if err != nil {
		tp.logger.Error(tp.ctx, "Error updating workflow status", "error", err)
		if rerr := tx.Rollback(); rerr != nil {
			tp.logger.Error(tp.ctx, "Error rolling back transaction", "error", rerr)
		}
		return "", fmt.Errorf("error updating workflow status: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		tp.logger.Error(tp.ctx, "Error committing transaction", "error", err)
		return "", fmt.Errorf("error committing transaction: %w", err)
	}

	tp.logger.Info(tp.ctx, "Workflow replay created successfully", "workflowID", workflowID, "newExecutionID", newExecution.ID)
	return workflowID, nil
}
