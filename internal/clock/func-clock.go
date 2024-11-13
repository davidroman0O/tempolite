package clock

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// FuncID is a unique identifier for a function ticker
type FuncID interface{}

// FuncOption configures a funcEntry
type FuncOption func(*funcEntry)

// WithFuncCleanup sets a cleanup function for the funcEntry
func WithFuncCleanup(cleanup func()) FuncOption {
	return func(entry *funcEntry) {
		entry.cleanup = cleanup
	}
}

// funcEntry holds a function pointer and its cleanup function
type funcEntry struct {
	id      FuncID
	fn      interface{} // Will verify this is func(context.Context) error
	cleanup func()
}

// FuncClock manages function pointers that implement tick behavior
type FuncClock struct {
	mu       sync.RWMutex
	funcs    map[FuncID]*funcEntry
	status   sync.Map // map[FuncID]bool
	interval time.Duration
	onError  func(error)
	done     chan struct{}
	wg       sync.WaitGroup
	inc      atomic.Int32

	ctx    context.Context
	cancel context.CancelFunc

	closing atomic.Bool
}

// NewFuncClock creates a new FuncClock with specified interval and error handler
func NewFuncClock(interval time.Duration, onError func(error)) *FuncClock {
	ctx, cancel := context.WithCancel(context.Background())
	return &FuncClock{
		ctx:      ctx,
		cancel:   cancel,
		funcs:    make(map[FuncID]*funcEntry),
		interval: interval,
		onError:  onError,
		done:     make(chan struct{}),
		status:   sync.Map{},
	}
}

// AddFunc adds a function pointer to the clock
// The function must have signature: func(context.Context) error
func (c *FuncClock) AddFunc(id FuncID, fn interface{}, options ...FuncOption) error {
	if c.closing.Load() {
		return nil
	}

	// Verify function signature
	fnType := reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		return fmt.Errorf("fn must be a function")
	}
	if fnType.NumIn() != 1 || fnType.NumOut() != 1 {
		return fmt.Errorf("fn must have signature: func(context.Context) error")
	}
	if !fnType.In(0).Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
		return fmt.Errorf("first parameter must be context.Context")
	}
	if !fnType.Out(0).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		return fmt.Errorf("return value must be error")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	entry := &funcEntry{
		id: id,
		fn: fn,
	}
	for _, opt := range options {
		opt(entry)
	}
	c.status.Store(id, false)
	c.funcs[id] = entry

	return nil
}

// RemoveFunc removes a function from the clock
func (c *FuncClock) RemoveFunc(id FuncID) {
	if c.closing.Load() {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if entry, ok := c.funcs[id]; ok {
		if entry.cleanup != nil {
			entry.cleanup()
		}
		delete(c.funcs, id)
	}
}

func (c *FuncClock) GetFuncs() []FuncID {
	cp := []FuncID{}
	c.mu.RLock()
	funcs := make([]*funcEntry, 0, len(c.funcs))
	for _, entry := range c.funcs {
		funcs = append(funcs, entry)
	}
	c.mu.RUnlock()
	for k := range funcs {
		cp = append(cp, k)
	}
	return cp
}

func (c *FuncClock) GetMapFuncs() map[FuncID]bool {
	cp := map[FuncID]bool{}
	c.mu.RLock()
	funcs := make([]*funcEntry, 0, len(c.funcs))
	for _, entry := range c.funcs {
		funcs = append(funcs, entry)
	}
	c.mu.RUnlock()
	for _, f := range funcs {
		c.mu.RLock()
		running, ok := c.status.Load(f.id)
		c.mu.RUnlock()
		if ok {
			cp[f.id] = running.(bool)
		}
	}
	return cp
}

// Start begins the ticking process
func (c *FuncClock) Start() {
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
				c.cleanupFuncs()
				return
			case <-c.done:
				c.cleanupFuncs()
				return
			default:
				now := time.Now()
				if !now.Before(nextTick) {
					c.triggerFuncs(c.ctx)
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
						c.cleanupFuncs()
						return
					}
				} else {
					runtime.Gosched()
				}
			}
		}
	}()
}

func (c *FuncClock) HasFuncID(id FuncID) bool {
	c.mu.RLock()
	_, ok := c.funcs[id]
	c.mu.RUnlock()
	return ok
}

// Stop halts the ticking process
func (c *FuncClock) Stop() {
	c.closing.Store(true)
	close(c.done)
	c.cancel()
	c.wg.Wait()
}

// triggerFuncs calls all registered functions
func (c *FuncClock) triggerFuncs(ctx context.Context) {
	c.mu.RLock()
	funcs := make([]*funcEntry, 0, len(c.funcs))
	for _, entry := range c.funcs {
		funcs = append(funcs, entry)
	}
	c.mu.RUnlock()

	for _, entry := range funcs {
		if _, ok := c.status.Load(entry.id); ok {
			c.status.Store(entry.id, true)
		}

		// Call the function via reflection
		fnVal := reflect.ValueOf(entry.fn)
		args := []reflect.Value{reflect.ValueOf(ctx)}
		results := fnVal.Call(args)

		// Check for error
		if err := results[0].Interface(); err != nil {
			if errVal, ok := err.(error); ok && c.onError != nil {
				c.onError(errVal)
			}
		}

		if _, ok := c.status.Load(entry.id); ok {
			c.status.Store(entry.id, false)
		}
	}
}

// cleanupFuncs calls cleanup functions for all registered functions
func (c *FuncClock) cleanupFuncs() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, entry := range c.funcs {
		if entry.cleanup != nil {
			entry.cleanup()
		}
	}
}
