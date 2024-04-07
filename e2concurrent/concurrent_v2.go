package e2concurrent

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Trace struct {
	Id       string        `json:"id,omitempty"`
	Refer    string        `json:"refer,omitempty"`
	StartAt  time.Time     `json:"start_at"`
	Duration time.Duration `json:"duration,omitempty"`
}

type Result struct {
	Value any   `json:"value,omitempty"`
	Err   error `json:"err,omitempty"`
	Trace `json:"trace"`
}

type Arg struct {
	Refer string `json:"refer,omitempty"`
	Value any    `json:"value,omitempty"`
}

type WorkFunc interface {
	Run(arg Arg) Result
}

type Task struct {
	Ctx     context.Context `json:"ctx,omitempty"`
	Timeout time.Duration   `json:"timeout,omitempty"`
	Fn      WorkFunc        `json:"fn,omitempty"`
	Arg     Arg             `json:"arg"`
}

func taskWorker(uuid string, task Task, r func(resultChan Result), wg *sync.WaitGroup) {
	ctx, cancel := context.WithTimeout(task.Ctx, task.Timeout)
	defer cancel()
	defer wg.Done()

	startTime := time.Now()
	localResultChan := make(chan Result, 1)
	go func() {
		localResultChan <- task.Fn.Run(task.Arg)
	}()

	select {
	case <-ctx.Done():
		// Task timeout
		r(Result{Err: ctx.Err(), Trace: Trace{
			Id:       uuid,
			Refer:    task.Arg.Refer,
			StartAt:  startTime,
			Duration: time.Since(startTime),
		}})
	case result := <-localResultChan:
		// Task completed
		result.Trace = Trace{
			Id:       uuid,
			Refer:    task.Arg.Refer,
			StartAt:  startTime,
			Duration: time.Since(startTime),
		}
		r(result)
	}

}

func DefaultExec(ctx context.Context, taskFn func(tasks chan<- Task), resultFn func(r Result)) {
	Exec(ctx, runtime.NumCPU(), taskFn, resultFn)
}

func Exec(ctx context.Context, maxConcurrency int, taskFn func(tasks chan<- Task), resultFn func(r Result)) {
	var wg sync.WaitGroup
	wg.Add(1)

	tasksChan := make(chan Task)

	go func() {
		defer wg.Done()
		taskFn(tasksChan)
	}()

	go func() {
		wg.Wait()
		close(tasksChan)
	}()

	semaphore := make(chan struct{}, maxConcurrency)
	for task := range tasksChan {
		semaphore <- struct{}{}
		wg.Add(1)
		go func(t Task) {
			defer func() { <-semaphore }()
			taskWorker(uuid.NewString(), t, resultFn, &wg)
		}(task)
	}

}
