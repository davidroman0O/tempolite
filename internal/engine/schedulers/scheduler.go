package schedulers

import (
	"context"
	"log"
	"time"

	"github.com/davidroman0O/tempolite/internal/engine/queues"
	"github.com/davidroman0O/tempolite/internal/engine/registry"
	"github.com/davidroman0O/tempolite/internal/persistence/repository"
)

type Scheduler struct {
	clock    *Clock
	cerr     chan error
	db       repository.Repository
	getQueue func(queue string) *queues.Queue
	registry *registry.Registry
}

func New(
	ctx context.Context,
	db repository.Repository,
	registry *registry.Registry,
	getQueue func(queue string) *queues.Queue,
) (*Scheduler, error) {
	s := &Scheduler{
		db:       db,
		cerr:     make(chan error),
		getQueue: getQueue,
		registry: registry,
		clock: NewClock(
			ctx,
			WithInterval(time.Millisecond),
			WithName("scheduler"),
			WithOnError(func(err error) {
				// TODO: i don't know what to do with the errors
				log.Println(err)
			}),
		),
	}

	s.clock.Start()

	return s, nil
}

func (s *Scheduler) Stop() {
	s.clock.Stop()
}

func (s *Scheduler) addScheduler(ticker Ticker) {
	s.clock.Add(ticker, BestEffort)
}

func (s *Scheduler) AddQueue(queue string) {
	s.addScheduler(
		SchedulerWorkflowsPending{
			Scheduler: s,
		})
}
