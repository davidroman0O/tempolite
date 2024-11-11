package main

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	chute "github.com/davidroman0O/tempolite/playground"
)

// Message represents our test data
type Message struct {
	ID        uint64
	Timestamp int64
	Data      [64]byte // Simulate some payload
}

// Metrics tracks performance statistics
type Metrics struct {
	messagesProduced uint64
	messagesConsumed uint64
	producerDrops    uint64
	consumerLatency  atomic.Int64  // Sum of latencies in microseconds
	latencyCount     atomic.Uint64 // Number of latency measurements
}

func (m *Metrics) recordLatency(latencyMicros int64) {
	m.consumerLatency.Add(latencyMicros)
	m.latencyCount.Add(1)
}

func (m Metrics) String() string {
	duration := 10.0 // seconds
	avgLatency := float64(m.consumerLatency.Load()) / float64(m.latencyCount.Load())

	return fmt.Sprintf(
		"\nPerformance Metrics (%.0fs run):\n"+
			"├─ Messages Produced: %d (%.2f msg/s)\n"+
			"├─ Messages Consumed: %d (%.2f msg/s)\n"+
			"├─ Producer Drops: %d\n"+
			"├─ Average Latency: %.2f µs\n"+
			"└─ Throughput: %.2f MB/s\n",
		duration,
		m.messagesProduced,
		float64(m.messagesProduced)/duration,
		m.messagesConsumed,
		float64(m.messagesConsumed)/duration,
		m.producerDrops,
		avgLatency,
		// Throughput calculation: msgs/s * msg size / (1024*1024) for MB/s
		float64(m.messagesConsumed)*float64(unsafe.Sizeof(Message{}))/duration/(1024*1024),
	)
}

func producer(q *chute.Queue[Message], metrics *Metrics, done <-chan struct{}) {
	runtime.LockOSThread() // Pin to OS thread for better performance
	var msgID uint64

	for {
		select {
		case <-done:
			return
		default:
			msg := Message{
				ID:        msgID,
				Timestamp: time.Now().UnixNano(),
			}

			// Try to produce
			q.Push(msg)
			atomic.AddUint64(&metrics.messagesProduced, 1)
			msgID++
		}
	}
}

func consumer(q *chute.Queue[Message], metrics *Metrics, done <-chan struct{}) {
	runtime.LockOSThread()
	reader := q.NewReader()
	if reader == nil {
		// Handle error or retry
		return
	}
	defer reader.Close()

	for {
		select {
		case <-done:
			return
		default:
			if msg, ok := reader.Next(); ok {
				latency := time.Now().UnixNano() - msg.Timestamp
				metrics.recordLatency(latency / 1000)
				atomic.AddUint64(&metrics.messagesConsumed, 1)
			} else {
				runtime.Gosched()
			}
		}
	}
}

func runBenchmark(producers, consumers int, duration time.Duration) Metrics {
	q := chute.NewQueue[Message]()
	metrics := Metrics{}
	done := make(chan struct{})
	var wg sync.WaitGroup

	// Start producers
	for i := 0; i < producers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			producer(q, &metrics, done)
		}()
	}

	// Start consumers
	for i := 0; i < consumers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			consumer(q, &metrics, done)
		}()
	}

	// Run for specified duration
	time.Sleep(duration)
	close(done)
	wg.Wait()

	return metrics
}

func main() {
	fmt.Printf("Running queue benchmarks on %d CPUs\n", runtime.NumCPU())
	runtime.GOMAXPROCS(runtime.NumCPU())

	scenarios := []struct {
		name      string
		producers int
		consumers int
	}{
		{"1P-1C", 1, 1},
		{"1P-4C", 1, 4},
		{"4P-1C", 4, 1},
		{"4P-4C", 4, 4},
		{"8P-8C", 8, 8},
		{fmt.Sprintf("%dP-%dC", runtime.NumCPU(), runtime.NumCPU()), runtime.NumCPU(), runtime.NumCPU()},
	}

	duration := 10 * time.Second

	for _, s := range scenarios {
		fmt.Printf("\n=== Scenario: %s ===\n", s.name)
		metrics := runBenchmark(s.producers, s.consumers, duration)
		fmt.Println(metrics)

		// Force GC between scenarios
		runtime.GC()
		time.Sleep(time.Second) // Let system settle
	}
}
