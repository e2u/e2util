package e2context

import (
	"context"
	"testing"
	"time"
)

func TestCheckAndCancelContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	CheckAndCancelContext(ctx, cancel)
	if ctx.Err() == nil {
		t.Fatal("expected context to be canceled")
	}

	already, alreadyCancel := context.WithCancel(context.Background())
	alreadyCancel()
	CheckAndCancelContext(already, alreadyCancel)
	if already.Err() == nil {
		t.Fatal("expected already-canceled context to stay canceled")
	}
}

func TestCheckAndCancelContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	time.Sleep(5 * time.Millisecond)
	CheckAndCancelContext(ctx, cancel)
	if ctx.Err() == nil {
		t.Fatal("expected timed-out context to be done")
	}
}
