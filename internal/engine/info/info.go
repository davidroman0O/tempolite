package info

import (
	"context"
	"fmt"
	"time"

	"github.com/davidroman0O/tempolite/internal/clock"
)

type InfoClock struct {
	clock *clock.Clock
}

func New(ctx context.Context) *InfoClock {
	infoClock := &InfoClock{
		clock: clock.NewClock(
			ctx,
			time.Nanosecond,
			func(err error) {
				fmt.Println(err)
			},
		),
	}
	infoClock.clock.Start()
	return infoClock
}

func (i *InfoClock) Stop() {
	fmt.Println("Stopping info clock")
	i.clock.Stop()
}

func (i *InfoClock) AddInfo(id clock.TickerID, ticker clock.Ticker) {
	i.clock.Add(id, ticker, clock.BestEffort)
}

func (i *InfoClock) Remove(id clock.TickerID) chan struct{} {
	return i.clock.Remove(id)
}
