package tempolite

import (
	"log"
	"runtime"

	"github.com/davidroman0O/go-tempolite/ent/handlertask"
)

// Will pull tasks from db to be dispatched
type Scheduler struct {
	tp   *Tempolite
	done chan struct{}
}

func NewScheduler(tp *Tempolite) *Scheduler {

	s := &Scheduler{
		tp:   tp,
		done: make(chan struct{}),
	}

	go s.run()

	return s
}

func (s *Scheduler) Close() {
	close(s.done)
}

func (s *Scheduler) run() {
	for {
		select {
		case <-s.done:
			return
		default:
			// TODO: fetch saga task, side effect task, compensation task

			totalHandlerTasks, err := s.tp.client.HandlerTask.Query().Where(handlertask.StatusEQ(handlertask.StatusPending)).Count(s.tp.ctx)
			if err != nil {
				log.Printf("Error counting pending tasks: %v", err)
				continue
			}
			if totalHandlerTasks > 0 {
				handlerTask, err := s.tp.client.
					HandlerTask.
					Query().
					Where(handlertask.StatusEQ(handlertask.StatusPending)).
					WithTaskContext().
					WithExecutionContext().
					First(s.tp.ctx)
				if err != nil {
					log.Printf("Error getting pending task: %v", err)
					continue
				}
				if err := s.tp.handlerTaskPool.pool.Dispatch(handlerTask); err != nil {
					log.Printf("Error dispatching task: %v", err)
					continue
				}
				if err = s.tp.client.HandlerTask.Update().Where(handlertask.IDEQ(handlerTask.ID)).SetStatus(handlertask.StatusInProgress).Exec(s.tp.ctx); err != nil {
					log.Printf("Error updating task status: %v", err)
					continue
				}
			}
			runtime.Gosched()
		}
	}
}
