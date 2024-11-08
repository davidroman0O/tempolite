package clock

import (
	"context"
	"sync"
	"time"
)

type Ticker interface {
	Tick() error
}

type ExecutionMode int

const (
	NonBlocking ExecutionMode = iota
	ManagedTimeline
	BestEffort
)

type TickerSubscriber struct {
	Ticker          Ticker
	Mode            ExecutionMode
	LastExecTime    time.Time
	Priority        int
	DynamicInterval func(elapsedTime time.Duration) time.Duration
	Interval        time.Duration
	Name            string
	OnError         func(error)
}

type TickerSubscriberOption func(*TickerSubscriber)

func WithPriority(priority int) TickerSubscriberOption {
	return func(ts *TickerSubscriber) {
		ts.Priority = priority
	}
}

func WithDynamicInterval(dynamicInterval func(elapsedTime time.Duration) time.Duration) TickerSubscriberOption {
	return func(ts *TickerSubscriber) {
		ts.DynamicInterval = dynamicInterval
	}
}

func WithInterval(interval time.Duration) TickerSubscriberOption {
	return func(ts *TickerSubscriber) {
		ts.Interval = interval
	}
}

func WithName(name string) TickerSubscriberOption {
	return func(ts *TickerSubscriber) {
		ts.Name = name
	}
}

func WithOnError(onError func(error)) TickerSubscriberOption {
	return func(ts *TickerSubscriber) {
		ts.OnError = onError
	}
}

type Clock struct {
	interval time.Duration
	ticker   *time.Ticker
	stopCh   chan struct{}
	subs     []TickerSubscriber
	mu       sync.Mutex
	ctx      context.Context
	cancel   context.CancelFunc
	onError  func(error)
}

func NewClock(ctx context.Context, interval time.Duration, onError func(error)) *Clock {
	ctx, cancel := context.WithCancel(ctx)
	return &Clock{
		interval: interval,
		ticker:   time.NewTicker(interval),
		stopCh:   make(chan struct{}),
		subs:     []TickerSubscriber{},
		ctx:      ctx,
		cancel:   cancel,
		onError:  onError,
	}
}

func (c *Clock) Add(ticker Ticker, mode ExecutionMode, opts ...TickerSubscriberOption) {
	c.mu.Lock()
	defer c.mu.Unlock()

	sub := TickerSubscriber{
		Ticker: ticker,
		Mode:   mode,
	}

	for _, opt := range opts {
		opt(&sub)
	}

	c.subs = append(c.subs, sub)
}

func (c *Clock) Remove(ticker Ticker) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, sub := range c.subs {
		if sub.Ticker == ticker {
			c.subs = append(c.subs[:i], c.subs[i+1:]...)
			break
		}
	}
}

func (c *Clock) Start() {
	go c.dispatchTicks()
}

func (c *Clock) Stop() {
	c.cancel()
	c.ticker.Stop()
	close(c.stopCh)
}

func (c *Clock) dispatchTicks() {
	for {
		select {
		case <-c.ticker.C:
			c.tick()
		case <-c.stopCh:
			return
		case <-c.ctx.Done():
			return
		}
	}
}

func (c *Clock) tick() {
	now := time.Now()
	c.mu.Lock()
	defer c.mu.Unlock()

	for i := range c.subs {
		sub := &c.subs[i]

		interval := c.interval
		if sub.Interval > 0 {
			interval = sub.Interval
		}
		if sub.DynamicInterval != nil {
			elapsedTime := now.Sub(sub.LastExecTime)
			interval = sub.DynamicInterval(elapsedTime)
		}

		switch sub.Mode {
		case NonBlocking:
			go func(sub *TickerSubscriber) {
				if err := sub.Ticker.Tick(); err != nil {
					if sub.OnError != nil {
						sub.OnError(err)
					} else if c.onError != nil {
						c.onError(err)
					}
				}
			}(sub)
		case ManagedTimeline:
			if now.Sub(sub.LastExecTime) >= interval {
				if err := sub.Ticker.Tick(); err != nil {
					if sub.OnError != nil {
						sub.OnError(err)
					} else if c.onError != nil {
						c.onError(err)
					}
				}
				sub.LastExecTime = now
			} else {
				go func(sub *TickerSubscriber) {
					time.Sleep(interval - now.Sub(sub.LastExecTime))
					if err := sub.Ticker.Tick(); err != nil {
						if sub.OnError != nil {
							sub.OnError(err)
						} else if c.onError != nil {
							c.onError(err)
						}
					}
					sub.LastExecTime = time.Now()
				}(sub)
			}
		case BestEffort:
			if now.Sub(sub.LastExecTime) >= interval {
				if err := sub.Ticker.Tick(); err != nil {
					if sub.OnError != nil {
						sub.OnError(err)
					} else if c.onError != nil {
						c.onError(err)
					}
				}
				sub.LastExecTime = now
			}
		}
	}
}
