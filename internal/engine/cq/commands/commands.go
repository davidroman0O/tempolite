package commands

import (
	"context"
	"fmt"

	"github.com/davidroman0O/tempolite/internal/engine/io"
	"github.com/davidroman0O/tempolite/internal/engine/registry"
	"github.com/davidroman0O/tempolite/internal/persistence/ent"
	"github.com/davidroman0O/tempolite/internal/persistence/ent/schema"
	"github.com/davidroman0O/tempolite/internal/persistence/repository"
	"github.com/davidroman0O/tempolite/internal/types"
	"github.com/davidroman0O/tempolite/pkg/logs"
)

type Commands struct {
	ctx      context.Context
	db       repository.Repository
	registry *registry.Registry
}

func New(ctx context.Context, db repository.Repository, registry *registry.Registry) *Commands {
	return &Commands{
		ctx:      ctx,
		db:       db,
		registry: registry,
	}
}

func (e *Commands) CommandSubWorkflow(workflowEntityID types.WorkflowID, workflowExecutionID types.WorkflowExecutionID, stepID string, workflowFunc interface{}, options types.WorkflowOptions, params ...any) (types.WorkflowID, error) {

	if workflowEntityID.IsNoID() {
		logs.Error(e.ctx, "CommandSubWorkflow workflowEntityID is required", "workflowEntityID", workflowEntityID, "workflowExecutionID", workflowExecutionID, "stepID", stepID)
		return types.NoWorkflowID, fmt.Errorf("workflowEntityID is required")
	}

	if workflowExecutionID.IsNoID() {
		logs.Error(e.ctx, "CommandSubWorkflow workflowExecutionID is required", "workflowEntityID", workflowEntityID, "workflowExecutionID", workflowExecutionID, "stepID", stepID)
		return types.NoWorkflowID, fmt.Errorf("workflowExecutionID is required")
	}

	var err error
	var identity types.HandlerIdentity
	if identity, err = e.registry.WorkflowIdentity(workflowFunc); err != nil {
		logs.Error(e.ctx, "CommandSubWorkflow workflow identity not found", "error", err)
		return types.NoWorkflowID, err
	}

	var workflow types.Workflow
	if workflow, err = e.registry.GetWorkflow(identity); err != nil {
		logs.Error(e.ctx, "CommandSubWorkflow workflow registry not found", "error", err)
		return types.NoWorkflowID, err
	}

	if err = e.registry.VerifyParamsMatching(types.HandlerInfo(workflow), params...); err != nil {
		logs.Error(e.ctx, "CommandSubWorkflow params not matching", "error", err)
		return types.NoWorkflowID, err
	}

	// TODO: check if the workflow already exist

	var tx *ent.Tx
	if tx, err = e.db.Tx(); err != nil {
		logs.Error(e.ctx, "CommandSubWorkflow creating transaction error", "error", err)
		return types.NoWorkflowID, err
	}

	var parentWorkflowInfo *repository.WorkflowInfo

	// Get the exact pair
	if parentWorkflowInfo, err = e.db.Workflows().GetWithExecution(tx, workflowEntityID, workflowExecutionID); err != nil {
		if err := tx.Rollback(); err != nil {
			logs.Error(e.ctx, "CommandSubWorkflow error getting parent workflow", "error", err)
			return types.NoWorkflowID, err
		}
		logs.Error(e.ctx, "CommandSubWorkflow error getting parent workflow", "error", err)
		return types.NoWorkflowID, err
	}

	retryPolicyConfig := schema.RetryPolicy{
		MaxAttempts: 1,
	}

	// TODO: options should have the queue name
	// queueName := "default"
	var queueInfo *repository.QueueInfo

	if queueInfo, err = e.db.Queues().Get(tx, parentWorkflowInfo.QueueID); err != nil {
		if err := tx.Rollback(); err != nil {
			logs.Error(e.ctx, "CommandSubWorkflow error getting queue", "error", err)
			return types.NoWorkflowID, err
		}
		logs.Error(e.ctx, "CommandSubWorkflow error getting queue", "error", err)
		return types.NoWorkflowID, err
	}

	duration := ""

	if options != nil {
		config := types.WorkflowConfig{}
		for _, opt := range options {
			opt(&config)
		}
		if config.QueueName != "" {
			// queueName = config.QueueName
		}
		if config.RetryMaximumAttempts >= 0 {
			retryPolicyConfig.MaxAttempts = config.RetryMaximumAttempts
		}
		if config.RetryInitialInterval >= 0 {
			retryPolicyConfig.InitialInterval = config.RetryInitialInterval
		}
		// TODO: implement layet
		// if config.RetryBackoffCoefficient >= 0 {
		// 	retryPolicyConfig.BackoffCoefficient = config.RetryBackoffCoefficient
		// }
		// if config.MaximumInterval >= 0 {
		// 	retryPolicyConfig.MaxAttempts = config.MaximumInterval
		// }
		if config.Duration != "" {
			duration = config.Duration
		}

	}

	// Extremely important conversion
	serializableParams, err := io.ConvertInputsForSerialization(params)
	if err != nil {
		logs.Error(e.ctx, "CommandSubWorkflow error converting params", "error", err)
		return types.NoWorkflowID, err
	}

	var workflowInfo *repository.WorkflowInfo
	if workflowInfo, err = e.db.
		Workflows().
		CreateSub(
			tx,
			repository.CreateSubWorkflowInput{
				ParentID:          workflowEntityID.ID(),
				ParentExecutionID: workflowExecutionID.ID(),
				RunID:             parentWorkflowInfo.RunID,
				HandlerName:       identity.String(),
				StepID:            stepID,
				RetryPolicy:       &retryPolicyConfig,
				Input:             serializableParams,
				QueueID:           queueInfo.ID,
				Duration:          duration, // if empty it won't be set
			}); err != nil {
		if err := tx.Rollback(); err != nil {
			logs.Error(e.ctx, "CommandSubWorkflow error creating sub workflow", "error", err)
			return types.NoWorkflowID, err
		}
		logs.Error(e.ctx, "CommandSubWorkflow error creating sub workflow", "error", err)
		return types.NoWorkflowID, err
	}

	logs.Debug(e.ctx, "CommandSubWorkflow created sub workflow", "workflowInfo", workflowInfo)
	return types.WorkflowID(workflowInfo.ID), nil
}

func (e *Commands) CommandWorkflow(workflowFunc interface{}, options types.WorkflowOptions, params ...any) (types.WorkflowID, error) {
	var err error
	var identity types.HandlerIdentity
	if identity, err = e.registry.WorkflowIdentity(workflowFunc); err != nil {
		logs.Error(e.ctx, "CommandWorkflow workflow identity not found", "error", err)
		return types.NoWorkflowID, err
	}
	logs.Debug(e.ctx, "CommandWorkflow checking", "identity", identity)

	logs.Debug(e.ctx, "CommandWorkflow getting workflow registry using identity", "identity", identity)
	var workflow types.Workflow
	if workflow, err = e.registry.GetWorkflow(identity); err != nil {
		logs.Error(e.ctx, "CommandWorkflow workflow registry not found using identity", "error", err, "identity", identity)
		return types.NoWorkflowID, err
	}

	logs.Debug(e.ctx, "CommandWorkflow verifying params matching", "workflow", types.HandlerInfo(workflow))
	if err = e.registry.VerifyParamsMatching(types.HandlerInfo(workflow), params...); err != nil {
		logs.Error(e.ctx, "CommandWorkflow params not matching", "error", err, "identity", identity, "workflow", workflow.HandlerName)
		return types.NoWorkflowID, err
	}

	var tx *ent.Tx
	if tx, err = e.db.Tx(); err != nil {
		logs.Error(e.ctx, "CommandWorkflow creating transaction error", "error", err)
		return types.NoWorkflowID, err
	}

	var workflowInfo *repository.WorkflowInfo

	var runInfo *repository.RunInfo
	if runInfo, err = e.db.Runs().Create(tx); err != nil {
		if err := tx.Rollback(); err != nil {
			logs.Error(e.ctx, "CommandWorkflow error creating run", "error", err)
			return types.NoWorkflowID, err
		}
		logs.Error(e.ctx, "CommandWorkflow error creating run", "error", err)
		return types.NoWorkflowID, err
	}

	retryPolicyConfig := schema.RetryPolicy{
		MaxAttempts: 1,
	}

	queueName := "default"
	var queueInfo *repository.QueueInfo

	if queueInfo, err = e.db.Queues().GetByName(tx, queueName); err != nil {
		if err := tx.Rollback(); err != nil {
			logs.Error(e.ctx, "CommandWorkflow error getting queue", "error", err)
			return types.NoWorkflowID, err
		}
		logs.Error(e.ctx, "CommandWorkflow error getting queue", "error", err)
		return types.NoWorkflowID, err
	}

	duration := ""

	if options != nil {
		config := types.WorkflowConfig{}
		for _, opt := range options {
			opt(&config)
		}
		if config.QueueName != "" {
			queueName = config.QueueName
		}
		if config.RetryMaximumAttempts >= 0 {
			retryPolicyConfig.MaxAttempts = config.RetryMaximumAttempts
		}
		if config.RetryInitialInterval >= 0 {
			retryPolicyConfig.InitialInterval = config.RetryInitialInterval
		}
		// TODO: implement layet
		// if config.RetryBackoffCoefficient >= 0 {
		// 	retryPolicyConfig.BackoffCoefficient = config.RetryBackoffCoefficient
		// }
		// if config.MaximumInterval >= 0 {
		// 	retryPolicyConfig.MaxAttempts = config.MaximumInterval
		// }
		if config.Duration != "" {
			duration = config.Duration
		}

	}

	// Extremely important conversion
	serializableParams, err := io.ConvertInputsForSerialization(params)
	if err != nil {
		logs.Error(e.ctx, "CommandWorkflow error converting params", "error", err)
		return types.NoWorkflowID, err
	}

	if workflowInfo, err = e.db.
		Workflows().
		Create(
			tx,
			repository.CreateWorkflowInput{
				RunID:       runInfo.ID,
				HandlerName: identity.String(),
				StepID:      "root",
				RetryPolicy: &retryPolicyConfig,
				Input:       serializableParams,
				QueueID:     queueInfo.ID,
				Duration:    duration, // if empty it won't be set
			}); err != nil {
		if err := tx.Rollback(); err != nil {
			logs.Error(e.ctx, "CommandWorkflow error creating workflow", "error", err)
			return types.NoWorkflowID, err
		}
		logs.Error(e.ctx, "CommandWorkflow error creating workflow", "error", err)
		return types.NoWorkflowID, err
	}

	if err = tx.Commit(); err != nil {
		logs.Error(e.ctx, "CommandWorkflow error committing transaction", "error", err)
		return types.NoWorkflowID, err
	}

	return types.WorkflowID(workflowInfo.ID), nil
}
