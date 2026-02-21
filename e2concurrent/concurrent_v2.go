package e2concurrent

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"
)

// Trace records metadata about a task's execution.
// It tracks the task's identity and timing details.
type Trace struct {
	Id       string        `json:"id,omitempty"`       // Unique identifier for this execution instance (e.g., UUID)
	Refer    string        `json:"refer,omitempty"`    // Task identifier inherited from Task.Refer
	StartAt  time.Time     `json:"start_at"`           // Time when the task started execution
	Duration time.Duration `json:"duration,omitempty"` // Duration of the task's execution
}

// Result represents the outcome of a task's execution.
// It includes the computed value, any error, execution trace, and optionally the original argument.
type Result[Out any] struct {
	Value Out            `json:"value,omitempty"` // The result value of the task, type-safe via generics
	Err   error          `json:"err,omitempty"`   // Error if the task failed, nil if successful
	Arg   Arg[any]       `json:"arg"`             // Original input argument, included only if Task.RetainInput is true
	Trace `json:"trace"` // Embedded Trace struct for execution metadata
}

// Arg holds the input data for a task.
// It contains only the value to be processed, without tracking metadata.
type Arg[In any] struct {
	Value In `json:"value,omitempty"` // Input value for the task, type-safe via generics
}

// WorkFunc defines the interface for task execution logic.
// It processes an input Arg of type In and returns a Result of type Out.
type WorkFunc[In, Out any] interface {
	ConcurrentRun(arg Arg[In]) Result[Out]
}

// Task encapsulates a single task to be executed.
// It includes execution context, timeout, identifier, function, input argument, and an option to retain input in Result.
type Task[In, Out any] struct {
	Ctx         context.Context   `json:"-"`                      // Context for task cancellation or timeout (not serialized)
	Timeout     time.Duration     `json:"timeout,omitempty"`      // Optional timeout duration for the task
	Refer       string            `json:"refer,omitempty"`        // Unique identifier for tracking this task
	Fn          WorkFunc[In, Out] `json:"fn,omitempty"`           // Function to execute the task
	Arg         Arg[In]           `json:"arg"`                    // Input argument for the task
	RetainInput bool              `json:"retain_input,omitempty"` // Whether to retain the input Arg in the Result
}

// NewTask creates a new Task with the given function and input value.
// It assigns a default unique Refer and applies optional configurations.
func NewTask[In, Out any](fn WorkFunc[In, Out], value In, opts ...TaskOption[In, Out]) Task[In, Out] {
	t := Task[In, Out]{
		Ctx:   context.Background(),
		Fn:    fn,
		Arg:   Arg[In]{Value: value},
		Refer: uuid.NewString(),
	}
	for _, opt := range opts {
		opt(&t)
	}
	return t
}

// TaskOption is a function type for configuring Task options.
// It uses generics to match the Task's input and output type parameters.
type TaskOption[In, Out any] func(*Task[In, Out])

// WithTimeout sets a timeout duration for the task.
func WithTimeout[In, Out any](d time.Duration) TaskOption[In, Out] {
	return func(t *Task[In, Out]) { t.Timeout = d }
}

// WithContext sets a custom context for the task.
func WithContext[In, Out any](ctx context.Context) TaskOption[In, Out] {
	return func(t *Task[In, Out]) { t.Ctx = ctx }
}

// WithRefer sets a custom reference ID for the task.
func WithRefer[In, Out any](refer string) TaskOption[In, Out] {
	return func(t *Task[In, Out]) { t.Refer = refer }
}

// WithRetainInput sets whether the Result should retain the original input Arg.
func WithRetainInput[In, Out any](retain bool) TaskOption[In, Out] {
	return func(t *Task[In, Out]) { t.RetainInput = retain }
}

// taskWorker executes a single task and sends the result to the provided channel.
// It manages task timing, cancellation, and tracing, respecting the RetainInput setting.
func taskWorker[In, Out any](wg *sync.WaitGroup, uuid string, task Task[In, Out], result chan<- Result[Out]) {
	var ctx context.Context
	var cancel context.CancelFunc
	if task.Timeout > 0 {
		ctx, cancel = context.WithTimeout(task.Ctx, task.Timeout)
	} else {
		ctx, cancel = task.Ctx, func() {}
	}
	defer cancel()
	defer wg.Done()

	startTime := time.Now()
	done := make(chan Result[Out], 1)

	go func() {
		done <- task.Fn.ConcurrentRun(task.Arg)
	}()

	select {
	case <-ctx.Done():
		r := Result[Out]{
			Err: ctx.Err(),
			Trace: Trace{
				Id:       uuid,
				Refer:    task.Refer,
				StartAt:  startTime,
				Duration: time.Since(startTime),
			},
		}
		if task.RetainInput {
			r.Arg = Arg[any]{Value: task.Arg.Value}
		}
		result <- r
	case r := <-done:
		r.Trace = Trace{
			Id:       uuid,
			Refer:    task.Refer,
			StartAt:  startTime,
			Duration: time.Since(startTime),
		}
		if task.RetainInput {
			r.Arg = Arg[any]{Value: task.Arg.Value}
		}
		result <- r
	}
}

// DefaultExec runs tasks with default concurrency and buffer size.
func DefaultExec[In, Out any](ctx context.Context, tasks <-chan Task[In, Out]) <-chan Result[Out] {
	return Exec(ctx, runtime.NumCPU(), tasks, runtime.NumCPU()*10)
}

// Exec executes tasks concurrently with a specified concurrency limit and buffer size.
// It returns a read-only channel for receiving task results.
func Exec[In, Out any](ctx context.Context, maxConcurrency int, tasks <-chan Task[In, Out], bufferSize int) <-chan Result[Out] {
	if ctx == nil {
		ctx = context.Background()
	}
	if maxConcurrency < 1 {
		maxConcurrency = runtime.NumCPU()
	}
	if bufferSize < 0 {
		bufferSize = 0
	}

	result := make(chan Result[Out], bufferSize)
	go func() {
		var wg sync.WaitGroup
		sem := semaphore.NewWeighted(int64(maxConcurrency))
		for {
			select {
			case <-ctx.Done():
				wg.Wait()
				logrus.WithField("func", "e2concurrent.Exec").Debug("Context done, closing result channel")
				close(result)
				return
			case task, ok := <-tasks:
				if !ok {
					wg.Wait()
					logrus.WithField("func", "e2concurrent.Exec").Debug("Tasks channel closed, closing result channel")
					close(result)
					return
				}
				if err := sem.Acquire(ctx, 1); err != nil {
					logrus.WithField("func", "e2concurrent.Exec").Errorf("Failed to acquire semaphore: %v", err)
					r := Result[Out]{
						Err: err,
						Trace: Trace{
							Id:      uuid.NewString(),
							Refer:   task.Refer,
							StartAt: time.Now(),
						},
					}
					if task.RetainInput {
						r.Arg = Arg[any]{Value: task.Arg.Value}
					}
					result <- r
					continue
				}
				wg.Add(1)
				go func(t Task[In, Out]) {
					defer sem.Release(1)
					taskWorker(&wg, uuid.NewString(), t, result)
				}(task)
			}
		}
	}()
	return result
}
