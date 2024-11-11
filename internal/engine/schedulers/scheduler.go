package schedulers

import (
	"context"
	"time"

	"github.com/davidroman0O/tempolite/internal/clock"
	"github.com/davidroman0O/tempolite/internal/engine/queues"
	"github.com/davidroman0O/tempolite/internal/engine/registry"
	"github.com/davidroman0O/tempolite/internal/persistence/repository"
	"github.com/davidroman0O/tempolite/pkg/logs"
)

type Scheduler struct {
	ctx      context.Context
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
		ctx:      ctx,
		db:       db,
		cerr:     make(chan error),
		getQueue: getQueue,
		registry: registry,
		clock: clock.NewClock(
			context.Background(),
			time.Nanosecond,
			func(err error) {
				// TODO: i don't know what to do with the errors
				logs.Error(ctx, "Error in scheduler", "error", err)
			},
		),
	}

	s.clock.Start()

	return s, nil
}

func (s *Scheduler) Stop() {
	// TODO: could be nice to have a status of the things that aren't closed
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				logs.Debug(s.ctx, "Scheduler", "total", s.clock.TotalSubscribers())
			}
		}
	}()
	s.clock.Stop()
	ticker.Stop()
}

func (s *Scheduler) addScheduler(tickerID clock.TickerID, ticker clock.Ticker, opts ...clock.TickerSubscriberOption) {
	logs.Debug(s.ctx, "Adding scheduler", "tickerID", tickerID)
	s.clock.Add(tickerID, ticker, clock.BestEffort, opts...)
}

func (s *Scheduler) AddQueue(queue string) {
	s.incID++
	logs.Debug(s.ctx, "Adding queue schedulers", "queue", queue)
	s.addScheduler(
		s.incID,
		SchedulerWorkflowsPending{
			Scheduler: s,
		},
		clock.WithName("SchedulerWorkflowsPending"))
}
