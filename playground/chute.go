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
)

type alignedCounter struct {
	value atomic.Uint64
	_     [cacheLineSize - unsafe.Sizeof(atomic.Uint64{})]byte
}

type alignedPtr[T any] struct {
	value atomic.Pointer[T]
	_     [cacheLineSize - unsafe.Sizeof(atomic.Pointer[struct{}]{})]byte
}

type block[T any] struct {
	len       alignedCounter
	useCount  alignedCounter
	next      alignedPtr[block[T]]
	pool      *sync.Pool
	bitBlocks [blockSize / 64]atomic.Uint64
	mem       [blockSize]T
}

func (b *block[T]) incRef() uint64 {
	return b.useCount.value.Add(1)
}

func (b *block[T]) decRef() {
	prev := b.useCount.value.Add(^uint64(0))
	if prev == 1 {
		if next := b.next.value.Load(); next != nil {
			next.decRef()
		}
		if b.pool != nil {
			b.len.value.Store(0)
			b.useCount.value.Store(0)
			b.next.value.Store(nil)
			for i := range b.bitBlocks {
				b.bitBlocks[i].Store(0)
			}
			b.pool.Put(b)
		}
	}
}

type SPMCQueue[T any] struct {
	lastBlock alignedPtr[block[T]]
	pool      *sync.Pool
}

func NewSPMCQueue[T any]() *SPMCQueue[T] {
	q := &SPMCQueue[T]{
		pool: &sync.Pool{
			New: func() interface{} {
				return &block[T]{}
			},
		},
	}
	firstBlock := q.newBlock(1)
	q.lastBlock.value.Store(firstBlock)
	return q
}

func (q *SPMCQueue[T]) newBlock(refCount uint64) *block[T] {
	b := q.pool.Get().(*block[T])
	b.len.value.Store(0)
	b.useCount.value.Store(refCount)
	b.next.value.Store(nil)
	b.pool = q.pool
	return b
}

func (q *SPMCQueue[T]) Push(value T) {
	b := q.lastBlock.value.Load()
	idx := b.len.value.Load()

	if idx >= blockSize {
		newB := q.newBlock(2)
		b.next.value.Store(newB)
		q.lastBlock.value.Store(newB)
		b.decRef()
		b = newB
		idx = 0
	}

	b.mem[idx] = value
	b.len.value.Store(idx + 1)
}

type MPMCQueue[T any] struct {
	lastBlock alignedPtr[block[T]]
	pool      *sync.Pool
}

func NewMPMCQueue[T any]() *MPMCQueue[T] {
	q := &MPMCQueue[T]{
		pool: &sync.Pool{
			New: func() interface{} {
				return &block[T]{}
			},
		},
	}
	firstBlock := q.newBlock(1)
	q.lastBlock.value.Store(firstBlock)
	return q
}

func (q *MPMCQueue[T]) newBlock(refCount uint64) *block[T] {
	b := q.pool.Get().(*block[T])
	b.len.value.Store(0)
	b.useCount.value.Store(refCount)
	b.next.value.Store(nil)
	for i := range b.bitBlocks {
		b.bitBlocks[i].Store(0)
	}
	b.pool = q.pool
	return b
}

func (q *MPMCQueue[T]) Push(value T) {
	for {
		b := q.tryLockLastBlock()
		if b == nil {
			runtime.Gosched()
			continue
		}

		idx := b.len.value.Add(1) - 1
		if idx >= blockSize {
			b.len.value.Add(^uint64(0))
			newB := q.newBlock(3)
			b.next.value.Store(newB)
			q.lastBlock.value.Store(newB)
			b.decRef()
			q.unlockLastBlock(newB)
			continue
		}

		b.mem[idx] = value
		blockIdx := idx / 64
		bitIdx := idx % 64
		b.bitBlocks[blockIdx].Or(1 << bitIdx)
		q.unlockLastBlock(b)
		return
	}
}

func (q *MPMCQueue[T]) tryLockLastBlock() *block[T] {
	for i := uint(0); i < 32; i++ {
		if b := q.lastBlock.value.Swap(nil); b != nil {
			return b
		}
		if i > 10 {
			runtime.Gosched()
		} else {
			for j := uint(0); j < (1 << i); j++ {
				runtime.Gosched()
			}
		}
	}
	return nil
}

func (q *MPMCQueue[T]) unlockLastBlock(b *block[T]) {
	q.lastBlock.value.Store(b)
}

// SPMC Reader implementation
type SPMCReader[T any] struct {
	block *block[T]
	index uint64
	len   uint64
}

func (q *SPMCQueue[T]) NewReader() *SPMCReader[T] {
	b := q.lastBlock.value.Load()
	if b == nil {
		return nil
	}
	b.incRef()
	len := b.len.value.Load()
	return &SPMCReader[T]{
		block: b,
		index: len,
		len:   len,
	}
}

func (r *SPMCReader[T]) Next() (T, bool) {
	var zero T
	if r.block == nil {
		return zero, false
	}

	if r.index >= r.len {
		if r.len == blockSize {
			next := r.block.next.value.Load()
			if next == nil {
				return zero, false
			}
			next.incRef()
			oldBlock := r.block
			r.block = next
			oldBlock.decRef()
			r.index = 0
			r.len = next.len.value.Load()
			if r.len == 0 {
				return zero, false
			}
		} else {
			newLen := r.block.len.value.Load()
			if r.len == newLen {
				return zero, false
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

func (r *SPMCReader[T]) Close() {
	if r.block != nil {
		r.block.decRef()
		r.block = nil
	}
}

// MPMC Reader implementation
type MPMCReader[T any] struct {
	block         *block[T]
	index         uint64
	len           uint64
	bitBlockIndex uint64
}

func (q *MPMCQueue[T]) NewReader() *MPMCReader[T] {
	b := q.lastBlock.value.Load()
	if b == nil {
		return nil
	}
	b.incRef()
	len := b.len.value.Load()
	return &MPMCReader[T]{
		block:         b,
		index:         len,
		len:           len,
		bitBlockIndex: len / 64,
	}
}

func (r *MPMCReader[T]) Next() (T, bool) {
	var zero T
	if r.block == nil {
		return zero, false
	}

	if r.index >= r.len {
		if r.len == blockSize {
			next := r.block.next.value.Load()
			if next == nil {
				return zero, false
			}
			next.incRef()
			oldBlock := r.block
			r.block = next
			oldBlock.decRef()
			r.index = 0

			bits := next.bitBlocks[0].Load()
			r.len = uint64(trailingOnes(bits))
			r.bitBlockIndex = 0
			if bits == ^uint64(0) {
				r.bitBlockIndex = 1
			}
			if r.len == 0 {
				return zero, false
			}
		} else {
			if r.bitBlockIndex >= blockSize/64 {
				return zero, false
			}

			bits := r.block.bitBlocks[r.bitBlockIndex].Load()
			newLen := r.bitBlockIndex*64 + uint64(trailingOnes(bits))

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

func (r *MPMCReader[T]) Close() {
	if r.block != nil {
		r.block.decRef()
		r.block = nil
	}
}

func trailingOnes(x uint64) int {
	return bits.TrailingZeros64(^x)
}
