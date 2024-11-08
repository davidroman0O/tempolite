package schedulers

import (
	"context"
	"errors"

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
		if err := tx.Rollback(); err != nil {
			return err
		}
		return err
	}

	for _, q := range queues {
		queue := s.Scheduler.getQueue(q.Name)

		// fmt.Println("Queue", q.Name, "available workers", queue.AvailableWorkflowWorkers())
		if queue.AvailableWorkflowWorkers() <= 0 {
			continue
		}

		workflows, err := s.db.Workflows().ListExecutionsPending(tx, q.Name)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			if err := tx.Rollback(); err != nil {
				return err
			}
			return err
		}

		for _, w := range workflows {
			// TODO: if we fail, then we should retry while updating the workflow somehow cause we need to fail
			if err := queue.SubmitWorkflow(w); err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				}
				if err := tx.Rollback(); err != nil {
					return err
				}
			}
		}

	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
