package schedulers

import (
	"fmt"

	"github.com/davidroman0O/tempolite/internal/persistence/ent"
	"github.com/davidroman0O/tempolite/internal/persistence/repository"
)

type SchedulerWorkflowsPending struct {
	*Scheduler
}

func (s SchedulerWorkflowsPending) Tick() error {
	var tx *ent.Tx
	var err error
	var queues []*repository.QueueInfo

	fmt.Println("SchedulerWorkflowsPending tick")

	if tx, err = s.db.Tx(); err != nil {
		return err
	}

	if queues, err = s.db.Queues().List(tx); err != nil {
		return err
	}

	for _, q := range queues {
		queue := s.Scheduler.getQueue(q.Name)

		fmt.Println("Queue", q.Name, "available workers", queue.AvailableWorkflowWorkers())
		if queue.AvailableWorkflowWorkers() <= 0 {
			continue
		}

		// TODO: we need more methods to get pending workflows
		workflows, err := s.db.Workflows().ListExecutionsPending(tx, q.Name)
		if err != nil {
			return err
		}

		fmt.Println(workflows)
	}

	return nil
}
