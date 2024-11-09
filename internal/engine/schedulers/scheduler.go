package schedulers

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/davidroman0O/tempolite/internal/clock"
	"github.com/davidroman0O/tempolite/internal/engine/queues"
	"github.com/davidroman0O/tempolite/internal/engine/registry"
	"github.com/davidroman0O/tempolite/internal/persistence/repository"
)

type Scheduler struct {
	clock    *clock.Clock
	cerr     chan error
	db       repository.Repository
	getQueue func(queue string) *queues.Queue
	registry *registry.Registry
	incID    int
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
		clock: clock.NewClock(
			ctx,
			time.Nanosecond,
			func(err error) {
				// TODO: i don't know what to do with the errors
				log.Println(err)
			},
		),
	}

	s.clock.Start()

	return s, nil
}

func (s *Scheduler) Stop() {
	fmt.Println("Stopping scheduler")
	s.clock.Stop()
	defer fmt.Println("Scheduler stopped")
}

func (s *Scheduler) addScheduler(tickerID clock.TickerID, ticker clock.Ticker) {
	s.clock.Add(tickerID, ticker, clock.BestEffort)
}

func (s *Scheduler) AddQueue(queue string) {
	s.incID++
	s.addScheduler(
		s.incID,
		SchedulerWorkflowsPending{
			Scheduler: s,
		})
}
