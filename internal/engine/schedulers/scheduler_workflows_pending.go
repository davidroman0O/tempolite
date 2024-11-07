package schedulers

import (
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

	if tx, err = s.db.Tx(); err != nil {
		return err
	}

	if queues, err = s.db.Queues().List(tx); err != nil {
		return err
	}

	for _, q := range queues {
		queue := s.Scheduler.getQueue(q.Name)

		if queue.AvailableWorkflowWorkers() <= 0 {
			continue
		}

		// TODO: we need more methods to get pending workflows
	}

	return nil
}
