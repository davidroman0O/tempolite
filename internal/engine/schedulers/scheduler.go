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
			ctx,
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
	logs.Debug(s.ctx, "Stopping scheduler")
	s.clock.Stop()
	logs.Debug(s.ctx, "Scheduler stopped")
}

func (s *Scheduler) addScheduler(tickerID clock.TickerID, ticker clock.Ticker) {
	logs.Debug(s.ctx, "Adding scheduler", "tickerID", tickerID)
	s.clock.Add(tickerID, ticker, clock.BestEffort)
}

func (s *Scheduler) AddQueue(queue string) {
	s.incID++
	logs.Debug(s.ctx, "Adding queue schedulers", "queue", queue)
	s.addScheduler(
		s.incID,
		SchedulerWorkflowsPending{
			Scheduler: s,
		})
}
