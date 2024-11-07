package schedulers

import (
	"context"
	"time"
)

/// TODO: big big mutex problem

var DefaultClock = NewClock(context.Background(), WithInterval(100*time.Millisecond))

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
	lastExecTime    time.Time
	Priority        int
	DynamicInterval func(elapsedTime time.Duration) time.Duration
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

type Clock struct {
	name     string
	interval time.Duration
	ticker   *time.Ticker
	stopCh   chan struct{}
	subs     []TickerSubscriber
	ticking  bool
	started  bool
	// mu       sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc
	cerr   chan error

	onError func(error)
}

func (tm *Clock) Ticking() bool {
	// tm.mu.Lock()
	// defer tm.mu.Unlock()
	return tm.ticking
}

type ClockOption func(*Clock)

func WithName(name string) ClockOption {
	return func(c *Clock) {
		c.name = name
	}
}

func WithInterval(interval time.Duration) ClockOption {
	return func(c *Clock) {
		c.interval = interval
	}
}

func WithOnError(onError func(error)) ClockOption {
	return func(c *Clock) {
		c.onError = onError
	}
}

func NewClock(ctx context.Context, opts ...ClockOption) *Clock {
	ctx, cancel := context.WithCancel(ctx)
	tm := &Clock{
		cerr:   make(chan error),
		stopCh: make(chan struct{}),
		subs:   []TickerSubscriber{},
		ctx:    ctx,
		cancel: cancel,
	}
	for _, opt := range opts {
		opt(tm)
	}
	if tm.interval == 0 {
		tm.interval = 100 * time.Millisecond
	}
	tm.ticker = time.NewTicker(tm.interval)
	return tm
}

func (tm *Clock) Start() {
	// tm.mu.Lock()
	// defer tm.mu.Unlock()

	if tm.started {
		if !tm.ticking {
			go tm.dispatchTicks()
		}
		return
	}
	tm.started = true
	go tm.dispatchTicks()
}

func (tm *Clock) Clear() {
	// tm.mu.Lock()
	// defer tm.mu.Unlock()
	tm.subs = []TickerSubscriber{}
}

func (tm *Clock) Add(rb Ticker, mode ExecutionMode, opts ...TickerSubscriberOption) {
	// tm.mu.Lock()
	// defer tm.mu.Unlock()

	sub := TickerSubscriber{
		Ticker: rb,
		Mode:   mode,
	}

	for _, opt := range opts {
		opt(&sub)
	}

	tm.subs = append(tm.subs, sub)
}

func (tm *Clock) Remove(rb Ticker) {
	// tm.mu.Lock()
	// defer tm.mu.Unlock()

	for i, sub := range tm.subs {
		if sub.Ticker == rb {
			tm.subs = append(tm.subs[:i], tm.subs[i+1:]...)
			break
		}
	}
}

func (tm *Clock) dispatchTicks() {
	// tm.mu.Lock()
	tm.ticking = true
	// tm.mu.Unlock()

	defer func() {
		close(tm.stopCh)
		close(tm.cerr)
	}()

	for {
		select {
		case err := <-tm.cerr:
			if tm.onError != nil {
				tm.onError(err)
			}
		case <-tm.ticker.C:
			now := time.Now()
			// tm.mu.Lock()
			for i := range tm.subs {
				sub := &tm.subs[i]

				interval := tm.interval
				if sub.DynamicInterval != nil {
					elapsedTime := now.Sub(sub.lastExecTime)
					interval = sub.DynamicInterval(elapsedTime)
				}

				switch sub.Mode {
				case NonBlocking:
					go func() {
						tm.cerr <- sub.Ticker.Tick()
					}()
				case ManagedTimeline, BestEffort:
					if now.Sub(sub.lastExecTime) >= interval {
						tm.cerr <- sub.Ticker.Tick()
						sub.lastExecTime = now
					}
				}
			}
			// tm.mu.Unlock()

		case <-tm.stopCh:
			tm.ticker.Stop()
			tm.ticking = false
			return

		case <-tm.ctx.Done():
			tm.ticker.Stop()
			tm.ticking = false
			return
		}
	}
}

func (tm *Clock) Pause() {
	// tm.mu.Lock()
	tm.ticking = false
	// tm.mu.Unlock()
	tm.ticker.Stop()
}

func (tm *Clock) Stop() {
	// tm.mu.Lock()
	tm.cancel()
	// tm.mu.Unlock()
	tm.ticker.Stop()
	tm.ticking = false
}
