package clock

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type TickerID interface{}

type Ticker interface {
	Tick(ctx context.Context) error
}

type TickerOption func(*tickerEntry)

func WithCleanupCallback(cleanup func()) TickerOption {
	return func(te *tickerEntry) {
		te.cleanup = cleanup
	}
}

type tickerEntry struct {
	id      TickerID
	ticker  Ticker
	cleanup func()
}

type Clock struct {
	ctx              context.Context
	cancel           context.CancelFunc
	name             string
	interval         time.Duration
	onError          func(error)
	totalSubscribers int64
	done             chan struct{}

	activeTickers sync.Map // map[TickerID]*tickerEntry

	tickerAdd    chan *tickerEntry
	tickerRemove chan TickerID
}

type ClockOption func(*Clock)

func WithName(name string) ClockOption {
	return func(c *Clock) {
		c.name = name
	}
}

func NewClock(ctx context.Context, interval time.Duration, onError func(error), opts ...ClockOption) *Clock {
	if interval <= 0 {
		interval = time.Second
	}

	ctx, cancel := context.WithCancel(ctx)
	c := &Clock{
		ctx:          ctx,
		cancel:       cancel,
		name:         "clock",
		interval:     interval,
		onError:      onError,
		done:         make(chan struct{}),
		tickerAdd:    make(chan *tickerEntry, 1000), // Increased buffer size
		tickerRemove: make(chan TickerID, 1000),     // Increased buffer size
	}

	for _, opt := range opts {
		opt(c)
	}

	fmt.Println("Clock created", "name", c.name, "interval", interval)
	return c
}

func (c *Clock) AddTicker(id TickerID, ticker Ticker, opts ...TickerOption) {
	if ticker == nil {
		fmt.Println("Ticker cannot be nil", "name", c.name)
		return
	}

	te := &tickerEntry{
		id:     id,
		ticker: ticker,
	}
	for _, opt := range opts {
		opt(te)
	}

	// Non-blocking add
	select {
	case c.tickerAdd <- te:
	default:
		go func() { c.tickerAdd <- te }()
	}
}

func (c *Clock) RemoveTicker(id TickerID) {
	// Non-blocking remove
	select {
	case c.tickerRemove <- id:
	default:
		go func() { c.tickerRemove <- id }()
	}
}

func (c *Clock) TotalSubscribers() int {
	var count int
	c.activeTickers.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

func (c *Clock) Start() {
	go c.run()
}

func (c *Clock) run() {
	defer close(c.done)

	nextTick := time.Now()

	for {
		now := time.Now()
		if !now.Before(nextTick) {
			// Execute Tick on all active tickers
			c.activeTickers.Range(func(_, value interface{}) bool {
				te, ok := value.(*tickerEntry)
				if !ok {
					return true
				}
				if err := te.ticker.Tick(c.ctx); err != nil && c.onError != nil {
					c.onError(err)
				}
				return true
			})
			nextTick = nextTick.Add(c.interval)
			// Adjust nextTick if we've fallen behind
			if nextTick.Before(now) {
				nextTick = now.Add(c.interval)
			}
		}

		// Non-blocking read from tickerAdd and tickerRemove channels
		select {
		case te := <-c.tickerAdd:
			if existing, ok := c.activeTickers.Load(te.id); ok {
				if existingTe, ok := existing.(*tickerEntry); ok && existingTe.cleanup != nil {
					existingTe.cleanup()
				}
			} else {
				atomic.AddInt64(&c.totalSubscribers, 1)
			}
			c.activeTickers.Store(te.id, te)
		case id := <-c.tickerRemove:
			if value, ok := c.activeTickers.Load(id); ok {
				if te, ok := value.(*tickerEntry); ok && te.cleanup != nil {
					te.cleanup()
				}
				c.activeTickers.Delete(id)
				atomic.AddInt64(&c.totalSubscribers, -1)
			}
		case <-c.ctx.Done():
			c.activeTickers.Range(func(key, value interface{}) bool {
				if te, ok := value.(*tickerEntry); ok && te.cleanup != nil {
					te.cleanup()
				}
				c.activeTickers.Delete(key)
				return true
			})
			return
		default:
			// Sleep very briefly to yield processor
			time.Sleep(time.Nanosecond)
		}
	}
}

func (c *Clock) Stop() {
	c.cancel()
	<-c.done
	fmt.Println("Clock has stopped")
}
