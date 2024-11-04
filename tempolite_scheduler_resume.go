// package tempolite

// import (
// 	"context"
// 	"errors"

// 	"github.com/davidroman0O/tempolite/ent/workflow"
// )

// // One shot function at startup to resume all running workflows
// // Basically, we treat workflow with "Running false false" as paused!
// func (tp *Tempolite) resumeRunningWorkflows(queue string) error {
// 	// get all workflows with status Running, ispaused as false and isready as false
// 	workflows, err := tp.client.Workflow.Query().
// 		Where(workflow.And(
// 			workflow.StatusEQ(workflow.StatusRunning),
// 			workflow.IsPausedEQ(false),
// 			workflow.IsReadyEQ(false),
// 			workflow.QueueName(queue),
// 		)).
// 		All(tp.ctx)
// 	if err != nil {
// 		if errors.Is(err, context.Canceled) {
// 			tp.logger.Debug(tp.ctx, "resume workflow execution: context canceled", "queue", queue)
// 			return err
// 		}
// 		return err
// 	}

// 	for _, w := range workflows {
// 		// let's update as resumable
// 		_, err := tp.client.Workflow.UpdateOneID(w.ID).
// 			SetStatus(workflow.StatusPaused).
// 			SetIsPaused(false).
// 			SetIsReady(true).
// 			Save(tp.ctx)
// 		if err != nil {
// 			if errors.Is(err, context.Canceled) {
// 				tp.logger.Debug(tp.ctx, "resume workflow execution: context canceled", "queue", queue)
// 				return err
// 			}
// 			return err
// 		}
// 	}

// 	return nil
// }

package tempolite

import (
	"context"
	"errors"

	"github.com/davidroman0O/tempolite/ent/activity"
	"github.com/davidroman0O/tempolite/ent/executionrelationship"
	"github.com/davidroman0O/tempolite/ent/workflow"
	"github.com/davidroman0O/tempolite/ent/workflowexecution"
)

// One shot function at startup to resume all running workflows
// Basically, we treat workflow with "Running false false" as paused!
func (tp *Tempolite) resumeRunningWorkflows(queue string) error {
	// get all workflows with status Running, ispaused as false and isready as false
	workflows, err := tp.client.Workflow.Query().
		Where(workflow.And(
			workflow.StatusEQ(workflow.StatusRunning),
			workflow.IsPausedEQ(false),
			workflow.IsReadyEQ(false),
			workflow.QueueName(queue),
		)).
		All(tp.ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			tp.logger.Debug(tp.ctx, "resume workflow execution: context canceled", "queue", queue)
			return err
		}
		return err
	}

	for _, w := range workflows {
		// When you resume, you should also set the previous executions and activities to something
		wfExecs, err := tp.client.WorkflowExecution.Query().Where(workflowexecution.HasWorkflowWith(workflow.ID(w.ID))).All(tp.ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				tp.logger.Debug(tp.ctx, "resume workflow execution: context canceled", "queue", queue)
				return err
			}
			return err
		}

		// get the activities of the workflow
		execRelations, err := tp.client.ExecutionRelationship.Query().Where(executionrelationship.ParentEntityIDEQ(w.ID)).All(tp.ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				tp.logger.Debug(tp.ctx, "resume workflow execution: context canceled", "queue", queue)
				return err
			}
			return err
		}

		tx, err := tp.client.Tx(tp.ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				tp.logger.Debug(tp.ctx, "resume workflow execution: context canceled", "queue", queue)
				return err
			}
			return err
		}

		// let's update as resumable
		_, err = tx.Workflow.UpdateOneID(w.ID).
			SetStatus(workflow.StatusPaused).
			SetIsPaused(false).
			SetIsReady(true).
			Save(tp.ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				tp.logger.Debug(tp.ctx, "resume workflow execution: context canceled", "queue", queue)
				return err
			}
			return err
		}

		for _, wfexec := range wfExecs {
			if wfexec.Status == workflowexecution.StatusRunning {
				_, err = tx.WorkflowExecution.UpdateOneID(wfexec.ID).
					SetStatus(workflowexecution.StatusCancelled).
					Save(tp.ctx)
				if err != nil {
					if errors.Is(err, context.Canceled) {
						tp.logger.Debug(tp.ctx, "resume workflow execution: context canceled", "queue", queue)
						return err
					}
					return err
				}
			}
		}

		for _, relation := range execRelations {
			if relation.ChildType == executionrelationship.ChildTypeActivity {
				act, err := tx.Activity.Get(tp.ctx, relation.ChildEntityID)
				if err != nil {
					if errors.Is(err, context.Canceled) {
						tp.logger.Debug(tp.ctx, "resume workflow execution: context canceled", "queue", queue)
						return err
					}
					return err
				}
				if act.Status == activity.StatusRunning {
					_, err = tx.Activity.UpdateOneID(act.ID).
						SetStatus(activity.StatusCancelled).
						Save(tp.ctx)
					if err != nil {
						if errors.Is(err, context.Canceled) {
							tp.logger.Debug(tp.ctx, "resume workflow execution: context canceled", "queue", queue)
							return err
						}
						return err
					}
				}
			}
		}

	}

	return nil
}
