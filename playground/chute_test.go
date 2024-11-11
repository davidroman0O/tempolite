package chute

import (
	"fmt"
	"testing"
)

func TestMain(t *testing.T) {
	queue := NewQueue[int]()
	reader := queue.NewReader()
	defer reader.Close() // Properly cleanup resources

	// Producer
	queue.Push(1)
	queue.Push(2)

	// Consumer
	for {
		if val, ok := reader.Next(); ok {
			// Process val
			fmt.Println(val)
		} else {
			break
		}
	}
}
