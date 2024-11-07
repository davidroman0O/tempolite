package tempolite

import (
	"github.com/davidroman0O/tempolite/ent"
	"github.com/davidroman0O/tempolite/ent/workflow"
	"github.com/davidroman0O/tempolite/ent/workflowexecution"
)

func (tp *Tempolite) getAvailableWorkflowExecutionSlots(queueName string) int {
	queueWorkersRaw, ok := tp.queues.Load(queueName)
	if !ok {
		return 0
	}
	queueWorkers := queueWorkersRaw.(*QueueWorkers)

	workerIDs := queueWorkers.Workflows.GetWorkerIDs()
	availableSlots := len(workerIDs) - queueWorkers.Workflows.ProcessingCount()
	if availableSlots <= 0 {
		return 0
	}

	return availableSlots
}

func (tp *Tempolite) getAvailableWorkflowExecutionForQueue(queueName string, availableSlots int) ([]*ent.WorkflowExecution, error) {
	pendingWorkflows, err := tp.client.WorkflowExecution.Query().
		Where(
			workflowexecution.StatusEQ(workflowexecution.StatusPending),
			workflowexecution.HasWorkflowWith(workflow.QueueNameEQ(queueName)),
		).
		Order(ent.Asc(workflowexecution.FieldStartedAt)).
		WithWorkflow().
		Limit(availableSlots).
		All(tp.ctx)

	if err != nil {
		return nil, err
	}

	return pendingWorkflows, nil
}

func (tp *Tempolite) getWorkflowByID(id string) (*ent.Workflow, error) {
	workflowEntity, err := tp.client.Workflow.Get(tp.ctx, id)
	if err != nil {
		return nil, err
	}

	return workflowEntity, nil
}
