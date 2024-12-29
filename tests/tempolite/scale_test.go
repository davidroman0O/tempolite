package tests

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	tempolite "github.com/davidroman0O/tempolite"
)

func TestTempoliteScale(t *testing.T) {

	database := tempolite.NewMemoryDatabase()
	defer database.SaveAsJSON("./jsons/tempolite_workflows_execute.json")
	ctx := context.Background()

	tp, err := tempolite.New(
		ctx,
		database,
	)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("Scale up")
	if err := tp.Scale(tempolite.DefaultQueue, 2); err != nil {
		t.Fatal(err)
	}

	<-time.After(time.Second)

	fmt.Println("Scale up")
	if err := tp.Scale(tempolite.DefaultQueue, 4); err != nil {
		t.Fatal(err)
	}

	<-time.After(time.Second)

	fmt.Println("Scale down")
	if err := tp.Scale(tempolite.DefaultQueue, 0); err != nil {
		t.Fatal(err)
	}

	<-time.After(time.Second * 2)

	tp.Close()

}

func TestTempoliteScaleConcurrentWorkflows(t *testing.T) {
	ctx := context.Background()

	var workflowCounts []struct {
		scale int
		count int32
	}

	var executionsPerSecond []struct {
		scale  int
		second int
		count  int32
	}

	var startedWorkflows int32
	var globalStartedWorkflows int32

	workflowFunc := func(ctx tempolite.WorkflowContext) (int, error) {
		atomic.AddInt32(&startedWorkflows, 1)
		atomic.AddInt32(&globalStartedWorkflows, 1)
		return 1, nil
	}

	for scale := 1; scale <= 5; scale++ {
		database := tempolite.NewMemoryDatabase()
		tp, err := tempolite.New(
			ctx,
			database,
		)
		if err != nil {
			t.Fatal(err)
		}

		fmt.Printf("Scale %d\n", scale)
		if err := tp.Scale(tempolite.DefaultQueue, scale); err != nil {
			t.Fatal(err)
		}

		// Schedule 7500 workflows
		go func() {
			for i := 0; i < 7500; i++ {
				if _, err := tp.ExecuteDefault(workflowFunc, nil); err != nil {
					t.Fatal(err)
				}
			}
		}()

		go func() {
			for i := 0; i < 7500; i++ {
				if _, err := tp.ExecuteDefault(workflowFunc, nil); err != nil {
					t.Fatal(err)
				}
			}
		}()

		go func() {
			for i := 0; i < 7500; i++ {
				if _, err := tp.ExecuteDefault(workflowFunc, nil); err != nil {
					t.Fatal(err)
				}
			}
		}()

		time.Sleep(time.Second)

		start := time.Now()
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		done := make(chan bool)
		go func(s int) {
			second := 0
			for {
				select {
				case <-ticker.C:
					c := atomic.SwapInt32(&startedWorkflows, 0)
					executionsPerSecond = append(executionsPerSecond, struct {
						scale  int
						second int
						count  int32
					}{s, second, c})
					second++
				case <-done:
					return
				}
			}
		}(scale)

		for {
			queueCount, err := tp.CountQueue(tempolite.DefaultQueue, tempolite.StatusCompleted)
			if err != nil {
				t.Fatal(err)
			}
			if queueCount == 22500 {
				break
			}
			time.Sleep(time.Second)
			fmt.Println("Queue count:", queueCount)
		}
		duration := time.Since(start).Seconds()

		// Use globalStartedWorkflows to track how many total were actually started
		c := atomic.LoadInt32(&globalStartedWorkflows)
		workflowCounts = append(workflowCounts, struct {
			scale int
			count int32
		}{scale, c})
		atomic.StoreInt32(&globalStartedWorkflows, 0)

		done <- true

		fmt.Printf("Scale %d: %d workflows started in %.2f seconds\n", scale, c, duration)

		fmt.Println("Close")
		tp.Close()
	}

	fmt.Println("Workflow counts per scale change:")
	for _, wc := range workflowCounts {
		fmt.Printf("Scale %d: %d workflows started\n", wc.scale, wc.count)
	}

	fmt.Println("Executions per second (chronological):")
	for _, eps := range executionsPerSecond {
		fmt.Printf("Scale %d - Second %d: %d executions\n", eps.scale, eps.second, eps.count)
	}

	// Example of computing an average per second across all logs
	fmt.Println("Averages per scale:")
	currentScale := 1
	var sum, secs int32
	for _, eps := range executionsPerSecond {
		if eps.scale != currentScale {
			if secs > 0 {
				fmt.Printf("Scale %d average: %.2f executions/sec\n",
					currentScale,
					float64(sum)/float64(secs),
				)
			}
			currentScale = eps.scale
			sum, secs = 0, 0
		}
		sum += eps.count
		secs++
	}
	if secs > 0 {
		fmt.Printf("Scale %d average: %.2f executions/sec\n",
			currentScale,
			float64(sum)/float64(secs),
		)
	}
}
