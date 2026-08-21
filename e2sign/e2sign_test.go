package e2sign

import (
	"os"
	"syscall"
	"testing"
	"time"
)

func TestRegisterSignTask(t *testing.T) {
	done := make(chan struct{})
	RegisterSignTask(map[os.Signal]func(){
		syscall.SIGUSR1: func() { close(done) },
	})
	if err := syscall.Kill(os.Getpid(), syscall.SIGUSR1); err != nil {
		t.Fatalf("kill: %v", err)
	}
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("signal handler was not called")
	}
}
