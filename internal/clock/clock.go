package clock

import (
	"context"
	"runtime"
	"sync"
	"time"
)

type Ticker interface {
	Tick() error
}

type TickerID interface{}

type ExecutionMode int

type removeTask struct {
	id   TickerID
	done chan struct{}
}

const (
	NonBlocking ExecutionMode = iota
	ManagedTimeline
	BestEffort
)

type TickerSubscriber struct {
	ID              TickerID
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
	removeCh chan removeTask
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
	onError  func(error)
}

func NewClock(ctx context.Context, interval time.Duration, onError func(error)) *Clock {
	ctx, cancel := context.WithCancel(ctx)

	c := &Clock{
		interval: interval,
		ticker:   time.NewTicker(interval),
		stopCh:   make(chan struct{}),
		subs:     []TickerSubscriber{},
		removeCh: make(chan removeTask, 100),
		ctx:      ctx,
		cancel:   cancel,
		onError:  onError,
	}
	go c.handleRemovals()
	return c
}

func (c *Clock) Add(id TickerID, ticker Ticker, mode ExecutionMode, opts ...TickerSubscriberOption) {
	c.mu.Lock()
	defer c.mu.Unlock()

	sub := TickerSubscriber{
		ID:     id,
		Ticker: ticker,
		Mode:   mode,
	}

	for _, opt := range opts {
		opt(&sub)
	}

	c.subs = append(c.subs, sub)
}

func (c *Clock) handleRemovals() {
	for {
		select {
		case task := <-c.removeCh:
			c.mu.Lock()
			for i, sub := range c.subs {
				if sub.ID == task.id {
					c.subs = append(c.subs[:i], c.subs[i+1:]...)
					break
				}
			}
			close(task.done)
			c.mu.Unlock()
		case <-c.ctx.Done():
			return
		default:
			runtime.Gosched()
		}
	}
}

func (c *Clock) Remove(id TickerID) chan struct{} {
	done := make(chan struct{})
	select {
	case c.removeCh <- removeTask{id: id, done: done}:
	case <-c.ctx.Done():
	default:
		// Fallback to synchronous removal
		c.mu.Lock()
		for i, sub := range c.subs {
			if sub.ID == id {
				c.subs = append(c.subs[:i], c.subs[i+1:]...)
				break
			}
		}
		c.mu.Unlock()
	}
	return done
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
	c.mu.RLock()
	snapshot := make([]TickerSubscriber, len(c.subs))
	copy(snapshot, c.subs)
	c.mu.RUnlock()

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
