// Package chute provides high-performance lock-free multicast queues
package chute

import (
	"math/bits"
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"
)

const (
	blockSize     = 4096
	cacheLineSize = 64
	maxRetries    = 100 // For spinning operations
)

// CacheAligned helps prevent false sharing
type CacheAligned[T any] struct {
	value T
	_     [cacheLineSize - unsafe.Sizeof(atomic.Uint64{}) - (unsafe.Sizeof(atomic.Uint64{}) % cacheLineSize)]byte
}

// block represents a single block in the queue
type block[T any] struct {
	len       CacheAligned[atomic.Uint64] // Number of elements in block
	useCount  CacheAligned[atomic.Uint64] // Reference count
	next      CacheAligned[atomic.Pointer[block[T]]]
	bitBlocks [blockSize / 64]atomic.Uint64 // Bitmap for MPMC
	mem       [blockSize]T                  // Actual data storage
	pool      *sync.Pool
}

// blockPool manages a pool of blocks
type blockPool[T any] struct {
	pool sync.Pool
}

func newBlockPool[T any]() *blockPool[T] {
	bp := &blockPool[T]{}
	bp.pool.New = func() interface{} {
		return &block[T]{}
	}
	return bp
}

func (bp *blockPool[T]) get(counter uint64) *block[T] {
	b := bp.pool.Get().(*block[T])
	b.len.value.Store(0)
	b.useCount.value.Store(counter)
	b.next.value.Store(nil)
	// Clear only the first bitBlock as we'll start writing there
	b.bitBlocks[0].Store(0)
	b.pool = &bp.pool
	return b
}

type Queue[T any] struct {
	lastBlock CacheAligned[atomic.Pointer[block[T]]]
	pool      *blockPool[T]
}

func NewQueue[T any]() *Queue[T] {
	q := &Queue[T]{
		pool: newBlockPool[T](),
	}
	firstBlock := q.pool.get(1)
	q.lastBlock.value.Store(firstBlock)
	return q
}

// tryLockLastBlock attempts to atomically swap the last block pointer
func (q *Queue[T]) tryLockLastBlock() *block[T] {
	return q.lastBlock.value.Swap(nil)
}

// lockLastBlock spins until it can acquire the last block
func (q *Queue[T]) lockLastBlock() *block[T] {
	for i := 0; ; i++ {
		if b := q.tryLockLastBlock(); b != nil {
			return b
		}
		if i < maxRetries {
			runtime.Gosched()
		} else {
			runtime.Gosched()
			i = 0
		}
	}
}

func (q *Queue[T]) unlockLastBlock(b *block[T]) {
	q.lastBlock.value.Store(b)
}

func (b *block[T]) recycleBlock() {
	if next := b.next.value.Load(); next != nil {
		next.decUseCount()
	}
	if b.pool != nil {
		b.pool.Put(b)
	}
}

func (b *block[T]) incUseCount() {
	b.useCount.value.Add(1)
}

func (b *block[T]) decUseCount() {
	if b.useCount.value.Add(^uint64(0)) == 0 {
		b.recycleBlock()
	}
}

// Push adds an item to the queue with optimized retry logic
func (q *Queue[T]) Push(value T) {
	for i := 0; ; i++ {
		b := q.lockLastBlock()

		if q.tryPush(b, value) {
			q.unlockLastBlock(b)
			return
		}

		// Current block is full, try to add new block
		newB := q.pool.get(3)
		b.next.value.Store(newB)
		b.decUseCount()

		if q.tryPush(newB, value) {
			q.unlockLastBlock(newB)
			return
		}

		if i < maxRetries {
			runtime.Gosched()
		} else {
			runtime.Gosched()
			i = 0
		}
	}
}

// tryPush optimized for better memory access patterns
func (q *Queue[T]) tryPush(b *block[T], value T) bool {
	idx := b.len.value.Load()
	if idx >= blockSize {
		return false
	}

	// Store value before updating metadata
	b.mem[idx] = value

	// Update bitmap using a single atomic operation
	blockIdx := idx / 64
	bitIdx := idx % 64
	mask := uint64(1) << bitIdx
	b.bitBlocks[blockIdx].Or(mask)

	// Update length last
	b.len.value.Store(idx + 1)
	return true
}

type Reader[T any] struct {
	block         *block[T]
	index         uint64
	len           uint64
	bitBlockIndex uint64
}

// NewReader creates a new reader with proper synchronization
func (q *Queue[T]) NewReader() *Reader[T] {
	b := q.lockLastBlock()
	if b == nil {
		b = q.pool.get(1)
	}
	b.incUseCount() // Increment before creating reader
	len := b.len.value.Load()
	q.unlockLastBlock(b)

	return &Reader[T]{
		block:         b,
		index:         len,
		len:           len,
		bitBlockIndex: len / 64,
	}
}

// Optimized bit operations
func getTrailingOnes(value uint64) uint64 {
	return uint64(bits.TrailingZeros64(^value))
}

// Next returns the next item with optimized block transitions
func (r *Reader[T]) Next() (T, bool) {
	var zero T

	if r.block == nil {
		return zero, false
	}

	if r.index == r.len {
		if r.len == blockSize {
			// Move to next block
			if next := r.block.next.value.Load(); next != nil {
				next.incUseCount()
				r.block.decUseCount()
				r.block = next
				r.index = 0

				bits := next.bitBlocks[0].Load()
				r.len = getTrailingOnes(bits)
				r.bitBlockIndex = 0

				if bits == ^uint64(0) {
					r.bitBlockIndex = 1
				}

				if r.len == 0 {
					return zero, false
				}
			} else {
				return zero, false
			}
		} else {
			// Check current block for new items
			if r.bitBlockIndex >= blockSize/64 {
				return zero, false
			}

			bits := r.block.bitBlocks[r.bitBlockIndex].Load()
			newLen := r.bitBlockIndex*64 + getTrailingOnes(bits)

			if r.len == newLen {
				return zero, false
			}

			if bits == ^uint64(0) {
				r.bitBlockIndex++
			}

			r.len = newLen
		}
	}

	if r.index >= blockSize {
		return zero, false
	}

	value := r.block.mem[r.index]
	r.index++
	return value, true
}

func (r *Reader[T]) Clone() *Reader[T] {
	if r.block == nil {
		return nil
	}
	r.block.incUseCount()
	return &Reader[T]{
		block:         r.block,
		index:         r.index,
		len:           r.len,
		bitBlockIndex: r.bitBlockIndex,
	}
}

func (r *Reader[T]) Close() {
	if r.block != nil {
		r.block.decUseCount()
		r.block = nil
	}
}
