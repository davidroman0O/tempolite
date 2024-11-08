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
	defer fmt.Println("Info clock stopped")
}

func (i *InfoClock) AddInfo(ticker clock.Ticker) {
	i.clock.Add(ticker, clock.BestEffort)
}
