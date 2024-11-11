package info

import (
	"context"
	"fmt"
	"time"

	"github.com/davidroman0O/tempolite/internal/clock"
	"github.com/davidroman0O/tempolite/pkg/logs"
)

type InfoClock struct {
	ctx   context.Context
	clock *clock.Clock
}

func New(ctx context.Context) *InfoClock {
	infoClock := &InfoClock{
		ctx: ctx,
		clock: clock.NewClock(
			ctx,
			time.Nanosecond,
			func(err error) {
				fmt.Println(err)
			},
		),
	}
	logs.Debug(ctx, "Starting info clock")
	infoClock.clock.Start()
	return infoClock
}

func (i *InfoClock) Stop() {
	logs.Debug(i.ctx, "Stopping info clock")
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				logs.Debug(i.ctx, "Scheduler ", "total", i.clock.TotalSubscribers())
			}
		}
	}()
	i.clock.Stop()
	ticker.Stop()
	logs.Debug(i.ctx, "Info clock stopped")
}

func (i *InfoClock) AddInfo(id clock.TickerID, ticker clock.Ticker) {
	logs.Debug(i.ctx, "Adding info", "tickerID", id)
	i.clock.Add(id, ticker, clock.BestEffort)
}

func (i *InfoClock) Remove(id clock.TickerID) chan struct{} {
	logs.Debug(i.ctx, "Removing info", "tickerID", id)
	return i.clock.Remove(id)
}
