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
		return types.NoWorkflowID, fmt.Errorf("workflowEntityID is required")
	}

	if workflowExecutionID.IsNoID() {
		return types.NoWorkflowID, fmt.Errorf("workflowExecutionID is required")
	}

	var err error
	var identity types.HandlerIdentity
	if identity, err = e.registry.WorkflowIdentity(workflowFunc); err != nil {
		return types.NoWorkflowID, err
	}

	var workflow types.Workflow
	if workflow, err = e.registry.GetWorkflow(identity); err != nil {
		return types.NoWorkflowID, err
	}

	if err = e.registry.VerifyParamsMatching(types.HandlerInfo(workflow), params...); err != nil {
		return types.NoWorkflowID, err
	}

	// TODO: check if the workflow already exist

	var tx *ent.Tx
	if tx, err = e.db.Tx(); err != nil {
		return types.NoWorkflowID, err
	}

	var parentWorkflowInfo *repository.WorkflowInfo

	// Get the exact pair
	if parentWorkflowInfo, err = e.db.Workflows().GetWithExecution(tx, workflowEntityID, workflowExecutionID); err != nil {
		if err := tx.Rollback(); err != nil {
			return types.NoWorkflowID, err
		}
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
			return types.NoWorkflowID, err
		}
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
			return types.NoWorkflowID, err
		}
		return types.NoWorkflowID, err
	}

	return types.WorkflowID(workflowInfo.ID), nil
}

func (e *Commands) CommandWorkflow(workflowFunc interface{}, options types.WorkflowOptions, params ...any) (types.WorkflowID, error) {
	var err error
	var identity types.HandlerIdentity
	if identity, err = e.registry.WorkflowIdentity(workflowFunc); err != nil {
		return types.NoWorkflowID, err
	}

	var workflow types.Workflow
	if workflow, err = e.registry.GetWorkflow(identity); err != nil {
		return types.NoWorkflowID, err
	}

	if err = e.registry.VerifyParamsMatching(types.HandlerInfo(workflow), params...); err != nil {
		return types.NoWorkflowID, err
	}

	var tx *ent.Tx
	if tx, err = e.db.Tx(); err != nil {
		return types.NoWorkflowID, err
	}

	var workflowInfo *repository.WorkflowInfo

	var runInfo *repository.RunInfo
	if runInfo, err = e.db.Runs().Create(tx); err != nil {
		if err := tx.Rollback(); err != nil {
			return types.NoWorkflowID, err
		}
		return types.NoWorkflowID, err
	}

	retryPolicyConfig := schema.RetryPolicy{
		MaxAttempts: 1,
	}

	queueName := "default"
	var queueInfo *repository.QueueInfo

	if queueInfo, err = e.db.Queues().GetByName(tx, queueName); err != nil {
		if err := tx.Rollback(); err != nil {
			return types.NoWorkflowID, err
		}
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
			return types.NoWorkflowID, err
		}
		return types.NoWorkflowID, err
	}

	if err = tx.Commit(); err != nil {
		return types.NoWorkflowID, err
	}

	return types.WorkflowID(workflowInfo.ID), nil
}
