package e2concurrent

import (
	"context"
	"sync"
	"time"
)

type Trace struct {
	Id       string        `json:"id,omitempty"`
	StartAt  time.Time     `json:"start_at"`
	Duration time.Duration `json:"duration,omitempty"`
}

type Result struct {
	Value any   `json:"value,omitempty"`
	Err   error `json:"err,omitempty"`
	Trace `json:"trace"`
}

type Arg struct {
	Id    string `json:"id,omitempty"`
	Value any    `json:"value,omitempty"`
}

type TaskFunc interface {
	Run(arg Arg) Result
}

type Task struct {
	Ctx     context.Context `json:"ctx,omitempty"`
	Timeout time.Duration   `json:"timeout,omitempty"`
	Fn      TaskFunc        `json:"fn,omitempty"`
	Arg     Arg             `json:"arg"`
}

func ExecuteTasks(tasks []Task, maxConcurrency int, resultChan chan Result) {
	var wg sync.WaitGroup
	taskChan := make(chan Task, len(tasks))

	// Worker goroutines
	for i := 0; i < maxConcurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			worker(taskChan, resultChan)
		}()
	}

	// Send tasks to task channel
	go func() {
		for _, task := range tasks {
			taskChan <- task
		}
		close(taskChan)
	}()

	// Wait for all worker goroutines to finish
	wg.Wait()
	close(resultChan)
}

func worker(taskChan <-chan Task, resultChan chan<- Result) {
	for task := range taskChan {
		ctx, cancel := context.WithTimeout(task.Ctx, task.Timeout)

		startTime := time.Now()
		localResultChan := make(chan Result, 1)
		go func() {
			localResultChan <- task.Fn.Run(task.Arg)
		}()

		select {
		case <-ctx.Done():
			// Task timeout
			resultChan <- Result{Err: ctx.Err(), Trace: Trace{
				Id:       task.Arg.Id,
				StartAt:  startTime,
				Duration: time.Since(startTime),
			}}
		case result := <-localResultChan:
			// Task completed
			result.Trace = Trace{
				Id:       task.Arg.Id,
				StartAt:  startTime,
				Duration: time.Since(startTime),
			}
			resultChan <- result
		}
		cancel()
	}
}
