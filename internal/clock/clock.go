package clock

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// TickerID is a unique identifier for a ticker.
type TickerID interface{}

// Ticker defines the interface that tickers must implement.
type Ticker interface {
	Tick(ctx context.Context) error
}

// TickerOption is a function that configures a tickerEntry.
type TickerOption func(*tickerEntry)

// WithCleanup sets a cleanup function for the tickerEntry.
func WithCleanup(cleanup func()) TickerOption {
	return func(entry *tickerEntry) {
		entry.cleanup = cleanup
	}
}

// tickerEntry holds a ticker and its optional cleanup function.
type tickerEntry struct {
	id      TickerID
	ticker  Ticker
	cleanup func()
}

// Clock is the main struct that manages tickers and timing.
type Clock struct {
	mu       sync.RWMutex
	tickers  map[TickerID]*tickerEntry
	status   sync.Map // map[TickerID]bool
	interval time.Duration
	onError  func(error)
	done     chan struct{}
	wg       sync.WaitGroup
	inc      atomic.Int32

	ctx    context.Context
	cancel context.CancelFunc

	closing atomic.Bool
}

// NewClock creates a new Clock with the specified interval and error handler.
func NewClock(interval time.Duration, onError func(error)) *Clock {
	ctx, cancel := context.WithCancel(context.Background())
	return &Clock{
		ctx:      ctx,
		cancel:   cancel,
		tickers:  make(map[TickerID]*tickerEntry),
		interval: interval,
		onError:  onError,
		done:     make(chan struct{}),
		status:   sync.Map{},
	}
}

// AddTicker adds a ticker to the Clock with the given TickerID and options.
func (c *Clock) AddTicker(id TickerID, ticker Ticker, options ...TickerOption) {
	if c.closing.Load() {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	entry := &tickerEntry{
		id:     id,
		ticker: ticker,
	}
	for _, opt := range options {
		opt(entry)
	}
	c.status.Store(id, false)
	c.tickers[id] = entry
}

// RemoveTicker removes a ticker from the Clock using its TickerID.
func (c *Clock) RemoveTicker(id TickerID) {
	if c.closing.Load() {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if entry, ok := c.tickers[id]; ok {
		if entry.cleanup != nil {
			entry.cleanup()
		}
		delete(c.tickers, id)
	}
}

func (c *Clock) GetTickers() []TickerID {
	cp := []TickerID{}
	c.mu.RLock()
	tickers := make([]*tickerEntry, 0, len(c.tickers))
	for _, entry := range c.tickers {
		tickers = append(tickers, entry)
	}
	c.mu.RUnlock()
	for k, _ := range tickers {
		cp = append(cp, k)
	}
	return cp
}

func (c *Clock) GetMapTickers() map[TickerID]bool {
	cp := map[TickerID]bool{}
	c.mu.RLock()
	tickers := make([]*tickerEntry, 0, len(c.tickers))
	for _, entry := range c.tickers {
		tickers = append(tickers, entry)
	}
	c.mu.RUnlock()
	for _, t := range tickers {
		c.mu.RLock()
		running, ok := c.status.Load(t.id)
		c.mu.RUnlock()
		if ok {
			cp[t.id] = running.(bool)
		}
	}
	return cp
}

// Start begins the ticking process of the Clock.
func (c *Clock) Start() {
	if c.closing.Load() {
		return
	}
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer c.cancel()
		nextTick := time.Now()
		for {
			select {
			case <-c.ctx.Done():
				// fmt.Println("\tCLOCK DONE")
				c.cleanupTickers()
				return
			case <-c.done:
				// fmt.Println("\tCLOCK STOPPED")
				c.cleanupTickers()
				return
			default:
				now := time.Now()
				if !now.Before(nextTick) {
					c.triggerTickers(c.ctx)
					nextTick = nextTick.Add(c.interval)
					if nextTick.Before(now) {
						nextTick = now.Add(c.interval)
					}
				}
				sleepDuration := nextTick.Sub(time.Now())
				if sleepDuration > 0 {
					select {
					case <-time.After(sleepDuration):
					case <-c.done:
						c.cleanupTickers()
						return
					}
				} else {
					runtime.Gosched()
				}
			}
		}
	}()
}

func (c *Clock) HasTickerID(id TickerID) bool {
	c.mu.RLock()
	_, ok := c.tickers[id]
	c.mu.RUnlock()
	return ok
}

// Stop halts the ticking process and cleans up resources.
func (c *Clock) Stop() {
	c.closing.Store(true)
	close(c.done)
	c.cancel()
	c.wg.Wait()
}

// triggerTickers invokes the Tick method on all registered tickers.
func (c *Clock) triggerTickers(ctx context.Context) {
	c.mu.RLock()
	tickers := make([]*tickerEntry, 0, len(c.tickers))
	for _, entry := range c.tickers {
		tickers = append(tickers, entry)
	}
	c.mu.RUnlock()

	for _, entry := range tickers {

		if _, ok := c.status.Load(entry.id); ok {
			c.status.Store(entry.id, true)
		}

		if err := entry.ticker.Tick(ctx); err != nil && c.onError != nil {
			c.onError(err)
		}

		if _, ok := c.status.Load(entry.id); ok {
			c.status.Store(entry.id, false)
		}

	}

}

// cleanupTickers calls the cleanup function for all registered tickers.
func (c *Clock) cleanupTickers() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, entry := range c.tickers {
		if entry.cleanup != nil {
			entry.cleanup()
		}
	}
}
