package e2concurrent

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/e2u/e2util/e2crypto"
	"github.com/e2u/e2util/e2time"
)

type TArg struct {
	A int
	B int
	C string
}

type TVResult struct {
	ReqArg TArg
	N      int
	NS     string
	Err    error
}

type MT struct {
}

func (t *MT) Run(arg Arg) Result {
	slog.Info("Run begin", "arg", arg)
	rn := e2time.SleepRandom(1*time.Second, 3*time.Second)
	slog.Info("sleep random", "sleep_time", rn)
	a := arg.Value.(TArg)

	return Result{
		Value: TVResult{
			ReqArg: a,
			N:      a.A + a.B + 10,
			NS:     "Hi, " + a.C,
		},
	}
}

func newTask(i int) Task {
	fn := &MT{}
	ctx := context.Background()
	return Task{
		Ctx:     ctx,
		Timeout: 10 * time.Second,
		Fn:      fn,
		Arg: Arg{
			Refer: fmt.Sprintf("%000d", i),
			Value: TArg{
				A: int(e2crypto.RandomUint(0, 100)),
				B: int(e2crypto.RandomUint(50, 200)),
				C: e2crypto.RandomString(5),
			},
		},
	}
}

func resultFunc(result Result) {
	if result.Err == nil {
		slog.Info("Success", "Value", result.Value, "Trace", result.Trace)
	} else {
		slog.Info("Error", "Result", result)
	}
}

func Test_a02(t *testing.T) {
	ctx := context.Background()
	taskChan := make(chan Task)
	go func() {
		for i := 0; i < 20; i++ {
			taskChan <- newTask(i)
		}
		close(taskChan)
	}()

	DefaultExec(ctx, taskChan, resultFunc)
}
