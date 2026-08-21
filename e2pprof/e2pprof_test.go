package e2pprof

import (
	"context"
	"testing"
	"time"
)

func TestInit(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	Init(ctx)
	// Init starts the pprof HTTP server on a random local port in a goroutine.
	time.Sleep(50 * time.Millisecond)
}
