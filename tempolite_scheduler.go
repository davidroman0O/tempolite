package tempolite

import (
	"fmt"
	"log"
	"runtime"

	"github.com/davidroman0O/go-tempolite/ent"
	"github.com/davidroman0O/go-tempolite/ent/handlertask"
	"github.com/davidroman0O/go-tempolite/ent/node"
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
					WithNode().
					First(s.tp.ctx)
				if err != nil {
					log.Printf("Error getting pending task: %v", err)
					continue
				}

				node, err := s.tp.client.Node.
					Query().
					Where(node.HasHandlerTaskWith(handlertask.ID(handlerTask.ID))).
					Only(s.tp.ctx)

				if err != nil {
					if ent.IsNotFound(err) {
						fmt.Printf("no node found for handler task ID %s", handlerTask.ID)
						continue
					}
					fmt.Printf("error querying node: %v", err)
					continue
				}

				handlerTask.Edges.Node = node

				fmt.Println(handlerTask.ID)
				fmt.Println(handlerTask.Edges.Node)
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
