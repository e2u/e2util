package e2pprof

import (
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"sync"

	"github.com/sirupsen/logrus"
)

func Init() {
	var once sync.Once
	go func() {
		once.Do(func() {
			logrus.Infof("-----------------------------------")
			logrus.Infof("pprof init")

			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				logrus.Errorf("tcp listen error: %v", err)
				return
			}
			port := listener.Addr().(*net.TCPAddr).Port
			logrus.Infof("pprof port: %v", port)
			pprofUrl := fmt.Sprintf("http://127.0.0.1:%d/debug/pprof", port)
			logrus.Info(pprofUrl)

			if err := http.Serve(listener, nil); err != nil { // #nosec G114
				logrus.Infof("run pprof error: %v", err)
				return
			}
		})
	}()
}
