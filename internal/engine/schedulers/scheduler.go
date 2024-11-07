package schedulers

import (
	"context"
	"time"

	"github.com/davidroman0O/tempolite/internal/persistence/repository"
)

type Scheduler struct {
	clock *Clock
	db    repository.Repository
	cerr  chan error
}

func New(
	ctx context.Context,
	db repository.Repository,
) (*Scheduler, error) {
	s := &Scheduler{
		db:   db,
		cerr: make(chan error),
		clock: NewClock(
			ctx,
			WithInterval(time.Millisecond), WithName("scheduler")),
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
			db: s.db,
		})
}
