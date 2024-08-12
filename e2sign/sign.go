package e2sign

import (
	"os"
	"os/signal"

	"golang.org/x/exp/maps"
)

type Task struct {
	Signal os.Signal
	Func   func()
}

func RegisterSignTask(task map[os.Signal]func()) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, maps.Keys(task)...)
	go func() {
		for {
			sig := <-sigs
			if f, ok := task[sig]; ok {
				f()
			}
		}
	}()
}
