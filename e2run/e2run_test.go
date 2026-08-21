package e2run

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestGoRunner(t *testing.T) {
	done := make(chan struct{})
	GoRunner(func() { close(done) })
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("GoRunner did not run")
	}
}

func TestGoLoopRunnerContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var n atomic.Int32
	GoLoopRunnerContext(ctx, func() {
		n.Add(1)
		cancel()
	}, 1)
	deadline := time.After(2 * time.Second)
	for n.Load() == 0 {
		select {
		case <-deadline:
			t.Fatal("GoLoopRunnerContext did not run")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func TestGoLoopRunnerWithoutSleep(t *testing.T) {
	var n atomic.Int32
	done := make(chan struct{})
	GoLoopRunnerWithoutSleep(func() {
		if n.Add(1) == 1 {
			close(done)
		}
		time.Sleep(50 * time.Millisecond)
	})
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("GoLoopRunnerWithoutSleep did not run")
	}
}
