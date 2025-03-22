package e2concurrent

import (
	"context"
	"math/rand"
	"testing"
	"time"
)

// processResult is a sample output type for testing.
type processResult struct {
	Output string
}

// TestWorker tests a worker with different input and output types.
func TestWorker(t *testing.T) {
	tasks := make(chan Task[string, processResult], 3)

	go func() {
		tasks <- NewTask[string, processResult](
			&worker{},
			"test1",
			WithRefer[string, processResult]("task-1"),
			WithRetainInput[string, processResult](true),
		)
		tasks <- NewTask[string, processResult](
			&worker{},
			"test2",
			WithRefer[string, processResult]("task-2"),
			WithRetainInput[string, processResult](false),
		)
		tasks <- NewTask[string, processResult](
			&worker{},
			"test3",
			WithRefer[string, processResult]("task-3"),
			WithRetainInput[string, processResult](true),
		)
		close(tasks)
	}()

	results := Exec(context.Background(), 2, tasks, 3)
	count := 0
	for r := range results {
		if r.Err != nil {
			t.Errorf("Unexpected error for task %s: %v", r.Refer, r.Err)
		}
		// If Arg.Value is not nil, RetainInput was true, so validate the output
		if r.Arg.Value != nil {
			expectedOutput := "Processed: " + r.Arg.Value.(string)
			if r.Value.Output != expectedOutput {
				t.Errorf("Expected output %q, got %q for task %s", expectedOutput, r.Value.Output, r.Refer)
			}
		}
		// If Arg.Value is nil, RetainInput was false, so just check the task completed
		if r.Arg.Value == nil && r.Value.Output == "" {
			t.Errorf("Expected non-empty output for task %s, got %q", r.Refer, r.Value.Output)
		}
		count++
	}
	if count != 3 {
		t.Errorf("Expected 3 results, got %d", count)
	}
}

// worker is a sample implementation with string input and processResult output.
type worker struct{}

func (w *worker) Run(arg Arg[string]) Result[processResult] {
	sleepTime := time.Duration(rand.Intn(100)) * time.Millisecond
	time.Sleep(sleepTime)
	return Result[processResult]{
		Value: processResult{Output: "Processed: " + arg.Value},
	}
}
