package e2concurrent

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

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
	rn := e2time.SleepRandom(1*time.Second, 15*time.Second)
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

func Test_ExecuteTasks(t *testing.T) {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))
	ctx := context.Background()

	fn := &MT{}
	tasks := []Task{
		{Ctx: ctx, Timeout: 10 * time.Second, Fn: fn, Arg: Arg{Id: "t1", Value: TArg{A: 100, B: 10, C: "Hello"}}},
		{Ctx: ctx, Timeout: 10 * time.Second, Fn: fn, Arg: Arg{Id: "t2", Value: TArg{A: 200, B: 30, C: "AA"}}},
		{Ctx: ctx, Timeout: 10 * time.Second, Fn: fn, Arg: Arg{Id: "t3", Value: TArg{A: 300, B: 80, C: "BB"}}},
		{Ctx: ctx, Timeout: 10 * time.Second, Fn: fn, Arg: Arg{Id: "t4", Value: TArg{A: 400, B: 120, C: "CC"}}},
	}

	resultChan := make(chan Result)
	go ExecuteTasks(tasks, 20, resultChan)

	for result := range resultChan {
		if result.Err == nil {
			slog.Info("Success", "Value", result.Value, "Trace", result.Trace)
		} else {
			slog.Info("Error", "Result", result)
		}
	}
}
