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
		float64(m.messagesConsumed)*float64(unsafe.Sizeof(Message{}))/duration/(1024*1024),
	)
}

func producerSPMC(q *chute.SPMCQueue[Message], metrics *Metrics, done <-chan struct{}) {
	runtime.LockOSThread()
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
			q.Push(msg)
			atomic.AddUint64(&metrics.messagesProduced, 1)
			msgID++
		}
	}
}

func producerMPMC(q *chute.MPMCQueue[Message], metrics *Metrics, done <-chan struct{}) {
	runtime.LockOSThread()
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
			q.Push(msg)
			atomic.AddUint64(&metrics.messagesProduced, 1)
			msgID++
		}
	}
}

func consumerSPMC(q *chute.SPMCQueue[Message], metrics *Metrics, done <-chan struct{}) {
	runtime.LockOSThread()
	reader := q.NewReader()
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

func consumerMPMC(q *chute.MPMCQueue[Message], metrics *Metrics, done <-chan struct{}) {
	runtime.LockOSThread()
	reader := q.NewReader()
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

func runBenchmarkSPMC(producers, consumers int, duration time.Duration) Metrics {
	q := chute.NewSPMCQueue[Message]()
	metrics := Metrics{}
	done := make(chan struct{})
	var wg sync.WaitGroup

	// Single producer for SPMC
	wg.Add(1)
	go func() {
		defer wg.Done()
		producerSPMC(q, &metrics, done)
	}()

	// Multiple consumers
	for i := 0; i < consumers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			consumerSPMC(q, &metrics, done)
		}()
	}

	time.Sleep(duration)
	close(done)
	wg.Wait()

	return metrics
}

func runBenchmarkMPMC(producers, consumers int, duration time.Duration) Metrics {
	q := chute.NewMPMCQueue[Message]()
	metrics := Metrics{}
	done := make(chan struct{})
	var wg sync.WaitGroup

	// Multiple producers
	for i := 0; i < producers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			producerMPMC(q, &metrics, done)
		}()
	}

	// Multiple consumers
	for i := 0; i < consumers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			consumerMPMC(q, &metrics, done)
		}()
	}

	time.Sleep(duration)
	close(done)
	wg.Wait()

	return metrics
}

func main() {
	fmt.Printf("Running queue benchmarks on %d CPUs\n", runtime.NumCPU())
	runtime.GOMAXPROCS(runtime.NumCPU())

	duration := 10 * time.Second

	// SPMC Scenarios
	spmcScenarios := []struct {
		name      string
		consumers int
	}{
		{"SPMC-1C", 1},
		{"SPMC-4C", 4},
		{"SPMC-8C", 8},
		{fmt.Sprintf("SPMC-%dC", runtime.NumCPU()), runtime.NumCPU()},
	}

	fmt.Println("\n=== SPMC Benchmarks ===")
	for _, s := range spmcScenarios {
		fmt.Printf("\n=== Scenario: %s ===\n", s.name)
		metrics := runBenchmarkSPMC(1, s.consumers, duration)
		fmt.Println(metrics)
		runtime.GC()
		time.Sleep(time.Second)
	}

	// MPMC Scenarios
	mpmcScenarios := []struct {
		name      string
		producers int
		consumers int
	}{
		{"MPMC-1P-1C", 1, 1},
		{"MPMC-4P-4C", 4, 4},
		{"MPMC-8P-8C", 8, 8},
		{fmt.Sprintf("MPMC-%dP-%dC", runtime.NumCPU(), runtime.NumCPU()), runtime.NumCPU(), runtime.NumCPU()},
	}

	fmt.Println("\n=== MPMC Benchmarks ===")
	for _, s := range mpmcScenarios {
		fmt.Printf("\n=== Scenario: %s ===\n", s.name)
		metrics := runBenchmarkMPMC(s.producers, s.consumers, duration)
		fmt.Println(metrics)
		runtime.GC()
		time.Sleep(time.Second)
	}
}
