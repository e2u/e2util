package e2run

import (
	"context"
	"time"
)

func GoRunner(fn func()) {
	go func() {
		fn()
	}()
}

func GoLoopRunner(fn func(), sleepSec int64) {
	if sleepSec <= 0 {
		sleepSec = 1
	}
	go func() {
		for {
			fn()
			time.Sleep(time.Duration(sleepSec) * time.Second)
		}
	}()
}

func GoLoopRunnerContext(ctx context.Context, fn func(), sleepSec int64) {
	if sleepSec <= 0 {
		sleepSec = 1
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				fn()
				time.Sleep(time.Duration(sleepSec) * time.Second)
			}
		}
	}()
}

func GoLoopRunnerWithoutSleep(fn func()) {
	go func() {
		for {
			fn()
		}
	}()
}
